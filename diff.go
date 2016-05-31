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
	"strconv"

	"github.com/tmtk75/cli"
)

func findModIn(mods []Mod, m Mod) (Mod, error) {
	for _, i := range mods {
		if i.name == m.name {
			return i, nil
		}
	}
	return Mod{}, fmt.Errorf("missing for %s", m.name)
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

func normalize(c *cli.Context, a string) {
	modulepath = c.String("modulepath")
	mods := parse(a)
	for _, m := range mods {
		fmt.Println(m.Format())
	}
}

func bumpUp(c *cli.Context, a, b string) {
	modulepath = c.String("modulepath")

	am := parse(a)
	bm := parse(b)

	for _, n := range bm {
		m, err := findModIn(am, n)
		if err != nil {
			fmt.Println(n.Format())
			continue
		}
		aref := m.opts["ref"]
		bref := n.opts["ref"]
		if aref == "" || bref == "" {
			fmt.Println(n.Format())
			continue
		}
		d := gitDiff(n.Dest(), aref, bref)
		if d != "" {
			p := Mod{name: n.name, version: n.version, opts: map[string]string{"git": n.opts["git"], "ref": m.opts["ref"]}}
			fmt.Println(p.Format())
		} else {
			newref, err := increment(m.opts["ref"])
			if err != nil {
				log.Printf("WARN: %v for %v\n", err, n.name)
				fmt.Println(n.Format())
				continue
			}
			p := Mod{name: n.name, version: n.version, opts: map[string]string{"git": n.opts["git"], "ref": newref}}
			fmt.Println(p.Format())
		}
	}
}

func gitDiff(wd, aref, bref string) string {
	buf := bytes.NewBuffer([]byte{})
	w := bufio.NewWriter(buf)
	run2(w, wd, "git", []string{"--no-pager", "diff", "-w", aref, bref})
	return buf.String()
}

func printGitPush(c *cli.Context, a, b string) {
	diff(c, a, b, func(m, n Mod, aref, bref string) {
		d := gitDiff(n.Dest(), aref, bref)
		if d == "" {
			return
		}
		v, err := increment(aref)
		if err != nil {
			log.Printf("WARN: %v for %v\n", err, m.name)
			return
		}
		fmt.Printf("(cd %v; git push origin %v %v)\n", n.Dest(), bref, v)
	})
}

func Diff(c *cli.Context, a, b string) {
	diff(c, a, b, func(m, n Mod, aref, bref string) {
		fmt.Println(n.Dest(), aref, bref)
		run2(os.Stdout, n.Dest(), "git", []string{"--no-pager", "diff", "-w", aref, bref})
		//fmt.Printf("git push origin %v %v\n", aref, bref)
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
