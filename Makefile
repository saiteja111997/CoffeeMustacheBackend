clean:
	go clean
	rm -rf bootstrap

build : clean
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bootstrap main.go

deploy_prod: build 
	serverless deploy --stage prod --aws-profile saiteja