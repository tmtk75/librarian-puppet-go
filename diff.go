package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/tmtk75/cli"
)

func findModIn(mods []Mod, m Mod) (Mod, error) {
	for _, i := range mods {
		if i.name == m.name {
			return i, nil
		}
	}
	return Mod{}, fmt.Errorf("missing %s", m.name)
}

type DiffFunc func(m, n Mod, aref, bref string)

func run2(w io.Writer, wd, s string, args []string) {
	cmd := exec.Command(s, args...)
	cmd.Dir = wd
	cmd.Stderr = os.Stderr
	cmd.Stdout = w
	err := cmd.Run()
	if err != nil {
		log.Fatalln(err)
	}
}

const releaseBranchPattern = `release/0.([0-9]+)`

func increment(s string) (string, error) {
	re := regexp.MustCompile(releaseBranchPattern).FindAllStringSubmatch(s, -1)
	if len(re) == 0 {
		return "", fmt.Errorf("%v didn't match '%v'", s, releaseBranchPattern)
	}
	minor := re[0][1]
	v, err := strconv.Atoi(minor)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("release/0.%d", v+1), nil
}

type Mods []Mod

func (v Mods) Len() int {
	return len(v)
}

func (v Mods) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v Mods) Less(i, j int) bool {
	e := strings.Compare(v[i].name, v[j].name)
	return e < 0
}

func format(c *cli.Context, a string) {
	modulepath = c.String("modulepath")
	mods := Mods(parse(a))
	sort.Sort(mods)
	for _, m := range mods {
		fmt.Println(m.Format())
	}
}

func bumpUp(c *cli.Context, a, b string) {
	modulepath = c.String("modulepath")

	am := parse(a)
	bm := parse(b)

	for _, n := range bm {
		fmt.Println(bumpUpMod(n, am, c.String("initial-release-branch"), a, gitDiff))
	}
}

func bumpUpMod(n Mod, mods []Mod, rel, filename string, diff func(wd, a, b string) string) string {
	m, err := findModIn(mods, n)
	if err != nil {
		log.Printf("INFO: %v in %v. old one is used\n", err, filename)
		if n.opts["ref"] != "" {
			opts := map[string]string{"git": n.opts["git"], "ref": rel}
			p := Mod{name: n.name, version: n.version, opts: opts, user: n.user}
			return p.Format()
		} else {
			return n.Format()
		}
	}
	aref := m.opts["ref"]
	bref := n.opts["ref"]
	if aref == "" || bref == "" {
		return n.Format()
	}

	d := diff(n.Dest(), aref, bref)
	if d == "" {
		opts := map[string]string{"git": n.opts["git"], "ref": m.opts["ref"]}
		p := Mod{name: n.name, version: n.version, opts: opts}
		return p.Format()
	}

	newref, err := increment(m.opts["ref"])
	if err != nil {
		log.Printf("INFO: %v for %v. new one is used as is\n", err, n.name)
		return n.Format()
	}

	opts := map[string]string{"git": n.opts["git"], "ref": newref}
	p := Mod{name: n.name, version: n.version, opts: opts}
	return p.Format()
}

func gitDiff(wd, aref, bref string) string {
	buf := bytes.NewBuffer([]byte{})
	w := bufio.NewWriter(buf)
	run2(w, wd, "git", []string{"--no-pager", "diff", "-w", aref, bref})
	return buf.String()
}

func gitLog(wd, ref string) string {
	buf := bytes.NewBuffer([]byte{})
	w := bufio.NewWriter(buf)
	run2(w, wd, "git", []string{"log", ref, "-s", "--format=%H", "-n1"})
	return strings.TrimSpace(buf.String())
}

func printGitPush(c *cli.Context, a, b string) {
	remote := c.String("remote-name")
	diff(c, a, b, func(m, n Mod, aref, bref string) {
		d := gitDiff(n.Dest(), aref, bref)
		if d == "" {
			return
		}
		newref, err := increment(aref)
		if err != nil {
			log.Printf("WARN: %v for %v\n", err, m.name)
			newref = c.String("initial-release-branch")
		}
		oldref := gitLog(n.Dest(), bref)
		fmt.Printf("(cd %v; git push %v %v %v)\n", remote, n.Dest(), oldref, newref)
	})
}

func Diff(c *cli.Context, a, b string) {
	diff(c, a, b, func(m, n Mod, aref, bref string) {
		fmt.Println(n.Dest(), aref, bref)
		run2(os.Stdout, n.Dest(), "git", []string{"--no-pager", "diff", "-w", aref, bref})
	})
}

func parse(f string) []Mod {
	ar := newReader(f)
	mods, err := parsePuppetfile(ar)
	if err != nil {
		log.Fatalln(err)
	}
	return mods
}

func diff(c *cli.Context, a, b string, f DiffFunc) {
	modulepath = c.String("modulepath")

	am := parse(a)
	bm := parse(b)

	re, err := regexp.Compile(c.String("includes"))
	if err != nil {
		log.Fatalln(err)
		return
	}

	for _, n := range bm {
		if !re.MatchString(n.name) {
			continue
		}
		m, err := findModIn(am, n)
		if err != nil {
			log.Printf("WARN: %v in %s\n", err, a)
			continue
		}
		bref := n.opts["ref"]
		if bref == "" {
			continue
		}
		aref := m.opts["ref"]
		if aref == "" {
			continue
		}
		f(m, n, m.opts["ref"], n.opts["ref"])
	}
}
