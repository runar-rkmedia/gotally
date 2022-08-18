repo=github.com/runar-rkmedia/gotallly
version := $(shell git describe --tags)
gitHash := $(shell git rev-parse --short HEAD)
buildDate := $(shell TZ=UTC date +"%Y-%m-%dT%H:%M:%SZ")
ldflags=-X 'main.version=$(version)' -X 'main.date=$(buildDate)' -X 'main.commit=$(gitHash)' -X 'main.IsDevStr=0'

live-client:
	fd | entr -r go run live_client/main.go
test-watch:
	fd | entr -r gotest ./...
nterfaces:
	ifacemaker -f tallylogic/board.go -s TableBoard -i BoardController -p tallylogic -o tallylogic/board_controller.go
ontainer: 
	docker build . \
	-t runardocker/gotally:latest \
	-t runardocker/gotally:$(version) \
	--target scratch \
	--label "org.opencontainers.image.title=gotally" \
	--label "org.opencontainers.image.revision=$(gitHash)" \
	--label "org.opencontainers.image.created=$(buildDate)" \
	--label "org.opencontainers.image.version=$(version)" \
	--build-arg "ldflags=$(ldflags)"
run-container:
	docker run --rm -it -p 8080:8080 runardocker/gotally:latest
container-publish: 
	docker push runardocker/gotally:latest 
	docker push runardocker/gotally:$(version) 
fly:
	fly deploy 
