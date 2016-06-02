package librarianpuppetgo

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/tmtk75/cli"
)

func printGitPush(c *cli.Context, a, b string) {
	remote := c.String("remote-name")
	diff(c, a, b, func(m, n Mod, aref, bref string) {
		d := gitDiff(n.Dest(), aref, bref)
		if d == "" {
			return
		}
		newref, err := increment(aref)
		if err != nil {
			log.Printf("WARN: %v for %v\n", err, m.name)
			newref = c.String("initial-release-branch")
		}
		oldref := bref
		if c.Bool("use-sha1") {
			oldref = gitLog(n.Dest(), bref)
		}
		fmt.Printf("(cd %v; git push %v %v %v)\n", remote, n.Dest(), oldref, newref)
	})
}

func gitLog(wd, ref string) string {
	buf := bytes.NewBuffer([]byte{})
	w := bufio.NewWriter(buf)
	run2(w, wd, "git", []string{"log", ref, "-s", "--format=%H", "-n1"})
	return strings.TrimSpace(buf.String())
}
