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
              if input[idx].(string) == "false" {
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
              }
       }
       return res
}

//This func helps to build the correct map[str]interface{}
//based on the input string x
func retMapInf(x string, y interface{}) map[string]interface{} {
       switch x {
              case "id", "name", "category", "parentID", 
              "description", "domain", "parentid", "parentId":
                     return map[string]interface{}{x:y}
              
              default:
              return map[string]interface{}{
                     "attributes":map[string]interface{}{x:y}}
       }
}

func resMap(x *string) map[string]interface{} {
       resarr := strings.Split(*x, "=")
       res := make(map[string]interface{})
       attrs := make(map[string]string)

	for i := 0; i+1 < len(resarr); {
              if i+1 < len(resarr) {
                     switch resarr[i] {
                            case "id", "name", "category", "parentID", 
                            "description", "domain", "parentid", "parentId":
                                   res[resarr[i]] = resarr[i+1]
                            
                            default:
                            attrs[resarr[i]] = resarr[i+1]
                     }
			i += 2
		}
	}
       res["attributes"] = attrs
       return res
}

func replaceOCLICurrPath(x string) string {
       return strings.Replace(x, "_", cmd.State.CurrPath, 1)
}

//Gets node from Tree Hierarchy using a map[string]interface
func getNodeFromMapInf(x map[string]interface{}) *cmd.Node {
       ent := x["category"]
       pid,_ := x["parentId"].(string)
       id,_ := x["id"].(string)

       entNum := cmd.EntityStrToInt(ent.(string))
       nodes := cmd.GetNodes(&cmd.State.TreeHierarchy, entNum)
       for i := range nodes {
              if nodes[i].PID == pid && nodes[i].ID == id {
                     return nodes[i]
              }
       }
       return nil
}

func genNodeFromCommonRes(x node) node {
       val := x.execute()
       switch val.(type) {
              case string:
              return &strNode{STR, val.(string)}

              case []map[string]interface{}:
              return &jsonObjArrNode{JSONND, 
              len(val.([]map[string]interface{})), val.([]map[string]interface{})}
       }
       return nil
} 
%}

%union {
  n int
  s string
  sarr []string
  ast *ast
  node node
  nodeArr []node
  elifArr []elifNode
  arr []interface{}
  mapArr []map[int]interface{}
}

%token <n> TOK_NUM
%token <s> TOK_WORD TOK_TENANT TOK_SITE TOK_BLDG TOK_ROOM
%token <s> TOK_RACK TOK_DEVICE TOK_SUBDEVICE TOK_SUBDEVICE1
%token <s> TOK_CORIDOR TOK_GROUP TOK_WALL
%token <s> TOK_ATTR TOK_PLUS TOK_OCDEL TOK_BOOL
%token
       TOK_CREATE TOK_GET TOK_UPDATE TOK_DELETE TOK_SEARCH
       TOK_BASHTYPE TOK_EQUAL TOK_CMDFLAG TOK_SLASH 
       TOK_EXIT TOK_DOC TOK_CD TOK_PWD
       TOK_CLR TOK_GREP TOK_LS TOK_TREE
       TOK_LSOG TOK_LSTEN TOK_LSSITE TOK_LSBLDG
       TOK_LSROOM TOK_LSRACK TOK_LSDEV
       TOK_LSSUBDEV TOK_LSSUBDEV1 TOK_OCBLDG TOK_OCDEV
       TOK_OCRACK TOK_OCROOM TOK_ATTRSPEC TOK_OCSITE TOK_OCTENANT
       TOK_OCSDEV TOK_OCSDEV1 TOK_COL TOK_SELECT TOK_LBRAC TOK_RBRAC
       TOK_COMMA TOK_DOT TOK_CMDS TOK_TEMPLATE TOK_VAR TOK_DEREF
       TOK_SEMICOL TOK_IF TOK_FOR TOK_WHILE
       TOK_ELSE TOK_LBLOCK TOK_RBLOCK
       TOK_LPAREN TOK_RPAREN TOK_OR TOK_AND TOK_IN TOK_PRNT TOK_QUOT
       TOK_NOT TOK_DIV TOK_MULT TOK_GREATER TOK_LESS TOK_THEN TOK_FI TOK_DONE
       TOK_UNSET TOK_ELIF TOK_DO TOK_LEN
       TOK_OCGROUP TOK_OCWALL TOK_OCCORIDOR
       
