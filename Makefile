build:
	export GOPATH=/home/ubuntu/go
	go get -d -v
	go build -o bin/server
