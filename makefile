BASH=/bin/bash
GOPATH=$$(go env | grep "GOPATH=" | grep -oE '\/[^"]*')
GOYACC=$(GOPATH)/bin/goyacc
NEX=$(GOPATH)/bin/nex
DATE=$$(date +%Y.%m.%d//%T)

main: interpreter main.go lexer.nn.go y.go
	go build -ldflags="-X cli/controllers.BuildTime=$(DATE)" main.go lexer.nn.go y.go

interpreter: parser lexer buildTimeScript

parser: interpreter/parser.y controllers/commandController.go
	$(GOYACC) "interpreter/parser.y" 

lexer: interpreter/lexer.nex
	$(NEX) "interpreter/lexer.nex"; mv interpreter/lexer.nn.go .

buildTimeScript:
	other/injectionscript.py

clean:
	rm main y.go lexer.nn.go y.output

