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
	"syscall"
	"time"

	"code.google.com/p/go.crypto/ssh/terminal"

	"github.com/tmtk75/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "librarian-puppet-go"
	app.Version = "0.1.5"
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "verbose", Usage: "Show logs verbosely"},
	}
	app.Before = func(c *cli.Context) error {
		if c.Bool("verbose") {
			logger = log.New(os.Stderr, "", log.LstdFlags)
		}
		return nil
	}
	app.Commands = []cli.Command{
		cli.Command{
			Name:  "install",
			Usage: "Install modules with a Puppetfile given thru stdin",
			Args:  "[filename]",
			Flags: flags,
			Action: func(c *cli.Context) {
				realMain(c, "install")
			},
		},
		cli.Command{
			Name:  "checkout",
			Usage: "Checkout modules without network access",
			Args:  "[filename]",
			Flags: flags,
			Action: func(c *cli.Context) {
				onlyCheckout = true
				realMain(c, "checkout")
			},
		},
	}
	app.Run(os.Args)
}

var flags = []cli.Flag{
	cli.StringFlag{Name: "modulepath", Value: "modules", Usage: "Path to be for modules"},
	cli.IntFlag{Name: "throttle", Value: 0, Usage: `Throttle number of concurrent processes.
                                Max is number of mod, min is 1. Max is used if 0 or negative number is given.`},
}

func readFromFile(n string) io.ReadCloser {
	r, err := os.OpenFile(n, os.O_RDONLY, 0660)
	if err != nil {
		log.Fatalf("%v", err)
	}
	return r
}

var newReader func(string) io.ReadCloser = readFromFile

func realMain(c *cli.Context, cmdName string) {
	p := c.String("modulepath")
	n, b := c.ArgFor("filename")
	t := c.Int("throttle")
	if b {
		r := newReader(n)
		defer r.Close()
		install(p, bufio.NewReader(r), t)
	} else {
		if !terminal.IsTerminal(int(os.Stdin.Fd())) {
			install(p, os.Stdin, t)
		} else {
			cli.ShowCommandHelp(c, cmdName)
		}
	}
}

var logger = log.New(ioutil.Discard, "", log.LstdFlags)
var modulepath string
var onlyCheckout bool

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

func (m *Mod) Replace(e *Mod) {
	m.user = e.user
	m.version = e.version
	for k, v := range e.opts {
		m.opts[k] = v
	}
}

