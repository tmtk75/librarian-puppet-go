./librarian-puppet-go: *.go cmd/librarian-puppet-go/main.go
	go build -o librarian-puppet-go cmd/librarian-puppet-go/main.go

install:
	(cd cmd/librarian-puppet-go; go install)

XC_ARCH=amd64
XC_OS=linux darwin
version=`./librarian-puppet-go --version`

build:
	for arch in $(XC_ARCH); do \
	  for os in $(XC_OS); do \
	    echo $$arch $$os; \
	    GOARCH=$$arch GOOS=$$os go build -o pkg/librarian-puppet-go_$${os}_$$arch \
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
	rm -rf librarian-puppet-go

distclean: clean
	rm -rf pkg
