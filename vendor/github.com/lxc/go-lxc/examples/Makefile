NO_COLOR=\033[0m
OK_COLOR=\033[0;32m

ALL_GO_FILES = $(wildcard *.go)
ALL_FILES = $(patsubst %.go,%,$(ALL_GO_FILES))

all: $(ALL_FILES)

define PROGRAM_template
$(1): format vet lint
	@echo "$(OK_COLOR)==> Building $(1) $(NO_COLOR)"
	@go build $(1).go
endef

$(foreach prog,$(ALL_FILES),$(eval $(call PROGRAM_template,$(prog))))

clean:
	@$(foreach file,$(ALL_FILES),rm -f $(file);)

format:
	@echo "$(OK_COLOR)==> Formatting the code $(NO_COLOR)"
	@gofmt -s -w *.go
	@goimports -w *.go

vet:
	@echo "$(OK_COLOR)==> Running go vet $(NO_COLOR)"
	@`which go` vet .

lint:
	@echo "$(OK_COLOR)==> Running golint $(NO_COLOR)"
	@`which golint` .

.PHONY: all clean format vet lint
