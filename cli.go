package librarianpuppetgo

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/tmtk75/cli"
)

func CliMain() {
	app := cli.NewApp()
	app.Name = "librarian-puppet-go"
	app.Version = "0.3.0"
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "verbose", Usage: "Show logs verbosely"},
	}
	app.Before = func(c *cli.Context) error {
		if c.Bool("verbose") {
			logger = log.New(os.Stderr, "", log.LstdFlags)
		}
		return nil
	}
	app.Commands = []cli.Command{
		cli.Command{
			Name:  "install",
			Usage: "Install modules with a Puppetfile given thru stdin",
			Args:  "[filename]",
			Flags: flags,
			Action: func(c *cli.Context) {
				realMain(c, "install")
			},
		},
		cli.Command{
			Name:  "checkout",
			Usage: "Checkout modules without network access",
			Args:  "[filename]",
			Flags: flags,
			Action: func(c *cli.Context) {
				onlyCheckout = true
				realMain(c, "checkout")
			},
		},
		cli.Command{
			Name:  "format",
			Usage: "Normalize a Puppetfile",
			Args:  "<file>",
			Description: `Format a PUppetfile by removing comments/blank lines, good whitespacing,
   and sorting with mod name.

   e.g) norm Puppetfile`,
			Flags: []cli.Flag{
				modulepathOpt,
				includesOpt,
			},
			Action: func(c *cli.Context) {
				a, _ := c.ArgFor("file")
				Format(c, a)
			},
		},
		cli.Command{
			Name:  "diff",
			Usage: "Compare two files with local branches",
			Args:  "<file-a> <file-b>",
			Description: `Compare two files with local branches which are chekced out.
   You need to check out local branches to be used beforehand.
   To do that, 'install' command can be used.

   e.g) diff Puppetfile.staging Puppetfile.development`,
			Flags: []cli.Flag{
				modulepathOpt,
				includesOpt,
			},
			Action: func(c *cli.Context) {
				a, _ := c.ArgFor("file-a")
				b, _ := c.ArgFor("file-b")
				Diff(c, a, b)
			},
		},
		cli.Command{
			Name:        "git-push",
			Usage:       "Print git commands to release",
			Args:        "<file-a> <file-b>",
			Description: `e.g) release Puppetfile.staging Puppetfile.development`,
			Flags: []cli.Flag{
				modulepathOpt,
				includesOpt,
				relBranchOpt,
				remoteNameOpt,
				cli.BoolFlag{Name: "use-sha1", Usage: "Use SHA1 instead of branch name"},
			},
			Action: func(c *cli.Context) {
				a, _ := c.ArgFor("file-a")
				b, _ := c.ArgFor("file-b")
				printGitPush(c, a, b)
			},
		},
		cli.Command{
			Name:        "bump-up",
			Usage:       "Print bumped-up Puppetfile based on file-a",
			Args:        "<file-a> <file-b>",
			Description: `e.g) bump-up Puppetfile.staging Puppetfile.development`,
			Flags: []cli.Flag{
				modulepathOpt,
				includesOpt,
				relBranchOpt,
			},
			Action: func(c *cli.Context) {
				a, _ := c.ArgFor("file-a")
				b, _ := c.ArgFor("file-b")
				bumpUp(c, a, b)
			},
		},
	}
	app.Run(os.Args)

}

var (
	modulepathOpt = cli.StringFlag{Name: "modulepath", Value: "modules", Usage: "Path to be for modules"}
	includesOpt   = cli.StringFlag{Name: "includes", Value: ".*", Usage: "Regexp pattern to include"}
	relBranchOpt  = cli.StringFlag{Name: "initial-release-branch", Value: "release/0.1", Usage: "Initial release branch"}
	remoteNameOpt = cli.StringFlag{Name: "remote-name", Value: "origin", Usage: "Remote name"}
)

var flags = []cli.Flag{
	modulepathOpt,
	cli.IntFlag{Name: "throttle", Value: 0, Usage: `Throttle number of concurrent processes.
                                Max is number of mod, min is 1. Max is used if 0 or negative number is given.`},
	cli.BoolFlag{Name: "force,f", Usage: "checkout with --force"},
}

var logger = log.New(ioutil.Discard, "", log.LstdFlags)
var modulepath string
var onlyCheckout bool
var forceCheckout bool
