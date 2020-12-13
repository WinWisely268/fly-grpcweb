GO_LDFLAGS = CGO_ENABLED=0 go build -ldflags "-X main.build=${VERSION_GITHASH}" -a -tags netgo

.PHONY: all

all:
	@go mod vendor
	$(GO_LDFLAGS) -o app .