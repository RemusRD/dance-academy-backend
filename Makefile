.PHONY: build clean deploy

build:
	go get -u "github.com/aws/aws-lambda-go/lambda"
	go get -u "github.com/aws/aws-sdk-go"
	env GOOS=linux
	GOARCH=amd64
	go build -v -ldflags="-s -w" -a -o bin/batch-user-creation batch-user-creation/main.go
	go build -v -ldflags="-s -w" -a -o bin/attendance-clearance attendance-clearance/main.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
