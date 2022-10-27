base_dir = $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))

.PHONEY: test
test: lint unit-test

.PHONEY: lint
lint:
	"$(base_dir)/scripts/check_gofmt.sh" "$(base_dir)"
	go install honnef.co/go/tools/cmd/staticcheck@latest
	cd "$(base_dir)" && staticcheck -f stylish ./...
	cd "$(base_dir)" && go vet ./...
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	cd "$(base_dir)" && gosec -exclude-generated ./...

.PHONEY: unit-test
unit-test:
	cd "$(base_dir)" && go test -timeout=10s -coverpkg=./... -coverprofile=coverage.out ./...

.PHONEY: generate
generate:
	go install github.com/golang/mock/mockgen@v1.6
	cd "$(base_dir)" && go generate ./...
