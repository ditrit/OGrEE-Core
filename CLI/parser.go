package main

import (
	c "cli/controllers"
	"fmt"
	"strings"
)

type parseCommandFunc func() node

var commandDispatch map[string]parseCommandFunc
var createObjDispatch map[string]parseCommandFunc

var lsCommands = []string{"lssite", "lsbldg", "lsroom", "lsrack", "lsdev", "lsac",
	"lspanel", "lscabinet", "lscorridor", "lssensor"}
var noArgsCommands map[string]node

var manCommands = []string{
	"get", "getu", "getslot",
	"+", "-", "=", ">",
	".cmds", ".template", ".var",
	"ui", "camera",
	"link", "unlink",
	"lssite", "lsbldg", "lsroom", "lsrack", "lsdev", "lsac",
	"lspanel", "lscabinet", "lscorridor", "lssensor", "lsenterprise",
	"drawable", "draw", "undraw",
	"tree", "lsog", "env", "cd", "pwd", "clear", "grep", "ls", "exit", "len", "man", "hc",
	"print", "unset", "selection",
	"for", "while", "if",
}

func sliceContains(slice []string, s string) bool {
	for _, str := range slice {
		if str == s {
			return true
		}
	}
	return false
}

func indexOf(arr []string, val string) int {
	for pos, v := range arr {
		if v == val {
			return pos
		}
	}
	return -1
}

type traceItem struct {
	cursor  int
	message string
}

type parser struct {
	buf         string
	stackTrace  []traceItem
	startCursor int
	cursor      int
	err         string
	tok         token
}

func un(p *parser) {
	if p.err == "" {
		p.stackTrace = p.stackTrace[:len(p.stackTrace)-1]
		if len(p.stackTrace) > 0 {
			p.startCursor = p.stackTrace[len(p.stackTrace)-1].cursor
		} else {
			p.startCursor = 0
		}
	}
}

func trace(p *parser, name string) *parser {
	p.stackTrace = append(p.stackTrace, traceItem{p.cursor, name})
	p.startCursor = p.cursor
	return p
}

func (p *parser) reset() {
	p.cursor = p.startCursor
	p.tok = token{}
}

func (p *parser) unlex() {
	p.cursor = p.tok.start
	p.tok = token{}
}

func (p *parser) item(trim bool) string {
	var cur int
	if p.cursor > len(p.buf) {
		cur = len(p.buf)
	} else {
		cur = p.cursor
	}
	s := p.buf[p.startCursor:cur]
	if trim {
		return strings.Trim(s, " ")
	}
	return s
}

func (p *parser) remaining() string {
	if p.cursor >= len(p.buf) {
		return ""
	}
	return p.buf[p.cursor:]
}

func (p *parser) peek() byte {
	if p.cursor >= len(p.buf) {
		return eof
	}
	return p.buf[p.cursor]
}

func (p *parser) next() byte {
	var c byte
	if p.cursor >= len(p.buf) {
		c = eof
	} else {
		c = p.buf[p.cursor]
	}
	p.cursor++
	return c
}

func (p *parser) forward(n int) string {
	p.cursor += n
	return p.item(false)
}

func (p *parser) backward(n int) {
	if p.cursor-n < p.startCursor {
		panic("cannot go backward")
	}
	p.cursor -= n
}

func (p *parser) error(message string) {
	errorStr := ""
	for i := 0; i < len([]rune(c.State.BlankPrompt))+p.cursor; i++ {
		errorStr += " "
	}
	errorStr += "\033[31m" + "^" + "\033[0m" + "\n"
	if len(p.stackTrace) > 1 {
		errorStr += "parsing stack : "
		emptyStack := true
		for i := range p.stackTrace {
			if p.stackTrace[i].message != "" {
				if !emptyStack {
					errorStr += " -> "
				}
				errorStr += p.stackTrace[i].message
				emptyStack = false
			}
		}
		errorStr += "\n"
	}
	errorStr += "\033[31m" + "Error : " + "\033[0m" + message
	p.err = message
	panic(errorStr)
}

func (p *parser) skipWhiteSpaces() {
	defer un(trace(p, ""))
	for p.cursor < len(p.buf) && (p.peek() == ' ' || p.peek() == '\t') {
		p.forward(1)
	}
}

