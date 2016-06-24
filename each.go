package librarianpuppetgo

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"text/template"
)

type eachOpts struct {
	header, prefix string
}

func (g *Git) Each(path string, cmds []string, opts eachOpts) {
	mods := parse(path)
	for _, mod := range mods {
		c, err := makeEachArgs(cmds, mod)
		if err != nil {
			log.Fatalln(err)
		}

		r, err := replaceWithMod(opts.prefix, mod)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Print(r)
		run2(os.Stdout, mod.Dest(), c[0], c[1:])
	}
}

func replaceWithMod(templ string, m Mod) (string, error) {
	s, err := makeEachArgs([]string{templ}, m)
	return s[0], err
}

func makeEachArgs(args []string, m Mod) ([]string, error) {
	c := make([]string, len(args))
	for i, e := range args {
		t, err := template.New("").Parse(e)
		if err != nil {
			return []string{e}, err
		}
		b := bytes.NewBuffer([]byte{})
		t.Execute(b, struct {
			Name string
			Ref  string
		}{m.name, m.opts["ref"]})
		c[i] = b.String()
	}
	return c, nil
}
