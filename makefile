BASH=/bin/bash
GOPATH=$$(go env | grep "GOPATH=" | grep -oE '\/[^"]*')
GOYACC=$(GOPATH)/bin/goyacc
NEX=$(GOPATH)/bin/nex
DATE=$$(date +%Y.%m.%d//%T)
GITHASH=$$(git rev-parse HEAD)
GITBRANCH=$$(git branch --show-current)

main: interpreter main.go lexer.nn.go y.go
	go build \-ldflags="-X  cli/controllers.BuildHash=$(GITHASH) \
	-X cli/controllers.BuildTree=$(GITBRANCH) \
	-X cli/controllers.BuildTime=$(DATE)" main.go lexer.nn.go y.go

interpreter: parser lexer buildTimeScript

parser: interpreter/parser.y controllers/commandController.go
	$(GOYACC) "interpreter/parser.y" 

lexer: interpreter/lexer.nex
	$(NEX) "interpreter/lexer.nex"; mv interpreter/lexer.nn.go .

buildTimeScript:
	other/injectionscript.py

clean:
	rm main y.go lexer.nn.go y.output

