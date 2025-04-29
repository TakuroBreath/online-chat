# Переменные сборки
BINARY_NAME=chatik
INSTALL_PATH=/usr/local/bin

# Go параметры по умолчанию
GOOS?=darwin
GOARCH?=amd64

.PHONY: all build build-linux build-windows build-mac clean install uninstall

all: build

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BINARY_NAME) ./cmd/main.go

# Кросс-компиляция для разных платформ
build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux ./cmd/main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME).exe ./cmd/main.go

build-mac:
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-mac ./cmd/main.go

# Сборка для всех платформ
build-all: build-linux build-windows build-mac

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-linux $(BINARY_NAME)-mac $(BINARY_NAME).exe

install: build
	sudo mv $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)

uninstall:
	sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	