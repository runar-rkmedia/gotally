repo=github.com/runar-rkmedia/gotallly
version = $(shell git describe --tags)
gitHash = $(shell git rev-parse --short HEAD)
buildDate = $(shell TZ=UTC date +"%Y-%m-%dT%H:%M:%SZ")
ldflags=-X 'main.Version=$(version)' -X 'main.Date=$(buildDate)' -X 'main.Commit=$(gitHash)' -X 'main.IsDevStr=0'

hasBufCli = $(shell command -v bxuf 2> /dev/null)
gotester = $(shell command -v gotest 2>/dev/null || printf "go test")

ifndef VERBOSE
MAKEFLAGS += --no-print-directory
endif
ifeq (, $(shell which buf 2> /dev/null))
$(error "No buf in PATH, consider adding the buf-cli from https://docs.buf.build/installation")
endif

# Starts The server, and the webserver.
# Installs any dependecies first
start:
	$(MAKE) -s deps
	$(MAKE) -s -j2 web server

# Watchmode for server, web, tests and buf
watch:
ifeq (, $(shell which fd 2> /dev/null))
$(error "No fd in PATH, which is required for watch-mode consider adding fd from https://github.com/sharkdp/fd ")
endif
ifeq (, $(shell which entr 2> /dev/null))
$(error "No entr in PATH, which is required for watch-mode consider adding the entr from https://github.com/eradman/entr ")
endif
	$(MAKE) -s deps
	$(MAKE) -s -j4 web server-watch test-watch buf-watch

# Install dependencies
deps:
	$(MAKE) -s -j3 frontend/node_modules deps_go generate
frontend/node_modules: frontend/package.json
	@cd frontend && npm install
deps_go:
	go mod tidy

# Runs buf generate
generate:
	buf generate
	$(MAKE) sqlc
# Runs sqc generation
sqlc:
	sqlc generate
	echo -e "-- This file is generated\n-- Please do not edit.\n-- The file to edit should be ../schema-sqlite.sql" > ./sqlite/schema.sql
	cat ./schema-sqlite.sql >> ./sqlite/schema.sql
	# cp ./schema-sqlite.sql ./sqlite/schema.sql
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
cover: cover-go
cover-html: cover-go-html
cover-go:
	go test ./...  -cover -json | tparse -all
cover-go-html:
	go test ./... -coverprofile out.cover
	go tool cover -html=out.cover
go-test:
	@ echo Using $(gotester) as tester
	$(gotester) -race ./... -count 1
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
	fd '.go' | entr -cr richgo test ./... 

# web and servers
web:
	cd frontend && npm run dev --host -- --clearScreen false 
web_public_api:
	@echo "Using the public api-server. For local testing, it is adviced to use the local server instead."
	cd frontend && VITE_DEV_API="https://gotally.fly.dev" npm run dev --host -- --clearScreen false 
server:
	go run ./api/cmd/ --development
server-watch:
	fd '.go' | entr -cr go run ./api/cmd/ --development

build-web:
	@echo "VITE_API: '$$VITE_API' $VITE_API"
	cd frontend && VITE_API="/" npm run build
	# Copy the static frontend files to the expected folder for the api
	# The api will then bundle the files within its binary
	rm -rf static/static
	mkdir -p static/static
	cp  -r ./frontend/build/* static/static

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
fly-get-db:
	fly sftp get /app/data/db.sqlite ./data/bk-fly-$$(date +"%F-%H%M").sqlite
