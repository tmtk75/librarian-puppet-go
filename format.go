package librarianpuppetgo

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/tmtk75/cli"
)

type Mods []Mod

func (v Mods) Len() int {
	return len(v)
}

func (v Mods) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v Mods) Less(i, j int) bool {
	e := strings.Compare(v[i].name, v[j].name)
	return e < 0
}

func Format(c *cli.Context, a string) {
	mods := parse(a)
	s := format(mods)
	if c.Bool("overwrite") {
		f, err := os.Open(a)
		defer f.Close()
		if err != nil {
			log.Fatalln(err)
		}
		f.Write([]byte(s))
	} else {
		fmt.Println(s)
	}
}

func format(mods []Mod) string {
	sort.Sort(Mods(mods))
	buf := bytes.NewBuffer([]byte{})
	w := buf //bufio.NewWriter(buf)
	for _, m := range mods {
		_, err := fmt.Fprintln(w, m.Format())
		//fmt.Println(i)
		if err != nil {
			log.Fatalln(err)
		}
	}
	//fmt.Println("mod:", len(mods))
	//fmt.Println("len:", len(buf.String()))
	return buf.String()
}
