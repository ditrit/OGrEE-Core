# '$$' refers to shell variable not make variable
# https://ftp.gnu.org/old-gnu/Manuals/make-3.79.1/html_chapter/make_6.html
GOPATH=$(shell go env GOPATH)

BUILD_COMMAND=go build

LINUX_COMMAND=GOOS=linux $(BUILD_COMMAND) -o ogree_app_backend
MAC_COMMAND=GOOS=darwin $(BUILD_COMMAND) -o ogree_app_backend.mac
WIN_COMMAND=GOOS=windows $(BUILD_COMMAND) -o ogree_app_backend.exe
CROSS_COMMAND=$(LINUX_COMMAND) ; $(MAC_COMMAND) ; $(WIN_COMMAND)
DOCKER_PREFIX=docker run --rm -v $(shell pwd):/workdir -w /workdir golang:1.20 bash -c

.PHONY: main mac win linux allos

main:
	eval $(BUILD_COMMAND)
mac: 
	eval $(MAC_COMMAND)
win: 
	eval $(WIN_COMMAND)
linux: 
	eval $(LINUX_COMMAND)
allos: 
	eval $(MAC_COMMAND)
	eval $(WIN_COMMAND)
	eval $(LINUX_COMMAND)

docker:
	$(DOCKER_PREFIX) "$(LINUX_COMMAND)"
dockermac:
	$(DOCKER_PREFIX) "$(MAC_COMMAND)"
dockerwin:
	$(DOCKER_PREFIX) "$(WIN_COMMAND)"
dockercross:
	$(DOCKER_PREFIX) "$(CROSS_COMMAND)"

clean:
	rm ogree_app_backend*

