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
DOCKER_PREFIX=docker run --rm -v $(shell pwd):/workdir -w /workdir golang:1.19.6 bash -c

.PHONY: main mac win

main: protobuf
	eval $(LINUX_COMMAND)
mac: protobuf
	eval $(MAC_COMMAND)
win: protobuf
	eval $(WIN_COMMAND)

protobuf: 
	poetry run python -W ignore:pkg_resources -m grpc_tools.protoc --proto_path=. ./com.proto --python_out=. --grpc_python_out=.
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./com.proto

docker:
	$(DOCKER_PREFIX) "$(LINUX_COMMAND)"
dockermac:
	$(DOCKER_PREFIX) "$(MAC_COMMAND)"
dockerwin:
	$(DOCKER_PREFIX) "$(WIN_COMMAND)"
dockercross:
	$(DOCKER_PREFIX) "$(CROSS_COMMAND)"

clean:
	rm cli
