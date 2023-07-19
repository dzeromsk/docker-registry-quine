IMAGE=dzeromsk/quine:latest

all: docker-build

docker-build:
	docker build -t $(IMAGE) .

docker-push:
	docker push $(IMAGE)

kind-load:
	kind load docker-image $(IMAGE)
