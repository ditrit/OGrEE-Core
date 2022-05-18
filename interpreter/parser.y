%{
package main
import (
cmd "cli/controllers"
"strings"
"strconv"
)

var root node 

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
                            x := map[int]interface{}{0: &boolNode{BOOL, false}}
                            res = append(res, x)
                     }
                     if input[idx].(string) == "true" {
                            x := map[int]interface{}{0: &boolNode{BOOL, true}}
                            res = append(res, x)
                     } else {
                            x := map[int]interface{}{0: &strNode{STR, input[idx].(string)}}
                            res = append(res, x)
                     }
                     case int:
                     x := map[int]interface{}{0: &numNode{NUM, input[idx].(int)}}
                     res = append(res, x)

                     default: //Most likely a node
                     res = append(res, map[int]interface{}{0:input[idx]})
              }
              /*if input[idx].(string) == "false" {
                     x := map[int]interface{}{0: &boolNode{BOOL, false}}
                     res = append(res, x)
              } else if input[idx].(string) == "true" {
                     x := map[int]interface{}{0: &boolNode{BOOL, true}}
                     res = append(res, x)
              } else if v,e := strconv.Atoi(input[idx].(string)); e == nil {
                     x := map[int]interface{}{0: &numNode{NUM, v}}
                     res = append(res, x)
              } else {
                     x := map[int]interface{}{0: &strNode{STR, input[idx].(string)}}
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
                 dCatchNodePtr=&boolNode{BOOL, item.(bool)}
                 if item.(bool) == false {return "false"} else { return "true"}
              case string:
                 dCatchNodePtr=&strNode{STR, item.(string)}
                 return item.(string)
              case int:
                 dCatchNodePtr=&numNode{NUM, item.(int)}
                 return strconv.Itoa(item.(int))
              /*case map[string]interface{}:
                 //dCatchNodePtr=&symbolReferenceNode{REFERENCE, }
                 return item.(map[string]interface{})[subIdx].(string)*/
              case *commonNode:
                     dCatchNodePtr=item
                     args := ""
                     for i := range item.(*commonNode).args {
                            args += item.(*commonNode).args[i].(string)
                     }
                     return item.(*commonNode).val +" "+ args
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
       TOK_CAM TOK_UI TOK_HIERARCH TOK_DRAW TOK_ENV TOK_ORPH
       
%type <s> F E P P1 WORDORNUM CDORFG NTORIENTATION
%type <arr> WNARG NODEGETTER NODEACC
%type <sarr> GETOBJS
%type <elifArr> EIF
%type <node> OCSEL OCLISYNTX OCDEL OCGET NT_CREATE NT_GET NT_DEL 
%type <node> NT_UPDATE K Q BASH OCUPDATE OCCHOOSE OCCR OCDOT
%type <node> EXPR REL OPEN_STMT CTRL nex factor unary EQAL term
%type <node> stmnt JOIN
%type <node> st2 FUNC HANDLEUI HANDLELINKS
%left TOK_MULT TOK_OCDEL TOK_SLASH TOK_PLUS 
%right TOK_EQUAL


%%

start: st2 {root = $1}

st2: stmnt {$$=&ast{BLOCK,[]node{$1} }}
       |stmnt TOK_SEMICOL st2 {$$=&ast{BLOCK,[]node{$1, $3}}}
       |CTRL {$$=&ast{IF,[]node{$1}}}
;



stmnt: K {$$=$1}
       |Q {$$=$1}
       |OCLISYNTX {$$=$1}
       |FUNC {$$=$1}
       |{$$=nil}
;

CTRL: OPEN_STMT {$$=$1}
       ;

