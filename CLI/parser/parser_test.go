package parser

import (
	"cli/models"
	"reflect"
	"runtime/debug"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
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
	tests := []struct {
		name      string
		word      string
		remaining string
	}{
		{"ParseWordSingleLetter", "a", "42"},
		{"ParseWordMultipleLetters", "test", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer recoverFunc(t)
			p := newParser(tt.word + " " + tt.remaining)
			word := p.parseSimpleWord("")
			if word != tt.word {
				t.Errorf("wrong word parsed")
			}
			if p.remaining() != tt.remaining {
				t.Errorf("wrong stop, remaining buf : %s", p.remaining())
			}
		})
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
	expr := p.parseText(p.parseUnquotedStringToken, false, false)
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
	buffer := "lsbuilding -s height -a attr1:attr2 plouf.plaf attr1=a, attr2=b"
	path := &pathNode{path: &valueNode{"plouf.plaf"}}
	sort := "height"
	attrList := []string{"attr1", "attr2"}
	filters := map[string]node{
		"category": &valueNode{"building"},
		"attr1":    &valueNode{"a"},
		"attr2":    &valueNode{"b"},
	}
	expected := &lsNode{path: path, filters: filters, sortAttr: sort, attrList: attrList}
	testCommand(buffer, expected, t)
	buffer = "lsbuilding -s height -a \"attr1:attr2\" plouf.plaf attr1=a, attr2=b"
	testCommand(buffer, expected, t)
}

func TestParseLsRecursive(t *testing.T) {
	path := &pathNode{path: &valueNode{"#test"}}
	expected := &lsNode{path: path, filters: map[string]node{}, recursive: recursiveArgs{isRecursive: true}}
	buffer := "ls -r #test"
	testCommand(buffer, expected, t)

	expected = &lsNode{path: path, filters: map[string]node{}, recursive: recursiveArgs{isRecursive: true, minDepth: "1"}}
	buffer = "ls -r -m 1 #test"
	testCommand(buffer, expected, t)

	expected = &lsNode{path: path, filters: map[string]node{}, recursive: recursiveArgs{isRecursive: true, minDepth: "1", maxDepth: "2"}}
	buffer = "ls -r -m 1 -M 2 #test"
	testCommand(buffer, expected, t)

	expected = &lsNode{path: path, filters: map[string]node{}, recursive: recursiveArgs{isRecursive: false, minDepth: "1", maxDepth: "2"}}
	buffer = "ls -m 1 -M 2 #test"
	testCommand(buffer, expected, t)
}

