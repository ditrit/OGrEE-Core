%{
package main
import (
cmd "cli/controllers"
"path/filepath"
l "cli/logger"
"strings"
"strconv"
)

var root node 
var _ = l.GetInfoLogger() //Suppresses annoying Dockerfile build error

//Since the CFG will only execute rules
//when production is fully met.
//We need to catch values of array as they are coming,
//otherwise, only the last elt will be captured.
//The best way here is to catch array of strings
//then return array of maps
func retNodeArray(input []interface{}) []map[int]interface{}{
       res := []map[int]interface{}{}
       for idx := range input {
              switch input[idx].(type) {
                     case string:
                     if input[idx].(string) == "false" {
                            x := map[int]interface{}{0: &boolNode{false}}
                            res = append(res, x)
                     }
                     if input[idx].(string) == "true" {
                            x := map[int]interface{}{0: &boolNode{true}}
                            res = append(res, x)
                     } else {
                            x := map[int]interface{}{0: &strNode{input[idx].(string)}}
                            res = append(res, x)
                     }
                     case int:
                     x := map[int]interface{}{0: &numNode{input[idx].(int)}}
                     res = append(res, x)

                     default: //Most likely a node
                     res = append(res, map[int]interface{}{0:input[idx]})
              }
              /*if input[idx].(string) == "false" {
                     x := map[int]interface{}{0: &boolNode{false}}
                     res = append(res, x)
              } else if input[idx].(string) == "true" {
                     x := map[int]interface{}{0: &boolNode{true}}
                     res = append(res, x)
              } else if v,e := strconv.Atoi(input[idx].(string)); e == nil {
                     x := map[int]interface{}{0: &numNode{v}}
                     res = append(res, x)
              } else {
                     x := map[int]interface{}{0: &strNode{input[idx].(string)}}
                     res = append(res, x)
              }*/
       }
       return res
}

func resMap(x *string, ent string, isUpdate bool) map[string]interface{} {
	resarr := strings.Split(*x, "=")
	res := make(map[string]interface{})
	attrs := make(map[string]string)

	for i := 0; i+1 < len(resarr); {
		if isUpdate == true {
			res[resarr[i]] = resarr[i+1]
		} else if i+1 < len(resarr) {
			switch ent {
			case "sensor", "group":
				switch resarr[i] {
				case "id", "name", "category", "parentID",
					"description", "domain", "type",
					"parentid", "parentId":
					res[resarr[i]] = resarr[i+1]

				default:
					attrs[resarr[i]] = resarr[i+1]
				}
			case "room_template":
				switch resarr[i] {
				case "id", "slug", "orientation", "separators",
					"tiles", "colors", "rows", "sizeWDHm",
					"technicalArea", "reservedArea":
					res[resarr[i]] = resarr[i+1]

				default:
					attrs[resarr[i]] = resarr[i+1]
				}
			case "obj_template":
				switch resarr[i] {
				case "id", "slug", "description", "category",
					"slots", "colors", "components", "sizeWDHmm",
					"fbxModel":
					res[resarr[i]] = resarr[i+1]

				default:
					attrs[resarr[i]] = resarr[i+1]
				}

			default:
				switch resarr[i] {
				case "id", "name", "category", "parentID",
					"description", "domain", "parentid", "parentId":
					res[resarr[i]] = resarr[i+1]

				default:
					attrs[resarr[i]] = resarr[i+1]
				}

			}
		}
		i += 2
	}
       if len(attrs) > 0 {
              res["attributes"] = attrs
       }
	

	return res
}

func replaceOCLICurrPath(x string) string {
       return strings.Replace(x, "_/", cmd.State.CurrPath+"/", 1)
}


func resolveReference(ref string) string {
       /*Probably code to reference SymbolTable and return data*/
       idx := dynamicMap[ref];
       item := dynamicSymbolTable[idx];
       switch item.(type) {
              case bool:
                 dCatchNodePtr=&boolNode{item.(bool)}
                 if item.(bool) == false {return "false"} else { return "true"}
              case string:
                 dCatchNodePtr=&strNode{item.(string)}
                 return item.(string)
              case int:
                 dCatchNodePtr=&numNode{item.(int)}
                 return strconv.Itoa(item.(int))
              /*case map[string]interface{}:
                 //dCatchNodePtr=&symbolReferenceNode{}
                 return item.(map[string]interface{})[subIdx].(string)*/
              default:
                     println("Unable to deref your variable ")
                     return ""
                  }
}

func formActualPath(x string) string {
       if x == "" || x == "." {
		return cmd.State.CurrPath
	} else if string(x[0]) == "/" {
		return x

	} else {
		return cmd.State.CurrPath + "/" + x
	}
}

//This func is for distinguishing template from sizeU
//in the OCLI syntax for creating devices 
//refer to: 
//https://github.com/ditrit/OGrEE-3D/wiki/CLI-langage#Create-a-Device
func checkIfTemplate(x interface{}) bool {
       if s, ok := x.(string); ok {
              if m, _ := cmd.GetObject("/Logical/ObjectTemplates/"+s, true); m != nil {
                     return true
              }
       }
       
       return false
}

%}

%union {
  n int
  s string
  f float64
  sarr []string
  ast *ast
  node node
  nodeArr []node
  elifArr []elifNode
  arr []interface{}
  mapArr []map[int]interface{}
}

