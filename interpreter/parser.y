%{
package main
import (
cmd "cli/controllers"
"strings"
"strconv"
)

var dynamicVarLimit = []int{0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15}
var dynamicMap = make(map[string]int)
var dynamicSymbolTable = make(map[int]interface{})
var dCatchPtr interface{}
var varCtr = 0

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
%}

%union {
  n int
  s string
  sarr []string
}

%token <s> TOK_BOOL
%token <n> TOK_NUM
%token <s> TOK_WORD
%token <s> TOK_TENANT TOK_SITE TOK_BLDG TOK_ROOM
%token <s> TOK_RACK TOK_DEVICE TOK_SUBDEVICE TOK_SUBDEVICE1
%token <s> TOK_ATTR TOK_PLUS TOK_OCDEL
%token
       TOK_CREATE TOK_GET TOK_UPDATE
       TOK_DELETE TOK_SEARCH
       TOK_BASHTYPE TOK_EQUAL 
       TOK_CMDFLAG TOK_SLASH 
       TOK_EXIT TOK_DOC
       TOK_CD TOK_PWD
       TOK_CLR TOK_GREP TOK_LS TOK_TREE
       TOK_LSOG TOK_LSTEN TOK_LSSITE TOK_LSBLDG
       TOK_LSROOM TOK_LSRACK TOK_LSDEV
       TOK_LSSUBDEV TOK_LSSUBDEV1
       TOK_OCBLDG TOK_OCDEV
       TOK_OCRACK TOK_OCROOM TOK_ATTRSPEC
       TOK_OCSITE TOK_OCTENANT
       TOK_OCSDEV TOK_OCSDEV1 TOK_OCPSPEC
       TOK_SELECT TOK_LBRAC TOK_RBRAC
       TOK_COMMA TOK_DOT TOK_CMDS
       TOK_TEMPLATE TOK_VAR TOK_DEREF
       TOK_SEMICOL TOK_IF TOK_FOR TOK_WHILE
       TOK_ELSE TOK_LBLOCK TOK_RBLOCK
       TOK_LPAREN TOK_RPAREN TOK_OR TOK_AND
       TOK_NOT TOK_DIV TOK_MULT TOK_GREATER
       TOK_LESS TOK_THEN TOK_FI TOK_DONE
%type <s> F E P P1 ORIENTN WORDORNUM
%type <s> NT_CREATE NT_DEL NT_GET NT_UPDATE
%type <sarr> GETOBJS


%%

start: stmnt
       |stmnt TOK_SEMICOL start
       |CTRL
;

stmnt: K
       |Q
       |OCLISYNTX
       |
;

CTRL: OPEN_STMT
       |CLSD_STMT
       ;

OPEN_STMT:    TOK_IF TOK_LBLOCK EXPR TOK_RBLOCK TOK_THEN stmnt TOK_FI
              |TOK_IF TOK_LBLOCK EXPR TOK_RBLOCK TOK_THEN OPEN_STMT TOK_FI
              |TOK_IF TOK_LBLOCK EXPR TOK_RBLOCK TOK_THEN CLSD_STMT TOK_ELSE OPEN_STMT TOK_FI
              |TOK_WHILE TOK_LPAREN EXPR TOK_RPAREN OPEN_STMT TOK_DONE
              ;

CLSD_STMT: stmnt
              | TOK_IF TOK_LPAREN EXPR TOK_RPAREN TOK_THEN CLSD_STMT TOK_ELSE CLSD_STMT TOK_FI
              |TOK_WHILE TOK_LPAREN EXPR TOK_RPAREN CLSD_STMT TOK_DONE
              ;

EXPR: EXPR TOK_OR JOIN
       |JOIN
       ;

JOIN: JOIN TOK_AND EQAL
       |EQAL
       ;

EQAL: EQAL TOK_EQUAL TOK_EQUAL REL
       |EQAL TOK_NOT TOK_EQUAL REL
       |REL
       ;

REL: nex TOK_LESS nex
       |nex TOK_LESS TOK_EQUAL nex
       |nex TOK_GREATER TOK_EQUAL nex
       |nex TOK_GREATER TOK_EQUAL nex
       |nex TOK_GREATER nex
       |nex
       ;

nex: nex TOK_PLUS term
       |nex TOK_OCDEL term 
       |term
       ;

term: term TOK_MULT unary
       |term TOK_DIV unary
       |unary
       ;

unary: TOK_NOT unary 
       |TOK_OCDEL unary
       |factor
       ;

factor: TOK_LPAREN EXPR TOK_RPAREN
       |TOK_NUM
       |TOK_WORD
       |TOK_BOOL
       ;

