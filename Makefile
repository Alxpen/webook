.PHONY: docker
docker:
	@rm webook || true
	@GOOS=linux GOARCH=arm go build -tags=k8s -o webook .
	@docker rmi -f alxpen/webook:v0.0.1
	@docker pull ubuntu:20.04          
	@docker build -t alxpen/webook:v0.0.1 .  