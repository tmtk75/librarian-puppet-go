./librarian-puppet-go: *.go cmd/librarian-puppet-go/main.go
	go build -o librarian-puppet-go cmd/librarian-puppet-go/main.go

dep:
	dep ensure

VERSION := $(shell git describe --tags)
LDFLAGS := -ldflags "-X github.com/tmtk75/librarian-puppet-go.Version=$(VERSION)"
install:
	go install $(LDFLAGS) ./cmd/librarian-puppet-go

XC_ARCH=amd64
XC_OS=linux darwin
version=`./librarian-puppet-go --version 2>&1`

build:
	for arch in $(XC_ARCH); do \
	  for os in $(XC_OS); do \
	    echo $$arch $$os; \
	    GOARCH=$$arch GOOS=$$os go build \
	      $(LDFLAGS) \
	      -o pkg/librarian-puppet-go_$${os}_$$arch \
	      ./cmd/librarian-puppet-go/main.go; \
	  done; \
	done

release: compress ./librarian-puppet-go
	rm -f pkg/*_amd64 pkg/*.exe
	ghr -u tmtk75 v$(version) pkg

compress: pkg/librarian-puppet-go_darwin_amd64.tar.gz \
	  pkg/librarian-puppet-go_linux_amd64.tar.gz

pkg/librarian-puppet-go_darwin_amd64.tar.gz: build
	tar -C pkg -cz -f pkg/librarian-puppet-go_darwin_amd64.tar.gz librarian-puppet-go_darwin_amd64
pkg/librarian-puppet-go_linux_amd64.tar.gz: build
	tar -C pkg -cz -f pkg/librarian-puppet-go_linux_amd64.tar.gz librarian-puppet-go_linux_amd64

clean:
	rm -rf librarian-puppet-go *.out *.test

distclean: clean
	rm -rf pkg

cover: c.out
	go tool cover -html=c.out

c.out: *.go
	go test -coverprofile=c.out

bench:
	go test -bench=. -benchmem

cpu.out:
	go test -cpuprofile=cpu.out

block.out:
	go test -blockprofile=block.out

mem.out:
	go test -memprofile=mem.out

cpuprofile: cpu.out
	go tool pprof -text -nodecount=10 ./librarian-puppet-go.test cpu.out

blockprofile: block.out
	go tool pprof -text -nodecount=10 ./librarian-puppet-go.test block.out

memprofile: mem.out
	go tool pprof -text -nodecount=10 ./librarian-puppet-go.test mem.out

