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

func increment(s string) (string, error) {
	re := regexp.MustCompile(`release/0.([0-9]+)`).FindAllStringSubmatch(s, -1)
	if len(re) == 0 {
		return "", fmt.Errorf("%v didn't match", s)
	}
	minor := re[0][1]
	v, err := strconv.Atoi(minor)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("release/0.%d", v+1), nil
}

func Release(c *cli.Context, a, b string) {
	diff(c, a, b, func(m, n Mod, aref, bref string) {
		buf := bytes.NewBuffer([]byte{})
		w := bufio.NewWriter(buf)
		run2(w, n.Dest(), "git", []string{"--no-pager", "diff", "-w", aref, bref})
		if buf.String() == "" {
			return
		}
		v, err := increment(aref)
		if err != nil {
			log.Printf("WARN: %v\n", err)
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

func diff(c *cli.Context, a, b string, f DiffFunc) {
	modulepath = c.String("modulepath")

	ar := newReader(a)
	am, err := parsePuppetfile(ar)
	if err != nil {
		log.Fatalln(err)
		return
	}

	br := newReader(b)
	bm, err := parsePuppetfile(br)
	if err != nil {
		log.Fatalln(err)
		return
	}

	re, err := regexp.Compile(c.String("includes"))
	if err != nil {
		log.Fatalln(err)
		return
	}
	for _, m := range bm {
		if !re.MatchString(m.name) {
			continue
		}
		n, err := findModIn(am, m)
		if err != nil {
			log.Printf("WARN: %v in %s\n", err, a)
			continue
		}
		bref := m.opts["ref"]
		if bref == "" {
			continue
		}
		aref := n.opts["ref"]
		if aref == "" {
			continue
		}
		f(m, n, aref, bref)
	}
}
