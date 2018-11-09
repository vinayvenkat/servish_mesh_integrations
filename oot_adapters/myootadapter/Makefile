all: clean build_adapter docker_build docker_push
.PHONY: all 

clean:
	rm -rf bin
build_adapter: 
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -v -o bin/mygrpcadapter ./mygrpcadapter/cmd/

docker_build:
	docker build -t vinayvenkat/mygrpcadapter:latest .

docker_push:
	docker push vinayvenkat/mygrpcadapter:latest