OPEN_STMT:    TOK_IF TOK_LBLOCK EXPR TOK_RBLOCK TOK_THEN st2 TOK_FI {$$=&ifNode{IF, $3, $6, nil, nil}}
              |TOK_IF TOK_LBLOCK EXPR TOK_RBLOCK TOK_THEN st2 EIF TOK_ELSE st2 TOK_FI {$$=&ifNode{IF, $3, $6, $9, $7}}
              |TOK_WHILE TOK_LPAREN EXPR TOK_RPAREN st2 TOK_DONE {$$=&whileNode{WHILE, $3, $5}}
              |TOK_FOR TOK_LPAREN TOK_LPAREN TOK_WORD TOK_EQUAL WORDORNUM TOK_SEMICOL EXPR TOK_SEMICOL stmnt TOK_RPAREN TOK_RPAREN TOK_SEMICOL st2 TOK_DONE 
              {initnd:=&assignNode{ASSIGN, $4, dCatchNodePtr};$$=&forNode{FOR,initnd,$8,$10,$14}}
              |TOK_FOR TOK_WORD TOK_IN EXPR TOK_SEMICOL st2 TOK_DONE 
              {var incr *arithNode; var incrAssign *assignNode; 
              n1:=&numNode{NUM, 0};
              
              initd:=&assignNode{ASSIGN, $2, n1}; 
              iter:=&symbolReferenceNode{REFERENCE, $2, &numNode{NUM,0}, nil}; 
              cmp:=&comparatorNode{COMPARATOR, "<", iter, $4}
              incr=&arithNode{ARITHMETIC, "+", iter, &numNode{NUM, 1}}
              incrAssign=&assignNode{ASSIGN, iter,incr}
              $$=&forNode{FOR,initd, cmp, incrAssign, $6}
               
                }

              |TOK_FOR TOK_WORD TOK_IN TOK_LBRAC TOK_NUM TOK_DOT TOK_DOT TOK_NUM TOK_RBRAC TOK_SEMICOL st2 TOK_DONE 
              {n1:=&numNode{NUM, $5}; n2:= &numNode{NUM, $8};initnd:=&assignNode{ASSIGN, $2, n1};
               var cond *comparatorNode; var incr *arithNode; var iter *symbolReferenceNode;
               var incrAssign *assignNode;
              
              iter = &symbolReferenceNode{NUM, $2,&numNode{NUM,0}, nil}

              if $5 < $8 {
              cond=&comparatorNode{COMPARATOR, "<", iter, n2}
              incr=&arithNode{ARITHMETIC, "+", iter, &numNode{NUM, 1}}
              incrAssign=&assignNode{ASSIGN, iter, incr} //Maybe redundant
              } else if $5 == $8 {

              } else { //$5 > 8
              cond=&comparatorNode{COMPARATOR, ">", iter, n2}
              incr=&arithNode{ARITHMETIC, "-", iter, &numNode{NUM, 1}}
              incrAssign=&assignNode{ASSIGN, iter, incr}
              } 
              $$=&forNode{FOR, initnd, cond, incrAssign,$11 }
              }
              |TOK_FOR TOK_WORD TOK_IN TOK_DEREF TOK_LPAREN Q TOK_RPAREN TOK_DO st2 TOK_DONE 
              {
              arrNd:=$6
              arrRes:= arrNd.execute()
              qRes :=&assignNode{ASSIGN, "_internalRes", arrRes}
              varIter:=&assignNode{ASSIGN, $2, 
                     &symbolReferenceNode{REFERENCE, "_internalRes", &numNode{NUM,0}, nil}}
              init:=&ast{ASSIGN, []node{qRes, varIter}}


              offset := &symbolReferenceNode{REFERENCE, "_internalIdx", &numNode{NUM,0},nil}
              varIterAssign:=&assignNode{ASSIGN, 
              &symbolReferenceNode{REFERENCE, $2,&numNode{NUM,0},nil}, 
              &symbolReferenceNode{REFERENCE, "_internalRes", 
              offset, nil}}

              incr:=&ast{ASSIGN, []node{varIterAssign}}
              body:=&ast{BLOCK, []node{incr,$9}}
              $$=&rangeNode{FOR, init, arrRes,body }}
              
              |TOK_FOR TOK_WORD TOK_IN TOK_DEREF TOK_LPAREN factor TOK_RPAREN TOK_DO st2 TOK_DONE 
              {
              arrNd:= $6
              //This NonTerminal is broken, it is kept
              //here to show that eventuall the feature
              //must be added
              arrRes:= arrNd.execute()
              qRes :=&assignNode{ASSIGN, "_internalRes", arrRes}
              varIter:=&assignNode{ASSIGN, $2, 
                     &symbolReferenceNode{REFERENCE, "_internalRes", &numNode{NUM,0}, nil}}
              init:=&ast{ASSIGN, []node{qRes, varIter}}


              offset := &symbolReferenceNode{REFERENCE, "_internalIdx", &numNode{NUM,0},nil}
              varIterAssign:=&assignNode{ASSIGN, 
              &symbolReferenceNode{REFERENCE, $2,&numNode{NUM,0},nil}, 
              &symbolReferenceNode{REFERENCE, "_internalRes", 
              offset, nil}}

              incr:=&ast{ASSIGN, []node{varIterAssign}}
              body:=&ast{BLOCK, []node{incr,$9}}
              $$=&rangeNode{FOR, init, arrRes,body }}
              ;

EIF: TOK_ELIF TOK_LBLOCK EXPR TOK_RBLOCK TOK_THEN st2 EIF 
       {x:=elifNode{IF, $3, $6};f:=[]elifNode{x}; f = append(f,$7...);$$=f}
       | {$$=nil}
       ;

EXPR: EXPR TOK_OR JOIN
       |JOIN
       ;

JOIN: JOIN TOK_AND EQAL 
       |EQAL {$$=$1}
       ;

EQAL: EQAL TOK_EQUAL TOK_EQUAL REL {$$=&comparatorNode{COMPARATOR, "==", $1, $4}}
       |EQAL TOK_NOT TOK_EQUAL REL {$$=&comparatorNode{COMPARATOR, "!=", $1, $4}}
       |REL {$$=$1}
       ;

REL: nex TOK_LESS nex {$$=&comparatorNode{COMPARATOR, "<", $1, $3}}
       |nex TOK_LESS TOK_EQUAL nex {$$=&comparatorNode{COMPARATOR, "<=", $1, $4}}
       |nex TOK_GREATER TOK_EQUAL nex {$$=&comparatorNode{COMPARATOR, ">=", $1, $4}}
       |nex TOK_GREATER nex {$$=&comparatorNode{COMPARATOR, ">", $1, $3}}
       |nex {$$=$1}
       ;

