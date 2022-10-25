base_dir = $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))

.PHONEY: test
test: lint unit-test

.PHONEY: lint
lint:
	"$(base_dir)/scripts/check_gofmt.sh" "$(base_dir)"
	go install honnef.co/go/tools/cmd/staticcheck@latest
	staticcheck -f stylish "$(base_dir)/..."
	go vet "$(base_dir)/..."
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec -exclude-generated "$(base_dir)/..."

.PHONEY: unit-test
unit-test:
	go test -timeout=10s -coverprofile="$(base_dir)/coverage.out" "$(base_dir)/..."

.PHONEY: generate
generate:
	go install github.com/golang/mock/mockgen@v1.6
	go generate "$(base_dir)/pkg/..."