%token <n> TOK_NUM
%token <f> TOK_FLOAT
%token <s> TOK_WORD TOK_TENANT TOK_SITE TOK_BLDG TOK_ROOM
%token <s> TOK_RACK TOK_DEVICE TOK_STR
%token <s> TOK_CORIDOR TOK_GROUP TOK_WALL
%token <s> TOK_AC TOK_CABINET TOK_PANEL TOK_ROW
%token <s> TOK_TILE TOK_SENSOR
%token <s> TOK_ROOM_TMPL TOK_OBJ_TMPL
%token <s> TOK_PLUS TOK_OCDEL TOK_BOOL
%token
       TOK_CREATE TOK_GET TOK_UPDATE TOK_DELETE TOK_SEARCH
       TOK_EQUAL TOK_CMDFLAG TOK_SLASH 
       TOK_EXIT TOK_DOC TOK_CD TOK_PWD
       TOK_CLR TOK_GREP TOK_LS TOK_TREE
       TOK_LSOG TOK_LSTEN TOK_LSSITE TOK_LSBLDG TOK_LSROW
       TOK_LSTILE TOK_LSCAB TOK_LSSENSOR TOK_LSAC TOK_LSPANEL
       TOK_LSWALL TOK_LSCORRIDOR
       TOK_LSROOM TOK_LSRACK TOK_LSDEV
       TOK_ATTRSPEC
       TOK_COL TOK_SELECT TOK_LBRAC TOK_RBRAC
       TOK_COMMA TOK_DOT TOK_CMDS TOK_TEMPLATE TOK_VAR TOK_DEREF
       TOK_SEMICOL TOK_IF TOK_FOR TOK_WHILE
       TOK_ELSE TOK_LBLOCK TOK_RBLOCK
       TOK_LPAREN TOK_RPAREN TOK_OR TOK_AND TOK_IN TOK_PRNT TOK_QUOT
       TOK_NOT TOK_MULT TOK_GREATER TOK_LESS TOK_THEN TOK_FI TOK_DONE
       TOK_MOD
       TOK_UNSET TOK_ELIF TOK_DO TOK_LEN
       TOK_USE_JSON TOK_PARTIAL TOK_LINK TOK_UNLINK
       TOK_CAM TOK_UI TOK_HIERARCH TOK_DRAWABLE TOK_ENV TOK_ORPH
       TOK_DRAW TOK_SETENV
       
%type <n> LSOBJ_COMMAND
%type <s> F OBJ_TYPE P P1 WORDORNUM CDORFG NTORIENTATION COMMAND
%type <arr> WNARG NODEGETTER NODEACC
%type <sarr> GETOBJS
%type <elifArr> EIF
%type <node> OCSEL OCLISYNTX OCDEL OCGET NT_CREATE NT_GET NT_DEL 
%type <node> NT_UPDATE K Q BASH OCUPDATE OCCHOOSE OCCR OCDOT
%type <node> EXPR REL CTRL nex factor unary EQAL term
%type <node> stmnt JOIN
%type <node> st2 FUNC HANDLEUI HANDLELINKS
%left TOK_MULT TOK_OCDEL TOK_SLASH TOK_PLUS 
%right TOK_EQUAL


%%

start: st2 {root = $1}

st2: stmnt {$$=&ast{[]node{$1} }}
       |stmnt TOK_SEMICOL st2 {$$=&ast{[]node{$1, $3}}}
       |CTRL {$$=&ast{[]node{$1}}}
;

stmnt: K {$$=$1}
       |Q {$$=$1}
       |OCLISYNTX {$$=$1}
       |FUNC {$$=$1}
       |{$$=nil}
;