K: NT_CREATE     {println("@State start");}
       | NT_GET
       | NT_UPDATE
       | NT_DEL
;

NT_CREATE: TOK_CREATE E F {cmd.PostObj(cmd.EntityStrToInt($2),$2, resMap(&$3))}
       | TOK_CREATE E P F {$$=$4; cmd.Disp(resMap(&$4)); cmd.PostObj(cmd.EntityStrToInt($2),$2, resMap(&$4)) }
;

NT_GET: TOK_GET P {cmd.GetObject($2)}
       | TOK_GET E F {/*cmd.Disp(resMap(&$4)); */cmd.SearchObjects($2, resMap(&$3)) }
;

NT_UPDATE: TOK_UPDATE P F {$$=$3;/*cmd.Disp(resMap(&$4));*/ cmd.UpdateObj($2, resMap(&$3))}
;

NT_DEL: TOK_DELETE P {println("@State NT_DEL");cmd.DeleteObj($2)}
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

ORIENTN: TOK_PLUS {$$=$1}
         | TOK_OCDEL {$$=$1}
         | {$$=""}
         ;

WORDORNUM: TOK_WORD {$$=$1; dCatchPtr = $1;}
           |TOK_NUM {x := strconv.Itoa($1);$$=x;dCatchPtr = $1;}
           |ORIENTN TOK_WORD ORIENTN TOK_WORD {$$=$1+$2+$3+$4; dCatchPtr = $1+$2+$3+$4;}
           |TOK_BOOL {var x bool;if $1=="false"{x = false}else{x=true};dCatchPtr = x;}
           ;

F:     TOK_ATTR TOK_EQUAL WORDORNUM F {$$=string($1+"="+$3+"="+$4); println("So we got: ", $$)}
       | TOK_ATTR TOK_EQUAL WORDORNUM {$$=$1+"="+$3}
;


P:     P1
       | TOK_SLASH P1 {$$="/"+$2}
;

P1:    TOK_WORD TOK_SLASH P1 {$$=$1+"/"+$3}
       | TOK_WORD {$$=$1}
       | TOK_DOT TOK_DOT TOK_SLASH P1 {$$="../"+$4}
       | TOK_WORD {$$=$1}
       | TOK_DOT TOK_DOT {$$=".."}
       | TOK_OCDEL {$$="-"}
       | TOK_DEREF TOK_WORD {$$=""}
       | {$$=""}
;

Q:     TOK_CD P {cmd.CD($2)}
       | TOK_LS P {cmd.LS($2)}
       | TOK_LSTEN P {cmd.LSOBJECT($2, 0)}
       | TOK_LSSITE P {cmd.LSOBJECT($2, 1)}
       | TOK_LSBLDG P {cmd.LSOBJECT($2, 2)}
       | TOK_LSROOM P {cmd.LSOBJECT($2, 3)}
       | TOK_LSRACK P {cmd.LSOBJECT($2, 4)}
       | TOK_LSDEV P {cmd.LSOBJECT($2, 5)}
       | TOK_LSSUBDEV P {cmd.LSOBJECT($2, 6)}
       | TOK_LSSUBDEV1 P {cmd.LSOBJECT($2, 7)}
       | TOK_TREE TOK_NUM {cmd.Tree("", $2)}
       | TOK_TREE P {cmd.Tree($2, 0)}
       | TOK_TREE P TOK_NUM {cmd.Tree($2, $3)}
       | BASH     {cmd.Execute()}
;

BASH:  TOK_CLR
       | TOK_GREP {}
       | TOK_LSOG {cmd.LSOG()}
       | TOK_PWD {cmd.PWD()}
       | TOK_EXIT     {cmd.Exit()}
       | TOK_DOC {cmd.Help("")}
       | TOK_DOC TOK_LS {cmd.Help("ls")}
       | TOK_DOC TOK_CD {cmd.Help("cd")}
       | TOK_DOC TOK_CREATE {cmd.Help("create")}
       | TOK_DOC TOK_GET {cmd.Help("gt")}
       | TOK_DOC TOK_UPDATE {cmd.Help("update")}
       | TOK_DOC TOK_DELETE {cmd.Help("delete")}
       | TOK_DOC TOK_WORD {cmd.Help($2)}
       | TOK_DOC TOK_TREE {cmd.Help("tree")}
       | TOK_DOC TOK_LSOG {cmd.Help("lsog")}
;

OCLISYNTX:  TOK_PLUS OCCR
            |OCDEL
            |OCUPDATE
            |OCGET
            |OCCHOOSE
            |OCDOT
            |OCSEL
            ;


