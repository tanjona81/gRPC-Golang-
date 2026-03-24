# Variables - Easy to change in one place
IMAGE_NAME := grpc-server
VERSION    := $(shell git rev-parse --short HEAD)
CLUSTER    := dev-cluster
FULL_IMAGE = $(IMAGE_NAME):$(VERSION)

.PHONY: all deploy build load apply generate run

deploy: build load apply
	@echo "Deployment complete!"

build:
	@echo "Building Docker image: $(IMAGE_NAME):$(VERSION)..."
	docker build -t $(FULL_IMAGE) .

load:
	@echo "Loading image into Kind..."
	kind load docker-image $(FULL_IMAGE) --name $(CLUSTER)

apply:
	@echo "Applying Kubernetes manifests..."
	@echo "Updating manifest with tag $(VERSION)..."
	# This sed command finds the 'image:' line and replaces it with our new tag
	sed -i 's|image: $(IMAGE_NAME):.*|image: $(FULL_IMAGE)|' k8s/01-deployment.yaml
	kubectl apply -f k8s/01-deployment.yaml
	kubectl rollout restart deployment grpc-user-service -n go-grpc

applyall:
	@echo "Applying Kubernetes manifests..."
	@echo "Updating manifest with tag $(VERSION)..."
	# This sed command finds the 'image:' line and replaces it with our new tag
	sed -i 's|image: $(IMAGE_NAME):.*|image: $(FULL_IMAGE)|' k8s/01-deployment.yaml
	kubectl apply -f k8s/
	kubectl rollout restart deployment grpc-user-service -n go-grpc

generate:
	@echo "Generating Protobuf code..."
	buf generate

run:
	@echo "Starting local server..."
	go run cmd/server/main.go