CTRL:    TOK_IF TOK_LBLOCK EXPR TOK_RBLOCK TOK_THEN st2 TOK_FI {$$=&ifNode{$3, $6, nil, nil}}
              |TOK_IF TOK_LBLOCK EXPR TOK_RBLOCK TOK_THEN st2 EIF TOK_ELSE st2 TOK_FI {$$=&ifNode{$3, $6, $9, $7}}
              |TOK_WHILE TOK_LPAREN EXPR TOK_RPAREN st2 TOK_DONE {$$=&whileNode{$3, $5}}
              |TOK_FOR TOK_LPAREN TOK_LPAREN TOK_WORD TOK_EQUAL WORDORNUM TOK_SEMICOL EXPR TOK_SEMICOL stmnt TOK_RPAREN TOK_RPAREN TOK_SEMICOL st2 TOK_DONE 
              {initnd:=&assignNode{$4, dCatchNodePtr};$$=&forNode{initnd,$8,$10,$14}}
              |TOK_FOR TOK_WORD TOK_IN EXPR TOK_SEMICOL st2 TOK_DONE 
              {var incr *arithNode; var incrAssign *assignNode; 
              n1:=&numNode{0};
              
              initd:=&assignNode{$2, n1}; 
              iter:=&symbolReferenceNode{$2, &numNode{0}, nil}; 
              cmp:=&comparatorNode{"<", iter, $4}
              incr=&arithNode{"+", iter, &numNode{1}}
              incrAssign=&assignNode{iter,incr}
              $$=&forNode{initd, cmp, incrAssign, $6}
               
                }

              |TOK_FOR TOK_WORD TOK_IN TOK_LBRAC TOK_NUM TOK_DOT TOK_DOT TOK_NUM TOK_RBRAC TOK_SEMICOL st2 TOK_DONE 
              {n1:=&numNode{$5}; n2:= &numNode{$8};initnd:=&assignNode{$2, n1};
               var cond *comparatorNode; var incr *arithNode; var iter *symbolReferenceNode;
               var incrAssign *assignNode;
              
              iter = &symbolReferenceNode{$2,&numNode{0}, nil}

              if $5 < $8 {
              cond=&comparatorNode{"<", iter, n2}
              incr=&arithNode{"+", iter, &numNode{1}}
              incrAssign=&assignNode{iter, incr} //Maybe redundant
              } else if $5 == $8 {

              } else { //$5 > 8
              cond=&comparatorNode{">", iter, n2}
              incr=&arithNode{"-", iter, &numNode{1}}
              incrAssign=&assignNode{iter, incr}
              } 
              $$=&forNode{initnd, cond, incrAssign,$11 }
              }
              |TOK_FOR TOK_WORD TOK_IN TOK_DEREF TOK_LPAREN Q TOK_RPAREN TOK_DO st2 TOK_DONE 
              {
              arrNd:=$6
              arrRes:= arrNd.execute()
              qRes :=&assignNode{"_internalRes", arrRes}
              varIter:=&assignNode{$2, 
                     &symbolReferenceNode{"_internalRes", &numNode{0}, nil}}
              init:=&ast{[]node{qRes, varIter}}


              offset := &symbolReferenceNode{"_internalIdx", &numNode{0},nil}
              varIterAssign:=&assignNode{ 
              &symbolReferenceNode{$2,&numNode{0},nil}, 
              &symbolReferenceNode{"_internalRes", 
              offset, nil}}

              incr:=&ast{[]node{varIterAssign}}
              body:=&ast{[]node{incr,$9}}
              $$=&rangeNode{init, arrRes,body }}
              
              |TOK_FOR TOK_WORD TOK_IN TOK_DEREF TOK_LPAREN factor TOK_RPAREN TOK_DO st2 TOK_DONE 
              {
              arrNd:= $6
              //This NonTerminal is broken, it is kept
              //here to show that eventuall the feature
              //must be added
              arrRes:= arrNd.execute()
              qRes :=&assignNode{"_internalRes", arrRes}
              varIter:=&assignNode{$2, 
                     &symbolReferenceNode{"_internalRes", &numNode{0}, nil}}
              init:=&ast{[]node{qRes, varIter}}


              offset := &symbolReferenceNode{"_internalIdx", &numNode{0},nil}
              varIterAssign:=&assignNode{ 
              &symbolReferenceNode{$2,&numNode{0},nil}, 
              &symbolReferenceNode{"_internalRes", 
              offset, nil}}

              incr:=&ast{[]node{varIterAssign}}
              body:=&ast{[]node{incr,$9}}
              $$=&rangeNode{init, arrRes,body }}
              ;

EIF: TOK_ELIF TOK_LBLOCK EXPR TOK_RBLOCK TOK_THEN st2 EIF 
       {x:=elifNode{$3, $6};f:=[]elifNode{x}; f = append(f,$7...);$$=f}
       | {$$=nil}
       ;

EXPR: EXPR TOK_OR JOIN
       |JOIN
       ;

JOIN: JOIN TOK_AND EQAL 
       |EQAL {$$=$1}
       ;

EQAL: EQAL TOK_EQUAL TOK_EQUAL REL {$$=&comparatorNode{"==", $1, $4}}
       |EQAL TOK_NOT TOK_EQUAL REL {$$=&comparatorNode{"!=", $1, $4}}
       |REL {$$=$1}
       ;

REL: nex TOK_LESS nex {$$=&comparatorNode{"<", $1, $3}}
       |nex TOK_LESS TOK_EQUAL nex {$$=&comparatorNode{"<=", $1, $4}}
       |nex TOK_GREATER TOK_EQUAL nex {$$=&comparatorNode{">=", $1, $4}}
       |nex TOK_GREATER nex {$$=&comparatorNode{">", $1, $3}}
       |nex {$$=$1}
       ;

nex: nex TOK_PLUS term {$$=&arithNode{"+", $1, $3}}
       |nex TOK_OCDEL term {$$=&arithNode{"-", $1, $3}}
       |term {$$=$1}
       ;

term: term TOK_MULT unary {$$=&arithNode{"*", $1, $3}}
       |term TOK_SLASH unary {$$=&arithNode{"/", $1, $3}}
       |term TOK_MOD unary {$$=&arithNode{"%", $1, $3}}
       |unary {$$=$1}
       ;

unary: TOK_NOT unary {$$=&boolOpNode{"!", $2}}
       |TOK_OCDEL unary {left := &numNode{0};$$=&arithNode{"-",left,$2 }}
       |factor {$$=$1}
       ;

