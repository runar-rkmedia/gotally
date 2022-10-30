repo=github.com/runar-rkmedia/gotallly
version = $(shell git describe --tags)
gitHash = $(shell git rev-parse --short HEAD)
buildDate = $(shell TZ=UTC date +"%Y-%m-%dT%H:%M:%SZ")
ldflags=-X 'main.Version=$(version)' -X 'main.Date=$(buildDate)' -X 'main.Commit=$(gitHash)' -X 'main.IsDevStr=0'

hasGoTestDox = $(shell command -v gotestdox 2>/dev/null)
gotester = $(shell command -v gotest 2>/dev/null || printf "go test")

# gotester=gotestdox

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
sqlc:
	sqlc generate
	cp ./schema-sqlite.sql ./sqlite/schema.sql
model:
	@echo "Attempting to generate model with xo from local development-schema"
	@echo xo schema $$\{DSN\}
	@xo schema ${DSN}
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
	go test -test.run=none -bench=. -benchmem ./... > ./.bench/$(buildDate)-$(gitHash).bench
go-cover:
	go test ./...  -cover -json | tparse -all
go-test:
	@ echo Using $(gotester) as tester
	$(gotester) -race ./...
e2e-test:
	@echo "The development-api-server must be running prior to running e2e-test"
	@cd frontend && npm run test
web-unit-test:
	@echo "The development-api-server must be running prior to running e2e-test"
	@cd frontend && npm run test:unit -- --run
_test: go-test e2e-test web-unit-test
test:
	$(MAKE) -j3 go-test e2e-test web-unit-test
test-watch:
	fd '.go' | entr -r gotest ./... 

# web and servers
web:
	cd frontend && npm run dev --host -- --clearScreen false 
server:
	go run ./api/cmd/
server-watch:
	fd '.go' | entr -r sh -c "golangci-lint run & go run ./api/cmd/"

build-web:
	@echo "VITE_API: '$$VITE_API' $VITE_API"
	cd frontend && VITE_API="/" npm run build
	rm -rf static/static
	mkdir -p static/static
	cp -r frontend/.svelte-kit/output/client/* static/static
	# Not really sure why, but currently I have to copy the index file... I am sure I am doing something wrong....
	cp frontend/.svelte-kit/output/prerendered/pages/index.html static/static

build-api:
	CGO_ENABLED=0 GOOS=linux go build -v -ldflags="$(ldflags)" -o gotally ./api/cmd/
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
