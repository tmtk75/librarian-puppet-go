# README
librarian-puppet-go is a simple tool to clone modules and to checkout based on Puppetfile.

```
go run main.go install --modulepath /tmp/modules < Puppetfile
```

# Performance
* It takes about 30 seconds in order to clone about 80 modules
  although basically cloning modules strongly depends on the network speed :grin:
* It takes about 5 seconds in order to checkout about 80 moudles.

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

real    0m27.031s
user    0m7.944s
sys     0m7.078s
```
Afterward checkout completes within 5 seconds.
```
$ time ./librarian-puppet-go install --modulepath modules < Puppetfile.2

real    0m4.282s
user    0m3.266s
sys     0m1.788s
```