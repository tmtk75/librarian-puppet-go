package librarianpuppetgo

import (
	"log"
	"os"
	"text/template"
)

func (g *Git) Each(path, prefix string, cmds []string) {
	mods := parse(path)
	for _, mod := range mods {
		c := make([]string, len(cmds))
		for i, e := range cmds {
			switch e {
			case "{{.Name}}":
				c[i] = mod.name
			case "{{.Ref}}":
				c[i] = mod.opts["ref"]
			default:
				c[i] = e
			}
		}
		t, err := template.New(mod.name).Parse(prefix)
		if err != nil {
			log.Fatalln(err)
		}
		t.Execute(os.Stdout, struct {
			Name string
			Ref  string
		}{mod.name, mod.opts["ref"]})
		run2(os.Stdout, mod.Dest(), c[0], c[1:])
	}
}
