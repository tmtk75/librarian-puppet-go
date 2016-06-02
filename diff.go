package librarianpuppetgo

import (
	"fmt"
	"log"
	"os"
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
	return Mod{}, fmt.Errorf("missing %s", m.name)
}

type DiffFunc func(m, n Mod, aref, bref string)

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
