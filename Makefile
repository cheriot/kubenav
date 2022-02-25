precommit: fmt test
	git status

test:
	go test -v ./...

fmt:
	go mod tidy
	gofmt -w .
	goimports --local github.com/cheriot/netpoltool/ -w .

convey:
	$$(go env GOPATH)/bin/goconvey

run:
	go run cmd/localserver/main.go
