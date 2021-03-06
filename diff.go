package librarianpuppetgo

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
)

func findModIn(mods []Mod, m Mod) (Mod, error) {
	for _, i := range mods {
		if i.name == m.name {
			return i, nil
		}
	}
	return Mod{}, fmt.Errorf("missing %s", m.name)
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

const (
	STAT    = "STAT"
	FULL    = "FULL"
	SUMMARY = "SUMMARY"
)

func Diff(a, b string, dirs []string, mode string) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	mode = strings.ToUpper(mode)

	diff(a, b, func(oldm, newm Mod, oldref, newref string) {
		args := []string{"--no-pager", "diff", "-w"}
		if mode == STAT {
			args = append(args, "--stat")
		}
		args = append(args, []string{oldref, newref, "--"}...)
		args = append(args, dirs...)

		b := bytes.NewBuffer([]byte{})
		run2(b, newm.Dest(), "git", args)
		switch mode {
		case FULL, STAT:
			if b.String() == "" {
				fmt.Printf("# NO any diff: %v %v %v\n", newm.Dest(), oldref, newref)
			} else {
				fmt.Printf("# FOUND: %v %v %v\n", newm.Dest(), oldref, newref)
				fmt.Print(b.String())
			}
		case SUMMARY:
			sc := bufio.NewScanner(b)
			var add, del int
			for sc.Scan() {
				if len(sc.Text()) == 0 {
					continue
				}
				//fmt.Printf("%q\n", sc.Text())
				switch sc.Text()[0] {
				case '+':
					add++
				case '-':
					del++
				default:
				}
			}
			if add > 0 || del > 0 {
				fmt.Fprintf(w, "%v\t %v insertion(+), %v deletion(-)\tbetween %v and %v\n", newm.name, add, del, oldref, newref)
			}
		default:
			log.Fatalf("unknown mode: %v", mode)
		}
	})

	w.Flush()
}

func parse(f string) []Mod {
	ar := newReader(f)
	mods, err := parsePuppetfile(ar)
	if err != nil {
		log.Fatalln(err)
	}
	return mods
}

type DiffFunc func(oldm, newm Mod, oldref, newref string)

func diff(oldfile, newfile string, f DiffFunc) {
	oldmods := parse(oldfile)
	newmods := parse(newfile)

	for _, newm := range newmods {
		oldm, err := findModIn(oldmods, newm)
		if err != nil {
			log.Printf("INFO: %v in %s\n", err, oldfile)
			continue
		}
		newref := newm.Ref()
		if newref == "" {
			logger.Printf("INFO: missing ref in %v of %v", newm.name, newfile)
			continue
		}
		oldref := oldm.Ref()
		if oldref == "" {
			logger.Printf("INFO: missing ref in %v of %v", oldm.name, oldfile)
			continue
		}
		f(oldm, newm, oldref, newref)
	}
}
