IMAGE ?= sguerard/k8sviz
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
	go vet `go list ./... | grep -v test/e2e`


test-unit:
	@echo "[Running unit tests]"
	go test -cover `go list ./... | grep -v test/e2e`

test-e2e:
	@echo "[Running e2e tests]"
	./test/e2e/e2e.sh

build:
	@echo "[Build]"
	mkdir -p bin/
	GO111MODULE=on go build -o bin/k8sviz ./cmd/k8sviz

release: test build test-e2e

image-build:
	@echo "[Building image $(DEVEL_IMAGE):$(DEVEL_TAG)]"
	docker build -t $(DEVEL_IMAGE):$(DEVEL_TAG) .

image-push: image-build
	@echo "[Pushing image $(IMAGE):$(TAG)]"
	docker tag $(DEVEL_IMAGE):$(DEVEL_TAG) $(IMAGE):$(TAG)
	docker push $(IMAGE):$(TAG)

.PHONY: test test-lint test-fmt test-vet test-unit test-e2e build release image-build image-push

generate-graph:
	mkdir -p out
	./k8sviz.sh  -t dot -k config-k8sviz -o out -n "ns1,ns2" -f k8s-diagram.png

merge-graph:
	m4 merge.m4 > merged.gv
	sed -i -e "s/\/icons/icons/" merged.gv
	dot -n -Tpng merged.gv -o diagram_staging.png
	rm merged.gv
