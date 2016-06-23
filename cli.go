package librarianpuppetgo

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jawher/mow.cli"
)

func CliMain() {
	app := cli.App("librarian-puppet-go", "Support a workflow for puppet modules")
	app.Version("version", "0.3.2")
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
		forceOpt    = cli.BoolOpt{Name: "force f", Desc: "checkout with --force"}
		includesOpt = cli.StringOpt{Name: "includes-with-repository-name", Value: ".*", Desc: "Specify modules to be installed"}
	)
	f := func(b bool) func(c *cli.Cmd) {
		return func(c *cli.Cmd) {
			file := c.String(fileArg)
			throttle := c.Int(throttleOpt)
			force := c.Bool(forceOpt)
			includes := c.String(includesOpt)
			c.Spec = "[OPTIONS] FILE"
			c.Action = func() {
				c := installCmd{
					throttle:             *throttle,
					forceCheckout:        *force,
					onlyCheckout:         b,
					includesWithRepoName: *includes,
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
			m := c.String(cli.StringOpt{Name: "mode", Value: STAT, Desc: fmt.Sprintf("Specify diff mode. %v, %v and %v", STAT, FULL, SUMMARY)})
			c.Spec = "[OPTIONS] SRC DST [DIRS...]"
			c.Action = func() {
				Diff(*a, *b, *d, *m)
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
	app.Command(
		"semver",
		`Manage semver`,
		func(c *cli.Cmd) {
			c.Command(
				"sort",
				`Sort semver in ascending order`,
				func(c *cli.Cmd) {
					c.LongDesc = `Sort sever in ascending order

Accepted pattern is like this.

  v0.1.0
  v0.1
  0.1.0
  0.1

Output doesn't have "v" prefix if it has "v".
`
					c.Action = func() {
						SortCmd()
					}
				},
			)
		},
	)
	app.Command(
		"each",
		"Exec a command you want for all modules",
		func(c *cli.Cmd) {
			src := c.String(cli.StringArg{Name: "FILE", Desc: "A puppetfile"})
			args := c.Strings(cli.StringsArg{Name: "ARGS", Desc: "Command and args"})
			prefix := c.String(cli.StringOpt{Name: "prefix p", Value: "", Desc: "Prefix template"})
			c.LongDesc = `e.g)
  each -- Puppetfile git --no-pager show .

  each --prefix "{{.Name}}/{{.Ref}}    " -- Puppetfile git --no-pager log {{.Ref}} --format=%H -n1

  In prefix template:
          .Name    mod name
          .Ref     :ref

  In ARGS:
          {{.Name}}     replaced with mod name
          {{.Ref}}      replaced with :ref
	  `
			c.Spec = "[OPTIONS] FILE ARGS..."
			c.Action = func() {
				NewGit().Each(*src, *prefix, *args)
			}
		},
	)
	app.Run(os.Args)
}

var logger = log.New(ioutil.Discard, "", log.LstdFlags)
