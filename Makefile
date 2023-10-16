.DEFAULT_GOAL := vet

fmt:
	go fmt *.go
.PHONY:fmt

lint: fmt
	golint *.go
.PHONY:lint

vet: lint
	go vet *.go
.PHONY:vet
