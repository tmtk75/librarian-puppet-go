package librarianpuppetgo

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
)

type semver struct {
	major int
	minor int
	patch int
}

type semvers []semver

func (v semver) String() string {
	if v.patch == -1 {
		return fmt.Sprintf("%d.%d", v.major, v.minor)
	}
	return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
}

func (v semvers) Len() int {
	return len(v)
}

func (v semvers) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v semvers) Less(i, j int) bool {
	if v[i].major < v[j].major {
		return true
	} else if v[i].major > v[j].major {
		return false
	} else {
		if v[i].minor < v[j].minor {
			return true
		} else if v[i].minor > v[j].minor {
			return false
		} else {
			if v[i].patch < v[j].patch {
				return true
			} else if v[i].patch > v[j].patch {
				return false
			} else {
				return false
			}
		}
	}
}

func Sort(a string) (string, error) {
	b := make(semvers, 0)
	for _, e := range strings.Split(a, "\n") {
		s := strings.TrimSpace(e)
		if s == "" {
			continue
		}
		x, y, z, err := semanticVersion(s)
		if err != nil {
			return "", fmt.Errorf("%v", err)
		}
		b = append(b, semver{major: x, minor: y, patch: z})
	}
	sort.Sort(b)

	r := make([]string, 0)
	for _, e := range b {
		r = append(r, e.String())
	}
	return strings.Join(r, "\n"), nil
}

func SortCmd() {
	e, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalln(err)
	}
	r, err := Sort(string(e))
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Print(r)
}