%type <s> F E P P1 WORDORNUM STRARG CDORFG
%type <arr> WNARG 
%type <sarr> GETOBJS
%type <elifArr> EIF
%type <node> OCSEL OCLISYNTX OCDEL OCGET NT_CREATE NT_GET NT_DEL 
%type <node> NT_UPDATE K Q BASH OCUPDATE OCCHOOSE OCCR OCDOT
%type <node> EXPR REL OPEN_STMT CTRL nex factor unary EQAL term
%type <node> stmnt JOIN
%type <node> st2 FUNC 
%left TOK_MULT TOK_OCDEL TOK_DIV TOK_PLUS
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
                     println("WE MUST BE HERE")
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
       |term TOK_DIV unary {$$=&arithNode{ARITHMETIC, "/", $1, $3}}
       |unary {$$=$1}
       ;

unary: TOK_NOT unary {$$=&boolOpNode{BOOLOP, "!", $2}}
       |TOK_OCDEL unary {left := &numNode{NUM, 0};$$=&arithNode{ARITHMETIC, "-",left,$2 }}
       |factor {$$=$1}
       ;

factor: TOK_LPAREN EXPR TOK_RPAREN {$$=$2}
       |TOK_NUM {$$=&numNode{NUM, $1}}
       |TOK_DEREF TOK_WORD TOK_LBLOCK TOK_ATTR TOK_RBLOCK {$$=&symbolReferenceNode{REFERENCE, $2, &strNode{STR,$4}, nil}}
       |TOK_DEREF TOK_WORD TOK_LBLOCK TOK_WORD TOK_RBLOCK {$$=&symbolReferenceNode{REFERENCE, $2, &strNode{STR,$4}, nil}}
       |TOK_DEREF TOK_WORD {$$=&symbolReferenceNode{REFERENCE, $2, &numNode{NUM,0}, nil}}
       |TOK_DEREF TOK_WORD TOK_LBLOCK TOK_NUM TOK_RBLOCK {$$=&symbolReferenceNode{REFERENCE, $2, &numNode{NUM,$4}, nil}}
       |TOK_LEN TOK_LPAREN TOK_WORD TOK_RPAREN {x:=&symbolReferenceNode{REFERENCE, $3, &numNode{NUM, -1}, nil};$$=&numNode{NUM, x.execute().(int)}}
       |TOK_WORD {$$=&symbolReferenceNode{REFERENCE, $1,&numNode{NUM,0}, nil}}
       |TOK_QUOT STRARG TOK_QUOT {$$=&strNode{STR, $2}}
       |TOK_BOOL {var x bool;if $1=="false"{x = false}else{x=true};$$=&boolNode{BOOL, x}}
       ;

K: NT_CREATE     {if cmd.State.DebugLvl >= 3 {println("@State start");}}
       | NT_GET
       | NT_UPDATE 
       | NT_DEL 
;

NT_CREATE: TOK_CREATE E P F {cmd.Disp(resMap(&$4)); $$=&commonNode{COMMON, cmd.PostObj, "PostObj", []interface{}{cmd.EntityStrToInt($2),$2, resMap(&$4)}}}
;

NT_GET: TOK_GET P {$$=&commonNode{COMMON, cmd.GetObject, "GetObject", []interface{}{$2}}}
       | TOK_GET E F {/*cmd.Disp(resMap(&$4)); */$$=&commonNode{COMMON, cmd.SearchObjects, "SearchObjects", []interface{}{$2, resMap(&$3)}} }
;

NT_UPDATE: TOK_UPDATE P F {$$=&commonNode{COMMON, cmd.UpdateObj, "UpdateObj", []interface{}{$2, resMap(&$3)}}}
;

NT_DEL: TOK_DELETE P {if cmd.State.DebugLvl >= 3 {println("@State NT_DEL");}; $$=&commonNode{COMMON, cmd.DeleteObj, "DeleteObj", []interface{}{$2}}}
;

E:     TOK_TENANT 
       | TOK_SITE 
       | TOK_BLDG 
       | TOK_ROOM 
       | TOK_RACK 
       | TOK_DEVICE 
       | TOK_SUBDEVICE 
       | TOK_SUBDEVICE1 
;