func install(mpath string, src io.Reader, throttle int) {
	mp, err := filepath.Abs(mpath)
	if err != nil {
		logger.Fatalf("%v", err)
	}
	modulepath = mp
	logger.Printf("modulepath: %v", modulepath)

	mods, err := parsePuppetfile(src)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	if throttle < 1 || len(mods) < throttle {
		throttle = len(mods)
	}
	logger.Printf("mods size: %v, throttle: %v", len(mods), throttle)

	var wg sync.WaitGroup
	wg.Add(len(mods))

	tasks := make(chan Mod)
	errs := make(chan Mod)

	for i := 0; i < throttle; i++ {
		go func() {
			for m := range tasks {
				defer wg.Done()
				if err := installMod(m); err != nil {
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

func parsePuppetfile(i io.Reader) ([]Mod, error) {
	r := bufio.NewReader(i)
	incs := make([][]Mod, 0)
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

		if a := isInclude(s); a != "" {
			if len(mods) > 0 {
				return mods, fmt.Errorf("[error] include(s) must be at the top: %v", s)
			}
			logger.Printf("include: '%v'\n", a)

			r := newReader(a)
			defer r.Close()
			inc, err := parsePuppetfile(bufio.NewReader(r))
			if err != nil {
				return mods, err
			}
			incs = append(incs, inc)
			continue
		}

		m, err := parseMod(s)
		if err != nil {
			logger.Printf("[warn] %v\n", err)
			return mods, err
		}
		mods = append(mods, m)
	}

	return packMods(&incs, &mods), nil
}

// pack all mods as a slice
func packMods(incs *[][]Mod, mods *[]Mod) []Mod {
	all := append(*incs, *mods)
	n2m := map[string]*Mod{}
	result := make([]*Mod, 0)
	for _, i := range all {
		for _, e := range i {
			if n2m[e.name] == nil {
				a := e // copy
				n2m[e.name] = &a
				result = append(result, &a)
			} else {
				n2m[e.name].Replace(&e)
			}
		}
	}

	res := make([]Mod, len(result))
	for i, v := range result {
		res[i] = *v
	}
	return res
}

func isInclude(s string) string /* filename */ {
	re := regexp.MustCompile(`include\s+["'](.*?)["']`).FindAllStringSubmatch(s, -1)
	if len(re) == 0 {
		return ""
	}
	//log.Printf("[TRACE] %v\n", re)
	return re[0][1]
}

func parseMod(i string) (Mod, error) {
	s := regexp.MustCompile(`#.*$`).ReplaceAllLiteralString(i, "")
	re := regexp.MustCompile(`^mod\s+["']([a-z/_0-9]+)['"]\s*(,\s*["'](\d\.\d(\.\d)?)["'])?$`).FindAllStringSubmatch(s, -1)
	if len(re) > 0 {
		n := re[0][1]
		v := ""
		if len(re[0]) > 3 {
			v = re[0][3]
		}
		nn := strings.Split(n, "/")
		if len(nn) != 2 {
			return Mod{}, fmt.Errorf("'%v' should contain one '/'", n)
		}
		return Mod{name: nn[1], user: nn[0], version: v, opts: ModOpts{}}, nil
	}

	re = regexp.MustCompile(`^mod\s+([^,]+),(.*?)$`).FindAllStringSubmatch(s, -1)
	if len(re) > 0 {
		return Mod{name: unquote(re[0][1]), opts: parseOpts(re[0][2])}, nil
	}

	return Mod{}, fmt.Errorf("cannot parse: %v", s)
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
	ep := "https://forgeapi.puppetlabs.com/v3/modules/" + m.user + "-" + m.name
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

func installMod(m Mod) error {
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
		if !onlyCheckout {
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

	err = gitCheckout(m.Dest(), ver)
	m.cmd = "checkout"
	if err != nil {
		return err
	}
	if !isTag(m.Dest(), ver) && !onlyCheckout {
		err = gitPull(m.Dest(), ver)
		m.cmd = "pull"
	}
	return err
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

func isTag(dest, tag string) bool {
	cmd := exec.Command("git", "show-ref", "-q", "--verify", "refs/tags/"+tag)
	cmd.Dir = dest
	err := cmd.Run()
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus() == 0
		}
	}
	if err != nil {
		log.Fatalf("[error] %v\t%v\t%v\n", err, dest, tag)
	}
	return true
}

func gitClone(url, dest string) error {
	return run("", "git", []string{"clone", url, dest})
}

func gitFetch(dest string) error {
	return run(dest, "git", []string{"fetch", "-p"})
}

func gitPull(dest, ref string) error {
	return run(dest, "git", []string{"pull", "origin", ref})
}

func gitCheckout(dest, ref string) error {
	if ref == "" {
		ref = "master"
	}
	return run(dest, "git", []string{"checkout", ref})
}

func run(wd, s string, args []string) error {
	//logger.Printf("[debug] %v %v\n", s, args)
	cmd := exec.Command(s, args...)
	cmd.Dir = wd
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	buf := bytes.NewBuffer([]byte{})
	cmd.Stderr = bufio.NewWriter(buf)
	logger.Printf("start: %v %v in %v", s, args, wd)
	now := time.Now()
	err := cmd.Run()
	prefix := "done"
	if err != nil {
		prefix = "error"
		log.Printf("[error] %v\t%v\t%v\n", err, args, buf)
	}
	elapsed := time.Since(now)
	logger.Printf("%v: %v %v %v in %v", prefix, elapsed, s, args, wd)
	return err
}
