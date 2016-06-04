package librarianpuppetgo

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
)

func (g *Git) PushCmds(src, dst string) {
	srcmods := parse(src)
	dstmods := parse(dst)

	for _, srcm := range srcmods {
		newm, err := findModIn(dstmods, srcm)
		if err != nil {
			fmt.Fprintf(g.Writer, "# %v is missing in %v\n", srcm.name, dst)
			continue
		}
		if srcm.opts["ref"] != "" && newm.opts["ref"] != "" {
			if g.Diff(srcm.Dest(), srcm.opts["ref"], newm.opts["ref"]) == "" {
				continue
			}
		}
		s, err := g.PushCmd(srcm, newm)
		if err != nil {
			log.Printf("WARN: %v\n", err)
			continue
		}
		fmt.Fprintf(g.Writer, "%s\n", s)
		logger.Printf("%v: %v", srcm.name, s)
	}
}

func PrintGitPushCmds(remote, src, dst string) {
	g := NewGit()
	g.Remote = remote
	g.PushCmds(src, dst)
	fmt.Print()
}

type Git struct {
	Writer   io.Writer
	Remote   string
	IsCommit func(wd, sha1 string) bool
	IsBranch func(wd, name string) bool
	IsTag    func(wd, name string) bool
	Sha1     func(wd, ref string) string
	Diff     func(wd, srcref, dstref string) string
}

func NewGit() *Git {
	return &Git{
		Writer:   os.Stdout,
		Remote:   "origin",
		IsCommit: isCommit,
		IsBranch: isBranch,
		IsTag:    isTag,
		Sha1:     gitSha1,
		Diff:     gitDiff,
	}
}

func minorVersionNumber(s string) (int, error) {
	re := regexp.MustCompile(releaseBranchPattern).FindAllStringSubmatch(s, -1)
	if len(re) == 0 {
		return -1, fmt.Errorf("%v didn't match %v", s, releaseBranchPattern)
	}
	minor := re[0][1]
	v, err := strconv.Atoi(minor)
	if err != nil {
		return -1, fmt.Errorf("%v cannot be converted to int", minor)
	}
	return v, nil
}

func (g Git) PushCmd(oldm, newm Mod) (string, error) {
	oldref := oldm.opts["ref"]
	if oldref == "" {
		if oldm.version == "" {
			return fmt.Sprintf("# %v/%v doesn't have :ref", oldm.user, oldm.name), nil
		} else {
			return fmt.Sprintf("# %v/%v %v doesn't have :ref", oldm.user, oldm.name, oldm.version), nil
		}
	}

	srcref := newm.opts["ref"]
	if g.IsTag(newm.Dest(), oldref) && g.IsBranch(newm.Dest(), srcref) {
		if v, err := minorVersionNumber(srcref); err != nil {
			return fmt.Sprintf("# WARN: %s cannot be parsed minor version for %v", srcref, newm.name), nil
		} else {
			return fmt.Sprintf("(cd modules/%v; git push origin %v:v0.%d.0)", newm.name, srcref, v), nil
		}
	}

	dstref, err := increment(oldref)
	if err != nil {
		return fmt.Sprintf("# %v is referred at %v", newm.Dest(), newm.opts["ref"]), nil
	}

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
