# Переменные сборки
BINARY_NAME=chatik
INSTALL_PATH=/usr/local/bin

# Go параметры
GOOS=darwin
GOARCH=amd64

.PHONY: all build clean install uninstall

all: build

build:
	go build -o $(BINARY_NAME) ./cmd/main.go

clean:
	rm -f $(BINARY_NAME)

install: build
	sudo mv $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)

uninstall:
	sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)