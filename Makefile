#!make
.PHONY: check-tools reportVersion depend clean build package all test
.DEFAULT: all
.EXPORT_ALL_VARIABLES:

-include .env
PACKAGE_NAME = simple-node-health

cat := $(if $(filter $(OS),Windows_NT),type,cat)
PACKAGE_VERSION := $(shell $(cat) VERSION)
VERSION := $(PACKAGE_VERSION)

TARGET_DIR = build
BUILD_DIR = support
DEBIAN_DIR = $(BUILD_DIR)/DEBIAN
CONTROL_TEMPLATE = $(DEBIAN_DIR)/control.tpl
CONTROL_FILE = $(DEBIAN_DIR)/control

GOLD_FLAGS=-X github.com/shadowbq/simple-node-health/cmd.Version=$(PACKAGE_VERSION)

BUILD_DOCS := README.md LICENSE example_config.yml

OS := $(shell uname)

all: check-tools reportVersion depend clean build package
test: check-tools reportVersion depend clean build 

check-tools:
	@command -v dpkg-deb >/dev/null 2>&1 || { echo >&2 "dpkg-deb is required but it's not installed. Aborting."; exit 1; }

reportVersion: 
	@echo "\033[32mProduct Version $(PACKAGE_VERSION)"

build:
	@echo
	@echo "\033[32mBuilding ----> \033[m"
	
	env GOOS=linux GOARCH=amd64 go build -ldflags "$(GOLD_FLAGS) ${SILVER_FALGS}" -o $(TARGET_DIR)/simple-node-health_linux_amd64 main.go
	env GOOS=darwin GOARCH=amd64 go build -ldflags "$(GOLD_FLAGS) ${SILVER_FALGS}" -o $(TARGET_DIR)/simple-node-health_darwin_amd64 main.go
	

clean:
	@echo
	@echo "\033[32mCleaning Build ----> \033[m"
	$(RM) -rf pkg/*
	$(RM) -rf build/*
	$(RM) -rf tmp/*
	$(RM) -rf support/usr/local/bin/*

depend:
	@echo
	@echo "\033[32mChecking Build Dependencies ----> \033[m"

package:
	@echo
	@echo "\033[32mPackaging ----> \033[m"
	mkdir -p $(BUILD_DIR)/usr/local/bin/
	cp $(TARGET_DIR)/simple-node-health_linux_amd64 $(BUILD_DIR)/usr/local/bin/simple-node-health
	# Replace {{VERSION}} in the control template with the actual version
	sed 's/{{VERSION}}/$(VERSION)/g' $(CONTROL_TEMPLATE) > $(CONTROL_FILE)
	chmod 0644 $(CONTROL_FILE)
	# Build the .deb package
	dpkg-deb --build $(BUILD_DIR)
	mv $(BUILD_DIR).deb $(TARGET_DIR)/$(PACKAGE_NAME)_$(VERSION)_amd64.deb

ifndef PACKAGE_VERSION
	@echo "\033[1;33mPACKAGE_VERSION is not set. In order to build a package I need PACKAGE_VERSION=n\033[m"
	exit 1;
endif

ifndef GOPATH
	@echo "\033[1;33mGOPATH is not set. This means that you do not have go setup properly on this machine\033[m"
	@echo "$$ mkdir ~/gocode";
	@echo "$$ echo 'export GOPATH=~/gocode' >> ~/.bash_profile";
	@echo "$$ echo 'export PATH=\"\$$GOPATH/bin:\$$PATH\"' >> ~/.bash_profile";
	@echo "$$ source ~/.bash_profile";
	exit 1;
endif

	@type go >/dev/null 2>&1|| { \
		echo "\033[1;33mGo is required to build this application\033[m"; \
		echo "\033[1;33mIf you are using homebrew on OSX, run\033[m"; \
		echo "Recommend: $$ brew install go --cross-compile-all"; \
		exit 1; \
	}