func TestParseLsComplexFilter(t *testing.T) {
	path := &pathNode{path: &valueNode{"plouf.plaf"}}
	buffer := "ls plouf.plaf -f category=building, attr1=a, attr2=b"
	filters := map[string]node{
		"filter": &valueNode{"(category=building) & ((attr1=a) & (attr2=b))"},
	}
	expected := &lsNode{path: path, filters: filters}
	testCommand(buffer, expected, t)
	buffer = "ls plouf.plaf -f category=building & (attr1!=a | attr2>5)"
	filters = map[string]node{
		"filter": &valueNode{"category=building & (attr1!=a | attr2>5)"},
	}
	expected = &lsNode{path: path, filters: filters}
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
	"man":                               &helpNode{""},
	"man draw":                          &helpNode{"draw"},
	"man camera":                        &helpNode{"camera"},
	"man ui":                            &helpNode{"ui"},
	"ls":                                &lsNode{path: &pathNode{path: &valueNode{""}}, filters: map[string]node{}},
	"cd":                                &cdNode{&pathNode{path: &valueNode{"/"}}},
	"tree":                              &treeNode{&pathNode{path: &valueNode{"."}}, 1},
	"undraw":                            &undrawNode{nil},
	"undraw ${toto}/tata":               &undrawNode{testPath},
	"draw":                              &drawNode{&pathNode{path: &valueNode{""}}, 0, false},
	"draw ${toto}/tata":                 &drawNode{testPath, 0, false},
	"draw ${toto}/tata 4":               &drawNode{testPath, 4, false},
	"draw -f":                           &drawNode{&pathNode{path: &valueNode{""}}, 0, true},
	"draw -f ${toto}/tata":              &drawNode{testPath, 0, true},
	"draw -f ${toto}/tata 4 ":           &drawNode{testPath, 4, true},
	".cmds:../toto/tata.ocli":           &loadNode{&valueNode{"../toto/tata.ocli"}},
	".template:../toto/tata.ocli":       &loadTemplateNode{&valueNode{"../toto/tata.ocli"}},
	".var:a=42":                         &assignNode{"a", &valueNode{"42"}},
	".var:b= $(($a+3))":                 &assignNode{"b", &formatStringNode{&valueNode{"%v"}, []node{&arithNode{"+", &symbolReferenceNode{"a"}, &valueNode{3}}}}},
	"=${toto}/tata":                     &selectObjectNode{testPath},
	"=..":                               &selectObjectNode{&pathNode{path: &valueNode{".."}}},
	"={${toto}/tata}":                   &selectChildrenNode{[]node{testPath}},
	"={${toto}/tata, /toto/../tata}":    &selectChildrenNode{[]node{testPath, testPath2}},
	"-${toto}/tata":                     &deleteObjNode{testPath},
	">${toto}/tata":                     &focusNode{testPath},
	"+site:${toto}/tata":                &createSiteNode{testPath},
	"+si:${toto}/tata":                  &createSiteNode{testPath},
	"getu rackA 42":                     &getUNode{&pathNode{path: &valueNode{"rackA"}}, &valueNode{42}},
	"get ${toto}/tata":                  &getObjectNode{path: testPath, filters: map[string]node{}},
	"get -r ${toto}/tata":               &getObjectNode{path: testPath, filters: map[string]node{}, recursive: recursiveArgs{isRecursive: true}},
	"get -r ${toto}/tata category=room": &getObjectNode{path: testPath, filters: map[string]node{"category": &valueNode{"room"}}, recursive: recursiveArgs{isRecursive: true}},
	"get ${toto}/tata -f category=room, name=R1":                   &getObjectNode{path: testPath, filters: map[string]node{"filter": &valueNode{"(category=room) & (name=R1)"}}},
	"get ${toto}/tata -f category=room & (name!=R1 | height>5)":    &getObjectNode{path: testPath, filters: map[string]node{"filter": &valueNode{"category=room & (name!=R1 | height>5)"}}},
	"+building:${toto}/tata@[1., 2.]@3.@[.1, 2., 3.]":              &createBuildingNode{testPath, vec2(1., 2.), &valueNode{3.}, vec3(.1, 2., 3.)},
	"+room:${toto}/tata@[1., 2.]@3.@[.1, 2., 3.]@+x-y":             &createRoomNode{testPath, vec2(1., 2.), &valueNode{3.}, vec3(.1, 2., 3.), &valueNode{"+x-y"}, nil, nil},
	"+room:${toto}/tata@[1., 2.]@3.@[.1, 2., 3.]@+x-y@m":           &createRoomNode{testPath, vec2(1., 2.), &valueNode{3.}, vec3(.1, 2., 3.), &valueNode{"+x-y"}, &valueNode{"m"}, nil},
	"+room:${toto}/tata@[1., 2.]@3.@template":                      &createRoomNode{testPath, vec2(1., 2.), &valueNode{3.}, nil, nil, nil, &valueNode{"template"}},
	"+rack:${toto}/tata@[1., 2.]@t@front@[.1, 2., 3.]":             &createRackNode{testPath, vec2(1., 2.), &valueNode{"t"}, &valueNode{"front"}, vec3(.1, 2., 3.)},
	"+rack:${toto}/tata@[1., 2.]@m@front@template":                 &createRackNode{testPath, vec2(1., 2.), &valueNode{"m"}, &valueNode{"front"}, &valueNode{"template"}},
	"+rack:${toto}/tata@[1., 2.]@m@[.1, 2., 3.]@template":          &createRackNode{testPath, vec2(1., 2.), &valueNode{"m"}, vec3(.1, 2., 3.), &valueNode{"template"}},
	"+generic:${toto}/tata@[1., 2.]@t@front@[.1, 2., 3.]@cube@box": &createGenericNode{testPath, vec2(1., 2.), &valueNode{"t"}, &valueNode{"front"}, vec3(.1, 2., 3.), &valueNode{"cube"}, &valueNode{"box"}},
	"+generic:${toto}/tata@[1., 2.]@m@front@template":              &createGenericNode{testPath, vec2(1., 2.), &valueNode{"m"}, &valueNode{"front"}, &valueNode{"template"}, nil, nil},
	"+generic:${toto}/tata@[1., 2.]@m@[.1, 2., 3.]@template":       &createGenericNode{testPath, vec2(1., 2.), &valueNode{"m"}, vec3(.1, 2., 3.), &valueNode{"template"}, nil, nil},
	"+device:${toto}/tata@42@42":                                   &createDeviceNode{testPath, []node{&valueNode{"42"}}, &valueNode{"42"}, false, nil},
	"+device:${toto}/tata@42@template":                             &createDeviceNode{testPath, []node{&valueNode{"42"}}, &valueNode{"template"}, false, nil},
	"+device:${toto}/tata@42@template@true@frontflipped ":          &createDeviceNode{testPath, []node{&valueNode{"42"}}, &valueNode{"template"}, true, &valueNode{"frontflipped"}},
	"+device:${toto}/tata@[slot42]@42@true":                        &createDeviceNode{testPath, []node{&valueNode{"slot42"}}, &valueNode{"42"}, true, nil},
	"+device:${toto}/tata@[slot42,slot43]@template":                &createDeviceNode{testPath, []node{&valueNode{"slot42"}, &valueNode{"slot43"}}, &valueNode{"template"}, false, nil},
	"+device:${toto}/tata@[slot42]@template@false@frontflipped ":   &createDeviceNode{testPath, []node{&valueNode{"slot42"}}, &valueNode{"template"}, false, &valueNode{"frontflipped"}},
	"+vobj:${toto}/tata@vm ":                                       &createVirtualNode{testPath, &valueNode{"vm"}, nil, nil},
	"+vobj:${toto}/tata@vm@[myvlink,onemore]@proxmox ":             &createVirtualNode{testPath, &valueNode{"vm"}, []node{&valueNode{"myvlink"}, &valueNode{"onemore"}}, &valueNode{"proxmox"}},
	"+group:${toto}/tata@{c1, c2}":                                 &createGroupNode{testPath, []node{&pathNode{path: &valueNode{"c1"}}, &pathNode{path: &valueNode{"c2"}}}},
	"+corridor:${toto}/tata@[1., 2.]@t@front@[.1, 2., 3.]@cold":    &createCorridorNode{testPath, vec2(1., 2.), &valueNode{"t"}, &valueNode{"front"}, vec3(.1, 2., 3.), &valueNode{"cold"}},
	"${toto}/tata:areas=[1., 2., 3., 4.]@[1., 2., 3., 4.]":         &updateObjNode{testPathUpdate, "areas", []node{vec4(1., 2., 3., 4.), vec4(1., 2., 3., 4.)}, false},
	"${toto}/tata:separators+=name@[1., 2.]@[1., 2.]@wireframe":    &updateObjNode{testPathUpdate, "separators+", []node{&valueNode{"name"}, vec2(1., 2.), vec2(1., 2.), &valueNode{"wireframe"}}, false},
	"${toto}/tata:separators-=name":                                &updateObjNode{testPathUpdate, "separators-", []node{&valueNode{"name"}}, false},
	"${toto}/tata:pillars+=name@[1., 2.]@[1., 2.]@2.5":             &updateObjNode{testPathUpdate, "pillars+", []node{&valueNode{"name"}, vec2(1., 2.), vec2(1., 2.), &valueNode{"2.5"}}, false},
	"${toto}/tata:pillars-=name":                                   &updateObjNode{testPathUpdate, "pillars-", []node{&valueNode{"name"}}, false},
	"${toto}/tata:vlinks+=name":                                    &updateObjNode{testPathUpdate, "vlinks+", []node{&valueNode{"name"}}, false},
	"${toto}/tata:vlinks-=name":                                    &updateObjNode{testPathUpdate, "vlinks-", []node{&valueNode{"name"}}, false},
	"${toto}/tata:attr=42":                                         &updateObjNode{testPathUpdate, "attr", []node{&valueNode{"42"}}, false},
	"${toto}/tata:label=\"plouf\"":                                 &updateObjNode{testPathUpdate, "label", []node{&valueNode{"plouf"}}, false},
	"${toto}/tata:labelFont=bold":                                  &updateObjNode{testPathUpdate, "labelFont", []node{&valueNode{"bold"}}, false},
	"${toto}/tata:labelFont=color@42ff42":                          &updateObjNode{testPathUpdate, "labelFont", []node{&valueNode{"color"}, &valueNode{"42ff42"}}, false},
	"${toto}/tata:tilesName=true":                                  &updateObjNode{testPathUpdate, "tilesName", []node{&valueNode{"true"}}, false},
	"${toto}/tata:tilesColor=false":                                &updateObjNode{testPathUpdate, "tilesColor", []node{&valueNode{"false"}}, false},
	"${toto}/tata:U=false":                                         &updateObjNode{testPathUpdate, "U", []node{&valueNode{"false"}}, false},
	"${toto}/tata:slots=false":                                     &updateObjNode{testPathUpdate, "slots", []node{&valueNode{"false"}}, false},
	"${toto}/tata:localCS=false":                                   &updateObjNode{testPathUpdate, "localCS", []node{&valueNode{"false"}}, false},
	"${toto}/tata:displayContent=false":                            &updateObjNode{testPathUpdate, "displayContent", []node{&valueNode{"false"}}, false},
	"${toto}/tata:temperature_01-Inlet-Ambient=7":                  &updateObjNode{testPathUpdate, "temperature_01-Inlet-Ambient", []node{&valueNode{"7"}}, false},
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
	ifBody := &ast{[]node{&lsNode{path: &pathNode{path: &valueNode{""}}, filters: map[string]node{}}, nil}}
	elifBody := &ast{[]node{&treeNode{&pathNode{path: &valueNode{"."}}, 1}, nil}}
	elseBody := &ast{[]node{&pwdNode{}, nil}}
	elif := &ifNode{conditionElif, elifBody, elseBody}
	expected := &ifNode{condition, ifBody, elif}
	testCommand(command, expected, t)
}

