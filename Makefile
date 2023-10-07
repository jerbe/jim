.PHONY: build
.PHONY: before

# 设置变量
TARGETS := \
    darwin/amd64 \
    darwin/arm64 \
    freebsd/386 \
    freebsd/amd64 \
    freebsd/arm \
    freebsd/arm64 \
    linux/386 \
    linux/amd64 \
    linux/arm \
    linux/arm64 \
    linux/mips \
    linux/mips64 \
    windows/386 \
    windows/amd64 \
    windows/arm \
    windows/arm64
CC := go
FLAGS := -v
APP_NAME := jim
BIN := ./bin

build: all

# 默认目标，编译所有平台
all: $(TARGETS)

# 编译指定平台的目标
$(TARGETS):
	@echo "编译 $@"
	GOOS=$(word 1, $(subst /, ,$@)) GOARCH=$(word 2, $(subst /, ,$@)) $(CC) build -o $(BIN)/$(APP_NAME)-$(word 1, $(subst /, ,$@))-$(word 2, $(subst /, ,$@))$(if $(filter windows%,$@),.exe,)

# 清理生成的文件
clean:
	rm -f ./bin/$(APP_NAME)-*

# 帮助信息
help:
	@echo "Usage:"
	@echo "  make			编译项目到所有平台"
	@echo "  make <平台>		编译指定平台的项目"
	@echo "  make clean		清理生成的文件"
	@echo "  make help		显示帮助信息"