factor: TOK_LPAREN EXPR TOK_RPAREN {$$=$2}
       |TOK_NUM {$$=&numNode{$1}}
       |TOK_FLOAT {$$=&floatNode{$1}}
       |TOK_DEREF TOK_WORD {$$=&symbolReferenceNode{$2, &numNode{0}, nil}}
       |TOK_DEREF TOK_WORD TOK_LBLOCK EXPR TOK_RBLOCK {$$=&symbolReferenceNode{$2, $4, nil}}
       |TOK_LEN TOK_LPAREN TOK_WORD TOK_RPAREN {x:=&symbolReferenceNode{$3, &numNode{-1}, nil};
                                                 switch x.execute().(type) {
                                                        case int:
                                                        $$=&numNode{x.execute().(int)}
                                                        default: //Error, the array length is not an int
                                                        println("Error! Single element arrays are not supported")
                                                        $$=&numNode{-1}

                                                 }
                                                }
       //|TOK_WORD {$$=&symbolReferenceNode{$1,&numNode{0}, nil}}
       |TOK_WORD {$$=&strNode{$1}; dCatchPtr = $1; dCatchNodePtr=&strNode{$1}}
       |TOK_STR {$$=&strNode{$1}}
       |TOK_BOOL {var x bool;if $1=="false"{x = false}else{x=true};$$=&boolNode{x}}
       |TOK_LBLOCK WNARG TOK_RBLOCK {x:=retNodeArray($2);$$=&arrNode{len(x), x}}
       ;

K: NT_CREATE     
       | NT_GET
       | NT_UPDATE 
       | NT_DEL 
;
NT_CREATE: TOK_CREATE OBJ_TYPE P TOK_COL F {/*cmd.Disp(resMap(&$5, $2, false));*/ $$=&postObjNode{$2, resMap(&$5, $2, false)}}
           |TOK_CREATE OBJ_TYPE TOK_USE_JSON TOK_COL P {$$=&easyPostNode{$2, $5}}
;

NT_GET: TOK_GET P {$$=&getObjectNode{$2}}
       | TOK_GET OBJ_TYPE F {/*cmd.Disp(resMap(&$4)); */$$=&searchObjectsNode{$2, resMap(&$3, $2, false)}}
;

NT_UPDATE: TOK_UPDATE P TOK_COL F {$$=&updateObjNode{[]interface{}{$2, resMap(&$4, $2, true)}}}
           |TOK_UPDATE P TOK_COL F TOK_ATTRSPEC TOK_WORD {$$=&recursiveUpdateObjNode{$2, resMap(&$4, $2, true), $6}}
           |TOK_UPDATE P TOK_COL TOK_USE_JSON TOK_COL P {$$=&easyUpdateNode{$2, $6, true}}
           |TOK_UPDATE P TOK_COL TOK_USE_JSON TOK_PARTIAL TOK_COL P {$$=&easyUpdateNode{$2, $7, false}}
;

NT_DEL: TOK_DELETE P {if cmd.State.DebugLvl >= 3 {println("@State NT_DEL");}; $$=&deleteObjNode{$2}}
;

OBJ_TYPE: TOK_TENANT | TOK_SITE | TOK_BLDG | TOK_ROOM | TOK_RACK | TOK_DEVICE | TOK_AC | TOK_PANEL |TOK_CABINET | TOK_ROW 
       | TOK_TILE | TOK_WALL | TOK_SENSOR | TOK_CORIDOR | TOK_GROUP | TOK_OBJ_TMPL | TOK_ROOM_TMPL
;

WORDORNUM: TOK_WORD {$$=$1; dCatchPtr = $1; dCatchNodePtr=&strNode{$1}}
           |TOK_NUM {x := strconv.Itoa($1);$$=x;dCatchPtr = $1; dCatchNodePtr=&numNode{$1}}
           |TOK_FLOAT {x := strconv.FormatFloat($1, 'E', -1, 64);$$=x;dCatchPtr = $1; dCatchNodePtr=&floatNode{$1}}
           |TOK_PLUS TOK_WORD TOK_PLUS TOK_WORD {$$=$1+$2+$3+$4; dCatchPtr = $1+$2+$3+$4; dCatchNodePtr=&strNode{$1+$2+$3+$4}}
           |TOK_PLUS TOK_WORD TOK_OCDEL TOK_WORD {$$=$1+$2+$3+$4; dCatchPtr = $1+$2+$3+$4; dCatchNodePtr=&strNode{$1+$2+$3+$4}}
           |TOK_OCDEL TOK_WORD TOK_OCDEL TOK_WORD {$$=$1+$2+$3+$4; dCatchPtr = $1+$2+$3+$4; dCatchNodePtr=&strNode{$1+$2+$3+$4}}
           |TOK_OCDEL TOK_WORD TOK_PLUS TOK_WORD {$$=$1+$2+$3+$4; dCatchPtr = $1+$2+$3+$4; dCatchNodePtr=&strNode{$1+$2+$3+$4}}
           |TOK_BOOL {var x bool;if $1=="false"{x = false}else{x=true};dCatchPtr = x; dCatchNodePtr=&boolNode{x}}
           |TOK_DEREF TOK_WORD 
           {
                  $$=resolveReference($2)
           }
           ;

