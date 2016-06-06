package librarianpuppetgo

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/jawher/mow.cli"
)

func CliMain() {
	app := cli.App("librarian-puppet-go", "Support a workflow for puppet modules")
	app.Version("version", "0.3.1")
	var (
		verbose    = app.Bool(cli.BoolOpt{Name: "v verbose", EnvVar: "LP_VERBOSE", Desc: "Show logs verbosely"})
		modulepath = app.String(cli.StringOpt{Name: "module-path", Value: "modules", Desc: "Path to be for modules"})
	)
	app.Before = func() {
		if *verbose {
			logger = log.New(os.Stderr, "", log.LstdFlags)
		}
		modulePath = *modulepath
	}
	var (
		fileArg     = cli.StringArg{Name: "FILE", Desc: "A puppetfile path"}
		throttleOpt = cli.IntOpt{Name: "throttle", Value: 0, EnvVar: "LP_THROTTLE",
			Desc: `Throttle number of concurrent processes.
                 Max is number of mod, min is 1. Max is used if 0 or negative number is given.`}
		forceOpt = cli.BoolOpt{Name: "force f", Desc: "checkout with --force"}
	)
	f := func(b bool) func(c *cli.Cmd) {
		return func(c *cli.Cmd) {
			file := c.String(fileArg)
			throttle := c.Int(throttleOpt)
			force := c.Bool(forceOpt)
			c.Spec = "[OPTIONS] FILE"
			c.Action = func() {
				c := installCmd{
					throttle:      *throttle,
					forceCheckout: *force,
					onlyCheckout:  b,
				}
				c.Main(*file)
			}
		}
	}
	app.Command(
		"install",
		"Install modules with a puppetfile",
		f(false),
	)
	app.Command(
		"checkout",
		"Checkout modules without network access",
		f(true),
	)
	app.Command(
		"format",
		"Format a puppetfile",
		func(c *cli.Cmd) {
			c.LongDesc = `Format a puppetfile by removing comments/blank lines, good whitespacing,
and sorting with mod name.

e.g) norm puppetfile`
			a := c.String(cli.StringArg{Name: "FILE", Desc: "puppetfile to be formated"})
			b := c.Bool(cli.BoolOpt{Name: "w overwrite", Desc: "Overwrite"})
			c.Spec = "[OPTIONS] FILE"
			c.Action = func() {
				Format(*a, *b)
			}
		},
	)
	app.Command(
		"diff",
		"Compare two files with local branches which are chekced out",
		func(c *cli.Cmd) {
			c.LongDesc = `Compare two files with local branches which are chekced out.
You need to check out local branches to be used beforehand.
To do that, 'install' command can be used.

e.g) diff Puppetfile.staging Puppetfile.development
     diff Puppetfile.production Puppetfile.staging manifests templates`
			a := c.String(cli.StringArg{Name: "SRC", Desc: "Source puppetfile"})
			b := c.String(cli.StringArg{Name: "DST", Desc: "Destination puppetfile"})
			d := c.Strings(cli.StringsArg{Name: "DIRS", Desc: "Directories to be compared"})
			c.Spec = "SRC DST [DIRS...]"
			c.Action = func() {
				Diff(*a, *b, *d)
			}
		},
	)
	app.Command(
		"git-push",
		"Print git commands to release",
		func(c *cli.Cmd) {
			a := c.String(cli.StringArg{Name: "SRC", Desc: "Source puppetfile"})
			b := c.String(cli.StringArg{Name: "DST", Desc: "Destination puppetfile"})
			remoteName := c.String(cli.StringOpt{Name: "remote-name", Value: "origin", Desc: "Remote name"})
			c.Spec = "SRC DST"
			c.Action = func() {
				PrintGitPushCmds(*remoteName, *a, *b)
			}
		},
	)
	app.Command(
		"bump-up",
		`Print bumped-up puppetfile based on given file`,
		func(c *cli.Cmd) {
			c.LongDesc = `Print bumped-up puppetfile based on given file

		e.g) bump-up Puppetfile.staging Puppetfile.development`
			a := c.String(cli.StringArg{Name: "SRC", Desc: "Source puppetfile"})
			b := c.String(cli.StringArg{Name: "DST", Desc: "Destination puppetfile"})
			relBranch := c.String(cli.StringOpt{Name: "release-branch", Value: "release/0.1", Desc: "Release branch name used first"})
			c.Spec = "SRC DST"
			c.Action = func() {
				bumpUp(*a, *b, *relBranch)
			}
		},
	)
	app.Run(os.Args)
}

var logger = log.New(ioutil.Discard, "", log.LstdFlags)
