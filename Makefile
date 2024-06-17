BUILDER = ./bin/faucet-builder

.PHONY: znet-start
znet-start:
	$(BUILDER) znet start --profiles=1cored,faucet

.PHONY: znet-remove
znet-remove:
	$(BUILDER) znet remove

.PHONY: lint
lint:
	$(BUILDER) lint

.PHONY: test
test:
	$(BUILDER) test

.PHONY: build
build:
	$(BUILDER) build

.PHONY: images
images:
	$(BUILDER) images

.PHONY: release
release:
	$(BUILDER) release

.PHONY: release-images
release-images:
	$(BUILDER) release/images

.PHONY: integration-tests
integration-tests:
	$(BUILDER) integration-tests
