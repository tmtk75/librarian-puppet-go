package librarianpuppetgo

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
)

var newReader func(string) io.ReadCloser = readFromFile

func readFromFile(n string) io.ReadCloser {
	r, err := os.OpenFile(n, os.O_RDONLY, 0660)
	if err != nil {
		log.Fatalf("%v", err)
	}
	return r
}

type installCmd struct {
	throttle             int
	forceCheckout        bool
	onlyCheckout         bool
	includesWithRepoName string
}

func (c installCmd) Main(path string) {
	r := newReader(path)
	defer r.Close()
	c.install(bufio.NewReader(r))
}

func (c installCmd) install(src io.Reader) {
	ms, err := parsePuppetfile(src)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	logger.Printf("includes-with-repository-name: '%v'", c.includesWithRepoName)
	re, err := regexp.Compile(c.includesWithRepoName)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	mods := make([]Mod, 0)
	for _, e := range ms {
		if y := re.Match([]byte(e.opts["git"])); y {
			mods = append(mods, e)
		}
	}

	if c.throttle < 1 || len(mods) < c.throttle {
		c.throttle = len(mods)
	}
	logger.Printf("mods size: %v, throttle: %v", len(mods), c.throttle)

	var wg sync.WaitGroup
	wg.Add(len(mods))

	tasks := make(chan Mod)
	errs := make(chan Mod)

	for i := 0; i < c.throttle; i++ {
		go func() {
			for m := range tasks {
				defer wg.Done()
				if err := c.installMod(m); err != nil {
					m.err = err
					errs <- m
				}
			}
		}()
	}

	failed := make([]Mod, 0)
	go func() {
		for e := range errs {
			failed = append(failed, e)
		}
	}()

	for _, m := range mods {
		tasks <- m
	}
	close(tasks)

	wg.Wait()
	close(errs)

	for _, m := range failed {
		log.Printf("\t%v\t%v\t%v\n", m.err, m.cmd, m)
	}
	if len(failed) > 0 {
		os.Exit(1)
	}
}

func (c installCmd) installMod(m Mod) error {
	if m.opts["git"] == "" {
		m.opts["git"] = giturl(m)
		if m.opts["git"] == "" {
			log.Fatalf("[fatal] :git is empty %v", m)
		}
	}
	//logger.Printf("%v\n", m)

	// start git operations
	var err error
	if !exists(m.Dest()) {
		err = gitClone(m.opts["git"], m.Dest())
		m.cmd = "clone"
	} else {
		if !c.onlyCheckout {
			err = gitFetch(m.Dest())
			m.cmd = "fetch"
		}
	}
	if err != nil {
		return err
	}

	ver := m.version
	if m.opts["ref"] != "" {
		ver = m.opts["ref"]
	}

	err = gitCheckout(m.Dest(), ver, c.forceCheckout)
	m.cmd = "checkout"
	if err != nil {
		return err
	}
	if !isTag(m.Dest(), ver) && !c.onlyCheckout {
		err = gitPull(m.Dest(), ver)
		m.cmd = "pull"
	}
	return err
}

type Res struct {
	CurrentRelease struct {
		Metadata struct {
			Source      string `json:"source"`
			ProjectPage string `json:"project_page"`
		} `json:"metadata"`
	} `json:"current_release"`
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func giturl(m Mod) string {
	ep := "https://forgeapi.puppetlabs.com/v3/modules/" + m.user + "-" + m.name
	logger.Printf("%v", ep)
	req, err := http.NewRequest("GET", ep, nil)
	if err != nil {
		log.Fatalf("%v", err)
	}
	c := http.Client{}
	res, err := c.Do(req)
	if (res.StatusCode / 100) != 2 {
		log.Fatalf("%v\t%v\t%v\n", m, ep, res)
	}

	var v Res
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if err := json.Unmarshal(b, &v); err != nil {
		log.Fatalf("%v", err)
	}

	u := v.CurrentRelease.Metadata.Source
	if u == "UNKNOWN" {
		u = v.CurrentRelease.Metadata.ProjectPage
	}
	// NOTE: workaround because 301 comes via http for github.com
	//       and it's hard to handle it.
	return regexp.MustCompile(`^http://`).ReplaceAllString(u, "https://")
}
