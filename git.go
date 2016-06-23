package librarianpuppetgo

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

var modulePath string = "modules"

func isCommit(dest, sha1 string) bool {
	cmd := exec.Command("git", "show", "-q", sha1)
	a := checkExitCode(dest, cmd)
	b := !isBranch(dest, sha1)
	c := !isTag(dest, sha1)
	return a && b && c
}

func isBranch(dest, name string) bool {
	return isRef(dest, "heads", name)
}

func isTag(dest, tag string) bool {
	return isRef(dest, "tags", tag)
}

func isRef(dest, kind, tag string) bool {
	cmd := exec.Command("git", "show-ref", "-q", "--verify", "refs/"+kind+"/"+tag)
	return checkExitCode(dest, cmd)
}

func checkExitCode(dest string, cmd *exec.Cmd) bool {
	cmd.Dir = dest
	err := cmd.Run()
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus() == 0
		}
	}
	if err != nil {
		log.Fatalf("[error] %v\t%v\t%v\n", err, dest)
	}
	return true
}

func gitSha1(wd, ref string) string {
	buf := bytes.NewBuffer([]byte{})
	w := bufio.NewWriter(buf)
	run2(w, wd, "git", []string{"log", ref, "-s", "--format=%H", "-n1"})
	return strings.TrimSpace(buf.String())
}

func gitClone(url, dest string) error {
	return run("", "git", []string{"clone", url, dest})
}

func gitFetch(dest string) error {
	return run(dest, "git", []string{"fetch", "-p"})
}

func gitPull(dest, ref string) error {
	return run(dest, "git", []string{"pull", "origin", ref})
}

func gitCheckout(dest, ref string, force bool) error {
	if ref == "" {
		ref = "master"
	}
	if force {
		return run(dest, "git", []string{"checkout", "--force", ref})
	}
	return run(dest, "git", []string{"checkout", ref})
}

func run(wd, s string, args []string) error {
	cmd := exec.Command(s, args...)
	cmd.Dir = wd
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	buf := bytes.NewBuffer([]byte{})
	cmd.Stderr = bufio.NewWriter(buf)
	logger.Printf("start: %v %v in %v", s, args, wd)
	now := time.Now()
	err := cmd.Run()
	prefix := "done"
	if err != nil {
		prefix = "error"
		log.Printf("[error] %v\t%v\t%v\n", err, args, buf)
	}
	elapsed := time.Since(now)
	logger.Printf("%v: %v %v %v in %v", prefix, elapsed, s, args, wd)
	return err
}

func gitDiff(wd, aref, bref string) string {
	buf := bytes.NewBuffer([]byte{})
	w := bufio.NewWriter(buf)
	run2(w, wd, "git", []string{"--no-pager", "diff", "-w", aref, bref, "--"})
	return buf.String()
}

func run2(w io.Writer, wd, s string, args []string) {
	cmd := exec.Command(s, args...)
	cmd.Dir = wd
	cmd.Stderr = os.Stderr
	cmd.Stdout = w
	err := cmd.Run()
	if err != nil {
		//log.Fatalln(err)
		//fmt.Println(wd, s, args)
	}
}