func (p *parser) commandEnd() bool {
	p.skipWhiteSpaces()
	return p.cursor == len(p.buf) || strings.Contains(";})", string(p.peek()))
}

func (p *parser) parseExact(word string) bool {
	defer un(trace(p, "\""+word+"\""))
	ok := p.startCursor+len(word) <= len(p.buf) && p.forward(len(word)) == word
	if !ok {
		p.reset()
	}
	return ok
}

func (p *parser) expect(word string) {
	if !p.parseExact(word) {
		p.error(word + " expected")
	}
}

func isPrefix(prefix string, candidates []string) bool {
	if prefix == "" {
		return true
	}
	for _, candidate := range candidates {
		if strings.HasPrefix(candidate, prefix) {
			return true
		}
	}
	return false
}

func (p *parser) parseKeyWord(candidates []string) string {
	defer un(trace(p, "keyword"))
	for p.cursor < len(p.buf) {
		p.forward(1)
		if !isPrefix(p.item(false), candidates) {
			p.backward(1)
			break
		}
	}
	if sliceContains(candidates, p.item(false)) {
		return p.item(false)
	}
	return ""
}

func (p *parser) parseSimpleWord(name string) string {
	defer un(trace(p, name))
	p.skipWhiteSpaces()
	for {
		c := p.next()
		if isAlphaNumeric(c) {
			continue
		}
		p.backward(1)
		p.skipWhiteSpaces()
		return p.item(true)
	}
}

func (p *parser) parseComplexWord(name string) string {
	p.skipWhiteSpaces()
	defer un(trace(p, name))
	for {
		c := p.next()
		if isAlphaNumeric(c) || c == '-' || c == '_' {
			continue
		}
		p.backward(1)
		p.skipWhiteSpaces()
		return p.item(true)
	}
}

func (p *parser) parseInt(name string) int {
	defer un(trace(p, name))
	p.skipWhiteSpaces()
	tok := p.parseExprToken()
	if tok.t != tokInt {
		p.error("integer expected")
	}
	p.skipWhiteSpaces()
	return tok.val.(int)
}

func (p *parser) parseFloat(name string) float64 {
	defer un(trace(p, name))
	p.skipWhiteSpaces()
	tok := p.parseExprToken()
	switch tok.t {
	case tokFloat:
		p.skipWhiteSpaces()
		return tok.val.(float64)
	case tokInt:
		p.skipWhiteSpaces()
		return float64(tok.val.(int))
	default:
		p.error("float expected")
		return 0.
	}
}

func (p *parser) parseBool() bool {
	defer un(trace(p, "bool"))
	p.skipWhiteSpaces()
	tok := p.parseExprToken()
	if tok.t != tokBool {
		p.error("boolean expected")
	}
	p.skipWhiteSpaces()
	return tok.val.(bool)
}

func (p *parser) parseText(lexFunc func() token, trim bool) node {
	defer un(trace(p, ""))
	s := ""
	subExpr := []node{}
loop:
	for {
		tok := lexFunc()
		switch tok.t {
		case tokText:
			s += tok.str
		case tokDeref:
			s += "%v"
			subExpr = append(subExpr, &symbolReferenceNode{tok.val.(string)})
		case tokLeftEval:
			s += "%v"
			subExpr = append(subExpr, p.parseExpr(""))
			p.expect("))")
		case tokEOF:
			break loop
		default:
			p.error("unexpected token")
		}
	}
	if trim {
		s = strings.Trim(s, " ")
	}
	if len(subExpr) == 0 {
		return &valueNode{s}
	}
	return &formatStringNode{s, subExpr}
}

func (p *parser) parsePath(name string) node {
	if name != "" {
		name = name + " path"
	} else {
		name = "path"
	}
	defer un(trace(p, name))
	p.skipWhiteSpaces()
	path := p.parseText(p.parsePathToken, true)
	p.skipWhiteSpaces()
	return &pathNode{path}
}

func (p *parser) parsePathGroup() []node {
	defer un(trace(p, "path group"))
	paths := []node{}
	p.skipWhiteSpaces()
	p.expect("{")
	for {
		p.skipWhiteSpaces()
		paths = append(paths, p.parsePath(""))
		p.skipWhiteSpaces()
		if p.parseExact("}") {
			break
		}
		if !p.parseExact(",") {
			p.error(", or } expected")
		}
	}
	p.skipWhiteSpaces()
	return paths
}

