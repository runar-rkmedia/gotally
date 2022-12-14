repo=github.com/runar-rkmedia/gotallly
version := $(shell git describe --tags)
gitHash := $(shell git rev-parse --short HEAD)
buildDate := $(shell TZ=UTC date +"%Y-%m-%dT%H:%M:%SZ")
ldflags=-X 'main.version=$(version)' -X 'main.date=$(buildDate)' -X 'main.commit=$(gitHash)' -X 'main.IsDevStr=0'

ifndef VERBOSE
MAKEFLAGS += --no-print-directory
endif

dev:
	$(MAKE) -s -j4 web server-watch test-watch buf-watch

# Dependencies
deps:
	@cd frontend && npm install
generate:
	buf generate
buf-watch:
	fd '' ./proto | entr -r sh -c "make buf-lint && make generate"


# linters
lint: buf-lint go-lint
buf-lint:
	@cd proto && make lint
	# @echo "Buflinter returned ok"
go-lint:
	golangci-lint run
	# @echo "Golinter returned ok"

# tests
go-bench:
	gotest -test.run=none -bench=. -benchmem ./...
go-test:
	gotest -race ./...
test-watch:
	fd '.go' | entr -r sh -c "gotest -race ./..."

# web and servers
web:
	cd frontend && npm run dev --host -- --clearScreen false 
server:
	go run ./api/cmd/main.go
server-watch:
	fd '.go' | entr -r sh -c "golangci-lint run & go run ./api/cmd/main.go"

build-web:
	@echo "VITE_API: '$$VITE_API' $VITE_API"
	cd frontend && VITE_API="/" npm run build
	rm -rf static/static
	mkdir -p static/static
	cp -r frontend/.svelte-kit/output/client/* static/static
	# Not really sure why, but currently I have to copy the index file... I am sure I am doing something wrong....
	cp frontend/.svelte-kit/output/prerendered/pages/index.html static/static

# build a container with the application
build-container: 
	docker build . \
	-t runardocker/gotally:latest \
	-t runardocker/gotally:$(version) \
	--target scratch \
	--label "org.opencontainers.image.title=gotally" \
	--label "org.opencontainers.image.revision=$(gitHash)" \
	--label "org.opencontainers.image.created=$(buildDate)" \
	--label "org.opencontainers.image.version=$(version)" \
	--build-arg "ldflags=$(ldflags)"

# run the latest container
run-container:
	docker run --rm -it -p 8080:8080 runardocker/gotally:latest

# publish the container.
container-publish: 
	@echo "will now publish the contianer. Did you remember to log into docker-hub?"
	docker push runardocker/gotally:latest 
	docker push runardocker/gotally:$(version) 

# Deploy to fly.io
fly: build-web
	fly deploy 