nex: nex TOK_PLUS term {$$=&arithNode{ARITHMETIC, "+", $1, $3}}
       |nex TOK_OCDEL term {$$=&arithNode{ARITHMETIC, "-", $1, $3}}
       |term {$$=$1}
       ;

term: term TOK_MULT unary {$$=&arithNode{ARITHMETIC, "*", $1, $3}}
       |term TOK_SLASH unary {$$=&arithNode{ARITHMETIC, "/", $1, $3}}
       |term TOK_MOD unary {$$=&arithNode{ARITHMETIC, "%", $1, $3}}
       |unary {$$=$1}
       ;

unary: TOK_NOT unary {$$=&boolOpNode{BOOLOP, "!", $2}}
       |TOK_OCDEL unary {left := &numNode{NUM, 0};$$=&arithNode{ARITHMETIC, "-",left,$2 }}
       |factor {$$=$1}
       ;

factor: TOK_LPAREN EXPR TOK_RPAREN {$$=$2}
       |TOK_NUM {$$=&numNode{NUM, $1}}
       |TOK_FLOAT {$$=&floatNode{FLOAT,$1}}
       |TOK_DEREF TOK_WORD {$$=&symbolReferenceNode{REFERENCE, $2, &numNode{NUM,0}, nil}}
       |TOK_DEREF TOK_WORD TOK_LBLOCK EXPR TOK_RBLOCK {$$=&symbolReferenceNode{REFERENCE, $2, $4, nil}}
       |TOK_LEN TOK_LPAREN TOK_WORD TOK_RPAREN {x:=&symbolReferenceNode{REFERENCE, $3, &numNode{NUM, -1}, nil};
                                                 switch x.execute().(type) {
                                                        case int:
                                                        $$=&numNode{NUM, x.execute().(int)}
                                                        default: //Error, the array length is not an int
                                                        println("Error! Single element arrays are not supported")
                                                        $$=&numNode{NUM, -1}

                                                 }
                                                }
       //|TOK_WORD {$$=&symbolReferenceNode{REFERENCE, $1,&numNode{NUM,0}, nil}}
       |TOK_WORD {$$=&strNode{STR, $1}; dCatchPtr = $1; dCatchNodePtr=&strNode{STR, $1}}
       |TOK_STR {$$=&strNode{STR, $1}}
       |TOK_BOOL {var x bool;if $1=="false"{x = false}else{x=true};$$=&boolNode{BOOL, x}}
       |TOK_LBLOCK WNARG TOK_RBLOCK {x:=retNodeArray($2);$$=&arrNode{ARRAY, len(x), x}}
       ;

K: NT_CREATE     {if cmd.State.DebugLvl >= 3 {println("@State start");}}
       | NT_GET
       | NT_UPDATE 
       | NT_DEL 
;

NT_CREATE: TOK_CREATE E P TOK_COL F {/*cmd.Disp(resMap(&$5, $2, false));*/ $$=&commonNode{COMMON, cmd.PostObj, "PostObj", []interface{}{cmd.EntityStrToInt($2),$2, resMap(&$5, $2, false)}}}
           |TOK_CREATE E TOK_USE_JSON TOK_COL P {$$=&commonNode{COMMON, cmd.PostObj, "EasyPost", []interface{}{cmd.EntityStrToInt($2),$2, $5}}}
;

NT_GET: TOK_GET P {$$=&commonNode{COMMON, cmd.GetObject, "GetObject", []interface{}{$2}}}
       | TOK_GET E F {/*cmd.Disp(resMap(&$4)); */$$=&commonNode{COMMON, cmd.SearchObjects, "SearchObjects", []interface{}{$2, resMap(&$3, $2, false)}} }
;

NT_UPDATE: TOK_UPDATE P TOK_COL F {$$=&commonNode{COMMON, cmd.UpdateObj, "UpdateObj", []interface{}{$2, resMap(&$4, $2, true)}}}
           |TOK_UPDATE P TOK_COL F TOK_ATTRSPEC TOK_WORD {$$=&commonNode{COMMON, cmd.RecursivePatch, "RecursiveUpdateObj", []interface{}{$2, resMap(&$4, $2, true), $6}}}
           |TOK_UPDATE P TOK_COL TOK_USE_JSON TOK_COL P {$$=&commonNode{COMMON, cmd.UpdateObj, "EasyUpdate", []interface{}{$2, $6, true}}}
           |TOK_UPDATE P TOK_COL TOK_USE_JSON TOK_PARTIAL TOK_COL P {$$=&commonNode{COMMON, cmd.UpdateObj, "EasyUpdate", []interface{}{$2, $7, false}}}
;

NT_DEL: TOK_DELETE P {if cmd.State.DebugLvl >= 3 {println("@State NT_DEL");}; $$=&commonNode{COMMON, cmd.DeleteObj, "DeleteObj", []interface{}{$2}}}
;