func (p *parser) parsePrimaryExpr() node {
	defer un(trace(p, ""))
	tok := p.parseExprToken()
	switch tok.t {
	case tokBool:
		return &valueNode{tok.val.(bool)}
	case tokInt:
		return &valueNode{tok.val.(int)}
	case tokFloat:
		return &valueNode{tok.val.(float64)}
	case tokDoubleQuote:
		n := p.parseText(p.parseQuotedStringToken, false)
		p.expect("\"")
		return n
	case tokDeref:
		p.parseExprToken()
		switch p.tok.t {
		case tokLeftBrac:
			index := p.parseExpr("index")
			p.parseExprToken()
			if p.tok.t != tokRightBrac {
				p.error("square bracket opened but not closed")
			}
			return &arrayReferenceNode{tok.val.(string), index}
		}
		p.unlex()
		return &symbolReferenceNode{tok.val.(string)}
	case tokLeftParen:
		expr := p.parseExpr("")
		endTok := p.parseExprToken()
		if endTok.t != tokRightParen {
			p.error(") expected, got " + endTok.str)
		}
		return expr
	case tokLeftBrac:
		exprList := []node{}
		p.parseExprToken()
		if p.tok.t == tokRightBrac {
			return &arrNode{exprList}
		}
		p.unlex()
		for {
			expr := p.parseExpr("array element")
			exprList = append(exprList, expr)
			p.parseExprToken()
			if p.tok.t == tokRightBrac {
				return &arrNode{exprList}
			}
			if p.tok.t == tokComma {
				continue
			}
			p.error("] or comma expected")
		}
	}
	p.error("unexpected token : " + tok.str)
	return nil
}

func (p *parser) parseUnaryExpr() node {
	defer un(trace(p, ""))
	tok := p.parseExprToken()
	switch tok.t {
	case tokAdd:
		return p.parseUnaryExpr()
	case tokSub:
		return &negateNode{p.parseUnaryExpr()}
	case tokNot:
		return &negateBoolNode{p.parseUnaryExpr()}
	}
	p.unlex()
	return p.parsePrimaryExpr()
}

func (p *parser) parseBinaryExpr(leftOperand node, precedence int) node {
	defer un(trace(p, ""))
	if leftOperand == nil {
		leftOperand = p.parseUnaryExpr()
	}
	for {
		operator := p.parseExprToken()
		operatorPrecedence := operator.precedence()
		if operatorPrecedence < precedence {
			p.unlex()
			return leftOperand
		}
		rightOperand := p.parseBinaryExpr(nil, operatorPrecedence+1)
		switch operator.t {
		case tokAdd, tokSub, tokMul, tokDiv, tokMod:
			leftOperand = &arithNode{operator.str, leftOperand, rightOperand}
		case tokOr, tokAnd:
			leftOperand = &logicalNode{operator.str, leftOperand, rightOperand}
		case tokEq, tokNeq:
			leftOperand = &equalityNode{operator.str, leftOperand, rightOperand}
		case tokLeq, tokGeq, tokGtr, tokLss:
			leftOperand = &comparatorNode{operator.str, leftOperand, rightOperand}
		}
	}
}

func (p *parser) parseExpr(name string) node {
	defer un(trace(p, name))
	n := p.parseBinaryExpr(nil, 1)
	p.skipWhiteSpaces()
	return n
}

func (p *parser) parseString(name string) node {
	defer un(trace(p, name))
	p.skipWhiteSpaces()
	if p.parseExact("\"") {
		p.backward(1)
		return p.parseExpr("")
	}
	n := p.parseText(p.parseUnquotedStringToken, true)
	p.skipWhiteSpaces()
	return n
}

func (p *parser) parseValue() node {
	defer un(trace(p, "value"))
	p.skipWhiteSpaces()
	if p.parseExact("eval ") || p.parseExact("[") {
		p.backward(1)
		return p.parseExpr("")
	}
	if !p.parseExact("$((") && p.parseExact("$(") {
		n := p.parseCommand("")
		p.expect(")")
		return n
	}
	p.reset()
	return p.parseString("")
}