func TestParseUrl(t *testing.T) {
	url := "http://url.com/route"
	p := newParser(url + " other")
	parsedUrl := p.parseUrl("url")
	assert.Equal(t, url, parsedUrl)
}

func parserRecoverFunction(t *testing.T, p *parser, expectedErrorMessage string) {
	if panicInfo := recover(); panicInfo != nil {
		assert.Equal(t, expectedErrorMessage, p.err)
	} else {
		t.Errorf("The function should have ended with an error")
	}
}

func TestParseIntError(t *testing.T) {
	p := newParser("2s")
	defer parserRecoverFunction(t, p, "integer expected")
	p.parseInt("integer")
}

func TestParseFloat(t *testing.T) {
	p := newParser("2 2.5 2.g")
	defer parserRecoverFunction(t, p, "float expected")

	parsedFloat := p.parseFloat("float")
	assert.Equal(t, 2.0, parsedFloat)

	parsedFloat = p.parseFloat("float")
	assert.Equal(t, 2.5, parsedFloat)

	p.parseFloat("float")
}

func TestParseBoolError(t *testing.T) {
	p := newParser("tru")
	defer parserRecoverFunction(t, p, "boolean expected")
	p.parseBool()
}

func TestParseIndexing(t *testing.T) {
	p := newParser("[12]")
	parsedNode := p.parseIndexing().(*valueNode)
	assert.Equal(t, 12, parsedNode.val)
}

