package main

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestFindClosing(t *testing.T) {
	frame := newFrame("(a(a)(a()\")\"a))aa")
	i := findClosing(frame)
	if i != 14 {
		t.Errorf("cannot find the closing parenthesis")
	}
	frame = newFrame("(a(a)(a()\")\"a)aa")
	i = findClosing(frame)
	if i != 16 {
		t.Errorf("closing parenthesis should not be found")
	}
}

func TestParseExact(t *testing.T) {
	frame := newFrame("testabc")
	ok, nextFrame := parseExact("test", frame)
	if !ok {
		t.Errorf("parseExact should return true")
	}
	if nextFrame.str() != "abc" {
		t.Errorf("parseExact returns the wrong next frame")
	}
	frame = newFrame("abctest")
	ok, nextFrame = parseExact("test", frame)
	if ok {
		t.Errorf("parseExact should return false")
	}
	if nextFrame.str() != "abctest" {
		t.Errorf("parseExact should return the same next frame")
	}
	frame = newFrame("test")
	ok, nextFrame = parseExact("test", frame)
	if !ok {
		t.Errorf("parseExact should return true")
	}
	if nextFrame.str() != "" {
		t.Errorf("parseExact returns the wrong next frame")
	}
}

func TestParseWord(t *testing.T) {
	frame := newFrame("test abc")
	word, nextFrame, err := parseWord(frame)
	if err != nil {
		t.Errorf(err.Error())
	}
	if word != "test" {
		t.Errorf("wrong word parsed")
	}
	if nextFrame.str() != " abc" {
		t.Errorf("wrong next frame")
	}
}

func TestParsePathGroup(t *testing.T) {
	s := "{ test.plouf.plaf , test.plaf.plouf } a"
	frame := newFrame(s)
	paths, nextFrame, err := parsePathGroup(frame)
	if err != nil {
		t.Errorf(err.Error())
	}
	firstNode := &pathNode{&strLeaf{"test.plouf.plaf"}}
	secondNode := &pathNode{&strLeaf{"test.plaf.plouf"}}
	if !reflect.DeepEqual(paths, []node{firstNode, secondNode}) {
		t.Errorf("wrong path group parsed : %s", spew.Sdump(paths))
	}
	if nextFrame.str() != " a" {
		t.Errorf("wrong next frame")
	}
}

func TestParseWordSingleLetter(t *testing.T) {
	frame := newFrame("a 42")
	word, nextFrame, err := parseWord(frame)
	if err != nil {
		t.Errorf(err.Error())
	}
	if word != "a" {
		t.Errorf("wrong word parsed")
	}
	if nextFrame.str() != " 42" {
		t.Errorf("wrong next frame")
	}
}

func TestParseArgs(t *testing.T) {
	frame := newFrame("-a 42 -v -f -s dazd coucou.plouf")
	args, frame, err := parseArgs([]string{"a", "s"}, []string{"v", "f"}, frame)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if frame.start != 20 {
		t.Errorf("wrong end position for left arguments : %d", frame.start)
		return
	}
	if !reflect.DeepEqual(args, map[string]string{"a": "42", "s": "dazd", "v": "", "f": ""}) {
		t.Errorf("wrong args returned : %v", args)
		return
	}
	frame = newFrame(" -f toto.tata")
	args, frame, err = parseArgs([]string{}, []string{"f"}, frame)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if !reflect.DeepEqual(args, map[string]string{"f": ""}) {
		t.Errorf("wrong args returned : %v", args)
		return
	}
}

func TestParseExpr(t *testing.T) {
	frame := newFrame("\"plouf\" + (3 - 4.2) * $ab - ${a}  42")
	expr, cursor, err := parseExpr(frame)
	if err != nil {
		t.Errorf(err.Error())
	}
	expectedExpr := &arithNode{
		op: "-",
		left: &arithNode{
			op:   "+",
			left: &strLeaf{"plouf"},
			right: &arithNode{
				op: "*",
				left: &arithNode{
					op:    "-",
					left:  &intLeaf{3},
					right: &floatLeaf{4.2},
				},
				right: &symbolReferenceNode{"ab"},
			},
		},
		right: &symbolReferenceNode{"a"},
	}
	if !reflect.DeepEqual(expr, expectedExpr) {
		t.Errorf("unexpected expression : \n%s", spew.Sdump(expr))
	}
	if cursor.start != 34 {
		t.Errorf("unexpected cursor : %d", cursor.start)
	}
}

func TestParseExprRange(t *testing.T) {
	frame := newFrame("42..48")
	expr, _, _ := parseExpr(frame)
	expected := &intLeaf{42}
	if !reflect.DeepEqual(expr, expected) {
		t.Errorf("unexpected expression : \n%s", spew.Sdump(expr))
	}
}