F:     TOK_WORD TOK_EQUAL WORDORNUM F {$$=string($1+"="+$3+"="+$4); if cmd.State.DebugLvl >= 3 {println("So we got: ", $$);}}
       | TOK_WORD TOK_EQUAL WORDORNUM {$$=$1+"="+$3}
       | TOK_WORD TOK_EQUAL TOK_STR F{$$=$1+"="+$3+"="+$4}
       | TOK_WORD TOK_EQUAL TOK_STR {$$=$1+"="+$3}
       | TOK_WORD TOK_EQUAL OBJ_TYPE {$$=$1+"="+$3}
       | TOK_WORD TOK_EQUAL OBJ_TYPE F {$$=string($1+"="+$3+"="+$4); if cmd.State.DebugLvl >= 3 {println("So we got: ", $$);}}
;


P:     P1
       | TOK_SLASH P1 {$$="/"+$2}
;

P1:    TOK_WORD TOK_SLASH P1 {$$=$1+"/"+$3}
       | TOK_WORD TOK_DOT P1 {$$=$1+"."+$3}
       | TOK_WORD {$$=$1}
       | TOK_STR {$$=$1}
       | TOK_STR TOK_PLUS TOK_SLASH P1 {$$=$1+"/"+$4}
       | TOK_DOT TOK_DOT TOK_SLASH P1 {$$="../"+$4}
       | TOK_DOT TOK_DOT {$$=".."}
       | TOK_OCDEL {$$="-"}
       | TOK_DEREF TOK_WORD {$$= resolveReference($2)}
       | TOK_DEREF TOK_WORD TOK_PLUS TOK_SLASH P1  {$$=resolveReference($2)+"/"+$5}
       | TOK_DEREF TOK_WORD TOK_PLUS TOK_STR {$$=resolveReference($2)+$4}
       | TOK_DEREF TOK_WORD TOK_PLUS TOK_STR TOK_PLUS TOK_SLASH P1 {$$=resolveReference($2)+$4+"/"+$7}
       | TOK_DEREF TOK_LBRAC TOK_WORD TOK_RBRAC P {$$=resolveReference($3)+$5}
       | {$$=""}
;

LSOBJ_COMMAND: TOK_LSTEN {$$=0} | TOK_LSSITE {$$=1} | TOK_LSBLDG {$$=2} | TOK_LSROOM {$$=3} | TOK_LSRACK {$$=4}
       | TOK_LSDEV {$$=5} | TOK_LSROW {$$=10} | TOK_LSTILE {$$=11} | TOK_LSAC {$$=6} | TOK_LSPANEL {$$=7}
       | TOK_LSWALL {$$=8} | TOK_LSCAB {$$=9} | TOK_LSCORRIDOR {$$=12} | TOK_LSSENSOR{$$=13}
;


Q:      TOK_CD P {$$=&cdNode{$2}}
       | TOK_LS P {$$=&lsNode{$2}}
       | LSOBJ_COMMAND P {$$=&lsObjNode{$2, $1}}
       | TOK_TREE P {$$=&treeNode{$2, 0}}
       | TOK_TREE P TOK_NUM {$$=&treeNode{$2, $3}}
       | TOK_DRAW P {$$=&drawNode{$2, 0}}
       | TOK_DRAW P TOK_NUM {$$=&drawNode{$2, $3}}
       | TOK_HIERARCH P {$$=&hierarchyNode{$2, 1}}
       | TOK_HIERARCH P TOK_NUM {$$=&hierarchyNode{$2, $3}}
       | TOK_UNSET TOK_OCDEL TOK_WORD TOK_WORD {$$=&unsetNode{$2+$3, $4, nil, nil} }
       | TOK_UNSET TOK_DEREF TOK_WORD TOK_LBLOCK EXPR TOK_RBLOCK {
              v:=&symbolReferenceNode{$3, $5, nil}; 
              //$$=&assignNode{v, "deleteValue"}
              $$=&unsetNode{"","" ,v, nil}
              }
       | BASH     {$$=$1}
       | TOK_DRAWABLE TOK_LPAREN P TOK_RPAREN {$$=&isEntityDrawableNode{$3}}
       | TOK_DRAWABLE TOK_LPAREN P TOK_COMMA factor TOK_RPAREN {$$=&isAttrDrawableNode{$3, $5}}
       | TOK_SETENV TOK_WORD TOK_EQUAL EXPR {$$=&setEnvNode{$2, $4}}
;

COMMAND: TOK_LINK{$$="link"} | TOK_UNLINK{$$="unlink"} | TOK_CLR{$$="clear"} | TOK_LS{$$="ls"} 
       | TOK_PWD{$$="pwd"} | TOK_PRNT{$$="print"} | TOK_CD{$$="cd"} | TOK_CAM{$$="camera"} 
       | TOK_UI{$$="ui"} | TOK_CREATE{$$="create"} | TOK_GET{$$="gt"} | TOK_UPDATE{$$="update"} 
       | TOK_DELETE{$$="delete"} | TOK_HIERARCH{$$="hc"} | TOK_TREE{$$="tree"} | TOK_DRAW{$$="draw"} 
       | TOK_IF{$$="if"} | TOK_WHILE{$$="while"} | TOK_FOR{$$="for"} | TOK_UNSET{$$="unset"}
       | TOK_SELECT{$$="select"} | TOK_CMDS{$$="cmds"} | TOK_LSOG{$$="lsog"} | TOK_ENV{$$="env"} 
       | TOK_LSTEN{$$="lsten"} | TOK_LSSITE{$$="lssite"} | TOK_LSBLDG{$$="lsbldg"} | TOK_LSROOM{$$="lsroom"} 
       | TOK_LSRACK{$$="lsrack"} | TOK_LSDEV{$$="lsdev"} | TOK_OCDEL{$$="-"} | TOK_DOT TOK_TEMPLATE{$$=".template"}
       | TOK_DOT TOK_CMDS{$$=".cmds"} | TOK_VAR{$$=".var"} | TOK_PLUS{$$="+"} | TOK_EQUAL{$$="="} 
       | TOK_GREATER{$$=">"} | TOK_DRAWABLE{$$="drawable"}
