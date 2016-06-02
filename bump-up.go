package librarianpuppetgo

import (
	"fmt"
	"log"

	"github.com/tmtk75/cli"
)

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
		logger.Printf("INFO: %v in %v. old one is used\n", err, filename)
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