func TestParseExprCompare(t *testing.T) {
	frame := newFrame("$i<6 {print \"a\"}")
	expr, _, _ := parseExpr(frame)
	expected := &comparatorNode{"<", &symbolReferenceNode{"i"}, &intLeaf{6}}
	if !reflect.DeepEqual(expr, expected) {
		t.Errorf("unexpected expression : \n%s", spew.Sdump(expr))
	}
}

func TestParseExprString(t *testing.T) {
	frame := newFrame("\"${a}test\"")
	expr, _, err := parseExpr(frame)
	if err != nil {
		t.Errorf("error while parsing : %s", err.Error())
	}
	expected := &formatStringNode{"%vtest", []symbolReferenceNode{{"a"}}}
	if !reflect.DeepEqual(expr, expected) {
		t.Errorf("unexpected expression : \n%s", spew.Sdump(expr))
		t.Errorf("unexpected parsing : \ntree : %s\nexpected : %s",
			spew.Sdump(expr), spew.Sdump(expected))
	}
}

func TestParseExprArrayRef(t *testing.T) {
	frame := newFrame("$ab[42 + 1]")
	expr, _, err := parseExpr(frame)
	if err != nil {
		t.Errorf("error while parsing : %s", err.Error())
	}
	expected := &objReferenceNode{"ab", &arithNode{op: "+", left: &intLeaf{42}, right: &intLeaf{1}}}
	if !reflect.DeepEqual(expr, expected) {
		t.Errorf("unexpected parsing : \ntree : %s\nexpected : %s",
			spew.Sdump(expr), spew.Sdump(expected))
	}
}

func TestParseRawText(t *testing.T) {
	frame := newFrame("${a}a")
	expr, _, err := parseRawText(lexUnquotedString, frame)
	if err != nil {
		t.Errorf("error while parsing : %s", err.Error())
	}
	expected := &formatStringNode{"%va", []symbolReferenceNode{{"a"}}}
	if !reflect.DeepEqual(expr, expected) {
		t.Errorf("unexpected expression : \n%s", spew.Sdump(expr))
	}
}

func TestParseStringExpr(t *testing.T) {
	frame := newFrame("${a}a")
	expr, _, _ := parseStringExpr(frame)
	expected := &formatStringNode{"%va", []symbolReferenceNode{{"a"}}}
	if !reflect.DeepEqual(expr, expected) {
		t.Errorf("unexpected expression : \n%s", spew.Sdump(expr))
	}
}

func TestParseAssign(t *testing.T) {
	frame := newFrame("test= plouf")
	va, nextFrame, err := parseAssign(frame)
	if err != nil {
		t.Errorf("cannot parse assign : %s", err.Error())
	}
	if va != "test" {
		t.Errorf("wrong variable parserd : %s", va)
	}
	if nextFrame.str() != " plouf" {
		t.Errorf("wrong next frame : %s", nextFrame.str())
	}
}

func assertParsing(n node, expected node, t *testing.T) {
	if !reflect.DeepEqual(n, expected) {
		t.Errorf("unexpected parsing : \n\ntree : %s\nexpected : %s",
			spew.Sdump(n), spew.Sdump(expected))
	}
}

func testCommand(buffer string, expected node, t *testing.T) {
	n, err := Parse(buffer)
	if err != nil {
		t.Errorf("cannot parse command : %s", err.Error())
		return
	}
	assertParsing(n, expected, t)
}

func TestParseLsObj(t *testing.T) {
	buffer := "lsbldg -s height - f attr1:attr2 -r plouf.plaf "
	path := &pathNode{&strLeaf{"plouf.plaf"}}
	entity := 2
	recursive := true
	sort := "height"
	attrList := []string{"attr1", "attr2"}
	format := ""
	expected := &lsObjNode{path, entity, recursive, sort, attrList, format}
	testCommand(buffer, expected, t)

	buffer = "lsbldg -s height - f (\"height is %s\", height) -r plouf.plaf "
	attrList = []string{"height"}
	format = "height is %s"
	expected = &lsObjNode{path, entity, recursive, sort, attrList, format}
	testCommand(buffer, expected, t)
}

var testPath = &pathNode{&formatStringNode{"%v/tata", []symbolReferenceNode{{"toto"}}}}
var testPath2 = &pathNode{&strLeaf{"/toto/../tata"}}

func vec2(x float64, y float64) node {
	return &arrNode{[]node{&floatLeaf{x}, &floatLeaf{y}}}
}

func vec3(x float64, y float64, z float64) node {
	return &arrNode{[]node{&floatLeaf{x}, &floatLeaf{y}, &floatLeaf{z}}}
}