;

BASH:  TOK_CLR {$$=&clrNode{}}
       | TOK_GREP {$$=&grepNode{}}
       | TOK_PRNT NODEGETTER {$$=&printNode{$2}}
       | TOK_LSOG {$$=&lsogNode{}}
       | TOK_ENV {$$=&envNode{}}
       | TOK_PWD {$$=&pwdNode{}}
       | TOK_EXIT {$$=&exitNode{}}
       | TOK_DOC COMMAND {$$=&helpNode{$2}}
       | TOK_DOC {$$=&helpNode{""}}
       | TOK_DOC TOK_WORD {$$=&helpNode{$2}}
;

OCLISYNTX:  TOK_PLUS OCCR {$$=$2}
            |OCDEL {$$=$1}
            |OCUPDATE {$$=$1}
            |OCGET {$$=$1}
            |OCCHOOSE {$$=$1}
            |OCDOT {$$=$1}
            |OCSEL {$$=$1;}
            |HANDLEUI {$$=$1}
            |HANDLELINKS {$$=$1}
            ;


OCCR:   
        TOK_TENANT TOK_COL P TOK_ATTRSPEC EXPR {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"color":$5}} 
              $$=&getOCAttrNode{replaceOCLICurrPath($3), cmd.TENANT, attributes}
        }
        |TOK_SITE TOK_COL P TOK_ATTRSPEC EXPR {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"orientation":$5}}
              $$=&getOCAttrNode{replaceOCLICurrPath($3), cmd.SITE, attributes}
        } 
        |TOK_BLDG TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}}
              $$=&getOCAttrNode{replaceOCLICurrPath($3), cmd.BLDG, attributes}
        }
        |TOK_ROOM TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR{
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7, "orientation":$9, "floorUnit":$11}}
              $$=&getOCAttrNode{replaceOCLICurrPath($3), cmd.ROOM, attributes}
        }
        |TOK_ROOM TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7, "orientation":$9}}
              $$=&getOCAttrNode{replaceOCLICurrPath($3), cmd.ROOM, attributes}
        }
        |TOK_ROOM TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC NTORIENTATION {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7, "orientation":$9}}
              $$=&getOCAttrNode{replaceOCLICurrPath($3), cmd.ROOM, attributes}
        }
        |TOK_ROOM TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "template":$7}}
              $$=&getOCAttrNode{replaceOCLICurrPath($3), cmd.ROOM, attributes}
        }
        /* |TOK_RACK TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR{
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7, "orientation":$9, "template":$11}}
              $$=&getOCAttrNode{replaceOCLICurrPath($3),cmd.RACK, attributes}
        } */
        |TOK_RACK TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR{
              attr := map[string]interface{}{}; 
              if checkIfTemplate(($7).execute()) == false {
                     attr["size"] = $7
              } else {
                     attr["template"] = $7
              }
              attr["posXY"] = $5; attr["orientation"] = $9;
              attributes := map[string]interface{}{"attributes":attr};
              $$=&getOCAttrNode{replaceOCLICurrPath($3), cmd.RACK, attributes}
        }
        |TOK_DEVICE TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"slot":$5, "template":$7, "orientation":$9}}
              $$=&getOCAttrNode{replaceOCLICurrPath($3), cmd.DEVICE, attributes}
        }
        |TOK_DEVICE TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR {
              attr := map[string]interface{}{"posU/slot":$5}; 
              res := checkIfTemplate(($7).execute()); 
              if res == false {
                     attr["sizeU"] = $7
              } else {
                     attr["template"]=$7
              }
              attributes := map[string]interface{}{"attributes":attr}
              $$=&getOCAttrNode{replaceOCLICurrPath($3), cmd.DEVICE, attributes}
        }
        |TOK_CORIDOR TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR {
              attributes := map[string]interface{}{"name":$5, "leftRack":$7, "rightRack":$9, "temperature":$11}
              $$=&getOCAttrNode{replaceOCLICurrPath($3), cmd.CORIDOR, attributes}
        }
        |TOK_GROUP TOK_COL P TOK_ATTRSPEC EXPR CDORFG { 
              attributes:=map[string]interface{}{"name":$5,"racks":$6}; 
              $$=&getOCAttrNode{replaceOCLICurrPath($3), cmd.GROUP, attributes}
        }
        |TOK_WALL TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR {
              attributes := map[string]interface{}{"pos1":$5,"pos2":$7}
              $$=&getOCAttrNode{replaceOCLICurrPath($3), cmd.SEPARATOR, attributes}
        }
        |TOK_ORPH TOK_COL TOK_DEVICE TOK_COL P TOK_ATTRSPEC EXPR {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"template":$7}}
              $$=&getOCAttrNode{replaceOCLICurrPath($5), cmd.STRAY_DEV, attributes}
        }
        |TOK_ORPH TOK_COL TOK_SENSOR TOK_COL P TOK_ATTRSPEC EXPR {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"template":$7} }
              $$=&getOCAttrNode{replaceOCLICurrPath($5), cmd.STRAYSENSOR, attributes}
        }
       //EasyPost syntax STRAYSENSOR
       |OBJ_TYPE TOK_USE_JSON P {$$=&easyPostNode{$1, $3}}

       ; 
