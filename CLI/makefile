# '$$' refers to shell variable not make variable
# https://ftp.gnu.org/old-gnu/Manuals/make-3.79.1/html_chapter/make_6.html
GOPATH=$(shell go env GOPATH)

#Binary Stamping Vars
DATE=$(shell date +%Y.%m.%d//%T)
GITHASH=$(shell git rev-parse HEAD)
GITBRANCH=$(shell git branch --show-current)
GITHASHDATE=$(shell git show -s --format=%ci HEAD | sed 's/ /\//g')

BUILD_COMMAND=go build -ldflags=\"-X  cli/controllers.BuildHash=$(GITHASH) \
	-X cli/controllers.BuildTree=$(GITBRANCH) \
	-X cli/controllers.BuildTime=$(DATE) \
	-X cli/controllers.GitCommitDate=$(GITHASHDATE)\"

LINUX_COMMAND=$(BUILD_COMMAND) -o cli
MAC_COMMAND=GOOS=darwin $(BUILD_COMMAND) -o cli.mac
WIN_COMMAND=GOOS=windows $(BUILD_COMMAND) -o cli.exe
CROSS_COMMAND=$(LINUX_COMMAND) ; $(MAC_COMMAND) ; $(WIN_COMMAND)

.PHONY: main mac win

main:
	$(LINUX_COMMAND)
mac: 
	$(MAC_COMMAND)
win: 
	$(WIN_COMMAND)

docker:
	docker run --rm -v $(shell pwd):/workdir -w /workdir golang:1.19.6 bash -c "$(CROSS_COMMAND)"

clean:
	rm cli