func (p *parser) parseStringOrVec(name string) node {
	defer un(trace(p, name))
	p.skipWhiteSpaces()
	if p.parseExact("[") {
		p.backward(1)
		return p.parseExpr("")
	}
	return p.parseString("")
}

func (p *parser) parseArgValue() string {
	defer un(trace(p, "value"))
	if p.parseExact("\"") {
		for {
			c := p.next()
			if c == '"' {
				v := p.item(true)
				return v[1 : len(v)-1]
			}
			if c == eof {
				p.error("\" opened but not closed")
			}
		}
	}
	p.skipWhiteSpaces()
	for p.parseComplexWord("") != "" {
		if !p.parseExact(":") {
			break
		}
	}
	return p.item(true)
}

func (p *parser) parseSingleArg(allowedArgs []string, allowedFlags []string) (string, string) {
	defer un(trace(p, "single argument"))
	arg := p.parseSimpleWord("name")
	var value string
	if sliceContains(allowedArgs, arg) {
		value = p.parseArgValue()
	} else if sliceContains(allowedFlags, arg) {
		value = ""
	} else {
		p.error("unexpected argument : " + arg)
	}
	p.skipWhiteSpaces()
	return arg, value
}

func (p *parser) parseArgs(allowedArgs []string, allowedFlags []string) map[string]string {
	defer un(trace(p, "args"))
	args := map[string]string{}
	p.skipWhiteSpaces()
	for {
		if !p.parseExact("-") {
			break
		}
		arg, value := p.parseSingleArg(allowedArgs, allowedFlags)
		args[arg] = value
	}
	return args
}

func (p *parser) parseAssign() string {
	defer un(trace(p, "assign"))
	varName := p.parseSimpleWord("var name")
	p.expect("=")
	return varName
}

func (p *parser) parseIndexing() node {
	defer un(trace(p, "indexing"))
	p.skipWhiteSpaces()
	p.expect("[")
	index := p.parseExpr("index")
	p.expect("]")
	return index
}

func (p *parser) parseLsObjArguments() map[string]string {
	defer un(trace(p, "lsobj arguments"))
	return p.parseArgs([]string{"s", "f"}, []string{"r"})
}

func (p *parser) parseLsObj(lsIdx int) node {
	defer un(trace(p, ""))
	args := p.parseLsObjArguments()
	path := p.parsePath("")
	_, recursive := args["r"]
	sort := args["s"]
	var attrList []string
	if formatArg, ok := args["f"]; ok {
		attrList = strings.Split(formatArg, ":")
	}
	return &lsObjNode{path, lsIdx, recursive, sort, attrList}
}

func (p *parser) parseLsArguments() map[string]string {
	defer un(trace(p, "ls arguments"))
	return p.parseArgs([]string{"s", "f"}, []string{"r"})
}

func (p *parser) parseLs() node {
	defer un(trace(p, "ls"))
	args := p.parseLsArguments()
	path := p.parsePath("")
	if attr, ok := args["s"]; ok {
		return &lsAttrNode{path, attr}
	}
	return &lsNode{path}
}

func (p *parser) parseGet() node {
	defer un(trace(p, "get"))
	return &getObjectNode{p.parsePath("")}
}

func (p *parser) parseGetU() node {
	defer un(trace(p, "getu"))
	return &getUNode{p.parsePath(""), p.parseExpr("u")}
}

func (p *parser) parseGetSlot() node {
	defer un(trace(p, "getslot"))
	return &getSlotNode{p.parsePath(""), p.parseString("slot name")}
}

func (p *parser) parseUndraw() node {
	defer un(trace(p, "undraw"))
	if p.commandEnd() {
		return &undrawNode{nil}
	}
	return &undrawNode{p.parsePath("")}
}

func (p *parser) parseDrawArgs() map[string]string {
	defer un(trace(p, "draw args"))
	return p.parseArgs([]string{}, []string{"f"})
}

func (p *parser) parseDraw() node {
	defer un(trace(p, "draw"))
	args := p.parseDrawArgs()
	_, force := args["f"]
	path := p.parsePath("")
	depth := 0
	if !p.commandEnd() {
		depth = p.parseInt("depth")
	}
	return &drawNode{path, depth, force}
}

