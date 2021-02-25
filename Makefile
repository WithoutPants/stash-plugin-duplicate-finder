ifeq ($(OS), Windows_NT)
  EXT := .exe
endif

OUTPUT := plugin_duplicate_finder

.PHONY: build pre-cross-compile cross-compile-win cross-compile-osx cross-compile-linux cross-compile-pi cross-compile-all cross-compile-docker

build:
	go build -o $(OUTPUT)$(EXT)

LDFLAGS := -ldflags "-extldflags '-static -s -w'"
LDFLAGS_WIN := -ldflags "-extldflags '-static -s -w'"

cross-compile-win: 
	GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ go build $(LDFLAGS_WIN) -mod=vendor -o "dist/win/$(OUTPUT).exe"

cross-compile-osx:
	GOOS=darwin GOARCH=amd64 CC=o64-clang CXX=o64-clang++ go build $(LDFLAGS) -mod=vendor -o "dist/osx/$(OUTPUT)"

cross-compile-linux:
	go build -tags "osusergo netgo" $(LDFLAGS) -mod=vendor -o "dist/linux/$(OUTPUT)"

cross-compile-pi:
	GOOS=linux GOARCH=arm GOARM=5 CC=arm-linux-gnueabi-gcc go build -tags "osusergo netgo"  $(LDFLAGS) -mod=vendor -o "dist/pi/$(OUTPUT)"

cross-compile-all: cross-compile-win cross-compile-osx cross-compile-linux cross-compile-pi

pre-docker:
	docker pull stashapp/compiler:develop

RELEASE_INC := duplicate-finder.yml duplicate-finder.cfg

build-release-win: cross-compile-win
	tar -cf dist/stash-plugin-duplicate-finder-win.tar -C dist/win $(OUTPUT).exe
	tar -rf dist/stash-plugin-duplicate-finder-win.tar $(RELEASE_INC)

build-release-osx: cross-compile-osx
	tar -cf dist/stash-plugin-duplicate-finder-osx.tar -C dist/osx $(OUTPUT)
	tar -rf dist/stash-plugin-duplicate-finder-osx.tar $(RELEASE_INC)

build-release-linux: cross-compile-linux
	tar -cf dist/stash-plugin-duplicate-finder-linux.tar -C dist/linux $(OUTPUT)
	tar -rf dist/stash-plugin-duplicate-finder-linux.tar $(RELEASE_INC)

build-release-pi: cross-compile-pi
	tar -cf dist/stash-plugin-duplicate-finder-pi.tar -C dist/pi $(OUTPUT)
	tar -rf dist/stash-plugin-duplicate-finder-pi.tar $(RELEASE_INC)

RUN_DOCKER := docker run --rm --mount type=bind,source="$(shell pwd)",target=/stash -w /stash stashapp/compiler:develop /bin/bash -c

cross-compile-docker:
	$(RUN_DOCKER) "make cross-compile-all"

build-release-docker:
	$(RUN_DOCKER) "make build-release-win build-release-osx build-release-linux build-release-pi"
