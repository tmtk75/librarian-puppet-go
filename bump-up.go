package librarianpuppetgo

import (
	"fmt"
	"log"
)

func bumpUp(a, b string, init string) {
	am := parse(a)
	bm := parse(b)

	for _, n := range bm {
		e, _ := bumpUpMod(n, am, init, a, gitDiff)
		fmt.Println(e)
	}
}

type bumpDiff func(wd, a, b string) string

func bumpUpMod(n Mod, mods []Mod, rel, filename string, diff bumpDiff) (string, error) {
	g := NewGit()
	g.Diff = diff
	return g.bumpUpMod(n, mods, rel, filename)
}

func (g *Git) bumpUpMod(n Mod, mods []Mod, rel, filename string) (string, error) {
	m, err := findModIn(mods, n)
	if err != nil {
		logger.Printf("INFO: %v in %v. old one is used\n", err, filename)
		if n.Ref() != "" {
			opts := map[string]string{"git": n.opts["git"], "ref": rel}
			p := Mod{name: n.name, version: n.version, opts: opts, user: n.user}
			return p.Format(), nil
		} else {
			return n.Format(), nil
		}
	}
	aref := m.Ref()
	bref := n.Ref()
	if aref == "" || bref == "" {
		return n.Format(), nil
	}

	if g.IsTag(m.Dest(), aref) && g.IsBranch(n.Dest(), bref) {
		d := g.Diff(n.Dest(), aref, bref)
		if d == "" {
			opts := map[string]string{"git": n.opts["git"], "ref": m.Ref()}
			p := Mod{name: n.name, version: n.version, opts: opts}
			return p.Format(), nil
		}

		v, err := minorVersionNumber(bref)
		if err != nil {
			return "", err
		}
		_, b, c, err := semanticVersion(aref)
		if err != nil {
			return "", err
		}
		if v > b {
			r := fmt.Sprintf("v0.%v.0", v)
			opts := map[string]string{"git": n.opts["git"], "ref": r}
			return Mod{name: n.name, version: n.version, opts: opts}.Format(), nil
		}
		r := fmt.Sprintf("v0.%v.%v", b, c+1)
		opts := map[string]string{"git": n.opts["git"], "ref": r}
		return Mod{name: n.name, version: n.version, opts: opts}.Format(), nil
	}

	d := g.Diff(n.Dest(), aref, bref)
	if d == "" {
		if n.opts["git"] == "" {
			return n.Format(), nil
		}
		opts := map[string]string{"git": n.opts["git"], "ref": m.Ref()}
		p := Mod{name: n.name, version: n.version, opts: opts}
		return p.Format(), nil
	}

	newref, err := increment(m.Ref())
	if err != nil {
		log.Printf("INFO: %v for %v. new one is used as is\n", err, n.name)
		return n.Format(), nil
	}

	opts := map[string]string{"git": n.opts["git"], "ref": newref}
	p := Mod{name: n.name, version: n.version, opts: opts}
	return p.Format(), nil
}
