.PHONY: fmt
fmt:
	@echo go fmt
	go fmt $$(go list ./...)

.PHONY: test
test:
	@echo go test
	go test ./... -v

.PHONY: mockgen
mockgen:
	@echo ....... Generating mock .......
	~/go/bin/mockgen -source=interface.go -destination=interface_mock.go -package=helmclient