WORDORNUM: TOK_WORD {$$=$1; dCatchPtr = $1; dCatchNodePtr=&strNode{STR, $1}}
           |TOK_NUM {x := strconv.Itoa($1);$$=x;dCatchPtr = $1; dCatchNodePtr=&numNode{NUM, $1}}
           |TOK_PLUS TOK_WORD TOK_PLUS TOK_WORD {$$=$1+$2+$3+$4; dCatchPtr = $1+$2+$3+$4; dCatchNodePtr=&strNode{STR, $1+$2+$3+$4}}
           |TOK_PLUS TOK_WORD TOK_OCDEL TOK_WORD {$$=$1+$2+$3+$4; dCatchPtr = $1+$2+$3+$4; dCatchNodePtr=&strNode{STR, $1+$2+$3+$4}}
           |TOK_OCDEL TOK_WORD TOK_OCDEL TOK_WORD {$$=$1+$2+$3+$4; dCatchPtr = $1+$2+$3+$4; dCatchNodePtr=&strNode{STR, $1+$2+$3+$4}}
           |TOK_OCDEL TOK_WORD TOK_PLUS TOK_WORD {$$=$1+$2+$3+$4; dCatchPtr = $1+$2+$3+$4; dCatchNodePtr=&strNode{STR, $1+$2+$3+$4}}
           |TOK_BOOL {var x bool;if $1=="false"{x = false}else{x=true};dCatchPtr = x; dCatchNodePtr=&boolNode{BOOL, x}}
           ;

F:     TOK_ATTR TOK_EQUAL WORDORNUM F {$$=string($1+"="+$3+"="+$4); if cmd.State.DebugLvl >= 3 {println("So we got: ", $$);}}
       | TOK_ATTR TOK_EQUAL WORDORNUM {$$=$1+"="+$3}
;


P:     P1
       | TOK_SLASH P1 {$$="/"+$2}
;

P1:    TOK_WORD TOK_SLASH P1 {$$=$1+"/"+$3}
       | TOK_WORD {$$=$1}
       | TOK_DOT TOK_DOT TOK_SLASH P1 {$$="../"+$4}
       | TOK_WORD TOK_DOT TOK_WORD {$$=$1+"."+$3}
       | TOK_DOT TOK_DOT {$$=".."}
       | TOK_OCDEL {$$="-"}
       | TOK_DEREF TOK_WORD {$$=""}
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
       | TOK_LSSUBDEV P { $$=&commonNode{COMMON, cmd.LSOBJECT, "LSOBJ", []interface{}{$2, 6}}}
       | TOK_LSSUBDEV1 P {$$=&commonNode{COMMON, cmd.LSOBJECT, "LSOBJ", []interface{}{$2, 7}}}
       | TOK_TREE P {$$=&commonNode{COMMON, cmd.Tree, "Tree", []interface{}{$2, 0}}}
       | TOK_TREE P TOK_NUM {$$=&commonNode{COMMON, cmd.Tree, "Tree", []interface{}{$2, $3}}}
       | TOK_UNSET TOK_OCDEL TOK_WORD TOK_WORD {$$=&commonNode{COMMON,UnsetUtil, "Unset",[]interface{}{$2+$3, $4} }}
       | BASH     {$$=$1}
;

BASH:  TOK_CLR {$$=&commonNode{COMMON, nil, "CLR", nil}}
       | TOK_GREP {$$=&commonNode{COMMON, nil, "Grep", nil}}
       | TOK_PRNT TOK_QUOT STRARG TOK_QUOT{$$=&commonNode{COMMON, cmd.Print, "Print", []interface{}{$3}}}
       | TOK_LSOG {$$=&commonNode{COMMON, cmd.LSOG, "LSOG", nil}}
       | TOK_PWD {$$=&commonNode{COMMON, cmd.PWD, "PWD", nil}}
       | TOK_EXIT {$$=&commonNode{COMMON, cmd.Exit, "Exit", nil}}
       | TOK_DOC {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{""}}}
       | TOK_DOC TOK_LS {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"ls"}}}
       | TOK_DOC TOK_PWD {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"pwd"}}}
       | TOK_DOC TOK_PRNT {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"print"}}}
       | TOK_DOC TOK_CD {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"cd"}}}
       | TOK_DOC TOK_CREATE {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"create"}}}
       | TOK_DOC TOK_GET {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"gt"}}}
       | TOK_DOC TOK_UPDATE {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"update"}}}
       | TOK_DOC TOK_DELETE {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"delete"}}}
       | TOK_DOC TOK_WORD {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{$2}}}
       | TOK_DOC TOK_TREE {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"tree"}}}
       | TOK_DOC TOK_IF {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"if"}}}
       | TOK_DOC TOK_WHILE {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"while"}}}
       | TOK_DOC TOK_FOR {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"for"}}}
       | TOK_DOC TOK_UNSET {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"unset"}}}
       | TOK_DOC TOK_SELECT {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"select"}}}
       | TOK_DOC TOK_CMDS {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"cmds"}}}
       | TOK_DOC TOK_LSOG {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"lsog"}}}
       | TOK_DOC TOK_LSTEN {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"lsten"}}}
       | TOK_DOC TOK_LSSITE {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"lssite"}}} 
       | TOK_DOC TOK_LSBLDG {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"lsbldg"}}}
       | TOK_DOC TOK_LSROOM {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"lsroom"}}}
       | TOK_DOC TOK_LSRACK {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"lsrack"}}}
       | TOK_DOC TOK_LSDEV {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"lsdev"}}}
       | TOK_DOC TOK_LSSUBDEV {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"lssubdev"}}}
       | TOK_DOC TOK_LSSUBDEV1 {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"lssubdev1"}}}
       | TOK_DOC TOK_OCDEL {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"-"}}}
       | TOK_DOC TOK_DOT TOK_TEMPLATE {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{".template"}}}
       | TOK_DOC TOK_DOT TOK_CMDS {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{".cmds"}}}
       | TOK_DOC TOK_DOT TOK_VAR {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{".var"}}}
       | TOK_DOC TOK_PLUS {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"+"}}}
       | TOK_DOC TOK_EQUAL {$$=&commonNode{COMMON, cmd.Help, "Help", []interface{}{"="}}}

