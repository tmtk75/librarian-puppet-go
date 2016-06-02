package librarianpuppetgo

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

func parsePuppetfile(i io.Reader) ([]Mod, error) {
	r := bufio.NewReader(i)
	incs := make([][]Mod, 0)
	mods := make([]Mod, 0)
	for {
		b, _, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		s := string(b)
		if s == "" || s[0] == '#' {
			continue
		}

		if a := isInclude(s); a != "" {
			if len(mods) > 0 {
				return mods, fmt.Errorf("[error] include(s) must be at the top: %v", s)
			}
			logger.Printf("include: '%v'\n", a)

			r := newReader(a)
			defer r.Close()
			inc, err := parsePuppetfile(bufio.NewReader(r))
			if err != nil {
				return mods, err
			}
			incs = append(incs, inc)
			continue
		}

		m, err := parseMod(s)
		if err != nil {
			if _, ok := (err).(Ignorable); ok {
				continue
			}
			logger.Printf("[warn] %v\n", err)
			return mods, err
		}
		mods = append(mods, m)
	}

	return packMods(&incs, &mods), nil
}

func parseMod(i string) (Mod, error) {
	s := regexp.MustCompile(`#.*$`).ReplaceAllLiteralString(i, "")

	re := regexp.MustCompile(`\s*forge\s.*`).FindAllStringSubmatch(s, -1)
	if len(re) > 0 {
		return Mod{}, Ignorable{fmt.Errorf("'%v' '/'", s)}
	}

	re = regexp.MustCompile(`^mod\s+["']([a-z/_0-9]+)['"]\s*(,\s*["'](\d+\.\d+(\.\d+)?)["'])?$`).FindAllStringSubmatch(s, -1)
	if len(re) > 0 {
		n := re[0][1]
		v := ""
		if len(re[0]) > 3 {
			v = re[0][3]
		}
		nn := strings.Split(n, "/")
		if len(nn) != 2 {
			return Mod{}, fmt.Errorf("'%v' should contain one '/'", n)
		}
		return Mod{name: nn[1], user: nn[0], version: v, opts: ModOpts{}}, nil
	}

	re = regexp.MustCompile(`^mod\s+([^,]+),(.*?)$`).FindAllStringSubmatch(s, -1)
	if len(re) > 0 {
		return Mod{name: unquote(re[0][1]), opts: parseOpts(re[0][2])}, nil
	}

	return Mod{}, fmt.Errorf("cannot parse: %v", s)
}

// pack all mods as a slice
func packMods(incs *[][]Mod, mods *[]Mod) []Mod {
	all := append(*incs, *mods)
	n2m := map[string]*Mod{}
	result := make([]*Mod, 0)
	for _, i := range all {
		for _, e := range i {
			if n2m[e.name] == nil {
				a := e // copy
				n2m[e.name] = &a
				result = append(result, &a)
			} else {
				n2m[e.name].Replace(&e)
			}
		}
	}

	res := make([]Mod, len(result))
	for i, v := range result {
		res[i] = *v
	}
	return res
}

func isInclude(s string) string /* filename */ {
	re := regexp.MustCompile(`include\s+["'](.*?)["']`).FindAllStringSubmatch(s, -1)
	if len(re) == 0 {
		return ""
	}
	//log.Printf("[TRACE] %v\n", re)
	return re[0][1]
}

// Marker to ignore intentionally
type Ignorable struct {
	error
}

func unquote(s string) string {
	return regexp.MustCompile(`["']`).ReplaceAllString(s, "")
}

func parseOpts(s string) ModOpts {
	m := make(ModOpts)
	for _, e := range strings.Split(s, ",") {
		re := regexp.MustCompile(`:([a-z_]+)\s*=>\s*(.*)$`).FindAllStringSubmatch(e, -1)
		//fmt.Printf("%v", re)
		if len(re) == 0 {
			continue
		}
		m[re[0][1]] = unquote(re[0][2])
	}
	return m
}