func TestParseEnv(t *testing.T) {
	p := newParser("var=12")
	parsedNode := p.parseEnv().(*setEnvNode)
	assert.Equal(t, "var", parsedNode.arg)
	assert.Equal(t, 12, parsedNode.expr.(*valueNode).val)
}

func TestParseLink(t *testing.T) {
	sourcePath := models.StrayPath + "stray-device"
	destinationPath := models.PhysicalPath + "site/building/room/rack"
	p := newParser(sourcePath + "@" + destinationPath)
	parsedNode := p.parseLink().(*linkObjectNode)
	assert.Equal(t, sourcePath, parsedNode.source.(*pathNode).path.(*valueNode).val)
	assert.Equal(t, destinationPath, parsedNode.destination.(*pathNode).path.(*valueNode).val)

	p = newParser(sourcePath + "@" + destinationPath + "@slot=[slot1,slot2]@orientation=front")
	parsedNode = p.parseLink().(*linkObjectNode)
	assert.Equal(t, sourcePath, parsedNode.source.(*pathNode).path.(*valueNode).val)
	assert.Equal(t, destinationPath, parsedNode.destination.(*pathNode).path.(*valueNode).val)
	assert.Equal(t, []string{"orientation"}, parsedNode.attrs)
	assert.Len(t, parsedNode.values, 1)
	assert.Equal(t, "front", parsedNode.values[0].(*valueNode).val)

	assert.Len(t, parsedNode.slots, 2)
	assert.Equal(t, "slot1", parsedNode.slots[0].(*valueNode).val)
	assert.Equal(t, "slot2", parsedNode.slots[1].(*valueNode).val)
}