;

OCLISYNTX:  TOK_PLUS OCCR {$$=$2}
            |OCDEL {$$=$1}
            |OCUPDATE {$$=$1}
            |OCGET {$$=$1}
            |OCCHOOSE {$$=$1}
            |OCDOT {$$=$1}
            |OCSEL {$$=$1;}
            ;


OCCR:   TOK_OCTENANT TOK_COL P TOK_ATTRSPEC WORDORNUM {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{cmd.StrToStack(replaceOCLICurrPath($3)),cmd.TENANT,map[string]interface{}{"attributes":map[string]interface{}{"color":$5}} ,rlPtr}}}
        |TOK_TENANT TOK_COL P TOK_ATTRSPEC WORDORNUM {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{cmd.StrToStack(replaceOCLICurrPath($3)),cmd.TENANT,map[string]interface{}{"attributes":map[string]interface{}{"color":$5}} ,rlPtr}}}
        |TOK_OCSITE TOK_COL P TOK_ATTRSPEC WORDORNUM {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{cmd.StrToStack(replaceOCLICurrPath($3)),cmd.SITE,map[string]interface{}{"attributes":map[string]interface{}{"orientation":$5}} ,rlPtr}}}
        |TOK_SITE TOK_COL P TOK_ATTRSPEC WORDORNUM {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{cmd.StrToStack(replaceOCLICurrPath($3)),cmd.SITE,map[string]interface{}{"attributes":map[string]interface{}{"orientation":$5}} ,rlPtr}}}
        |TOK_OCBLDG TOK_COL P TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{cmd.StrToStack(replaceOCLICurrPath($3)),cmd.BLDG,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}} ,rlPtr}}}
        |TOK_BLDG TOK_COL P TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{cmd.StrToStack(replaceOCLICurrPath($3)),cmd.BLDG,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}} ,rlPtr}}}
        |TOK_OCROOM TOK_COL P TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{cmd.StrToStack(replaceOCLICurrPath($3)),cmd.ROOM,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}} ,rlPtr}}}
        |TOK_ROOM TOK_COL P TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{cmd.StrToStack(replaceOCLICurrPath($3)),cmd.ROOM,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}} ,rlPtr}}}
        |TOK_OCRACK TOK_COL P TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{cmd.StrToStack(replaceOCLICurrPath($3)),cmd.RACK,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}} ,rlPtr}}}
        |TOK_RACK TOK_COL P TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{cmd.StrToStack(replaceOCLICurrPath($3)),cmd.RACK,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}} ,rlPtr}}}
        |TOK_OCDEV TOK_COL P TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{cmd.StrToStack(replaceOCLICurrPath($3)),cmd.DEVICE,map[string]interface{}{"attributes":map[string]interface{}{"slot":$5, "sizeUnit":$7}} ,rlPtr}}}
        |TOK_DEVICE TOK_COL P TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{cmd.StrToStack(replaceOCLICurrPath($3)),cmd.DEVICE,map[string]interface{}{"attributes":map[string]interface{}{"slot":$5, "sizeUnit":$7}} ,rlPtr}}}
        |TOK_OCCORIDOR TOK_COL P TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{cmd.StrToStack(replaceOCLICurrPath($3)),cmd.CORRIDOR, map[string]interface{}{"name":$5, "leftRack":$7, "rightRack":$9, "temperature":$11},rlPtr}}}
        |TOK_CORIDOR TOK_COL P TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{cmd.StrToStack(replaceOCLICurrPath($3)),cmd.CORRIDOR, map[string]interface{}{"name":$5, "leftRack":$7, "rightRack":$9, "temperature":$11},rlPtr}}}
        |TOK_OCGROUP TOK_COL P TOK_ATTRSPEC WORDORNUM CDORFG { x:=map[string]interface{}{"name":$5,"racks":$6}; $$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{cmd.StrToStack(replaceOCLICurrPath($3)),cmd.GROUP,x,rlPtr}} }
        |TOK_GROUP TOK_COL P TOK_ATTRSPEC WORDORNUM CDORFG { x:=map[string]interface{}{"name":$5,"racks":$6}; $$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{cmd.StrToStack(replaceOCLICurrPath($3)),cmd.GROUP,x,rlPtr}} }
        |TOK_OCWALL TOK_COL P TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{cmd.StrToStack(replaceOCLICurrPath($3)), cmd.WALL, map[string]interface{}{"name":$5, "pos1":$7,"pos2":$9},rlPtr}}}
        |TOK_WALL TOK_COL P TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM {$$=&commonNode{COMMON, cmd.GetOCLIAtrributes, "GetOCAttr", []interface{}{cmd.StrToStack(replaceOCLICurrPath($3)), cmd.WALL, map[string]interface{}{"name":$5, "pos1":$7,"pos2":$9},rlPtr}}}
       ; 
