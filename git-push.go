package librarianpuppetgo

import (
	"bytes"
	"fmt"
	"log"

	"github.com/tmtk75/cli"
)

func (g *Git) PushCmds(src, dst string) string {
	srcmods := parse(src)
	dstmods := parse(dst)

	buf := bytes.NewBuffer([]byte{})
	for _, srcm := range srcmods {
		newm, err := findModIn(dstmods, srcm)
		if err != nil {
			fmt.Fprintf(buf, "# %v is missing in %v", srcm.name, dst)
			continue
		}
		s, err := g.PushCmd(srcm, newm)
		if err != nil {
			log.Printf("WARN: %v\n", err)
			continue
		}
		fmt.Fprintf(buf, "%s\n", s)
		logger.Printf("%v: %v", srcm.name, s)
	}

	return buf.String()
}

func PrintGitPushCmds(c *cli.Context, src, dst string) {
	modulepath = c.String("modulepath")
	fmt.Print(NewGit().PushCmds(src, dst))
}

type Git struct {
	Remote   string
	IsCommit func(wd, sha1 string) bool
	IsBranch func(wd, name string) bool
	Sha1     func(wd, ref string) string
}

func NewGit() *Git {
	return &Git{
		Remote:   "origin",
		IsCommit: isCommit,
		IsBranch: isBranch,
		Sha1:     gitSha1,
	}
}

func (g Git) PushCmd(oldm, newm Mod) (string, error) {
	dstref, err := increment(oldm.opts["ref"])
	if err != nil {
		return fmt.Sprintf("# %v is referred at %v", newm.Dest(), newm.opts["ref"]), nil
	}

	srcref := newm.opts["ref"]
	if g.IsCommit(newm.Dest(), srcref) {
		return fmt.Sprintf(
			"(cd %v; git branch %v %v; git push %v %v:%v)",
			newm.Dest(), srcref, dstref, g.Remote, dstref, dstref), nil
	}
	if !g.IsBranch(newm.Dest(), srcref) {
		return "", fmt.Errorf("%v is neither tag or branch in %v", srcref, newm.Dest())
	}
	if g.Sha1 != nil {
		srcsha1 := g.Sha1(newm.Dest(), srcref)
		return fmt.Sprintf(
			"(cd %v; git branch %v %v; git push %v %v:%v)",
			newm.Dest(), srcsha1, dstref, g.Remote, dstref, dstref), nil
	}
	return fmt.Sprintf(
		"(cd %v; git push %v %v:%v)",
		newm.Dest(), g.Remote, srcref, dstref), nil

}
