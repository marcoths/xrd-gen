NAME = xgen
GO_FILES := *.go
GO_CMD = go
GO_BUILD = $(GO_CMD) build
GO_TEST = $(GO_CMD) test
CONTROLLER_GEN = controller-gen

all: build

build: $(NAME)

$(NAME): $(GO_FILES)
	$(GO_BUILD) -o ./bin/$(NAME) $(GO_FILES)

.PHONY: controller-gen
controller-gen:
	$(GO_CMD) install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.12.0