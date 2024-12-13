.PHONY: init
init:
	$(MAKE) download-deps

	go env -w GOPRIVATE="github.com/v-bible/*"

.PHONY: download-deps
download-deps:
	@# Ref: https://github.com/golang/go/issues/25922#issuecomment-1038394599
	@# Ref: https://marcofranssen.nl/manage-go-tools-via-go-modules
	cat ./internal/tools/tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

.PHONY: lint
lint:
	golangci-lint run
