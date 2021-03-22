IMAGE ?= mkimuram/k8sviz
TAG ?= 0.3
DEVEL_IMAGE ?= k8sviz
DEVEL_TAG ?= devel

test: test-lint test-fmt test-vet test-unit
	@echo "[Running test]"

test-lint:
	@echo "[Running golint]"
	golint -set_exit_status cmd/... pkg/...

test-fmt:
	@echo "[Running gofmt]"
	if [ "$$(gofmt -l cmd/ pkg/ | wc -l)" -ne 0 ]; then \
		gofmt -d cmd/ pkg/ ;\
		false; \
	fi

test-vet:
	@echo "[Running go vet]"
	go vet `go list ./...`


test-unit:
	@echo "[Running unit tests]"
	go test -cover `go list ./...`

build:
	@echo "[Build]"
	mkdir -p bin/
	GO111MODULE=on go build -o bin/k8sviz ./cmd/k8sviz

release: test build

image-build:
	@echo "[Building image $(DEVEL_IMAGE):$(DEVEL_TAG)]"
	docker build -t $(DEVEL_IMAGE):$(DEVEL_TAG) .

image-push: image-build
	@echo "[Pushing image $(IMAGE):$(TAG)]"
	docker tag $(DEVEL_IMAGE):$(DEVEL_TAG) $(IMAGE):$(TAG)
	docker push $(IMAGE):$(TAG)

.PHONY: test test-lint test-fmt test-vet test-unit build release image-build image-push
