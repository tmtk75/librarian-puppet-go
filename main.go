package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/tmtk75/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "librarian-puppet-go"
	app.Version = "0.1.0"
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "verbose"},
	}
	app.Before = func(c *cli.Context) error {
		if c.Bool("verbose") {
			logger = log.New(os.Stderr, "", log.LstdFlags)
		}
		return nil
	}
	app.Commands = []cli.Command{
		cli.Command{
			Name: "install",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "modulepath", Value: "modules", Usage: "path to be for modules"},
			},
			Action: func(c *cli.Context) {
				p := c.String("modulepath")
				install(p)
			},
		},
	}
	app.Run(os.Args)
}

var logger = log.New(ioutil.Discard, "", log.LstdFlags)
var cwd string
var modulepath string

type ModOpts map[string]string

// mod 'puppetlabs/stdlib', '4.1.0'
// mod 'fiz', :git => 'git@github.com:foo/bar.git', :ref => 'v0.4.1'
type Mod struct {
	name    string  // stdlib, fiz
	user    string  // puppetlabs
	version string  // 4.1.0
	opts    ModOpts // git => git@github.com:foo/bar.git, ref => v0.4.1
	cmd     string  // clone, fetch, checkout
	err     error
}

func (m Mod) String() string {
	return fmt.Sprintf("name:%v\topts:%v\tuser:%v\tversion:%v", m.name, m.opts, m.user, m.version)
}

func (m Mod) Dest() string {
	return filepath.Join(modulepath, m.name)
}

func paths(mpath string) (modpath string, cwd string) {
	d, err := os.Getwd()
	if err != nil {
		logger.Fatalf("%v", err)
	}

	d, err = filepath.Abs(cwd)
	if err != nil {
		logger.Fatalf("%v", err)
	}

	mp, err := filepath.Abs(mpath)
	if err != nil {
		logger.Fatalf("%v", err)
	}
	return mp, d
}

func install(mpath string) {
	modulepath, cwd = paths(mpath)
	logger.Printf("cwd: %v", cwd)
	logger.Printf("modulepath: %v", modulepath)

	mods := parsePuppetfile(os.Stdin)

	errs := make(chan Mod)
	var wg sync.WaitGroup
	for _, m := range mods {
		wg.Add(1)
		go func(m Mod) {
			defer wg.Done()
			installMod(m, errs)
		}(m)
	}

	failed := make([]Mod, 0)
	go func() {
		for e := range errs {
			failed = append(failed, e)
		}
	}()
	wg.Wait()

	for _, m := range failed {
		fmt.Printf("\t%v\t%v\t%v\n", m.err, m.cmd, m)
	}
	if len(failed) > 0 {
		os.Exit(1)
	}
}

func parsePuppetfile(i io.Reader) []Mod {
	r := bufio.NewReader(i)
	mods := make([]Mod, 0)
	for {
		b, _, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		s := string(b)
		if s == "" || s[0] == '#' {
			continue
		}

		m, err := parseMod(s)
		if err != nil {
			logger.Printf("[warn] %v\n", err)
		}
		mods = append(mods, m)
	}
	return mods
}

func parseMod(s string) (Mod, error) {
	re := regexp.MustCompile(`^mod\s+'([a-z/_]+)'\s*(,\s*'(\d\.\d\.\d)')?$`).FindAllStringSubmatch(s, -1)
	if len(re) > 0 {
		n := re[0][1]
		v := ""
		if len(re[0]) > 3 {
			v = re[0][3]
		}
		nn := strings.Split(n, "/")
		return Mod{name: nn[1], user: nn[0], version: v, opts: ModOpts{}}, nil
	}

	re = regexp.MustCompile(`^mod\s+([^,]+),(.*?)$`).FindAllStringSubmatch(s, -1)
	if len(re) > 0 {
		return Mod{name: unquote(re[0][1]), opts: parseOpts(re[0][2])}, nil
	}

	return Mod{}, fmt.Errorf("[warn] ignore %v", s)
}

type Res struct {
	CurrentRelease struct {
		Metadata struct {
			Source      string `json:"source"`
			ProjectPage string `json:"project_page"`
		} `json:"metadata"`
	} `json:"current_release"`
}

func giturl(m Mod) string {
	ep := "http://forgeapi.puppetlabs.com/v3/modules/" + m.user + "-" + m.name
	req, err := http.NewRequest("GET", ep, nil)
	if err != nil {
		log.Fatalf("%v", err)
	}
	c := http.Client{}
	res, err := c.Do(req)

	var v Res
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if err := json.Unmarshal(b, &v); err != nil {
		log.Fatalf("%v", err)
	}
	if v.CurrentRelease.Metadata.Source == "UNKNOWN" {
		return v.CurrentRelease.Metadata.ProjectPage
	}
	return v.CurrentRelease.Metadata.Source
}

func installMod(m Mod, errs chan Mod) {
	logger.Printf("%v\n", m)
	if m.opts["git"] == "" {
		m.opts["git"] = giturl(m)
	}

	// start git operations
	var err error
	if m.opts["git"] == "" {
		log.Fatalf("[fatal] %v", m)
	}
	if !exists(m.Dest()) {
		err = gitClone(m.opts["git"], m.Dest())
		m.cmd = "clone"
	} else {
		err = gitFetch(m.Dest())
		m.cmd = "fetch"
	}
	m.err = err
	if err != nil {
		errs <- m
		return
	}

	//time.Sleep(250 * time.Millisecond) // to avoid git error, but this seems not to working
	ver := m.version
	if m.opts["ref"] != "" {
		ver = m.opts["ref"]
	}
	err = gitCheckout(m.Dest(), ver)
	m.cmd = "checkout"
	m.err = err
	if err != nil {
		errs <- m
	}
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func unquote(s string) string {
	return regexp.MustCompile(`["']`).ReplaceAllString(s, "")
}

func parseOpts(s string) ModOpts {
	m := make(ModOpts)
	for _, e := range strings.Split(s, ",") {
		re := regexp.MustCompile(`:([a-z_]+)\s*=>\s*(.*)$`).FindAllStringSubmatch(e, -1)
		//fmt.Printf("%v", re)
		if len(re) == 0 {
			continue
		}
		m[re[0][1]] = unquote(re[0][2])
	}
	return m
}

func gitClone(url, dest string) error {
	logger.Printf("git clone %v %v\n", url, dest)
	return run("git", []string{"clone", url, dest})
}

func gitFetch(dest string) error {
	os.Chdir(dest)
	logger.Printf("git fetch -p in %v\n", dest)
	return run("git", []string{"fetch", "-p"})
}

func gitCheckout(dest, ref string) error {
	os.Chdir(dest)
	if ref == "" {
		ref = "master"
	}
	logger.Printf("git checkout %v in %v\n", ref, dest)
	return run("git", []string{"checkout", ref})
}

func run(s string, args []string) error {
	//logger.Printf("[debug] %v %v\n", s, args)
	cmd := exec.Command(s, args...)
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	buf := bytes.NewBuffer([]byte{})
	cmd.Stderr = bufio.NewWriter(buf)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("[error] %v\t%v\n", err, buf)
	}
	return err
}