func vec4(x float64, y float64, z float64, w float64) node {
	return &arrNode{[]node{&floatLeaf{x}, &floatLeaf{y}, &floatLeaf{z}, &floatLeaf{w}}}
}

var commandsMatching = map[string]node{
	"ls":                             &lsNode{&pathNode{&strLeaf{""}}},
	"get ${toto}/tata":               &getObjectNode{testPath},
	"getu rackA 42":                  &getUNode{&pathNode{&strLeaf{"rackA"}}, &intLeaf{42}},
	"undraw":                         &undrawNode{nil},
	"undraw ${toto}/tata":            &undrawNode{testPath},
	"draw":                           &drawNode{&pathNode{&strLeaf{""}}, 0, false},
	"draw ${toto}/tata":              &drawNode{testPath, 0, false},
	"draw ${toto}/tata 4":            &drawNode{testPath, 4, false},
	"draw -f":                        &drawNode{&pathNode{&strLeaf{""}}, 0, true},
	"draw -f ${toto}/tata":           &drawNode{testPath, 0, true},
	"draw -f ${toto}/tata 4 ":        &drawNode{testPath, 4, true},
	".cmds:../toto/tata.ocli":        &loadNode{&strLeaf{"../toto/tata.ocli"}},
	".template:../toto/tata.ocli":    &loadTemplateNode{&strLeaf{"../toto/tata.ocli"}},
	".var:a=42":                      &assignNode{"a", &intLeaf{42}},
	"=${toto}/tata":                  &selectObjectNode{testPath},
	"=..":                            &selectObjectNode{&pathNode{&strLeaf{".."}}},
	"={${toto}/tata}":                &selectChildrenNode{[]node{testPath}},
	"={${toto}/tata, /toto/../tata}": &selectChildrenNode{[]node{testPath, testPath2}},
	"-${toto}/tata":                  &deleteObjNode{testPath},
	">${toto}/tata":                  &focusNode{testPath},
	"+tenant:${toto}/tata@42ff42":    &createTenantNode{testPath, &strLeaf{"42ff42"}},
	"+tn:${toto}/tata@42ff42":        &createTenantNode{testPath, &strLeaf{"42ff42"}},
	"+site:${toto}/tata":             &createSiteNode{testPath},
	"+si:${toto}/tata":               &createSiteNode{testPath},
	"+building:${toto}/tata@[1., 2.]@3.@[.1, 2., 3.]":      &createBuildingNode{testPath, vec2(1., 2.), &floatLeaf{3.}, vec3(.1, 2., 3.)},
	"+room:${toto}/tata@[1., 2.]@3.@[.1, 2., 3.]@+x-y":     &createRoomNode{testPath, vec2(1., 2.), &floatLeaf{3.}, vec3(.1, 2., 3.), &strLeaf{"+x-y"}, nil, nil},
	"+room:${toto}/tata@[1., 2.]@3.@[.1, 2., 3.]@+x-y@m":   &createRoomNode{testPath, vec2(1., 2.), &floatLeaf{3.}, vec3(.1, 2., 3.), &strLeaf{"+x-y"}, &strLeaf{"m"}, nil},
	"+room:${toto}/tata@[1., 2.]@3.@template":              &createRoomNode{testPath, vec2(1., 2.), &floatLeaf{3.}, nil, nil, nil, &strLeaf{"template"}},
	"+rack:${toto}/tata@[1., 2.]@[.1, 2., 3.]@front":       &createRackNode{testPath, vec2(1., 2.), vec3(.1, 2., 3.), &strLeaf{"front"}},
	"+rack:${toto}/tata@[1., 2.]@template@front":           &createRackNode{testPath, vec2(1., 2.), &strLeaf{"template"}, &strLeaf{"front"}},
	"+device:${toto}/tata@42@42":                           &createDeviceNode{testPath, &intLeaf{42}, &intLeaf{42}, nil},
	"+device:${toto}/tata@42@template":                     &createDeviceNode{testPath, &intLeaf{42}, &strLeaf{"template"}, nil},
	"+device:${toto}/tata@42@template@frontflipped ":       &createDeviceNode{testPath, &intLeaf{42}, &strLeaf{"template"}, &strLeaf{"frontflipped"}},
	"+device:${toto}/tata@slot42@42":                       &createDeviceNode{testPath, &strLeaf{"slot42"}, &intLeaf{42}, nil},
	"+device:${toto}/tata@slot42@template":                 &createDeviceNode{testPath, &strLeaf{"slot42"}, &strLeaf{"template"}, nil},
	"+device:${toto}/tata@slot42@template@frontflipped ":   &createDeviceNode{testPath, &strLeaf{"slot42"}, &strLeaf{"template"}, &strLeaf{"frontflipped"}},
	"+group:${toto}/tata@{c1, c2}":                         &createGroupNode{testPath, []node{&pathNode{&strLeaf{"c1"}}, &pathNode{&strLeaf{"c2"}}}},
	"+corridor:${toto}/tata@{r1, r2}@42.7":                 &createCorridorNode{testPath, &pathNode{&strLeaf{"r1"}}, &pathNode{&strLeaf{"r2"}}, &floatLeaf{42.7}},
	"${toto}/tata:areas=[1., 2., 3., 4.]@[1., 2., 3., 4.]": &updateObjNode{testPath, "areas", []node{vec4(1., 2., 3., 4.), vec4(1., 2., 3., 4.)}, false},
	"${toto}/tata:separator=[1., 2.]@[1., 2.]@wireframe":   &updateObjNode{testPath, "separator", []node{vec2(1., 2.), vec2(1., 2.), &strLeaf{"wireframe"}}, false},
	"${toto}/tata:attr=42":                                 &updateObjNode{testPath, "attr", []node{&intLeaf{42}}, false},
	"${toto}/tata:label=\"plouf\"":                         &updateObjNode{testPath, "label", []node{&strLeaf{"plouf"}}, false},
	"${toto}/tata:labelFont=bold":                          &updateObjNode{testPath, "labelFont", []node{&strLeaf{"bold"}}, false},
	"${toto}/tata:labelFont=color@42ff42":                  &updateObjNode{testPath, "labelFont", []node{&strLeaf{"color"}, &strLeaf{"42ff42"}}, false},
	"${toto}/tata:tilesName=true":                          &updateObjNode{testPath, "tilesName", []node{&boolLeaf{true}}, false},
	"${toto}/tata:tilesColor=false":                        &updateObjNode{testPath, "tilesColor", []node{&boolLeaf{false}}, false},
	"${toto}/tata:U=false":                                 &updateObjNode{testPath, "U", []node{&boolLeaf{false}}, false},
	"${toto}/tata:slots=false":                             &updateObjNode{testPath, "slots", []node{&boolLeaf{false}}, false},
	"${toto}/tata:localCS=false":                           &updateObjNode{testPath, "localCS", []node{&boolLeaf{false}}, false},
	"${toto}/tata:content=false":                           &updateObjNode{testPath, "content", []node{&boolLeaf{false}}, false},
	"ui.delay=15":                                          &uiDelayNode{15.},
	"ui.infos=true":                                        &uiToggleNode{"infos", true},
	"ui.debug=false":                                       &uiToggleNode{"debug", false},
	"ui.highlight=${toto}/tata":                            &uiHighlightNode{testPath},
	"ui.hl=${toto}/tata":                                   &uiHighlightNode{testPath},
	"camera.move=[1., 2., 3.]@[1., 2.]":                    &cameraMoveNode{"move", vec3(1., 2., 3.), vec2(1., 2.)},
	"camera.translate=[1., 2., 3.]@[1., 2.]":               &cameraMoveNode{"translate", vec3(1., 2., 3.), vec2(1., 2.)},
	"camera.wait=15":                                       &cameraWaitNode{15.},
	"clear":                                                &clrNode{},
	".cmds:${CUST}/DEMO.PERF.ocli":                         &loadNode{&formatStringNode{"%v/DEMO.PERF.ocli", []symbolReferenceNode{{"CUST"}}}},
	".cmds:${a}/${b}.ocli":                                 &loadNode{&formatStringNode{"%v/%v.ocli", []symbolReferenceNode{{"a"}, {"b"}}}},
	"while $i<6 {print \"a\"}":                             &whileNode{&comparatorNode{"<", &symbolReferenceNode{"i"}, &intLeaf{6}}, &printNode{&strLeaf{"a"}}},
}

func TestSimpleCommands(t *testing.T) {
	for command, tree := range commandsMatching {
		testCommand(command, tree, t)
	}
}

func TestParseUpdate(t *testing.T) {
	buffer := "coucou.plouf : attr = #val1 @ val2"
	expected := &updateObjNode{
		&pathNode{&strLeaf{"coucou.plouf"}},
		"attr",
		[]node{&strLeaf{"val1"}, &strLeaf{"val2"}},
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
		expected := &forRangeNode{"i", &intLeaf{0}, &intLeaf{42}, tree}
		testCommand(command, expected, t)
	}
}

func TestIf(t *testing.T) {
	for simpleCommand, tree := range commandsMatching {
		frame := newFrame("true { " + simpleCommand + " }e")
		expected := &ifNode{&boolLeaf{true}, tree, nil}
		result, _, err := parseIf(frame)
		if err != nil {
			t.Errorf("error parsing if frame : %s", err.Error())
			return
		}
		assertParsing(result, expected, t)
	}
}
