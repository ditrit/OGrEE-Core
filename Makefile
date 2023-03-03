# '$$' refers to shell variable not make variable
# https://ftp.gnu.org/old-gnu/Manuals/make-3.79.1/html_chapter/make_6.html
GOPATH=$(shell go env GOPATH)
DATE=$(shell date +%Y.%m.%d//%T)
GITHASH=$(shell git rev-parse HEAD)
GITBRANCH=$(shell git branch --show-current)
GITHASHDATE=$(shell git show -s --format=%ci HEAD | sed 's/ /\//g')

#.FORCE is purposely added as an empty dependecy so that 
#make will ALWAYS build our binary 
main: main.go .FORCE
	go build \-ldflags="-X  p3/utils.BuildHash=$(GITHASH) \
	-X p3/utils.BuildTree=$(GITBRANCH) \
	-X p3/utils.BuildTime=$(DATE) \
	-X p3/utils.GitCommitDate=$(GITHASHDATE)" \
	main.go 
	
.FORCE:

all: main run

run:
	./main

allos: linux windows mac

linux: main.go .FORCE
	go build -o OGrEE_API_Linux_x64 \
	-ldflags="-X  p3/utils.BuildHash=$(GITHASH) \
	-X p3/utils.BuildTree=$(GITBRANCH) \
	-X p3/utils.BuildTime=$(DATE) \
	-X p3/utils.GitCommitDate=$(GITHASHDATE)" \
	main.go 

windows: main.go .FORCE
	GOOS=windows GOARCH=amd64 go build -o OGrEE_API_Win_x64 \
	-ldflags="-X  p3/utils.BuildHash=$(GITHASH) \
	-X p3/utils.BuildTree=$(GITBRANCH) \
	-X p3/utils.BuildTime=$(DATE) \
	-X p3/utils.GitCommitDate=$(GITHASHDATE)" \
	main.go

mac: main.go .FORCE
	GOOS=darwin GOARCH=amd64 go build -o OGrEE_API_OSX_x64 \
	-ldflags="-X  p3/utils.BuildHash=$(GITHASH) \
	-X p3/utils.BuildTree=$(GITBRANCH) \
	-X p3/utils.BuildTime=$(DATE) \
	-X p3/utils.GitCommitDate=$(GITHASHDATE)" \
	main.go

clean:
	rm main OGrEE_API*