func (p *parser) parseDrawable() node {
	defer un(trace(p, "drawable"))
	path := p.parsePath("")
	if p.commandEnd() {
		return &isEntityDrawableNode{path}
	}
	attrName := p.parseComplexWord("attribute")
	return &isAttrDrawableNode{path, attrName}
}

func (p *parser) parseHc() node {
	defer un(trace(p, "hc"))
	path := p.parsePath("")
	if p.commandEnd() {
		return &hierarchyNode{path, 1}
	}
	depth := p.parseInt("depth")
	return &hierarchyNode{path, depth}
}

func (p *parser) parseUnset() node {
	defer un(trace(p, "unset"))
	args := p.parseArgs([]string{"f", "v"}, []string{})
	if len(args) == 0 {
		path := p.parsePath("")
		p.expect(":")
		attr := p.parseComplexWord("attribute")
		index := p.parseIndexing()
		return &unsetAttrNode{path, attr, index}
	}
	if funcName, ok := args["f"]; ok {
		return &unsetFuncNode{funcName}
	}
	if varName, ok := args["v"]; ok {
		return &unsetVarNode{varName}
	}
	panic("unexpected argument while unset command")
}

func (p *parser) parseEnv() node {
	defer un(trace(p, "env"))
	if p.commandEnd() {
		return &envNode{}
	}
	return &setEnvNode{p.parseAssign(), p.parseExpr("")}
}

func (p *parser) parseDelete() node {
	defer un(trace(p, "delete"))
	if p.parseExact("selection") {
		return &deleteSelectionNode{}
	}
	return &deleteObjNode{p.parsePath("")}
}

func (p *parser) parseEqual() node {
	defer un(trace(p, "="))
	if p.parseExact("{") {
		p.backward(1)
		return &selectChildrenNode{p.parsePathGroup()}
	}
	if p.commandEnd() {
		return &selectObjectNode{&valueNode{""}}
	}
	return &selectObjectNode{p.parsePath("")}
}

func (p *parser) parseVar() node {
	defer un(trace(p, "variable assignment"))
	varName := p.parseAssign()
	p.skipWhiteSpaces()
	value := p.parseValue()
	return &assignNode{varName, value}
}

func (p *parser) parseLoad() node {
	defer un(trace(p, "load"))
	return &loadNode{p.parseString("file path")}
}

func (p *parser) parseTemplate() node {
	defer un(trace(p, "template"))
	return &loadTemplateNode{p.parseString("template path")}
}

func (p *parser) parseLen() node {
	defer un(trace(p, "len"))
	return &lenNode{p.parseSimpleWord("variable")}
}

func (p *parser) parseLink() node {
	defer un(trace(p, "link"))
	sourcePath := p.parsePath("source path")
	p.expect("@")
	destPath := p.parsePath("destination path")
	if p.parseExact("@") {
		slot := p.parseString("slot name")
		return &linkObjectNode{sourcePath, destPath, slot}
	}
	return &linkObjectNode{sourcePath, destPath, nil}
}

func (p *parser) parseUnlink() node {
	defer un(trace(p, "unlink"))
	sourcePath := p.parsePath("source")
	if p.parseExact("@") {
		destPath := p.parsePath("destination")
		return &unlinkObjectNode{sourcePath, destPath}
	}
	return &unlinkObjectNode{sourcePath, nil}
}

func (p *parser) parsePrint() node {
	defer un(trace(p, "print"))
	return &printNode{p.parseValue()}
}

func (p *parser) parseMan() node {
	defer un(trace(p, "man"))
	if p.commandEnd() {
		return &helpNode{""}
	}
	commandName := p.parseKeyWord(manCommands)
	if !sliceContains(manCommands, commandName) {
		p.error("no manual for this command")
	}
	return &helpNode{commandName}
}

func (p *parser) parseCd() node {
	defer un(trace(p, "cd"))
	if p.commandEnd() {
		return &cdNode{&pathNode{&valueNode{"/"}}}
	}
	return &cdNode{p.parsePath("")}
}

func (p *parser) parseTree() node {
	defer un(trace(p, "tree"))
	if p.commandEnd() {
		return &treeNode{&pathNode{&valueNode{"."}}, 0}
	}
	path := p.parsePath("")
	if p.commandEnd() {
		return &treeNode{path, 0}
	}
	depth := p.parseInt("depth")
	return &treeNode{path, depth}
}

