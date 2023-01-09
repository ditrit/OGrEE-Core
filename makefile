# '$$' refers to shell variable not make variable
# https://ftp.gnu.org/old-gnu/Manuals/make-3.79.1/html_chapter/make_6.html
GOPATH=$(shell go env GOPATH)
GOYACC=$(GOPATH)/bin/goyacc
NEX=$(GOPATH)/bin/nex

#Binary Stamping Vars
DATE=$(shell date +%Y.%m.%d//%T)
GITHASH=$(shell git rev-parse HEAD)
GITBRANCH=$(shell git branch --show-current)
GITHASHDATE=$(shell git show -s --format=%ci HEAD | sed 's/ /\//g')

#File building dependencies
FILEDEPS = main.go ast.go semantic.go lexer.nn.go y.go repl.go ocli.go aststr.go \
 astnum.go astbool.go astflow.go astutil.go completer.go

main: interpreter $(FILEDEPS)
	go build \-ldflags="-X  cli/controllers.BuildHash=$(GITHASH) \
	-X cli/controllers.BuildTree=$(GITBRANCH) \
	-X cli/controllers.BuildTime=$(DATE) \
	-X cli/controllers.GitCommitDate=$(GITHASHDATE)" \
	$(FILEDEPS)
	

interpreter: parser lexer buildTimeScript

parser: interpreter/parser.y controllers/commandController.go
	$(GOYACC) "interpreter/parser.y" 

lexer: interpreter/lexer.nex
	$(NEX) -o "./lexer.nn.go" "interpreter/lexer.nex"

buildTimeScript:
	$(info Injecting build time code...)
	other/injectionscript.py

#OTHER PLATFORM COMPILATION BLOCK
mac: interpreter $(FILEDEPS)
	GOOS=darwin go build \-ldflags="-X  cli/controllers.BuildHash=$(GITHASH) \
	-X cli/controllers.BuildTree=$(GITBRANCH) \
	-X cli/controllers.BuildTime=$(DATE) \
	-X cli/controllers.GitCommitDate=$(GITHASHDATE)" \
	$(FILEDEPS)
	

win: interpreter $(FILEDEPS)
	GOOS=windows go build \-ldflags="-X  cli/controllers.BuildHash=$(GITHASH) \
	-X cli/controllers.BuildTree=$(GITBRANCH) \
	-X cli/controllers.BuildTime=$(DATE) \
	-X cli/controllers.GitCommitDate=$(GITHASHDATE)" \
	$(FILEDEPS)
	

clean:
	rm main y.go lexer.nn.go y.output parser.tab.c

