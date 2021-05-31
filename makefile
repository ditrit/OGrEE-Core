BASH=/bin/bash
GOPATH=$$(go env | grep "GOPATH=" | grep -oE '\/[^"]*')
GOYACC=$(GOPATH)/bin/goyacc
NEX=$(GOPATH)/bin/nex

main: interpreter
	go build main.go lexer.nn.go y.go 

interpreter: parser lexer

parser: 
	$(GOYACC) "interpreter/parser.y" 

lexer:
	$(NEX) "interpreter/lexer.nex"; cp interpreter/lexer.nn.go .


clean:
	rm main y.go lexer.nn.go y.output

