NO_COLOR=\033[0m
OK_COLOR=\033[0;32m

ALL_GO_FILES = $(wildcard */*.go)
ALL_BIN_FILES = $(patsubst %.go,%,$(ALL_GO_FILES))

all: $(ALL_GO_FILES)

define PROGRAM_template
$(1): vet
	@echo "$(OK_COLOR)==> Building $(1) $(NO_COLOR)"
	@cd $(dir $1); go build
endef

$(foreach prog,$(ALL_GO_FILES),$(eval $(call PROGRAM_template,$(prog))))

clean:
	@$(foreach file,$(ALL_BIN_FILES),rm -f $(file);)

format:
	@echo "$(OK_COLOR)==> Formatting the code $(NO_COLOR)"
	@$(foreach file,$(ALL_GO_FILES),gofmt -s -w $(file);)
	@$(foreach file,$(ALL_GO_FILES),goimports -w $(file);)

vet:
	@echo "$(OK_COLOR)==> Running go vet $(NO_COLOR)"
	@$(foreach file,$(ALL_GO_FILES),go tool vet -all $(file);)

lint:
	@echo "$(OK_COLOR)==> Running golint $(NO_COLOR)"
	@$(foreach file,$(ALL_GO_FILES),golint $(file);)

.PHONY: all clean format vet lint
