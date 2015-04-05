# README
librarian-puppet-go is a simple command to clone modules and to checkout based on Puppetfile.

```
go run main.go install --modulepath /tmp/modules < Puppetfile
```

# Feature
This command ensures that git repositories are checked out with tag, ref and version.

For instance, let's say you have a following Puppetfile.
```
mod 'bar', :git => 'git@github.com:tmtk75/tmtk75-bar.git', :ref => 'master'
mod 'foo', :git => 'git@bitbucket.org:tmtk75/tmtk75-foo.git', :ref => 'v0.1.2'
mod 'puppetlabs/stdlib', '4.1.0'
```

Then `modules` directory has these repositories after running the command.
```
modules
  |-- bar       # chekced out at the latest of master origin/branch
  |-- foo       # checked out at the specified tag, v0.1.2
  `-- stdlib    # checked out at the commit versioned as 4.1.0 in puppetlabs.com
```

* Check out the present latest commit of `origin`/${branchname} if `:ref` is a branch.
* Check out the tag if `:ref` is a tag.
* Check out the commit registered with a version in puppetlabs.com
  retrieving its URL of source repository using REST API. `puppet module` command is NOT used.

## Extensions
## include
`include` directive allows you to include several Puppetfiles like this.
```
include "Puppetfile.common"
include "Puppetfile.debug"
mod 'puppetlabs/stdlib'
```

- It can be only in the head of file. You cannot put `include` after `mod` directive.
- The latest mod is used if same module name appears. For example, next case is `1.0.0` is enabled.
```
mod 'puppetlabs/stdlib', '4.1.0'
mod 'puppetlabs/stdlib', '1.0.0'
```

# Performance
* It takes about 30 seconds in order to clone about 80 modules
  although basically cloning modules strongly depends on the network speed :grin:
* It takes about 7 seconds in order to fetch & checkout about 80 modules.

## Assumption
```
$ cat Puppetfile.1 | grep '^mod' | wc -l
      84
$ cat Puppetfile.2 | grep '^mod' | wc -l
      81
```
`Puppetfile.1` has 84 mod declarations.
For example,
```
mod 'puppetlabs/stdlib', '4.1.0'
mod 'nodejs', :git => 'git://github.com/danheberden/puppet-nodejs.git'
```

`Puppetfile.2` also has 81 mod declarations which each mod has a specific tag.
For instance,
```
mod 'foo', :git => 'git@github.com:tmtk75/tmtk75-foo.git', :ref => 'v0.1.2'
mod 'bar', :git => 'git@github.com:tmtk75/tmtk75-bar.git', :ref => 'v0.2.0'
```

## Time to cloning all & checkout all
At first there is no `modules`, which will be created,
so all modules are cloned.
```
$ time ./librarian-puppet-go install --modulepath modules < Puppetfile.1

real    0m32.639s
user    0m14.881s
sys     0m12.581s
```
Afterward fetch & checkout complete within 5 seconds.
```
$ time ./librarian-puppet-go install --modulepath modules < Puppetfile.2

real    0m7.282s
user    0m4.266s
sys     0m2.788s
```

# Limitation
Acceptable Puppetfile is NOT compatible to original Puppetfile written in DSL of Ruby.  
A mod definition MUST be one line which contains `mod`, version string, `:git` or `:ref`.
Please see [mod_test.go](./mod_test.go) for expected definitions.