func (p *parser) parseUi() node {
	defer un(trace(p, "ui"))
	if p.parseExact("clearcache") {
		return &uiClearCacheNode{}
	}
	key := p.parseAssign()
	if key == "delay" {
		return &uiDelayNode{p.parseFloat("delay")}
	}
	if key == "debug" || key == "infos" || key == "wireframe" {
		val := p.parseBool()
		return &uiToggleNode{key, val}
	}
	if key == "highlight" || key == "hl" {
		return &uiHighlightNode{p.parsePath("")}
	}
	p.error("unknown ui command")
	return nil
}

func (p *parser) parseCamera() node {
	defer un(trace(p, "camera"))
	key := p.parseAssign()
	if key == "move" || key == "translate" {
		position := p.parseExpr("position")
		p.expect("@")
		rotation := p.parseExpr("rotation")
		return &cameraMoveNode{key, position, rotation}
	}
	if key == "wait" {
		return &cameraWaitNode{p.parseFloat("waiting time")}
	}
	p.error("unknown ui command")
	return nil
}

func (p *parser) parseFocus() node {
	defer un(trace(p, "focus"))
	if p.commandEnd() {
		return &focusNode{&valueNode{""}}
	}
	return &focusNode{p.parsePath("")}
}

func (p *parser) parseWhile() node {
	defer un(trace(p, "while"))
	condition := p.parseExpr("condition")
	p.expect("{")
	body := p.parseCommand("body")
	p.skipWhiteSpaces()
	p.expect("}")
	return &whileNode{condition, body}
}

func (p *parser) parseFor() node {
	defer un(trace(p, "for"))
	varName := p.parseSimpleWord("variable")
	p.expect("in")
	start := p.parseExpr("start index")
	p.expect("..")
	end := p.parseExpr("end index")
	p.expect("{")
	body := p.parseCommand("body")
	p.skipWhiteSpaces()
	p.expect("}")
	return &forRangeNode{varName, start, end, body}
}

func (p *parser) parseIf() node {
	defer un(trace(p, "if"))
	condition := p.parseExpr("condition")
	p.expect("{")
	body := p.parseCommand("if body")
	p.expect("}")
	p.skipWhiteSpaces()
	keyword := p.parseKeyWord([]string{"else", "elif"})
	switch keyword {
	case "":
		return &ifNode{condition, body, nil}
	case "else":
		p.skipWhiteSpaces()
		p.expect("{")
		elseBody := p.parseCommand("else body")
		p.skipWhiteSpaces()
		p.expect("}")
		return &ifNode{condition, body, elseBody}
	case "elif":
		elseBody := p.parseIf()
		return &ifNode{condition, body, elseBody}
	default:
		panic("unexpected return from parseKeyWord : " + keyword)
	}
}

func (p *parser) parseAlias() node {
	defer un(trace(p, "alias"))
	name := p.parseSimpleWord("name")
	p.expect("{")
	p.skipWhiteSpaces()
	command := p.parseCommand("body")
	p.skipWhiteSpaces()
	p.expect("}")
	return &funcDefNode{name, command}
}

func (p *parser) parseObjType() string {
	defer un(trace(p, "object type"))
	candidates := []string{}
	for command := range createObjDispatch {
		candidates = append(candidates, command)
	}
	return p.parseKeyWord(candidates)
}

func (p *parser) parseCreate() node {
	defer un(trace(p, "create"))
	objType := p.parseObjType()
	if objType == "" {
		p.error("unknown object type")
	}
	p.skipWhiteSpaces()
	if objType == "orphan" {
		return p.parseCreateOrphan()
	}
	p.expect(":")
	p.skipWhiteSpaces()
	return createObjDispatch[objType]()
}

func (p *parser) parseCreateDomain() node {
	defer un(trace(p, "create domain"))
	path := p.parsePath("")
	p.expect("@")
	color := p.parseString("color")
	return &createDomainNode{path, color}
}

func (p *parser) parseCreateSite() node {
	defer un(trace(p, "create site"))
	path := p.parsePath("")
	return &createSiteNode{path}
}