OCDEL:  TOK_OCDEL P {$$=&commonNode{COMMON, cmd.DeleteObj, "DeleteObj", []interface{}{replaceOCLICurrPath($2)}}}
;

OCUPDATE:  P TOK_COL TOK_ATTR TOK_EQUAL WORDORNUM {val := $3+"="+$5; $$=&commonNode{COMMON, cmd.UpdateObj, "UpdateObj", []interface{}{replaceOCLICurrPath($1), resMap(&val)}};if cmd.State.DebugLvl >= 3 {println("Attribute Acquired");}}
;

OCGET: TOK_EQUAL P {$$=&commonNode{COMMON, cmd.GetObject, "GetObject", []interface{}{replaceOCLICurrPath($2)}}}
;

GETOBJS:      TOK_WORD TOK_COMMA GETOBJS {x := make([]string,0); x = append(x, cmd.State.CurrPath+"/"+$1); x = append(x, $3...); $$=x}
              | TOK_WORD {$$=[]string{cmd.State.CurrPath+"/"+$1}}
              ;

OCCHOOSE: TOK_EQUAL TOK_LBRAC GETOBJS TOK_RBRAC {$$=&commonNode{COMMON, cmd.SetClipBoard, "setCB", []interface{}{&$3}}; println("Selection made!")}
;

