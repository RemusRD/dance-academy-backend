.PHONY: build clean deploy

build:
	go get -u "github.com/aws/aws-lambda-go/lambda"
	go get -u "github.com/aws/aws-sdk-go"
	env GOOS=linux
	go build -ldflags="-s -w" -o bin/batch-user-creation batch-user-creation/main.go
	go build -ldflags="-s -w" -o bin/attendance-clearance attendance-clearance/main.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
