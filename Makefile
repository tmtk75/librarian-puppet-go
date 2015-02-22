XC_ARCH=amd64
XC_OS=linux darwin
version=`./librarian-puppet-go -v | sed 's/librarian-puppet-go version //g'`

build:
	gox \
	  -os="$(XC_OS)" \
	  -arch="$(XC_ARCH)" \
	  -output "pkg/{{.Dir}}_{{.OS}}_{{.Arch}}"

./librarian-puppet-go:
	go build

release: compress ./librarian-puppet-go
	rm -f pkg/*_amd64
	ghr -u tmtk75 v$(version) pkg

compress: pkg/librarian-puppet-go_darwin_amd64.tar.gz \
	  pkg/librarian-puppet-go_linux_amd64.tar.gz
pkg/librarian-puppet-go_darwin_amd64.tar.gz: build
	tar -C pkg -cz -f pkg/librarian-puppet-go_darwin_amd64.tar.gz librarian-puppet-go_darwin_amd64
pkg/librarian-puppet-go_linux_amd64.tar.gz: build
	tar -C pkg -cz -f pkg/librarian-puppet-go_linux_amd64.tar.gz librarian-puppet-go_linux_amd64

run:
	go run main.go