OCCR:   TOK_OCTENANT TOK_OCPSPEC P TOK_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.TENANT,map[string]interface{}{"attributes":map[string]interface{}{"color":$5}} ,rlPtr)}
        |TOK_TENANT TOK_OCPSPEC P TOK_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.TENANT,map[string]interface{}{"attributes":map[string]interface{}{"color":$5}} ,rlPtr)}
        |TOK_OCSITE TOK_OCPSPEC P TOK_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.SITE,map[string]interface{}{"attributes":map[string]interface{}{"orientation":$5}} ,rlPtr)}
        |TOK_SITE TOK_OCPSPEC P TOK_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.SITE,map[string]interface{}{"attributes":map[string]interface{}{"orientation":$5}} ,rlPtr)}
        |TOK_OCBLDG TOK_OCPSPEC P TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.BLDG,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}} ,rlPtr)}
        |TOK_BLDG TOK_OCPSPEC P TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.BLDG,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}} ,rlPtr)}
        |TOK_OCROOM TOK_OCPSPEC P TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.ROOM,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}} ,rlPtr)}
        |TOK_ROOM TOK_OCPSPEC P TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.ROOM,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}} ,rlPtr)}
        |TOK_OCRACK TOK_OCPSPEC P TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.RACK,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}} ,rlPtr)}
        |TOK_RACK TOK_OCPSPEC P TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.RACK,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}} ,rlPtr)}
        |TOK_OCDEV TOK_OCPSPEC P TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.DEVICE,map[string]interface{}{"attributes":map[string]interface{}{"slot":$5, "sizeUnit":$7}} ,rlPtr)}
        |TOK_DEVICE TOK_OCPSPEC P TOK_ATTRSPEC WORDORNUM TOK_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.DEVICE,map[string]interface{}{"attributes":map[string]interface{}{"slot":$5, "sizeUnit":$7}} ,rlPtr)}
       ; 
OCDEL:  TOK_OCDEL P {cmd.DeleteObj(replaceOCLICurrPath($2))}
;

OCUPDATE: P TOK_DOT TOK_ATTR TOK_EQUAL WORDORNUM {println("Attribute Acquired");val := $3+"="+$5;cmd.UpdateObj(replaceOCLICurrPath($1), resMap(&val))}
;

OCGET: TOK_EQUAL P {cmd.GetObject(replaceOCLICurrPath($2))}
;

GETOBJS:      TOK_WORD TOK_COMMA GETOBJS {x := make([]string,0); x = append(x, cmd.State.CurrPath+"/"+$1); x = append(x, $3...); $$=x}
              | TOK_WORD {$$=[]string{cmd.State.CurrPath+"/"+$1}}
              ;

OCCHOOSE: TOK_EQUAL TOK_LBRAC GETOBJS TOK_RBRAC {cmd.State.ClipBoard = &$3; println("Selection made!")}
;

OCDOT:      TOK_DOT TOK_VAR TOK_OCPSPEC TOK_WORD TOK_EQUAL WORDORNUM {dynamicMap[$4] = varCtr; dynamicSymbolTable[varCtr] = dCatchPtr; varCtr+=1;  switch dCatchPtr.(type) {
	case string:
              x := dCatchPtr.(string)
              println("You want to assign",$4, "with value of", x)
	case int:
		x := dCatchPtr.(int)
              println("You want to assign",$4, "with value of", x)
	case bool:
		x := dCatchPtr.(bool)
              println("You want to assign",$4, "with value of", x)
       case float64, float32:
		x := dCatchPtr.(float64)
              println("You want to assign",$4, "with value of", x)
	}}
            |TOK_DOT TOK_CMDS TOK_OCPSPEC P {cmd.LoadFile($4)}
            |TOK_DOT TOK_TEMPLATE TOK_OCPSPEC P {cmd.LoadFile($4)}
            |TOK_DEREF TOK_WORD {v := dynamicSymbolTable[dynamicMap[$2]]; switch v.(type) {
	case string:
              x := v.(string)
              println("So You want the value: ",x)
	case int:
		x := v.(int)
              println("So You want the value: ",x)
	case bool:
		x := v.(bool)
              println("So You want the value: ",x)
       case float64, float32:
		x := dCatchPtr.(float64)
              println("So You want the value: ",x)
	} }
;

OCSEL:      TOK_SELECT {cmd.ShowClipBoard()}
            |TOK_SELECT TOK_DOT TOK_ATTR TOK_EQUAL TOK_WORD {x := $3+"="+$5;cmd.UpdateSelection(resMap(&x))}

%%