BINARY=pi
SOURCE=cmd/pi/pi.go

OWNER=hyperhq
REPO=pi

# These are the values we want to pass for VERSION,COMMIT and BUILD
# git tag v1.0
# git commit -am "One more change after the tags"
VERSION=`git tag --points-at HEAD | head -n 1`
COMMIT=`git rev-parse --short HEAD`
BUILD=`date +%FT%T%z`

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS="-w -s -X github.com/hyperhq/pi.Version=${VERSION} -X github.com/hyperhq/pi.Build=${BUILD} -X github.com/hyperhq/pi.Commit=${COMMIT}"

# Builds the project
build:
	./tools/build.sh "${LDFLAGS}"


OSTYPE :=
OSARCH :=
ifeq ($(OS),Windows_NT)
	OSTYPE = windows
	ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
		OSARCH = amd64
	endif
	ifeq ($(PROCESSOR_ARCHITECTURE),x86)
		OSARCH = 386
	endif
else
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		OSTYPE = linux
	endif
	ifeq ($(UNAME_S),Darwin)
		OSTYPE = darwin
	endif

	UNAME_M := $(shell uname -m)
	ifeq ($(UNAME_M),x86_64)
		OSARCH = amd64
	endif
	ifneq ($(filter %86,$(UNAME_M)),)
		OSARCH = 386
	endif
	ifneq ($(filter arm%,$(UNAME_M)),)
		OSARCH = arm
	endif
endif

package:
	./tools/package.sh ${OSTYPE} ${OSARCH}

release:
	./tools/release.sh owner=${OWNER} repo=${REPO} tag=${VERSION} os=${OSTYPE} arch=${OSARCH}

# Cleans our project: deletes binaries
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
	rm -rf ${BINARY}*.{tar.gz,zip} ;



.PHONY: clean install
