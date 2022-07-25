precommit: fmt test
	git status

test:
	go test -v ./...

fmt:
	go mod tidy
	gofmt -w .
	goimports --local github.com/cheriot/netpoltool/ -w .

run:
	go run cmd/localserver/main.go

run-get:
	go run cmd/debug/*.go get pod -n back-end

run-desc:
	go run cmd/debug/*.go describe pod -n back-end product-a

int-cluster-create:
	kind create cluster --name test-cluster --wait 100s

int-cluster-delete:
	kind delete cluster --name test-cluster