func (p *parser) parseCreateBuilding() node {
	defer un(trace(p, "create building"))
	path := p.parsePath("")
	p.expect("@")
	posXY := p.parseExpr("posXY")
	p.expect("@")
	rotation := p.parseExpr("rotation")
	p.expect("@")
	sizeOrTemplate := p.parseStringOrVec("sizeOrTemplate")
	return &createBuildingNode{path, posXY, rotation, sizeOrTemplate}
}

func (p *parser) parseCreateRoom() node {
	defer un(trace(p, "create room"))
	path := p.parsePath("")
	p.expect("@")
	posXY := p.parseExpr("posXY")
	p.expect("@")
	rotation := p.parseExpr("rotation")
	p.expect("@")
	sizeOrTemplate := p.parseStringOrVec("sizeOrTemplate")
	if !p.parseExact("@") {
		return &createRoomNode{path, posXY, rotation, nil, nil, nil, sizeOrTemplate}
	}
	axisOrientation := p.parseString("axisOrientation")
	if !p.parseExact("@") {
		return &createRoomNode{path, posXY, rotation, sizeOrTemplate, axisOrientation, nil, nil}
	}
	floorUnit := p.parseString("floorUnit")
	return &createRoomNode{path, posXY, rotation, sizeOrTemplate, axisOrientation, floorUnit, nil}
}

func (p *parser) parseCreateRack() node {
	defer un(trace(p, "create rack"))
	path := p.parsePath("")
	p.expect("@")
	pos := p.parseExpr("position")
	p.expect("@")
	sizeOrTemplate := p.parseStringOrVec("sizeOrTemplate")
	p.expect("@")
	orientation := p.parseString("orientation")
	return &createRackNode{path, pos, sizeOrTemplate, orientation}
}

func (p *parser) parseCreateDevice() node {
	defer un(trace(p, "create device"))
	path := p.parsePath("")
	p.expect("@")
	posUOrSlot := p.parseString("posUOrSlot")
	p.expect("@")
	sizeUOrTemplate := p.parseString("sizeUOrTemplate")
	if !p.parseExact("@") {
		return &createDeviceNode{path, posUOrSlot, sizeUOrTemplate, nil}
	}
	side := p.parseString("side")
	return &createDeviceNode{path, posUOrSlot, sizeUOrTemplate, side}
}

func (p *parser) parseCreateGroup() node {
	defer un(trace(p, "create group"))
	path := p.parsePath("")
	p.expect("@")
	childs := p.parsePathGroup()
	return &createGroupNode{path, childs}
}

func (p *parser) parseCreateCorridor() node {
	defer un(trace(p, "create corridor"))
	path := p.parsePath("")
	p.expect("@")
	racks := p.parsePathGroup()
	if len(racks) != 2 {
		p.error("only 2 racks expected")
	}
	p.expect("@")
	temperature := p.parseString("temperature")
	return &createCorridorNode{path, racks[0], racks[1], temperature}
}

func (p *parser) parseCreateOrphan() node {
	defer un(trace(p, "create orphan"))
	if !p.parseExact("device") && !p.parseExact("dv") {
		p.error("device or dv keyword expected")
	}
	p.skipWhiteSpaces()
	p.expect(":")
	p.skipWhiteSpaces()
	path := p.parsePath("")
	p.expect("@")
	template := p.parseString("template")
	return &createOrphanNode{path, template}
}

func (p *parser) parseUpdate() node {
	defer un(trace(p, "update"))
	path := p.parsePath("")
	p.skipWhiteSpaces()
	p.expect(":")
	p.skipWhiteSpaces()
	attr := p.parseComplexWord("attribute")
	p.skipWhiteSpaces()
	p.expect("=")
	p.skipWhiteSpaces()
	sharpe := p.parseExact("#")
	values := []node{}
	moreValues := true
	for moreValues {
		value := p.parseValue()
		values = append(values, value)
		moreValues = p.parseExact("@")
	}
	return &updateObjNode{path, attr, values, sharpe}
}

func (p *parser) parseCommandKeyWord() string {
	defer un(trace(p, "command keyword"))
	candidates := []string{}
	for command := range commandDispatch {
		candidates = append(candidates, command)
	}
	for command := range noArgsCommands {
		candidates = append(candidates, command)
	}
	candidates = append(candidates, lsCommands...)
	return p.parseKeyWord(candidates)
}