OCDEL:  TOK_OCDEL P {$$=&deleteObjNode{replaceOCLICurrPath($2)}}
       |TOK_OCDEL TOK_SELECT {$$=&deleteSelectionNode{}}
;

OCUPDATE:  P TOK_COL TOK_WORD TOK_EQUAL EXPR {
              val := map[string]interface{}{$3:($5).(node).execute()};
              $$=&updateObjNode{[]interface{}{replaceOCLICurrPath($1), val}};
              if cmd.State.DebugLvl >= 3 {
                     println("Attribute Acquired");
              }}
           |P TOK_COL TOK_WORD TOK_EQUAL EXPR TOK_ATTRSPEC EXPR {
              if _, ok := ($7).(*arrNode); ok {
                     val := map[string]interface{}{$3 : map[string]interface{}{"reserved":($5).(node).execute(), "technical":($7).(node).execute()}};
                     $$=&updateObjNode{[]interface{}{replaceOCLICurrPath($1), val}}
              } else {
                     val := map[string]interface{}{$3:($5).(node).execute()};
                     $$=&recursiveUpdateObjNode{replaceOCLICurrPath($1), val, $7}
                     if cmd.State.DebugLvl >= 3 {
                            println("Attribute Acquired");
                     }
              }}
           //|P TOK_COL TOK_WORD TOK_EQUAL EXPR TOK_ATTRSPEC EXPR {}
;

OCGET: TOK_EQUAL P {$$=&getObjectNode{replaceOCLICurrPath($2)}}
;

GETOBJS:      P TOK_COMMA GETOBJS {x := make([]string,0); x = append(x, formActualPath($1)); x = append(x, $3...); $$=x}
              |P {$$=[]string{formActualPath($1)}}
              //| TOK_WORD {$$=[]string{cmd.State.CurrPath+"/"+$1}}
              ;

OCCHOOSE: TOK_EQUAL TOK_LBRAC GETOBJS TOK_RBRAC {$$=&setCBNode{&$3}; println("Selection made!")}
;

OCDOT:      //TOK_VAR TOK_COL TOK_WORD TOK_EQUAL WORDORNUM {$$=&assignNode{$3, dCatchNodePtr}}
            //|TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_QUOT STRARG TOK_QUOT{$$=&assignNode{$3, &strNode{$6}}}
            //TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_LPAREN WNARG TOK_RPAREN {$$=&assignNode{$3, &arrNode{len($6),retNodeArray($6)}}}
            TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN K TOK_RPAREN {$$=&assignNode{$3, ($7).(node).execute()}}
            |TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN Q  TOK_RPAREN {$$=&assignNode{$3, ($7).(node).execute()}}
            |TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN TOK_PLUS OCCR  TOK_RPAREN {$$=&assignNode{$3, ($8).(node).execute()}}
            |TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN OCDEL  TOK_RPAREN {$$=&assignNode{$3, ($7).(node).execute()}}
            |TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN OCUPDATE  TOK_RPAREN {$$=&assignNode{$3, ($7).(node).execute()}}
            |TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN OCGET  TOK_RPAREN {$$=&assignNode{$3, ($7).(node).execute()}}
            |TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN OCCHOOSE  TOK_RPAREN {$$=&assignNode{$3, ($7).(node).execute()}}
            |TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN OCSEL  TOK_RPAREN {$$=&assignNode{$3, ($7).(node).execute()}}
            |TOK_VAR TOK_COL TOK_WORD TOK_EQUAL EXPR {
                     val := ($5).(node).execute();
                     if _, ok := ($5).(*arithNode); ok && val == nil{
                            $$=nil
                     } else {
                            $$=&assignNode{$3, val}
                     }
                }
            |TOK_DOT TOK_CMDS TOK_COL P {$$=&loadNode{$4}}
            |TOK_DOT TOK_TEMPLATE TOK_COL P {$$=&loadTemplateNode{$4}}
            |TOK_VAR TOK_COL TOK_WORD TOK_EQUAL Q {$$=&assignNode{$3, $5}}
            |TOK_VAR TOK_COL TOK_WORD TOK_EQUAL K {$$=&assignNode{$3, $5}}
            //|TOK_VAR TOK_COL TOK_WORD TOK_EQUAL P {$$=&assignNode{$3, $5}}
            //|TOK_VAR TOK_COL TOK_WORD TOK_EQUAL OCLISYNTX {$$=&assignNode{$3, $5}}
            |TOK_DEREF TOK_WORD {$$=&symbolReferenceNode{$2, &numNode{0}, nil}}
              

            |TOK_DEREF TOK_WORD TOK_LBLOCK EXPR TOK_RBLOCK {$$=&symbolReferenceNode{$2, $4, nil}}
            |TOK_DEREF TOK_WORD TOK_LBLOCK EXPR TOK_RBLOCK TOK_EQUAL EXPR {v:=&symbolReferenceNode{$2, $4, nil}; $$=&assignNode{v, $7} }
            |TOK_DEREF TOK_WORD TOK_LBLOCK EXPR TOK_RBLOCK TOK_LBLOCK EXPR TOK_RBLOCK {$$=&symbolReferenceNode{$2, /*&numNode{$4}*/$4, /*&strNode{$7}*/ $7}}
            |TOK_DEREF TOK_WORD TOK_EQUAL EXPR {n:=&symbolReferenceNode{$2, &numNode{0}, nil};$$=&assignNode{n,$4 }}
