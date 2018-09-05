all: ready build

deps:
	@dep ensure -v
	@yarn

install: deps
	@go install github.com/universonic/panther/cmd/...

test: deps
	@go test -cpu 1,4 -timeout 5m github.com/universonic/panther/...
	@yarn run test

race: deps
	@go test -race -cpu 1,4 -timeout 7m github.com/universonic/panther/...

benchmark: deps
	@go test -bench . -cpu 1,4 -timeout 10m github.com/universonic/panther/...

build: deps
	go build github.com/universonic/panther/cmd/...
	yarn run build

devel: deps
	@yarn run start

clean:
	@go clean -i github.com/universonic/panther/...

ready:
	go version
	go get -u github.com/golang/dep/cmd/dep
	node -v
	@npm -g install @angular/cli
	@npm -g install yarn

.PHONY: \
	deps \
	test \
	build \
	devel \
	ready \
	dist \