E:     TOK_TENANT 
       | TOK_SITE 
       | TOK_BLDG 
       | TOK_ROOM 
       | TOK_RACK 
       | TOK_DEVICE 
       | TOK_AC
       | TOK_PANEL
       | TOK_CABINET
       | TOK_ROW
       | TOK_TILE
       | TOK_WALL
       | TOK_SENSOR
       | TOK_CORIDOR
       | TOK_GROUP
       | TOK_OBJ_TMPL
       | TOK_ROOM_TMPL
;


WORDORNUM: TOK_WORD {$$=$1; dCatchPtr = $1; dCatchNodePtr=&strNode{STR, $1}}
           |TOK_NUM {x := strconv.Itoa($1);$$=x;dCatchPtr = $1; dCatchNodePtr=&numNode{NUM, $1}}
           |TOK_FLOAT {x := strconv.FormatFloat($1, 'E', -1, 64);$$=x;dCatchPtr = $1; dCatchNodePtr=&floatNode{FLOAT, $1}}
           |TOK_PLUS TOK_WORD TOK_PLUS TOK_WORD {$$=$1+$2+$3+$4; dCatchPtr = $1+$2+$3+$4; dCatchNodePtr=&strNode{STR, $1+$2+$3+$4}}
           |TOK_PLUS TOK_WORD TOK_OCDEL TOK_WORD {$$=$1+$2+$3+$4; dCatchPtr = $1+$2+$3+$4; dCatchNodePtr=&strNode{STR, $1+$2+$3+$4}}
           |TOK_OCDEL TOK_WORD TOK_OCDEL TOK_WORD {$$=$1+$2+$3+$4; dCatchPtr = $1+$2+$3+$4; dCatchNodePtr=&strNode{STR, $1+$2+$3+$4}}
           |TOK_OCDEL TOK_WORD TOK_PLUS TOK_WORD {$$=$1+$2+$3+$4; dCatchPtr = $1+$2+$3+$4; dCatchNodePtr=&strNode{STR, $1+$2+$3+$4}}
           |TOK_BOOL {var x bool;if $1=="false"{x = false}else{x=true};dCatchPtr = x; dCatchNodePtr=&boolNode{BOOL, x}}
           |TOK_DEREF TOK_WORD 
           {
                  $$=resolveReference($2)
           }
           ;

F:     TOK_WORD TOK_EQUAL WORDORNUM F {$$=string($1+"="+$3+"="+$4); if cmd.State.DebugLvl >= 3 {println("So we got: ", $$);}}
       | TOK_WORD TOK_EQUAL WORDORNUM {$$=$1+"="+$3}
       | TOK_WORD TOK_EQUAL TOK_STR F{$$=$1+"="+$3+"="+$4}
       | TOK_WORD TOK_EQUAL TOK_STR {$$=$1+"="+$3}
       | TOK_WORD TOK_EQUAL E {$$=$1+"="+$3}
       | TOK_WORD TOK_EQUAL E F {$$=string($1+"="+$3+"="+$4); if cmd.State.DebugLvl >= 3 {println("So we got: ", $$);}}
;


P:     P1
       | TOK_SLASH P1 {$$="/"+$2}
;

P1:    TOK_WORD TOK_SLASH P1 {$$=$1+"/"+$3}
       | TOK_WORD {$$=$1}
       | TOK_STR {$$=$1}
       | TOK_STR TOK_PLUS TOK_SLASH P1 {$$=$1+"/"+$4}
       | TOK_DOT TOK_DOT TOK_SLASH P1 {$$="../"+$4}
       | TOK_WORD TOK_DOT TOK_WORD {$$=$1+"."+$3}
       | TOK_DOT TOK_DOT {$$=".."}
       | TOK_OCDEL {$$="-"}
       | TOK_DEREF TOK_WORD {$$= resolveReference($2)}
       | TOK_DEREF TOK_WORD TOK_PLUS TOK_SLASH P1  {$$=resolveReference($2)+"/"+$5}
       | TOK_DEREF TOK_WORD TOK_PLUS TOK_STR {$$=resolveReference($2)+$4}
       | TOK_DEREF TOK_WORD TOK_PLUS TOK_STR TOK_PLUS TOK_SLASH P1 {$$=resolveReference($2)+$4+"/"+$7}
       | {$$=""}
;