func TestParseUnlink(t *testing.T) {
	path := models.PhysicalPath + "site/building/room/rack"
	p := newParser(path)
	parsedNode := p.parseUnlink().(*unlinkObjectNode)
	assert.Equal(t, path, parsedNode.source.(*pathNode).path.(*valueNode).val)
}

func TestParseAlias(t *testing.T) {
	p := newParser("aliasName { print $i }")
	parsedNode := p.parseAlias().(*funcDefNode)
	assert.Equal(t, "aliasName", parsedNode.name)
	assert.Equal(t, "%v", parsedNode.body.(*printNode).expr.(*formatStringNode).str.(*valueNode).val)
	assert.Len(t, parsedNode.body.(*printNode).expr.(*formatStringNode).vals, 1)
	assert.Equal(t, "i", parsedNode.body.(*printNode).expr.(*formatStringNode).vals[0].(*symbolReferenceNode).va)
}

func TestParseCreateDomain(t *testing.T) {
	p := newParser("domain@00000A")
	parsedNode := p.parseCreateDomain().(*createDomainNode)
	assert.Equal(t, "domain", parsedNode.path.(*pathNode).path.(*valueNode).val)
	assert.Equal(t, "00000A", parsedNode.color.(*valueNode).val)
}

func TestParseCreateTag(t *testing.T) {
	p := newParser("tag@00000A")
	parsedNode := p.parseCreateTag().(*createTagNode)
	assert.Equal(t, "tag", parsedNode.slug.(*valueNode).val)
	assert.Equal(t, "00000A", parsedNode.color.(*valueNode).val)
}

func TestParseCreateLayer(t *testing.T) {
	p := newParser("layer@site.building.room@category=rack")
	parsedNode := p.parseCreateLayer().(*createLayerNode)
	assert.Equal(t, "layer", parsedNode.slug.(*valueNode).val)
	assert.Equal(t, "site.building.room", parsedNode.applicability.(*pathNode).path.(*valueNode).val)
	assert.Equal(t, "category=rack", parsedNode.filterValue.(*valueNode).val)
}

func TestParseCreateOrphan(t *testing.T) {
	path := models.StrayPath + "orphan"
	templateName := "my-template"
	p := newParser("device : " + path + "@" + templateName)
	parsedNode := p.parseCreateOrphan().(*createOrphanNode)
	assert.Equal(t, path, parsedNode.path.(*pathNode).path.(*valueNode).val)
	assert.Equal(t, templateName, parsedNode.template.(*valueNode).val)
}

func TestParseCreateUser(t *testing.T) {
	email := "email@mail.com"
	role := "my-role"
	domain := "my-domain"
	p := newParser(`"` + email + `"` + "@" + role + "@" + domain)
	parsedNode := p.parseCreateUser().(*createUserNode)
	assert.Equal(t, email, parsedNode.email.(*valueNode).val)
	assert.Equal(t, role, parsedNode.role.(*valueNode).val)
	assert.Equal(t, domain, parsedNode.domain.(*valueNode).val)
}

