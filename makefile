package = github.com/cinqict/havoc

.PHONY: install release

install:
	go get -t -v ./...

release:
	mkdir -p release
	rm -rf release/*
	GOOS=linux GOARCH=amd64 go build -o release/havoc-linux $(package)
	GOOS=darwin GOARCH=amd64 go build -o release/havoc $(package)