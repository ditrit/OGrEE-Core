BASH=/bin/bash
GOPATH=$$(go env | grep "GOPATH=" | grep -oE '\/[^"]*')
GOYACC=$(GOPATH)/bin/goyacc
NEX=$(GOPATH)/bin/nex

main: interpreter main.go lexer.nn.go y.go
	go build main.go lexer.nn.go y.go 

interpreter: parser lexer

parser: interpreter/parser.y controllers/commandController.go
	$(GOYACC) "interpreter/parser.y" 

lexer: interpreter/lexer.nex
	$(NEX) "interpreter/lexer.nex"; mv interpreter/lexer.nn.go .


clean:
	rm main y.go lexer.nn.go y.output

