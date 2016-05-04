export GO15VENDOREXPERIMENT=1

DEPS = $(go list -f '{{range .TestImports}}{{.}} {{end}}' ./... | grep -v /vendor/)
WEBSITE="github.com/mephux/envdb"
DESCRIPTION="Ask your environment questions."
NAME="envdb"

BUILDVERSION=$(shell cat VERSION)
GO_VERSION=$(shell go version)

# Get the git commit
SHA=$(shell git rev-parse --short HEAD)
BUILD_COUNT=$(shell git rev-list --count HEAD)

BUILD_TAG="${BUILD_COUNT}.${SHA}"

build: lint generate
	@echo "Building..."
	@mkdir -p bin/
	@go build \
    -ldflags "-X main.Build=${BUILD_TAG}" \
		${ARGS} \
    -o bin/envdb

server-linux: generate
	@echo "Building server..."
	@mkdir -p bin/
	@GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X main.Build=${BUILD_TAG}" \
    -o bin/csos-server

generate: clean
	@echo "Running go generate..."
	@go generate $$(go list ./... | grep -v /vendor/)

lint:
	@go vet  $$(go list ./... | grep -v /vendor/)
	# @for pkg in $$(go list ./... |grep -v /vendor/ |grep -v /kuber/) ; do \
		# golint -min_confidence=1 $$pkg ; \
		# done

package: setup strip rpm64

setup:
	@mkdir -p package/root/usr/bin
	@mkdir -p package/output/
	@cp -R ./bin/envdb package/root/usr/bin
	@./bin/envdb --version 2> VERSION

test:
	go list ./... | xargs -n1 go test

strip:
	strip bin/csos-server

rpm64:
	fpm -s dir -t rpm -n $(NAME) -v $(BUILDVERSION) -p package/output/csos-server-$(BUILDVERSION)-amd64.rpm \
		--rpm-compression bzip2 --rpm-os linux \
		--force \
		--url $(WEBSITE) \
		--description $(DESCRIPTION) \
		-m "Dustin Willis Webber <dustin.webber@gmail.com>" \
		--vendor "Dustin Willis Webber" -a amd64 \
		--exclude */**.gitkeep \
		package/root/=/

clean:
	rm -rf doc/
	rm -rf package/
	rm -rf bin/
	rm -rf envdb/bindata.go
	rm -rf VERSION

.PHONY: build