Q:     TOK_CD P {/*cmd.CD($2);*/ $$=&commonNode{COMMON, cmd.CD, "CD", []interface{}{$2}};}
       | TOK_LS P {/*cmd.LS($2)*/$$=&commonNode{COMMON, cmd.LS, "LS", []interface{}{$2}};}
       | TOK_LSTEN P {$$=&commonNode{COMMON, cmd.LSOBJECT, "LSOBJ", []interface{}{$2, 0}}}
       | TOK_LSSITE P { $$=&commonNode{COMMON, cmd.LSOBJECT, "LSOBJ", []interface{}{$2, 1}}}
       | TOK_LSBLDG P { $$=&commonNode{COMMON, cmd.LSOBJECT, "LSOBJ", []interface{}{$2, 2}}}
       | TOK_LSROOM P { $$=&commonNode{COMMON, cmd.LSOBJECT, "LSOBJ", []interface{}{$2, 3}}}
       | TOK_LSRACK P { $$=&commonNode{COMMON, cmd.LSOBJECT, "LSOBJ", []interface{}{$2, 4}}}
       | TOK_LSDEV P {$$=&commonNode{COMMON, cmd.LSOBJECT, "LSOBJ", []interface{}{$2, 5}}}
       | TOK_LSROW P {$$=&commonNode{COMMON, cmd.LSOBJECT, "LSOBJ", []interface{}{$2, 10}}}
       | TOK_LSTILE P {$$=&commonNode{COMMON, cmd.LSOBJECT, "LSOBJ", []interface{}{$2, 11}}}
       | TOK_LSAC P {$$=&commonNode{COMMON, cmd.LSOBJECT, "LSOBJ", []interface{}{$2, 6}}}
       | TOK_LSPANEL P {$$=&commonNode{COMMON, cmd.LSOBJECT, "LSOBJ", []interface{}{$2, 7}}}
       | TOK_LSWALL P {$$=&commonNode{COMMON, cmd.LSOBJECT, "LSOBJ", []interface{}{$2, 8}}}
       | TOK_LSCAB P {$$=&commonNode{COMMON, cmd.LSOBJECT, "LSOBJ", []interface{}{$2, 9}}}
       | TOK_LSCORRIDOR P {$$=&commonNode{COMMON, cmd.LSOBJECT, "LSOBJ", []interface{}{$2, 12}}}
       | TOK_LSSENSOR P {$$=&commonNode{COMMON, cmd.LSOBJECT, "LSOBJ", []interface{}{$2, 13}}}

       | TOK_TREE P {$$=&commonNode{COMMON, cmd.Tree, "Tree", []interface{}{$2, 0}}}
       | TOK_TREE P TOK_NUM {$$=&commonNode{COMMON, cmd.Tree, "Tree", []interface{}{$2, $3}}}
       | TOK_HIERARCH P {$$=&commonNode{COMMON, cmd.GetHierarchy, "Hierarchy", []interface{}{$2, 1}}}
       | TOK_HIERARCH P TOK_NUM {$$=&commonNode{COMMON, cmd.GetHierarchy, "Hierarchy", []interface{}{$2, $3}}}
       | TOK_UNSET TOK_OCDEL TOK_WORD TOK_WORD {$$=&commonNode{COMMON,UnsetUtil, "Unset",[]interface{}{$2+$3, $4, nil, nil} }}
       | TOK_UNSET TOK_DEREF TOK_WORD TOK_LBLOCK EXPR TOK_RBLOCK {
              v:=&symbolReferenceNode{REFERENCE, $3, $5, nil}; 
              //$$=&assignNode{ASSIGN, v, "deleteValue"}
              $$=&commonNode{COMMON, UnsetUtil, "Unset", []interface{}{"","" ,v, nil}}
              
              }
       | BASH     {$$=$1}
       | TOK_DRAW TOK_LPAREN P TOK_RPAREN {$$=&commonNode{COMMON, cmd.IsEntityDrawable, "IsEntityDrawable", []interface{}{$3}}}
       | TOK_DRAW TOK_LPAREN P TOK_COMMA factor TOK_RPAREN {$$=&commonNode{COMMON, cmd.IsAttrDrawable, "IsAttrDrawable", []interface{}{$3, $5}}}
       |TOK_WORD TOK_EQUAL EXPR {$$=&commonNode{COMMON, cmd.SetEnv, "SetEnv", []interface{}{$1, $3}}}
;

