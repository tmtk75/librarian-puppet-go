package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/hcl"
)

func parse(conf string) ([]Mod, error) {
	obj, err := hcl.Parse(conf)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	mods := make([]Mod, 0)

	for _, mod := range obj.Get("mod", true).Elem(false) {
		m := Mod{opts: map[string]string{}}
		for _, e := range mod.Elem(true) {
			switch e.Type {
			case hcl.ValueTypeObject:
				for _, a := range e.Elem(true) {
					switch a.Type {
					case hcl.ValueTypeObject:
						s := strings.Split(e.Key, "/")
						m.name = s[1]
						m.user = s[0]
						m.version = a.Key
					case hcl.ValueTypeString:
						m.name = e.Key
						m.opts[a.Key] = a.Value.(string)
					}
				}
			case hcl.ValueTypeString:
				switch e.Key {
				case "user":
					m.user = e.Value.(string)
				case "name":
					m.name = e.Value.(string)
				case "version":
					m.version = e.Value.(string)
				}
			}
		}
		mods = append(mods, m)
		continue
	}

	return mods, nil
}

func assert(m Mod, name, git, ref, user, version string) bool {
	return m.name == name &&
		m.opts["git"] == git &&
		m.opts["ref"] == ref &&
		m.user == user &&
		m.version == version
}

func Test_parse(t *testing.T) {
	var mods []Mod
	var err error

	mods, err = parse(`mod "puppetlabs/stdlib" "4.1.0" {}`)
	if !(len(mods) == 1 && err == nil && assert(mods[0], "stdlib", "", "", "puppetlabs", "4.1.0")) {
		t.Errorf("%v", mods)
	}

	mods, err = parse(`mod { user = "puppetlabs" name = "stdlib" version = "4.1.0" }`)
	if !(len(mods) == 1 && err == nil && assert(mods[0], "stdlib", "", "", "puppetlabs", "4.1.0")) {
		t.Errorf("%v", mods)
	}

	mods, err = parse(`mod "foobar" { git = "a url" ref = "master" }`)
	if !(len(mods) == 1 && err == nil && assert(mods[0], "foobar", "a url", "master", "", "")) {
		t.Errorf("%v", mods)
	}
}
