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
	body           string
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
		b := bytes.NewBuffer([]byte{})
		run2(b, mod.Dest(), c[0], c[1:])
		if opts.body == "" {
			fmt.Fprint(os.Stdout, b)
		} else {
			v := struct{ Name, Ref, Value string }{mod.name, mod.opts["ref"], b.String()}
			s, err := replaceWith(opts.body, v)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Fprint(os.Stdout, s)
		}
		fmt.Fprint(os.Stdout, s)
	}
}

func replaceWithMod(t string, m Mod) (string, error) {
	return replaceWith(t, struct{ Name, Ref string }{m.name, m.opts["ref"]})
}

func replaceWith(templ string, v interface{}) (string, error) {
	t, err := template.New("").Parse(templ)
	if err != nil {
		return "", err
	}
	b := bytes.NewBuffer([]byte{})
	t.Execute(b, v)
	s := strings.Replace(b.String(), "\\n", "\n", -1)
	s = strings.Replace(s, "\\t", "\t", -1)
	return s, err
}

func makeEachArgs(args []string, m Mod) ([]string, error) {
	c := make([]string, len(args))
	for i, e := range args {
		s, err := replaceWithMod(e, m)
		if err != nil {
			return []string{e}, err
		}
		c[i] = s
	}
	return c, nil
}