BASH:  TOK_CLR {$$=&commonNode{COMMON, cmd.Clear, "CLR", nil}}
       | TOK_GREP {$$=&commonNode{COMMON, nil, "Grep", nil}}
       | TOK_PRNT NODEGETTER {$$=&commonNode{COMMON, cmd.Print, "Print", $2}}
       | TOK_LSOG {$$=&commonNode{COMMON, cmd.LSOG, "LSOG", nil}}
       | TOK_ENV {$$=&commonNode{COMMON, cmd.Env, "Env", nil}}
       | TOK_PWD {$$=&commonNode{COMMON, cmd.PWD, "PWD", nil}}
       | TOK_EXIT {$$=&commonNode{COMMON, cmd.Exit, "Exit", nil}}
       | TOK_DOC TOK_LINK {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"link"}}}
       | TOK_DOC TOK_UNLINK {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"unlink"}}}
       | TOK_DOC TOK_CLR {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"clear"}}}
       | TOK_DOC {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{""}}}
       | TOK_DOC TOK_LS {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"ls"}}}
       | TOK_DOC TOK_PWD {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"pwd"}}}
       | TOK_DOC TOK_PRNT {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"print"}}}
       | TOK_DOC TOK_CD {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"cd"}}}
       | TOK_DOC TOK_CAM {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"camera"}}}
       | TOK_DOC TOK_UI {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"ui"}}}
       | TOK_DOC TOK_CREATE {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"create"}}}
       | TOK_DOC TOK_GET {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"gt"}}}
       | TOK_DOC TOK_UPDATE {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"update"}}}
       | TOK_DOC TOK_DELETE {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"delete"}}}
       | TOK_DOC TOK_HIERARCH {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"hc"}}}
       | TOK_DOC TOK_WORD {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{$2}}}
       | TOK_DOC TOK_TREE {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"tree"}}}
       | TOK_DOC TOK_IF {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"if"}}}
       | TOK_DOC TOK_WHILE {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"while"}}}
       | TOK_DOC TOK_FOR {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"for"}}}
       | TOK_DOC TOK_UNSET {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"unset"}}}
       | TOK_DOC TOK_SELECT {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"select"}}}
       | TOK_DOC TOK_CMDS {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"cmds"}}}
       | TOK_DOC TOK_LSOG {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"lsog"}}}
       | TOK_DOC TOK_ENV {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"env"}}}
       | TOK_DOC TOK_LSTEN {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"lsten"}}}
       | TOK_DOC TOK_LSSITE {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"lssite"}}} 
       | TOK_DOC TOK_LSBLDG {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"lsbldg"}}}
       | TOK_DOC TOK_LSROOM {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"lsroom"}}}
       | TOK_DOC TOK_LSRACK {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"lsrack"}}}
       | TOK_DOC TOK_LSDEV {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"lsdev"}}}
       | TOK_DOC TOK_OCDEL {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"-"}}}
       | TOK_DOC TOK_DOT TOK_TEMPLATE {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{".template"}}}
       | TOK_DOC TOK_DOT TOK_CMDS {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{".cmds"}}}
       | TOK_DOC TOK_DOT TOK_VAR {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{".var"}}}
       | TOK_DOC TOK_PLUS {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"+"}}}
       | TOK_DOC TOK_EQUAL {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"="}}}
       | TOK_DOC TOK_GREATER {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{">"}}}
       | TOK_DOC TOK_DRAW {$$=&commonNode{COMMON, cmd.Help,"Help", []interface{}{"drawable"} }}
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
        TOK_TENANT TOK_COL P TOK_ATTRSPEC EXPR {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{(replaceOCLICurrPath($3)),cmd.TENANT,map[string]interface{}{"attributes":map[string]interface{}{"color":$5}} }}}
        |TOK_SITE TOK_COL P TOK_ATTRSPEC EXPR {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{(replaceOCLICurrPath($3)),cmd.SITE,map[string]interface{}{"attributes":map[string]interface{}{"orientation":$5}} }}}
        
        |TOK_BLDG TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{(replaceOCLICurrPath($3)),cmd.BLDG,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}} }}}
        

        |TOK_ROOM TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR{$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{(replaceOCLICurrPath($3)),cmd.ROOM,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7, "orientation":$9, "floorUnit":$11}} }}}
        |TOK_ROOM TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{(replaceOCLICurrPath($3)),cmd.ROOM,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7, "orientation":$9}} }}}
        |TOK_ROOM TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC NTORIENTATION {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{(replaceOCLICurrPath($3)),cmd.ROOM,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7, "orientation":$9}} }}}
        |TOK_ROOM TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{(replaceOCLICurrPath($3)),cmd.ROOM,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "template":$7}} }}}

        //|TOK_RACK TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR{$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{(replaceOCLICurrPath($3)),cmd.RACK,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7, "orientation":$9, "template":$11}} }}}
        |TOK_RACK TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR{attr := map[string]interface{}{}; if checkIfTemplate(($7).execute()) == false {attr["size"] = $7} else {attr["template"] = $7};attr["posXY"] = $5; attr["orientation"] = $9;$$=&commonNode{COMMON,cmd.GetOCLIAtrributes, "GetOCAttr",[]interface{}{(replaceOCLICurrPath($3)),cmd.RACK,map[string]interface{}{"attributes":attr} }}}
        
        
        
        |TOK_DEVICE TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{(replaceOCLICurrPath($3)),cmd.DEVICE,map[string]interface{}{"attributes":map[string]interface{}{"slot":$5, "template":$7, "orientation":$9}} }}}
        |TOK_DEVICE TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR {attr := map[string]interface{}{"posU/slot":$5}; res := checkIfTemplate(($7).execute()); if res == false {attr["sizeU"] = $7} else {attr["template"]=$7}; $$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{(replaceOCLICurrPath($3)),cmd.DEVICE,map[string]interface{}{"attributes":attr} }}}



        |TOK_CORIDOR TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{(replaceOCLICurrPath($3)),cmd.CORIDOR, map[string]interface{}{"name":$5, "leftRack":$7, "rightRack":$9, "temperature":$11}}}}
        |TOK_GROUP TOK_COL P TOK_ATTRSPEC EXPR CDORFG { x:=map[string]interface{}{"name":$5,"racks":$6}; $$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{(replaceOCLICurrPath($3)),cmd.GROUP,x}} }
        |TOK_WALL TOK_COL P TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{(replaceOCLICurrPath($3)), cmd.SEPARATOR, map[string]interface{}{"pos1":$5,"pos2":$7}}}}
        
        |TOK_ORPH TOK_COL TOK_DEVICE TOK_COL P TOK_ATTRSPEC EXPR {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{replaceOCLICurrPath($5),cmd.STRAY_DEV,map[string]interface{}{"attributes":map[string]interface{}{"template":$7} } }}}

       //EasyPost syntax
        |E TOK_USE_JSON P {$$=&commonNode{COMMON, cmd.PostObj, "EasyPost", []interface{}{cmd.EntityStrToInt($1),$1, $3}}}

       ; 