OCDOT:      TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL WORDORNUM {$$=&assignNode{ASSIGN, $4, dCatchNodePtr}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_QUOT STRARG TOK_QUOT{$$=&assignNode{ASSIGN, $4, &strNode{STR, $7}}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_LPAREN WNARG TOK_RPAREN {$$=&assignNode{ASSIGN, $4, &arrNode{ARRAY, len($7),retNodeArray($7)}}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN K TOK_RPAREN {$$=&assignNode{ASSIGN, $4, ($8).(node).execute()}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN Q  TOK_RPAREN {$$=&assignNode{ASSIGN, $4, ($8).(node).execute()}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN TOK_PLUS OCCR  TOK_RPAREN {$$=&assignNode{ASSIGN, $4, ($9).(node).execute()}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN OCDEL  TOK_RPAREN {$$=&assignNode{ASSIGN, $4, ($8).(node).execute()}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN OCUPDATE  TOK_RPAREN {$$=&assignNode{ASSIGN, $4, ($8).(node).execute()}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN OCGET  TOK_RPAREN {$$=&assignNode{ASSIGN, $4, ($8).(node).execute()}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN OCCHOOSE  TOK_RPAREN {$$=&assignNode{ASSIGN, $4, ($8).(node).execute()}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN OCSEL  TOK_RPAREN {$$=&assignNode{ASSIGN, $4, ($8).(node).execute()}}
            |TOK_DOT TOK_CMDS TOK_COL P {$$=&commonNode{COMMON, cmd.LoadFile, "Load", []interface{}{$4}};}
            |TOK_DOT TOK_TEMPLATE TOK_COL P {$$=&commonNode{COMMON, cmd.LoadFile, "Load", []interface{}{$4}}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL Q {$$=&assignNode{ASSIGN, $4, $6}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL K {$$=&assignNode{ASSIGN, $4, $6}}
            |TOK_DOT TOK_VAR TOK_COL TOK_WORD TOK_EQUAL OCLISYNTX {$$=&assignNode{ASSIGN, $4, $6}}
            |TOK_DEREF TOK_WORD {$$=&symbolReferenceNode{REFERENCE, $2, &numNode{NUM,0}, nil}}
            |TOK_DEREF TOK_WORD TOK_LBLOCK TOK_ATTR TOK_RBLOCK TOK_EQUAL EXPR 
            {println("YAO");
            y:=&symbolReferenceNode{REFERENCE, $2, &strNode{STR, $4}, nil}; 
            x:=&assignNode{ASSIGN, y, $7};
            mp:=&symbolReferenceNode{REFERENCE, $2, &numNode{NUM, -1},nil}; 
            if mp.execute() == nil { // Edge case for ranged iterator loops
              z:=&commonNode{COMMON, cmd.UpdateObj, "UpdateObj", []interface{}{mp, $4, $7, true}}
              $$=&ast{ASSIGN, []node{x, z}}
              
              } else {

                   q:=getNodeFromMapInf(mp.execute().(map[string]interface{})) ;
            z:=&commonNode{COMMON,cmd.UpdateObj, "UpdateObj", []interface{}{q.Path, retMapInf($4,($7).execute() ),false }};
            $$=&ast{ASSIGN, []node{x, z}}}
            }
              
            
            
            
            |TOK_DEREF TOK_WORD TOK_LBLOCK TOK_ATTR TOK_RBLOCK {$$=&symbolReferenceNode{REFERENCE, $2, &strNode{STR, $4}, nil}}
            |TOK_DEREF TOK_WORD TOK_LBLOCK TOK_NUM TOK_RBLOCK {$$=&symbolReferenceNode{REFERENCE, $2, &numNode{NUM,$4}, nil}}
            |TOK_DEREF TOK_WORD TOK_LBLOCK TOK_NUM TOK_RBLOCK TOK_EQUAL EXPR {v:=&symbolReferenceNode{REFERENCE, $2, &numNode{NUM,$4}, nil}; $$=&assignNode{ASSIGN, v, $7} }
            |TOK_DEREF TOK_WORD TOK_LBLOCK TOK_NUM TOK_RBLOCK TOK_LBLOCK TOK_ATTR TOK_RBLOCK {$$=&symbolReferenceNode{REFERENCE, $2, &numNode{NUM,$4}, &strNode{STR, $7}}}
            |TOK_DEREF TOK_WORD TOK_EQUAL EXPR {n:=&symbolReferenceNode{REFERENCE, $2, &numNode{NUM,0}, nil};$$=&assignNode{ASSIGN,n,$4 }}
;

OCSEL:      TOK_SELECT {$$=&commonNode{COMMON, cmd.ShowClipBoard, "select", nil};}
            |TOK_SELECT TOK_DOT TOK_ATTR TOK_EQUAL TOK_WORD {x := $3+"="+$5; $$=&commonNode{COMMON, cmd.UpdateSelection, "UpdateSelect", []interface{}{resMap(&x)}};}
;

STRARG: WORDORNUM STRARG {if $2 != "" {$$=$1+" "+$2} else {$$=$1};}
       | {$$=""}
;

WNARG: WORDORNUM WNARG {x:=[]interface{}{$1}; $$=append(x, $2...)}
       |TOK_QUOT WORDORNUM TOK_QUOT WNARG {x:=[]interface{}{$2}; $$=append(x, $4...)}
       | {$$=nil}
       ;

FUNC: TOK_WORD TOK_LPAREN TOK_RPAREN TOK_LBRAC st2 TOK_RBRAC {$$=nil;funcTable[$1]=&funcNode{FUNC, $5}}
       |TOK_WORD {x:=funcTable[$1]; if _,ok:=x.(node); ok {$$=x.(node)}else{$$=nil};}


//Child devices of rack for group 
//Since the OCLI syntax defines no limit
//for the number of devices 
//a NonTerminal state is neccessary
CDORFG: TOK_ATTRSPEC WORDORNUM CDORFG {x:=$2; $$=x+","+$3}
       | {$$=""}
       ;
%%