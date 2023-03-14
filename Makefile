.PHONY: check doc examples test test-race

DOC_PORT ?= 5050
COVERAGE_OUT ?= coverage.out
TEST_FLAGS ?= -v -count 1
PACKAGES = $(go list ./... | grep -v examples)

test:
	env MEMBER_COUNT=$(MEMBER_COUNT) go test $(TEST_FLAGS) $(PACKAGES) ./...

test-race:
	env MEMBER_COUNT=$(MEMBER_COUNT) go test $(TEST_FLAGS) -race $(PACKAGES)

test-cover:
	env TEST_FLAGS="$(TEST_FLAGS)" bash ./coverage.sh

view-cover:
	go tool cover -func $(COVERAGE_OUT) | grep total:
	go tool cover -html $(COVERAGE_OUT) -o coverage.html

doc:
	godoc -http=localhost:$(DOC_PORT)

check:
	bash ./check.sh

