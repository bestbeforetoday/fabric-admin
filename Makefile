base_dir = $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))
TMPDIR ?= /tmp

client_cert = $(TMPDIR)/certificate.pem
client_key = $(TMPDIR)/private-key.pem
ca_cert = $(TMPDIR)/ca.pem

.PHONEY: test
test: lint unit-test integration-test

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
	cd "$(base_dir)" && go test -timeout=10s -coverprofile=coverage.out ./...

.PHONEY: generate
generate:
	go install github.com/golang/mock/mockgen@v1.6
	cd "$(base_dir)" && go generate ./...

.PHONEY: fabric-up
fabric-up:
	"$(base_dir)/scripts/microfab.sh" down
	"$(base_dir)/scripts/microfab.sh" up "$(base_dir)/test/microfab/two-org.json"

.PHONEY: fabric-down
fabric-down:
	"$(base_dir)/scripts/microfab.sh" down

.PHONEY: integration-test
integration-test: fabric-up
	curl http://console.127.0.0.1.nip.io:8080/ak/api/v1/components | jq '.[] | select(.type == "identity") | select(.id == "org1admin") | .cert' | sed -e 's/"//g' | base64 --decode > "$(client_cert)"
	curl http://console.127.0.0.1.nip.io:8080/ak/api/v1/components | jq '.[] | select(.type == "identity") | select(.id == "org1admin") | .private_key' | sed -e 's/"//g' | base64 --decode > "$(client_key)"
	curl http://console.127.0.0.1.nip.io:8080/ak/api/v1/components | jq '.[] | select(.type == "identity") | select(.id == "org1admin") | .ca' | sed -e 's/"//g' | base64 --decode > "$(ca_cert)"
	CHAINCODE_PACKAGE="$(base_dir)/test/chaincode/basic.tar.gz" CLIENT_CERT="$(client_cert)" CLIENT_KEY="$(client_key)" CA_CERT="$(ca_cert)" ENDPOINT="org1peer-api.127-0-0-1.nip.io:8080" go run "$(base_dir)/test/cmd"