;

OCSEL:      TOK_SELECT {$$=&selectNode{}}
            |TOK_SELECT TOK_DOT TOK_WORD TOK_EQUAL EXPR {
                     /*x := $3+"="+$5;*/ 
                     val:=($5).(node).execute(); 
                     x:=map[string]interface{}{$3:val};
                     $$=&updateSelectNode{x};
                     }
;

HANDLEUI: TOK_UI TOK_DOT TOK_WORD TOK_EQUAL EXPR {$$=&handleUnityNode{[]interface{}{"ui", $3, ($5).(node).execute()}}}
          |TOK_CAM TOK_DOT TOK_WORD TOK_EQUAL EXPR TOK_ATTRSPEC EXPR {
              _, firstIsArr := ($5).(*arrNode)
              _, secondIsArr := ($7).(*arrNode)
              if firstIsArr && secondIsArr {
                     $$=nil;
              } else {
                     $$=&handleUnityNode{[]interface{}{"camera", $3, ($5).(node).execute(), ($7).(node).execute()}}
              }}
          |TOK_CAM TOK_DOT TOK_WORD TOK_EQUAL EXPR {
              if _, ok := ($5).(*arrNode); ok && $3 != "wait" {
                     $$=nil;
              } else  {
                     $$=&handleUnityNode{[]interface{}{"camera", $3, ($5).(node).execute()}}
              }}
          |TOK_GREATER P {$$=&focusNode{$2}}
;

HANDLELINKS:  TOK_LINK TOK_COL P TOK_ATTRSPEC P {$$=&linkObjectNode{[]interface{}{$3, $5}}}
              |TOK_LINK TOK_COL P TOK_ATTRSPEC P TOK_ATTRSPEC EXPR {$$=&linkObjectNode{[]interface{}{$3, $5, $7}}}
              |TOK_UNLINK TOK_COL P {$$=&unlinkObjectNode{[]interface{}{$3}}}
              |TOK_UNLINK TOK_COL P TOK_ATTRSPEC P {$$=&unlinkObjectNode{[]interface{}{$3,$5}}}
            ;
//For making array types
WNARG: EXPR TOK_COMMA WNARG {x:=[]interface{}{$1}; $$=append(x, $3...)}
       |EXPR  {x:=[]interface{}{$1}; $$=x}
       ;

FUNC: TOK_WORD TOK_LPAREN TOK_RPAREN TOK_LBRAC st2 TOK_RBRAC {x:=&funcNode{$5}; $$=&assignNode{$1, x};}
       |TOK_WORD {x:=funcTable[$1]; if _,ok:=x.(node); ok {$$=x.(node)}else{$$=nil};}
       ;

//Special nonterminal for print
NODEGETTER: NODEACC TOK_PLUS NODEGETTER {if len($3) != 0 {$$=append($1, $3...)} else {$$=$1};}
       |NODEACC {$$=$1}
       | {$$=nil}
       ;


NODEACC:  factor {$$=[]interface{}{$1}}
           ;

//Child devices of rack for group 
//Since the OCLI syntax defines no limit
//for the number of devices 
//a NonTerminal state is neccessary
CDORFG: TOK_ATTRSPEC WORDORNUM CDORFG {x:=$2; $$=x+","+$3}
       | {$$=""}
       ;

//Nonterminal for the OCLI syntax object creation
//because team doesn't want to type "+N+W" 
//but rather +N+W
NTORIENTATION: TOK_PLUS TOK_WORD TOK_PLUS TOK_WORD {$$=$1+$2+$3+$4; dCatchPtr = $1+$2+$3+$4; dCatchNodePtr=&strNode{$1+$2+$3+$4}}
              |TOK_PLUS TOK_WORD TOK_OCDEL TOK_WORD {$$=$1+$2+$3+$4; dCatchPtr = $1+$2+$3+$4; dCatchNodePtr=&strNode{$1+$2+$3+$4}}
              |TOK_OCDEL TOK_WORD TOK_OCDEL TOK_WORD {$$=$1+$2+$3+$4; dCatchPtr = $1+$2+$3+$4; dCatchNodePtr=&strNode{$1+$2+$3+$4}}
              |TOK_OCDEL TOK_WORD TOK_PLUS TOK_WORD {$$=$1+$2+$3+$4; dCatchPtr = $1+$2+$3+$4; dCatchNodePtr=&strNode{$1+$2+$3+$4}}
              ;