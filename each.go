package librarianpuppetgo

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
)

type eachOpts struct {
	prefix, suffix string
}

func (g *Git) Each(path string, cmds []string, opts eachOpts) {
	mods := parse(path)
	for _, mod := range mods {
		c, err := makeEachArgs(cmds, mod)
		if err != nil {
			log.Fatalln(err)
		}
		p, err := replaceWithMod(opts.prefix, mod)
		if err != nil {
			log.Fatalln(err)
		}
		s, err := replaceWithMod(opts.suffix, mod)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Fprint(os.Stdout, p)
		run2(os.Stdout, mod.Dest(), c[0], c[1:])
		fmt.Fprint(os.Stdout, s)
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
		s := strings.Replace(b.String(), "\\n", "\n", -1)
		s = strings.Replace(s, "\\t", "\t", -1)
		c[i] = s
	}
	return c, nil
}