func (p *parser) parseSingleCommand() node {
	defer un(trace(p, ""))
	p.skipWhiteSpaces()
	if p.commandEnd() {
		return nil
	}
	commandKeyWord := p.parseCommandKeyWord()
	if commandKeyWord != "" {
		if lsIdx := indexOf(lsCommands, commandKeyWord); lsIdx != -1 {
			p.skipWhiteSpaces()
			return p.parseLsObj(lsIdx)
		}
		parseFunc, ok := commandDispatch[commandKeyWord]
		if ok {
			p.skipWhiteSpaces()
			return parseFunc()
		}
		result, ok := noArgsCommands[commandKeyWord]
		if ok {
			return result
		}
	}
	funcName := p.parseSimpleWord("function name")
	if funcName != "" && p.commandEnd() {
		return &funcCallNode{funcName}
	}
	p.reset()
	return p.parseUpdate()
}

func (p *parser) parseCommand(name string) node {
	defer un(trace(p, name))
	commandDispatch = map[string]parseCommandFunc{
		"ls":         p.parseLs,
		"get":        p.parseGet,
		"getu":       p.parseGetU,
		"getslot":    p.parseGetSlot,
		"undraw":     p.parseUndraw,
		"draw":       p.parseDraw,
		"drawable":   p.parseDrawable,
		"hc":         p.parseHc,
		"unset":      p.parseUnset,
		"env":        p.parseEnv,
		"+":          p.parseCreate,
		"-":          p.parseDelete,
		"=":          p.parseEqual,
		".var:":      p.parseVar,
		".cmds:":     p.parseLoad,
		".template:": p.parseTemplate,
		"len":        p.parseLen,
		"link:":      p.parseLink,
		"unlink":     p.parseUnlink,
		"print":      p.parsePrint,
		"man":        p.parseMan,
		"cd":         p.parseCd,
		"tree":       p.parseTree,
		"ui.":        p.parseUi,
		"camera.":    p.parseCamera,
		">":          p.parseFocus,
		"while":      p.parseWhile,
		"for":        p.parseFor,
		"if":         p.parseIf,
		"alias":      p.parseAlias,
	}
	createObjDispatch = map[string]parseCommandFunc{
		"domain":   p.parseCreateDomain,
		"dm":       p.parseCreateDomain,
		"site":     p.parseCreateSite,
		"si":       p.parseCreateSite,
		"bldg":     p.parseCreateBuilding,
		"building": p.parseCreateBuilding,
		"bd":       p.parseCreateBuilding,
		"room":     p.parseCreateRoom,
		"ro":       p.parseCreateRoom,
		"rack":     p.parseCreateRack,
		"rk":       p.parseCreateRack,
		"device":   p.parseCreateDevice,
		"dv":       p.parseCreateDevice,
		"corridor": p.parseCreateCorridor,
		"co":       p.parseCreateCorridor,
		"group":    p.parseCreateGroup,
		"gr":       p.parseCreateGroup,
		"orphan":   p.parseCreateOrphan,
	}
	noArgsCommands = map[string]node{
		"selection":    &selectNode{},
		"clear":        &clrNode{},
		"grep":         &grepNode{},
		"lsog":         &lsogNode{},
		"lsenterprise": &lsenterpriseNode{},
		"pwd":          &pwdNode{},
		"exit":         &exitNode{},
	}
	commands := []node{}
	var command node
	for {
		command = p.parseSingleCommand()
		commands = append(commands, command)
		p.skipWhiteSpaces()
		if !p.parseExact(";") {
			if len(commands) > 1 {
				return &ast{commands}
			}
			return command
		}
		p.skipWhiteSpaces()
	}
}

func newParser(buffer string) *parser {
	return &parser{
		buf:        buffer,
		stackTrace: []traceItem{},
	}
}

func Parse(buffer string) (n node, err error) {
	commentIdx := strings.Index(buffer, "//")
	if commentIdx != -1 {
		buffer = buffer[:commentIdx]
	}
	p := newParser(buffer)
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf(r.(string))
		}
	}()
	n = p.parseCommand("")
	if !p.commandEnd() {
		p.error("unexpected character")
	}
	return n, nil
}
