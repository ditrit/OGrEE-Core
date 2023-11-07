package main

import (
	"reflect"
	"runtime/debug"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func (p *parser) remaining() string {
	if p.cursor >= len(p.buf) {
		return ""
	}
	return p.buf[p.cursor:]
}

func recoverFunc(t *testing.T) {
	if r := recover(); r != nil {
		t.Errorf("Error while parsing : %s\n%s", r, string(debug.Stack()))
	}
}

func TestParseExact(t *testing.T) {
	defer recoverFunc(t)
	p := newParser("testabc")
	if !p.parseExact("test") {
		t.Errorf("should return true")
	}
	if p.remaining() != "abc" {
		t.Errorf("wrong stop, remaining buf : %s", p.remaining())
	}
	p = newParser("abctest")
	if p.parseExact("test") {
		t.Errorf("should return false")
	}
	if p.remaining() != "abctest" {
		t.Errorf("wrong stop, remaining buf : %s", p.remaining())
	}
	p = newParser("test")
	if !p.parseExact("test") {
		t.Errorf("should return true")
	}
	if p.remaining() != "" {
		t.Errorf("wrong stop, remaining buf : %s", p.remaining())
	}
}

func TestParseWord(t *testing.T) {
	defer recoverFunc(t)
	p := newParser("test abc")
	word := p.parseSimpleWord("")
	if word != "test" {
		t.Errorf("wrong word parsed")
	}
	if p.remaining() != "abc" {
		t.Errorf("wrong stop, remaining buf : %s", p.remaining())
	}
}

func TestParsePathGroup(t *testing.T) {
	defer recoverFunc(t)
	s := "{ test.plouf.plaf , test.plaf.plouf } a"
	p := newParser(s)
	paths := p.parsePathGroup()
	firstNode := &pathNode{path: &valueNode{"test.plouf.plaf"}}
	secondNode := &pathNode{path: &valueNode{"test.plaf.plouf"}}
	if !reflect.DeepEqual(paths, []node{firstNode, secondNode}) {
		t.Errorf("wrong path group parsed : %s", spew.Sdump(paths))
	}
	if p.remaining() != "a" {
		t.Errorf("wrong stop, remaining buf : %s", p.remaining())
	}
}

func TestParseWordSingleLetter(t *testing.T) {
	defer recoverFunc(t)
	p := newParser("a 42")
	word := p.parseSimpleWord("")
	if word != "a" {
		t.Errorf("wrong word parsed")
	}
	if p.remaining() != "42" {
		t.Errorf("wrong stop, remaining buf : %s", p.remaining())
	}
}

func TestParseArgs(t *testing.T) {
	defer recoverFunc(t)
	p := newParser("-a 42 -v -f -s dazd coucou.plouf")
	args := p.parseArgs([]string{"a", "s"}, []string{"v", "f"}, "")
	if p.cursor != 20 {
		t.Errorf("wrong end position for left arguments : %d", p.cursor)
		return
	}
	if !reflect.DeepEqual(args, map[string]string{"a": "42", "s": "dazd", "v": "", "f": ""}) {
		t.Errorf("wrong args returned : %s", spew.Sdump(args))
		return
	}
	p = newParser(" -f toto.tata")
	args = p.parseArgs([]string{}, []string{"f"}, "")
	if !reflect.DeepEqual(args, map[string]string{"f": ""}) {
		t.Errorf("wrong args returned : %v", args)
		return
	}
}

func TestParseExpr(t *testing.T) {
	defer recoverFunc(t)
	s := "\"plouf\" + (3 - 4.2) * $ab - ${a} + format(\"%03d\", 47)  42"
	p := newParser(s)
	expr := p.parseExpr("")
	expectedExpr := &arithNode{
		op: "+",
		left: &arithNode{
			op: "-",
			left: &arithNode{
				op:   "+",
				left: &valueNode{"plouf"},
				right: &arithNode{
					op: "*",
					left: &arithNode{
						op:    "-",
						left:  &valueNode{3},
						right: &valueNode{4.2},
					},
					right: &symbolReferenceNode{"ab"},
				},
			},
			right: &symbolReferenceNode{"a"},
		},
		right: &formatStringNode{&valueNode{"%03d"}, []node{&valueNode{47}}},
	}
	if !reflect.DeepEqual(expr, expectedExpr) {
		t.Errorf("unexpected expression : \n%s", spew.Sdump(expr))
	}
	if p.cursor != len(s)-2 {
		t.Errorf("unexpected cursor : %d", p.cursor)
	}
	p = newParser("$a+3))")
	expr = p.parseExpr("")
	expectedExpr = &arithNode{
		op:    "+",
		left:  &symbolReferenceNode{"a"},
		right: &valueNode{3},
	}
	if !reflect.DeepEqual(expr, expectedExpr) {
		t.Errorf("unexpected expression : \n%s", spew.Sdump(expr))
	}
	if p.cursor != 4 {
		t.Errorf("unexpected cursor : %d", p.cursor)
	}
}

func TestParseExprRange(t *testing.T) {
	defer recoverFunc(t)
	p := newParser("42..48")
	expr := p.parseExpr("")
	expected := &valueNode{42}
	if !reflect.DeepEqual(expr, expected) {
		t.Errorf("unexpected expression : \n%s", spew.Sdump(expr))
	}
}

func TestParseExprCompare(t *testing.T) {
	defer recoverFunc(t)
	p := newParser("$i<6 {print \"a\"}")
	expr := p.parseExpr("")
	expected := &comparatorNode{"<", &symbolReferenceNode{"i"}, &valueNode{6}}
	if !reflect.DeepEqual(expr, expected) {
		t.Errorf("unexpected expression : \n%s", spew.Sdump(expr))
	}
}

func TestParseExprString(t *testing.T) {
	defer recoverFunc(t)
	p := newParser("\"${a}test\"")
	expr := p.parseExpr("")
	expected := &formatStringNode{&valueNode{"%vtest"}, []node{&symbolReferenceNode{"a"}}}
	if !reflect.DeepEqual(expr, expected) {
		t.Errorf("unexpected expression : \n%s", spew.Sdump(expr))
		t.Errorf("unexpected parsing : \ntree : %s\nexpected : %s",
			spew.Sdump(expr), spew.Sdump(expected))
	}
}

func TestParseExprArrayRef(t *testing.T) {
	defer recoverFunc(t)
	p := newParser("$ab[42 + 1]")
	expr := p.parseExpr("")
	expected := &arrayReferenceNode{"ab", &arithNode{op: "+", left: &valueNode{42}, right: &valueNode{1}}}
	if !reflect.DeepEqual(expr, expected) {
		t.Errorf("unexpected parsing : \ntree : %s\nexpected : %s",
			spew.Sdump(expr), spew.Sdump(expected))
	}
}

func TestParseRawText(t *testing.T) {
	defer recoverFunc(t)
	p := newParser("${a}a")
	expr := p.parseText(p.parseUnquotedStringToken, false)
	expected := &formatStringNode{&valueNode{"%va"}, []node{&symbolReferenceNode{"a"}}}
	if !reflect.DeepEqual(expr, expected) {
		t.Errorf("unexpected expression : \n%s", spew.Sdump(expr))
	}
}

func TestParseString(t *testing.T) {
	defer recoverFunc(t)
	p := newParser("${a}a")
	expr := p.parseString("")
	expected := &formatStringNode{&valueNode{"%va"}, []node{&symbolReferenceNode{"a"}}}
	if !reflect.DeepEqual(expr, expected) {
		t.Errorf("unexpected expression : \n%s", spew.Sdump(expr))
	}
}

func TestParseAssign(t *testing.T) {
	defer recoverFunc(t)
	p := newParser("test= plouf")
	va := p.parseAssign("")
	if va != "test" {
		t.Errorf("wrong variable parserd : %s", va)
	}
	if p.remaining() != " plouf" {
		t.Errorf("wrong stop, remaining buf : %s", p.remaining())
	}
}

func assertParsing(buf string, n node, expected node, t *testing.T) {
	if !reflect.DeepEqual(n, expected) {
		t.Errorf("unexpected parsing : \n%s\n\ntree : %s\nexpected : %s",
			buf, spew.Sdump(n), spew.Sdump(expected))
	}
}

func testCommand(buffer string, expected node, t *testing.T) bool {
	n, err := Parse(buffer)
	if err != nil {
		t.Errorf("cannot parse command : \n%s\n%s", buffer, err.Error())
		return false
	}
	assertParsing(buffer, n, expected, t)
	return true
}

func TestParseLs(t *testing.T) {
	buffer := "lsbuilding -s height - f attr1:attr2 plouf.plaf attr1=a, attr2=b"
	path := &pathNode{path: &valueNode{"plouf.plaf"}}
	sort := "height"
	attrList := []string{"attr1", "attr2"}
	filters := map[string]node{
		"category": &valueNode{"building"},
		"attr1":    &valueNode{"a"},
		"attr2":    &valueNode{"b"},
	}
	expected := &lsNode{path, filters, sort, attrList}
	testCommand(buffer, expected, t)
	buffer = "lsbuilding -s height - f \"attr1:attr2\" plouf.plaf attr1=a, attr2=b"
	testCommand(buffer, expected, t)
}

var testPath = &pathNode{path: &formatStringNode{&valueNode{"%v/tata"}, []node{&symbolReferenceNode{"toto"}}}}
var testPathUpdate = &pathNode{path: &formatStringNode{&valueNode{"%v/tata"}, []node{&symbolReferenceNode{"toto"}}}, acceptSelection: true}
var testPath2 = &pathNode{path: &valueNode{"/toto/../tata"}}

func vec2(x float64, y float64) node {
	return &arrNode{[]node{&valueNode{x}, &valueNode{y}}}
}

func vec3(x float64, y float64, z float64) node {
	return &arrNode{[]node{&valueNode{x}, &valueNode{y}, &valueNode{z}}}
}

func vec4(x float64, y float64, z float64, w float64) node {
	return &arrNode{[]node{&valueNode{x}, &valueNode{y}, &valueNode{z}, &valueNode{w}}}
}

var commandsMatching = map[string]node{
	"man":                            &helpNode{""},
	"man draw":                       &helpNode{"draw"},
	"man camera":                     &helpNode{"camera"},
	"man ui":                         &helpNode{"ui"},
	"ls":                             &lsNode{&pathNode{path: &valueNode{""}}, map[string]node{}, "", nil},
	"cd":                             &cdNode{&pathNode{path: &valueNode{"/"}}},
	"tree":                           &treeNode{&pathNode{path: &valueNode{"."}}, 1},
	"get ${toto}/tata":               &getObjectNode{testPath},
	"getu rackA 42":                  &getUNode{&pathNode{path: &valueNode{"rackA"}}, &valueNode{42}},
	"undraw":                         &undrawNode{nil},
	"undraw ${toto}/tata":            &undrawNode{testPath},
	"draw":                           &drawNode{&pathNode{path: &valueNode{""}}, 0, false},
	"draw ${toto}/tata":              &drawNode{testPath, 0, false},
	"draw ${toto}/tata 4":            &drawNode{testPath, 4, false},
	"draw -f":                        &drawNode{&pathNode{path: &valueNode{""}}, 0, true},
	"draw -f ${toto}/tata":           &drawNode{testPath, 0, true},
	"draw -f ${toto}/tata 4 ":        &drawNode{testPath, 4, true},
	".cmds:../toto/tata.ocli":        &loadNode{&valueNode{"../toto/tata.ocli"}},
	".template:../toto/tata.ocli":    &loadTemplateNode{&valueNode{"../toto/tata.ocli"}},
	".var:a=42":                      &assignNode{"a", &valueNode{"42"}},
	".var:b= $(($a+3))":              &assignNode{"b", &formatStringNode{&valueNode{"%v"}, []node{&arithNode{"+", &symbolReferenceNode{"a"}, &valueNode{3}}}}},
	"=${toto}/tata":                  &selectObjectNode{testPath},
	"=..":                            &selectObjectNode{&pathNode{path: &valueNode{".."}}},
	"={${toto}/tata}":                &selectChildrenNode{[]node{testPath}},
	"={${toto}/tata, /toto/../tata}": &selectChildrenNode{[]node{testPath, testPath2}},
	"-${toto}/tata":                  &deleteObjNode{testPath},
	">${toto}/tata":                  &focusNode{testPath},
	"+site:${toto}/tata":             &createSiteNode{testPath},
	"+si:${toto}/tata":               &createSiteNode{testPath},
	"+building:${toto}/tata@[1., 2.]@3.@[.1, 2., 3.]":           &createBuildingNode{testPath, vec2(1., 2.), &valueNode{3.}, vec3(.1, 2., 3.)},
	"+room:${toto}/tata@[1., 2.]@3.@[.1, 2., 3.]@+x-y":          &createRoomNode{testPath, vec2(1., 2.), &valueNode{3.}, vec3(.1, 2., 3.), &valueNode{"+x-y"}, nil, nil},
	"+room:${toto}/tata@[1., 2.]@3.@[.1, 2., 3.]@+x-y@m":        &createRoomNode{testPath, vec2(1., 2.), &valueNode{3.}, vec3(.1, 2., 3.), &valueNode{"+x-y"}, &valueNode{"m"}, nil},
	"+room:${toto}/tata@[1., 2.]@3.@template":                   &createRoomNode{testPath, vec2(1., 2.), &valueNode{3.}, nil, nil, nil, &valueNode{"template"}},
	"+rack:${toto}/tata@[1., 2.]@t@front@[.1, 2., 3.]":          &createRackNode{testPath, vec2(1., 2.), &valueNode{"t"}, &valueNode{"front"}, vec3(.1, 2., 3.)},
	"+rack:${toto}/tata@[1., 2.]@m@front@template":              &createRackNode{testPath, vec2(1., 2.), &valueNode{"m"}, &valueNode{"front"}, &valueNode{"template"}},
	"+rack:${toto}/tata@[1., 2.]@m@[.1, 2., 3.]@template":       &createRackNode{testPath, vec2(1., 2.), &valueNode{"m"}, vec3(.1, 2., 3.), &valueNode{"template"}},
	"+device:${toto}/tata@42@42":                                &createDeviceNode{testPath, &valueNode{"42"}, &valueNode{"42"}, nil},
	"+device:${toto}/tata@42@template":                          &createDeviceNode{testPath, &valueNode{"42"}, &valueNode{"template"}, nil},
	"+device:${toto}/tata@42@template@frontflipped ":            &createDeviceNode{testPath, &valueNode{"42"}, &valueNode{"template"}, &valueNode{"frontflipped"}},
	"+device:${toto}/tata@slot42@42":                            &createDeviceNode{testPath, &valueNode{"slot42"}, &valueNode{"42"}, nil},
	"+device:${toto}/tata@slot42@template":                      &createDeviceNode{testPath, &valueNode{"slot42"}, &valueNode{"template"}, nil},
	"+device:${toto}/tata@slot42@template@frontflipped ":        &createDeviceNode{testPath, &valueNode{"slot42"}, &valueNode{"template"}, &valueNode{"frontflipped"}},
	"+group:${toto}/tata@{c1, c2}":                              &createGroupNode{testPath, []node{&pathNode{path: &valueNode{"c1"}}, &pathNode{path: &valueNode{"c2"}}}},
	"+corridor:${toto}/tata@[1., 2.]@t@front@[.1, 2., 3.]@cold": &createCorridorNode{testPath, vec2(1., 2.), &valueNode{"t"}, &valueNode{"front"}, vec3(.1, 2., 3.), &valueNode{"cold"}},
	"${toto}/tata:areas=[1., 2., 3., 4.]@[1., 2., 3., 4.]":      &updateObjNode{testPathUpdate, "areas", []node{vec4(1., 2., 3., 4.), vec4(1., 2., 3., 4.)}, false},
	"${toto}/tata:separator=[1., 2.]@[1., 2.]@wireframe":        &updateObjNode{testPathUpdate, "separator", []node{vec2(1., 2.), vec2(1., 2.), &valueNode{"wireframe"}}, false},
	"${toto}/tata:attr=42":                                      &updateObjNode{testPathUpdate, "attr", []node{&valueNode{"42"}}, false},
	"${toto}/tata:label=\"plouf\"":                              &updateObjNode{testPathUpdate, "label", []node{&valueNode{"plouf"}}, false},
	"${toto}/tata:labelFont=bold":                               &updateObjNode{testPathUpdate, "labelFont", []node{&valueNode{"bold"}}, false},
	"${toto}/tata:labelFont=color@42ff42":                       &updateObjNode{testPathUpdate, "labelFont", []node{&valueNode{"color"}, &valueNode{"42ff42"}}, false},
	"${toto}/tata:tilesName=true":                               &updateObjNode{testPathUpdate, "tilesName", []node{&valueNode{"true"}}, false},
	"${toto}/tata:tilesColor=false":                             &updateObjNode{testPathUpdate, "tilesColor", []node{&valueNode{"false"}}, false},
	"${toto}/tata:U=false":                                      &updateObjNode{testPathUpdate, "U", []node{&valueNode{"false"}}, false},
	"${toto}/tata:slots=false":                                  &updateObjNode{testPathUpdate, "slots", []node{&valueNode{"false"}}, false},
	"${toto}/tata:localCS=false":                                &updateObjNode{testPathUpdate, "localCS", []node{&valueNode{"false"}}, false},
	"${toto}/tata:content=false":                                &updateObjNode{testPathUpdate, "content", []node{&valueNode{"false"}}, false},
	"${toto}/tata:temperature_01-Inlet-Ambient=7":               &updateObjNode{testPathUpdate, "temperature_01-Inlet-Ambient", []node{&valueNode{"7"}}, false},
	"ui.delay=15":                            &uiDelayNode{15.},
	"ui.infos=true":                          &uiToggleNode{"infos", true},
	"ui.debug=false":                         &uiToggleNode{"debug", false},
	"ui.highlight=${toto}/tata":              &uiHighlightNode{testPath},
	"ui.hl=${toto}/tata":                     &uiHighlightNode{testPath},
	"camera.move=[1., 2., 3.]@[1., 2.]":      &cameraMoveNode{"move", vec3(1., 2., 3.), vec2(1., 2.)},
	"camera.translate=[1., 2., 3.]@[1., 2.]": &cameraMoveNode{"translate", vec3(1., 2., 3.), vec2(1., 2.)},
	"camera.wait=15":                         &cameraWaitNode{15.},
	"camera.wait = 15":                       &cameraWaitNode{15.},
	"clear":                                  &clrNode{},
	".cmds:${CUST}/DEMO.PERF.ocli":           &loadNode{&formatStringNode{&valueNode{"%v/DEMO.PERF.ocli"}, []node{&symbolReferenceNode{"CUST"}}}},
	".cmds:${a}/${b}.ocli":                   &loadNode{&formatStringNode{&valueNode{"%v/%v.ocli"}, []node{&symbolReferenceNode{"a"}, &symbolReferenceNode{"b"}}}},
	"while $i<6 {print \"a\"}":               &whileNode{&comparatorNode{"<", &symbolReferenceNode{"i"}, &valueNode{6}}, &printNode{&valueNode{"a"}}},
	"printf \"coucou %d\", 12":               &printNode{&formatStringNode{&valueNode{"coucou %d"}, []node{&valueNode{12}}}},
}

func TestSimpleCommands(t *testing.T) {
	for command, tree := range commandsMatching {
		if !testCommand(command, tree, t) {
			break
		}
	}
}

func TestParseUpdate(t *testing.T) {
	buffer := "coucou.plouf : attr = #val1 @ val2"
	expected := &updateObjNode{
		&pathNode{path: &valueNode{"coucou.plouf"}, acceptSelection: true},
		"attr",
		[]node{&valueNode{"val1"}, &valueNode{"val2"}},
		true,
	}
	testCommand(buffer, expected, t)
}

func TestSequence(t *testing.T) {
	for command1, tree1 := range commandsMatching {
		for command2, tree2 := range commandsMatching {
			seq := command1 + "; " + command2
			expected := &ast{[]node{tree1, tree2}}
			testCommand(seq, expected, t)
		}
	}
}

func TestFor(t *testing.T) {
	for simpleCommand, tree := range commandsMatching {
		command := "for i in 0..42 { " + simpleCommand + " }"
		expected := &forRangeNode{"i", &valueNode{0}, &valueNode{42}, tree}
		testCommand(command, expected, t)
	}
}

func TestIf(t *testing.T) {
	defer recoverFunc(t)
	for simpleCommand, tree := range commandsMatching {
		buf := "true { " + simpleCommand + " }e"
		p := newParser(buf)
		expected := &ifNode{&valueNode{true}, tree, nil}
		result := p.parseIf()
		assertParsing(buf, result, expected, t)
	}
}

func TestElif(t *testing.T) {
	command := "if 5 == 6  {ls;} elif 5 == 4 {tree;} else {pwd;}"
	condition := &equalityNode{"==", &valueNode{5}, &valueNode{6}}
	conditionElif := &equalityNode{"==", &valueNode{5}, &valueNode{4}}
	ifBody := &ast{[]node{&lsNode{&pathNode{path: &valueNode{""}}, map[string]node{}, "", nil}, nil}}
	elifBody := &ast{[]node{&treeNode{&pathNode{path: &valueNode{"."}}, 1}, nil}}
	elseBody := &ast{[]node{&pwdNode{}, nil}}
	elif := &ifNode{conditionElif, elifBody, elseBody}
	expected := &ifNode{condition, ifBody, elif}
	testCommand(command, expected, t)
}