func TestParseAddRole(t *testing.T) {
	email := "email@mail.com"
	role := "my-role-2"
	domain := "my-domain"
	p := newParser(`"` + email + `"` + "@" + role + "@" + domain)
	parsedNode := p.parseAddRole().(*addRoleNode)
	assert.Equal(t, email, parsedNode.email.(*valueNode).val)
	assert.Equal(t, role, parsedNode.role.(*valueNode).val)
	assert.Equal(t, domain, parsedNode.domain.(*valueNode).val)
}

func TestParseCp(t *testing.T) {
	source := models.LayersPath + "layer1"
	destination := "layer2"
	p := newParser(source + " " + destination)
	parsedNode := p.parseCp().(*cpNode)
	assert.Equal(t, source, parsedNode.source.(*pathNode).path.(*valueNode).val)
	assert.Equal(t, destination, parsedNode.dest.(*valueNode).val)
}

func TestParseExprList(t *testing.T) {
	p := newParser("-1")
	parsedNode := p.parseUnaryExpr().(*negateNode)
	assert.Equal(t, 1, parsedNode.val.(*valueNode).val)

	p = newParser("!true")
	parsedNode2 := p.parseUnaryExpr().(*negateBoolNode)
	assert.Equal(t, true, parsedNode2.expr.(*valueNode).val)

	p = newParser("+1")
	parsedNode3 := p.parseUnaryExpr().(*valueNode)
	assert.Equal(t, 1, parsedNode3.val)
}

func TestParseLsStarError(t *testing.T) {
	p := newParser("-r /*")
	defer parserRecoverFunction(t, p, "unexpected character in path: '*'")
	p.parseLs("")
}

func TestParseLsPathError(t *testing.T) {
	p := newParser("-r path/$ra")
	defer parserRecoverFunction(t, p, "path expected")
	p.parseLs("")
}

func TestParseDrawable(t *testing.T) {
	path := "/path/to/draw"
	p := newParser(path)
	parsedNode := p.parseDrawable().(*isEntityDrawableNode)
	assert.Equal(t, path, parsedNode.path.(*pathNode).path.(*valueNode).val)

	attribute := "color"
	p = newParser(path + " " + attribute)
	parsedNodes := p.parseDrawable().(*isAttrDrawableNode)
	assert.Equal(t, path, parsedNodes.path.(*pathNode).path.(*valueNode).val)
	assert.Equal(t, attribute, parsedNodes.attr)
}

func TestParseUnsetVariable(t *testing.T) {
	varName := "myVar"
	p := newParser("-v " + varName)
	parsedNode := p.parseUnset().(*unsetVarNode)
	assert.Equal(t, varName, parsedNode.varName)
}

func TestParseUnsetFunction(t *testing.T) {
	functionName := "myFunction"
	p := newParser("-f " + functionName)
	parsedNode := p.parseUnset().(*unsetFuncNode)
	assert.Equal(t, functionName, parsedNode.funcName)
}

func TestParseDeleteAttribute(t *testing.T) {
	path := "path/to/room"
	attribute := "template"
	p := newParser(path + ":" + attribute)
	parsedNode := p.parseDelete().(*deleteAttrNode)
	assert.Equal(t, path, parsedNode.path.(*pathNode).path.(*valueNode).val)
	assert.Equal(t, attribute, parsedNode.attr)
}

func TestParseTree(t *testing.T) {
	path := "/path"
	p := newParser(path)
	parsedNode := p.parseTree().(*treeNode)
	assert.Equal(t, path, parsedNode.path.(*pathNode).path.(*valueNode).val)
	assert.Equal(t, 1, parsedNode.depth)

	p = newParser(path + " 3")
	parsedNode = p.parseTree().(*treeNode)
	assert.Equal(t, path, parsedNode.path.(*pathNode).path.(*valueNode).val)
	assert.Equal(t, 3, parsedNode.depth)
}

func TestParseConnect3D(t *testing.T) {
	url := "url.com/path"
	p := newParser(url)
	parsedNode := p.parseConnect3D().(*connect3DNode)
	assert.Equal(t, url, parsedNode.url)
}
