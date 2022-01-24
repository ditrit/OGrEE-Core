# '$$' refers to shell variable not make variable
# https://ftp.gnu.org/old-gnu/Manuals/make-3.79.1/html_chapter/make_6.html
BASH=/bin/bash
GOPATH=$$(go env | grep "GOPATH=" | grep -oE '\/[^"]*')
GOYACC=$(GOPATH)/bin/goyacc
NEX=$(GOPATH)/bin/nex
DATE=$$(date +%Y.%m.%d//%T)
GITHASH=$$(git rev-parse HEAD)
GITBRANCH=$$(git branch --show-current)
GITHASHDATE=$$(git show -s --format=%ci HEAD | sed 's/ /\//g')


main: interpreter main.go ast.go lexer.nn.go y.go repl.go
	go build \-ldflags="-X  cli/controllers.BuildHash=$(GITHASH) \
	-X cli/controllers.BuildTree=$(GITBRANCH) \
	-X cli/controllers.BuildTime=$(DATE) \
	-X cli/controllers.GitCommitDate=$(GITHASHDATE)" main.go ast.go lexer.nn.go y.go repl.go
	

interpreter: parser lexer buildTimeScript

parser: interpreter/parser.y controllers/commandController.go
	$(GOYACC) "interpreter/parser.y" 

lexer: interpreter/lexer.nex
	$(NEX) "interpreter/lexer.nex"; mv interpreter/lexer.nn.go .

buildTimeScript:
	$(info Injecting build time code...)
	other/injectionscript.py

clean:
	rm main y.go lexer.nn.go y.output