OCDEL:  TOK_OCDEL P {$$=&commonNode{COMMON, cmd.DeleteObj, "DeleteObj", []interface{}{replaceOCLICurrPath($2)}}}
        |TOK_OCDEL TOK_SELECT {$$=&commonNode{COMMON, cmd.DeleteSelection, "DeleteSelection", nil }}
;

OCUPDATE:  P TOK_COL TOK_WORD TOK_EQUAL EXPR {val := map[string]interface{}{$3:($5).(node).execute()}; $$=&commonNode{COMMON, cmd.UpdateObj, "UpdateObj", []interface{}{replaceOCLICurrPath($1), val}};if cmd.State.DebugLvl >= 3 {println("Attribute Acquired");}}
           |P TOK_COL TOK_WORD TOK_EQUAL EXPR TOK_ATTRSPEC EXPR {if ($7).(node).getType() == ARRAY {val := map[string]interface{}{$3 : map[string]interface{}{"reserved":($5).(node).execute(), "technical":($7).(node).execute()}}; $$=&commonNode{COMMON, cmd.UpdateObj, "UpdateObj", []interface{}{replaceOCLICurrPath($1), val}}} else {val := map[string]interface{}{$3:($5).(node).execute()}; $$=&commonNode{COMMON, cmd.RecursivePatch, "RecursiveUpdateObj", []interface{}{replaceOCLICurrPath($1), val, $7}};if cmd.State.DebugLvl >= 3 {println("Attribute Acquired");}}}
           //|P TOK_COL TOK_WORD TOK_EQUAL EXPR TOK_ATTRSPEC EXPR {}
;

OCGET: TOK_EQUAL P {$$=&commonNode{COMMON, cmd.GetObject, "GetObject", []interface{}{replaceOCLICurrPath($2)}}}
;

GETOBJS:      P TOK_COMMA GETOBJS {x := make([]string,0); x = append(x, formActualPath($1)); x = append(x, $3...); $$=x}
              |P {$$=[]string{formActualPath($1)}}
              //| TOK_WORD {$$=[]string{cmd.State.CurrPath+"/"+$1}}
              ;

OCCHOOSE: TOK_EQUAL TOK_LBRAC GETOBJS TOK_RBRAC {$$=&commonNode{COMMON, cmd.SetClipBoard, "setCB", []interface{}{&$3}}; println("Selection made!")}
;

OCDOT:      //TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL WORDORNUM {$$=&assignNode{ASSIGN, $4, dCatchNodePtr}}
            //|TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_QUOT STRARG TOK_QUOT{$$=&assignNode{ASSIGN, $4, &strNode{STR, $7}}}
            //TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_LPAREN WNARG TOK_RPAREN {$$=&assignNode{ASSIGN, $4, &arrNode{ARRAY, len($7),retNodeArray($7)}}}
            TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN K TOK_RPAREN {$$=&assignNode{ASSIGN, $4, ($8).(node).execute()}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN Q  TOK_RPAREN {$$=&assignNode{ASSIGN, $4, ($8).(node).execute()}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN TOK_PLUS OCCR  TOK_RPAREN {$$=&assignNode{ASSIGN, $4, ($9).(node).execute()}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN OCDEL  TOK_RPAREN {$$=&assignNode{ASSIGN, $4, ($8).(node).execute()}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN OCUPDATE  TOK_RPAREN {$$=&assignNode{ASSIGN, $4, ($8).(node).execute()}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN OCGET  TOK_RPAREN {$$=&assignNode{ASSIGN, $4, ($8).(node).execute()}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN OCCHOOSE  TOK_RPAREN {$$=&assignNode{ASSIGN, $4, ($8).(node).execute()}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN OCSEL  TOK_RPAREN {$$=&assignNode{ASSIGN, $4, ($8).(node).execute()}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL EXPR {val := ($6).(node).execute(); if ($6).(node).getType() == ARITHMETIC && val == nil {$$=nil} else {$$=&assignNode{ASSIGN, $4, val}}}
            |TOK_DOT TOK_CMDS TOK_COL P {$$=&commonNode{COMMON, cmd.LoadFile, "Load", []interface{}{$4, "cmd"}};}
            |TOK_DOT TOK_TEMPLATE TOK_COL P {$$=&commonNode{COMMON, cmd.LoadTemplate, "Load", []interface{}{$4, "template"}}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL Q {$$=&assignNode{ASSIGN, $4, $6}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL K {$$=&assignNode{ASSIGN, $4, $6}}
            //|TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL OCLISYNTX {$$=&assignNode{ASSIGN, $4, $6}}
            |TOK_DEREF TOK_WORD {$$=&symbolReferenceNode{REFERENCE, $2, &numNode{NUM,0}, nil}}
              

            |TOK_DEREF TOK_WORD TOK_LBLOCK EXPR TOK_RBLOCK {$$=&symbolReferenceNode{REFERENCE, $2, $4, nil}}
            |TOK_DEREF TOK_WORD TOK_LBLOCK EXPR TOK_RBLOCK TOK_EQUAL EXPR {v:=&symbolReferenceNode{REFERENCE, $2, $4, nil}; $$=&assignNode{ASSIGN, v, $7} }
            |TOK_DEREF TOK_WORD TOK_LBLOCK EXPR TOK_RBLOCK TOK_LBLOCK EXPR TOK_RBLOCK {$$=&symbolReferenceNode{REFERENCE, $2, /*&numNode{NUM,$4}*/$4, /*&strNode{STR, $7}*/ $7}}
            |TOK_DEREF TOK_WORD TOK_EQUAL EXPR {n:=&symbolReferenceNode{REFERENCE, $2, &numNode{NUM,0}, nil};$$=&assignNode{ASSIGN,n,$4 }}
;

OCSEL:      TOK_SELECT {$$=&commonNode{COMMON, cmd.ShowClipBoard, "select", nil};}
            |TOK_SELECT TOK_DOT TOK_WORD TOK_EQUAL EXPR {/*x := $3+"="+$5;*/ val:=($5).(node).execute(); x:=map[string]interface{}{$3:val}; $$=&commonNode{COMMON, cmd.UpdateSelection, "UpdateSelect", []interface{}{x}};}
;

HANDLEUI: TOK_UI TOK_DOT TOK_WORD TOK_EQUAL EXPR {$$=&commonNode{COMMON, cmd.HandleUI, "HandleUnity", []interface{}{"ui", $3, ($5).(node).execute()}}}
          |TOK_CAM TOK_DOT TOK_WORD TOK_EQUAL EXPR TOK_ATTRSPEC EXPR {if ($5).(node).getType() != ARRAY || ($7).(node).getType() != ARRAY {$$=nil;} else {$$=&commonNode{COMMON, cmd.HandleUI, "HandleUnity", []interface{}{"camera", $3, ($5).(node).execute(), ($7).(node).execute()}}}}
          |TOK_CAM TOK_DOT TOK_WORD TOK_EQUAL EXPR {if ($5).(node).getType() != ARRAY && $3 != "wait" {$$=nil;} else  {$$=&commonNode{COMMON, cmd.HandleUI, "HandleUnity", []interface{}{"camera", $3, ($5).(node).execute()}}}}
          |TOK_GREATER P {$$=&commonNode{COMMON, cmd.FocusUI, "Focus", []interface{}{$2}}}
          ;

HANDLELINKS:  TOK_LINK TOK_COL P TOK_ATTRSPEC P {$$=&commonNode{COMMON, cmd.LinkObject, "LinkObject", []interface{}{$3, $5}}}
              |TOK_LINK TOK_COL P TOK_ATTRSPEC P TOK_ATTRSPEC EXPR {$$=&commonNode{COMMON, cmd.LinkObject, "LinkObject", []interface{}{$3, $5, $7}}}
              |TOK_UNLINK TOK_COL P {$$=&commonNode{COMMON, cmd.UnlinkObject, "UnlinkObject", []interface{}{$3}}}
              |TOK_UNLINK TOK_COL P TOK_ATTRSPEC P {$$=&commonNode{COMMON, cmd.UnlinkObject, "UnlinkObject", []interface{}{$3,$5}}}
            ;
//For making array types
WNARG: EXPR TOK_COMMA WNARG {x:=[]interface{}{$1}; $$=append(x, $3...)}
       |EXPR  {x:=[]interface{}{$1}; $$=x}
       ;

FUNC: TOK_WORD TOK_LPAREN TOK_RPAREN TOK_LBRAC st2 TOK_RBRAC {x:=&funcNode{FUNC, $5}; $$=&assignNode{ASSIGN, $1, x};}
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
NTORIENTATION: TOK_PLUS TOK_WORD TOK_PLUS TOK_WORD {$$=$1+$2+$3+$4; dCatchPtr = $1+$2+$3+$4; dCatchNodePtr=&strNode{STR, $1+$2+$3+$4}}
              |TOK_PLUS TOK_WORD TOK_OCDEL TOK_WORD {$$=$1+$2+$3+$4; dCatchPtr = $1+$2+$3+$4; dCatchNodePtr=&strNode{STR, $1+$2+$3+$4}}
              |TOK_OCDEL TOK_WORD TOK_OCDEL TOK_WORD {$$=$1+$2+$3+$4; dCatchPtr = $1+$2+$3+$4; dCatchNodePtr=&strNode{STR, $1+$2+$3+$4}}
              |TOK_OCDEL TOK_WORD TOK_PLUS TOK_WORD {$$=$1+$2+$3+$4; dCatchPtr = $1+$2+$3+$4; dCatchNodePtr=&strNode{STR, $1+$2+$3+$4}}
              ;