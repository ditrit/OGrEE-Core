%{
package main
import (
cmd "cli/controllers"
"path/filepath"      //Used by the injected code for error reporting
l "cli/logger"
)

var root node 
var _ = l.GetInfoLogger() //Suppresses annoying Dockerfile build error
%}

%union {
  n int
  s string
  f float64
  ast *ast
  node node
  nodeArr []node
  sArr []string
  mapVoid map[string]interface{}
}

%token <n> TOK_INT
%token <f> TOK_FLOAT
%token <s> TOK_WORD TOK_TENANT TOK_SITE TOK_BLDG TOK_ROOM
%token <s> TOK_RACK TOK_DEVICE TOK_STR
%token <s> TOK_CORIDOR TOK_GROUP 
%token <s> TOK_AC TOK_CABINET TOK_PANEL
%token <s> TOK_SENSOR 
%token <s> TOK_ROOM_TMPL TOK_OBJ_TMPL
%token <s> TOK_PLUS TOK_MINUS TOK_ORIENTATION
%token
       TOK_GET 
       TOK_EQUAL TOK_DOUBLE_EQUAL TOK_SLASH
       TOK_EXIT TOK_DOC TOK_CD TOK_PWD
       TOK_CLR TOK_GREP TOK_LS TOK_TREE
       TOK_LSOG TOK_LSTEN TOK_LSSITE TOK_LSBLDG
       TOK_LSCAB TOK_LSSENSOR TOK_LSAC TOK_LSPANEL
       TOK_LSCORRIDOR TOK_GETU
       TOK_LSROOM TOK_LSRACK TOK_LSDEV TOK_LSENTERPRISE
       TOK_ATTRSPEC TOK_GETSLOT
       TOK_COL TOK_SELECT TOK_LBRAC TOK_RBRAC
       TOK_COMMA TOK_DOT_DOT TOK_CMDS TOK_TEMPLATE TOK_VAR TOK_DEREF
       TOK_SEMICOL TOK_IF TOK_FOR TOK_WHILE
       TOK_ELSE TOK_LBLOCK TOK_RBLOCK
       TOK_LPAREN TOK_RPAREN TOK_OR TOK_AND TOK_IN TOK_PRNT
       TOK_NOT TOK_NOT_EQUAL 
       TOK_MULT TOK_GREATER TOK_GREATER_EQUAL TOK_LESS TOK_LESS_EQUAL 
       TOK_THEN TOK_FI TOK_DONE
       TOK_MOD
       TOK_UNSET TOK_ELIF TOK_DO TOK_LEN
       TOK_USE_JSON TOK_LINK TOK_UNLINK
       TOK_HIERARCH TOK_DRAWABLE TOK_ENV TOK_ORPH
       TOK_DRAW TOK_UNDRAW TOK_TRUE TOK_FALSE
       TOK_CAM_MOVE TOK_CAM_WAIT TOK_CAM_TRANSLATE TOK_CAM
       TOK_DOT TOK_SHARP
       TOK_UI_DELAY TOK_UI_WIREFRAME TOK_UI_INFOS TOK_UI_DEBUG TOK_UI_HIGHLIGHT TOK_UI
       
%type <n> LSOBJ_COMMAND
%type <s> OBJ_TYPE COMMAND UI_TOGGLE
%type <nodeArr> WNARG GETOBJS WORD_CONCAT
%type <mapVoid> ARGACC PRINTF
%type <sArr> WNARG2
%type <node> OCCR PATH PHYSICAL_PATH STRAY_DEV_PATH EXPR CONCAT CONCAT_TERM stmnt st2 IF 
       EXPR_NOQUOTE ARRAY ORIENTATION EXPR_NOQUOTE_NOCOL CONCAT_NOCOL CONCAT_TERM_NOCOL EXPR_NOQUOTE_COMMON
//%type <mapVoid> EQUAL_LIST

%right UNARY
%right TOK_NOT TOK_EQUAL TOK_GET TOK_CD TOK_LS TOK_TREE TOK_DRAW TOK_HIERARCH TOK_UNSET TOK_DRAWABLE TOK_VAR TOK_CMDS TOK_TEMPLATE TOK_SELECT TOK_LINK TOK_UNLINK TOK_LEN TOK_PRNT TOK_DOC
%left TOK_MULT TOK_SLASH TOK_MOD
%left TOK_OR
%left TOK_AND
%left TOK_DOUBLE_EQUAL TOK_NOT_EQUAL
%left TOK_LESS TOK_GREATER TOK_LESS_EQUAL TOK_GREATER_EQUAL
%left TOK_MINUS TOK_PLUS
//%left concat

%%

start: st2 {root = $1}

st2:    {$$=nil}
       | stmnt {$$=&ast{[]node{$1}}}
       | stmnt TOK_SEMICOL st2 {$$=&ast{[]node{$1, $3}}}
;

stmnt:   TOK_GET PATH {$$=&getObjectNode{$2}}
       //| TOK_GET OBJ_TYPE EQUAL_LIST {$$=&searchObjectsNode{$2, $3}}
       
       //NORMAL LSOBJ COMMANDS
       | LSOBJ_COMMAND PATH {$$=&lsObjNode{$2, $1,nil}}
       | LSOBJ_COMMAND {$$=&lsObjNode{&pathNode{&strLeaf{"."}, STD}, $1, nil}}
       
       //ARGUMENT LSOBJ COMMANDS 
       | LSOBJ_COMMAND PATH ARGACC {$$=&lsObjNode{$2, $1, $3}}
       | LSOBJ_COMMAND ARGACC {$$=&lsObjNode{&pathNode{&strLeaf{"."}, STD}, $1, $2}}

       //ARGUMENT TYPE LS COMMANDS
       | TOK_LS ARGACC {$$=&lsAttrGenericNode{&pathNode{&strLeaf{"."}, STD}, $2}}
       | TOK_LS PATH ARGACC {$$=&lsAttrGenericNode{$2, $3}}

       | TOK_GETU {x:=&pathNode{&strLeaf{"."}, STD}; y:=&intLeaf{0};$$=&getUNode{x, y}}
       | TOK_GETU PATH {$$=&getUNode{$2, &intLeaf{0}}}
       | TOK_GETU PATH TOK_INT {$$=&getUNode{$2, &intLeaf{$3}}}
       | TOK_GETSLOT PATH TOK_COMMA EXPR_NOQUOTE {$$=&getUNode{$2, $4}}

       | TOK_UNDRAW {$$=&undrawNode{nil}}
       | TOK_UNDRAW PATH {$$=&undrawNode{$2}}

       | TOK_DRAW PATH {$$=&drawNode{$2, 0,nil}}
       | TOK_DRAW PATH TOK_INT {$$=&drawNode{$2, $3,nil}}
       | TOK_DRAW {$$=&drawNode{&pathNode{&strLeaf{"."}, STD}, 0,nil}}

       | TOK_DRAW PATH ARGACC {$$=&drawNode{$2, 0,$3}}
       | TOK_DRAW PATH TOK_INT ARGACC {$$=&drawNode{$2, $3,$4}}
       | TOK_DRAW ARGACC {$$=&drawNode{&pathNode{&strLeaf{"."}, STD}, 0,$2}}
       
       | TOK_HIERARCH {$$=&hierarchyNode{&pathNode{&strLeaf{"."}, STD}, 1}}
       | TOK_HIERARCH PATH {$$=&hierarchyNode{$2, 1}}
       | TOK_HIERARCH PATH TOK_INT {$$=&hierarchyNode{$2, $3}}
       | TOK_UNSET PATH  {$$=&unsetAttrNode{$2}}
       | TOK_UNSET TOK_MINUS TOK_WORD TOK_WORD {$$=&unsetVarNode{$2+$3, $4}}
       
       | TOK_DRAWABLE {$$=&isEntityDrawableNode{&pathNode{&strLeaf{"."}, STD}}}
       | TOK_DRAWABLE PATH {$$=&isEntityDrawableNode{$2}}
       | TOK_ENV TOK_WORD TOK_EQUAL EXPR_NOQUOTE {$$=&setEnvNode{$2, $4}}
       | TOK_PLUS OCCR {$$=$2}
       | TOK_MINUS PATH {$$=&deleteObjNode{$2}}
       | TOK_MINUS TOK_SELECT {$$=&deleteSelectionNode{}}   


       // SELECTION COMMANDS
       | TOK_EQUAL PATH {$$=&selectObjectNode{$2}}
       | TOK_EQUAL {$$=&selectObjectNode{&strLeaf{""}}}
       | TOK_EQUAL TOK_LBRAC GETOBJS TOK_RBRAC {$$=&selectChildrenNode{$3}}

       // UPDATE / INTERACT
       | PHYSICAL_PATH TOK_COL TOK_WORD TOK_EQUAL TOK_SHARP EXPR_NOQUOTE {$$=&updateObjNode{$1, map[string]interface{}{$3:$6},true}}
       | PHYSICAL_PATH TOK_COL TOK_WORD TOK_EQUAL EXPR_NOQUOTE TOK_ATTRSPEC EXPR_NOQUOTE {$$=&specialUpdateNode{$1, $3, $5, $7,""}}
       | PHYSICAL_PATH TOK_COL TOK_WORD TOK_EQUAL EXPR_NOQUOTE TOK_ATTRSPEC EXPR_NOQUOTE TOK_ATTRSPEC TOK_WORD {$$=&specialUpdateNode{$1, $3, $5, $7,$9}}
       | PHYSICAL_PATH TOK_COL TOK_WORD TOK_EQUAL EXPR_NOQUOTE {
              /*Hack Case: we need to change the mode of the Path Node*/;
              ($1).(*pathNode).mode = STD;
              $$=&updateObjNode{$1, map[string]interface{}{$3:$5},false}}



       // ASSIGNMENT  
       | TOK_DEREF TOK_WORD TOK_EQUAL EXPR {$$=&assignNode{$2, $4}}
       | TOK_VAR TOK_COL TOK_WORD TOK_EQUAL EXPR_NOQUOTE {$$=&assignNode{$3, $5}}
       | TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN TOK_GET PATH TOK_RPAREN {$$=&assignNode{$3, &getObjectNode{$8}}}
       | TOK_VAR TOK_COL TOK_WORD TOK_EQUAL TOK_DEREF TOK_LPAREN TOK_PWD TOK_RPAREN {$$=&assignNode{$3, &pwdNode{}}}
       

       | TOK_CMDS TOK_COL EXPR_NOQUOTE {$$=&loadNode{$3}}
       | TOK_TEMPLATE TOK_COL EXPR_NOQUOTE {$$=&loadTemplateNode{$3}}
       | TOK_SELECT {$$=&selectNode{}}
       | TOK_DRAWABLE TOK_LPAREN PATH TOK_RPAREN {$$=&isEntityDrawableNode{$3}}
       | TOK_DRAWABLE TOK_LPAREN PATH TOK_COMMA EXPR_NOQUOTE TOK_RPAREN {$$=&isAttrDrawableNode{$3, $5}}
       | TOK_LEN TOK_LPAREN TOK_WORD TOK_RPAREN {$$=&lenNode{$3}}


       // LINKING
       | TOK_LINK TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR_NOQUOTE {$$=&linkObjectNode{$3, $5,nil}}
       | TOK_LINK TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR_NOQUOTE TOK_ATTRSPEC EXPR_NOQUOTE {$$=&linkObjectNode{$3, $5, $7}}
       | TOK_UNLINK TOK_COL PHYSICAL_PATH {$$=&unlinkObjectNode{$3,nil}}
       | TOK_UNLINK TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR_NOQUOTE {$$=&unlinkObjectNode{$3,$5}}

       // BASH
       | TOK_CLR {$$=&clrNode{}}
       | TOK_GREP {$$=&grepNode{}}
       | TOK_PRNT EXPR_NOQUOTE {$$=&printNode{$2}}
       | TOK_LSOG {$$=&lsogNode{}}
       | TOK_LSENTERPRISE {$$=&lsenterpriseNode{}}
       | TOK_ENV {$$=&envNode{}}
       | TOK_PWD {$$=&pwdNode{}}
       | TOK_EXIT {$$=&exitNode{}}
       | TOK_DOC COMMAND {$$=&helpNode{$2}}
       | TOK_DOC {$$=&helpNode{""}}
       | TOK_DOC TOK_WORD {$$=&helpNode{$2}}
       | TOK_CD PATH {$$=&cdNode{$2}}
       | TOK_CD {$$=&cdNode{strLeaf{"/"}}}
       | TOK_CD TOK_MINUS {$$=&cdNode{strLeaf{"-"}}}
       | TOK_LS PATH {$$=&lsNode{$2}}
       | TOK_LS {$$=&lsNode{&pathNode{&strLeaf{"."}, STD}}}
       | TOK_TREE PATH TOK_INT {$$=&treeNode{$2, $3}}
       | TOK_TREE PATH {$$=&treeNode{$2, 0}}
       | TOK_TREE {$$=&treeNode{&pathNode{&strLeaf{"."}, STD}, 0}}



       // UNITY
       | TOK_UI_DELAY TOK_EQUAL EXPR {$$=&uiDelayNode{$3}}
       | UI_TOGGLE TOK_EQUAL EXPR {$$=&uiToggleNode{$1, $3}}
       | TOK_UI_HIGHLIGHT TOK_EQUAL PATH {$$=&uiHighlightNode{$3}}
       | TOK_CAM_MOVE TOK_EQUAL EXPR TOK_ATTRSPEC EXPR {$$=&cameraMoveNode{"move", $3, $5}}
       | TOK_CAM_TRANSLATE TOK_EQUAL EXPR TOK_ATTRSPEC EXPR {$$=&cameraMoveNode{"translate", $3, $5}}
       | TOK_CAM_WAIT TOK_EQUAL EXPR {$$=&cameraWaitNode{$3}}
       | TOK_GREATER PATH {$$=&focusNode{$2}} 
       | TOK_GREATER  {$$=&focusNode{&strLeaf{""}}} 

       // FUNCTIONS
       | TOK_WORD TOK_LPAREN TOK_RPAREN TOK_LBRAC st2 TOK_RBRAC {$$=&funcDefNode{$1, $5}}
       | TOK_WORD {$$=&funcCallNode{$1}}

       // LOOPS
       | TOK_WHILE TOK_LPAREN EXPR TOK_RPAREN st2 TOK_DONE {$$=&whileNode{$3, $5}}

       // FOR
       | TOK_FOR TOK_WORD TOK_IN EXPR TOK_SEMICOL st2 TOK_DONE {$$=&forArrayNode{$2, $4, $6}}
       | TOK_FOR TOK_WORD TOK_IN TOK_LBRAC EXPR TOK_DOT_DOT EXPR TOK_RBRAC TOK_SEMICOL st2 TOK_DONE {$$=&forRangeNode{$2, $5, $7, $10}}
       | TOK_FOR TOK_LPAREN TOK_LPAREN TOK_WORD TOK_EQUAL EXPR TOK_SEMICOL EXPR TOK_SEMICOL st2 TOK_RPAREN TOK_RPAREN TOK_SEMICOL st2 TOK_DONE {
              $$=&forNode{&assignNode{$4, $6},$8,$10,$14}
       }

       // IF
       | TOK_IF IF {$$=$2}
       | TOK_IF TOK_LBLOCK EXPR TOK_RBLOCK TOK_THEN st2 TOK_ELIF IF {$$=&ifNode{$3, $6, $8}}
;

IF: TOK_LBLOCK EXPR TOK_RBLOCK TOK_THEN st2 TOK_FI {$$=&ifNode{$2, $5, nil}}
       | TOK_LBLOCK EXPR TOK_RBLOCK TOK_THEN st2 TOK_ELSE st2 TOK_FI {$$=&ifNode{$2, $5, $7}}
;

PHYSICAL_PATH: EXPR_NOQUOTE_NOCOL {$$=&pathNode{$1, PHYSICAL}}
STRAY_DEV_PATH: EXPR_NOQUOTE {$$=&pathNode{$1, STRAY_DEV}}
PATH: EXPR_NOQUOTE {$$=&pathNode{$1, STD}}

EXPR_NOQUOTE_NOCOL: TOK_DEREF TOK_LPAREN TOK_LPAREN EXPR TOK_RPAREN TOK_RPAREN {$$=$4}
       | CONCAT_NOCOL {$$=$1}
       | EXPR_NOQUOTE_COMMON {$$=$1}
;

EXPR_NOQUOTE: TOK_DEREF TOK_LPAREN TOK_LPAREN EXPR TOK_RPAREN TOK_RPAREN {$$=$4}
       | CONCAT {$$=$1}
       | EXPR_NOQUOTE_COMMON {$$=$1}
;

EXPR_NOQUOTE_COMMON: TOK_INT {$$=&floatLeaf{float64($1)}}
       | TOK_FLOAT {$$=&floatLeaf{$1}}
       | TOK_TRUE {$$=&boolLeaf{true}}
       | TOK_FALSE {$$=&boolLeaf{false}}
       | ARRAY {$$=$1}
       | TOK_MINUS EXPR_NOQUOTE_COMMON {$$=&arithNode{"-", &floatLeaf{0}, $2}}
;

CONCAT_NOCOL:  CONCAT_TERM_NOCOL {$$=$1}
       | CONCAT_TERM_NOCOL CONCAT_NOCOL {$$=&concatNode{[]node{$1, $2}}}
;

CONCAT:  CONCAT_TERM {$$=$1}
       | CONCAT_TERM CONCAT {$$=&concatNode{[]node{$1, $2}}}
;

CONCAT_TERM_NOCOL:  TOK_DEREF TOK_LBRAC TOK_WORD TOK_RBRAC {$$=&symbolReferenceNode{$3}}
       | TOK_DEREF TOK_WORD {$$=&symbolReferenceNode{$2}}
       | TOK_WORD {$$=&strLeaf{$1}}
       | TOK_ORIENTATION {$$=&strLeaf{$1}}
       | TOK_STR {$$=&strLeaf{$1}}
       | TOK_SLASH {$$=&strLeaf{"/"}}
       | TOK_DOT_DOT {$$=&strLeaf{".."}}
       | TOK_DOT {$$=&strLeaf{"."}}
;

CONCAT_TERM:  CONCAT_TERM_NOCOL {$$=$1}
       | TOK_COL {$$=&strLeaf{":"}}
;

EXPR: TOK_INT {$$=&floatLeaf{float64($1)}}
       | TOK_FLOAT {$$=&floatLeaf{$1}}
       | TOK_TRUE {$$=&boolLeaf{true}}
       | TOK_FALSE {$$=&boolLeaf{false}}
       | TOK_STR {$$=&strLeaf{$1}}
       | ARRAY {$$=$1}
       | TOK_DEREF TOK_LBRAC TOK_WORD TOK_RBRAC {$$=&symbolReferenceNode{$3}}
       | TOK_DEREF TOK_WORD {$$=&symbolReferenceNode{$2}}
       | TOK_DEREF TOK_WORD TOK_LBLOCK EXPR TOK_RBLOCK {$$=&objReferenceNode{$2,$4}}

       | TOK_LPAREN EXPR TOK_RPAREN {$$=$2}

       | EXPR TOK_OR EXPR {$$=&logicalNode{"||", $1, $3}}
       | EXPR TOK_AND EXPR {$$=&logicalNode{"&&", $1, $3}}
       | EXPR TOK_EQUAL TOK_EQUAL EXPR {$$=&equalityNode{"==", $1, $4}}
       | EXPR TOK_NOT_EQUAL EXPR {$$=&equalityNode{"!=", $1, $3}}
       | EXPR TOK_LESS EXPR {$$=&comparatorNode{"<", $1, $3}}
       | EXPR TOK_LESS_EQUAL EXPR {$$=&comparatorNode{"<=", $1, $3}}
       | EXPR TOK_GREATER_EQUAL EXPR {$$=&comparatorNode{">=", $1, $3}}
       | EXPR TOK_GREATER EXPR {$$=&comparatorNode{">", $1, $3}}
       | TOK_NOT EXPR {$$=&negateNode{$2}}

       | EXPR TOK_PLUS EXPR {$$=&arithNode{"+", $1, $3}}
       | EXPR TOK_MINUS EXPR {$$=&arithNode{"-", $1, $3}}
       | EXPR TOK_MULT EXPR {$$=&arithNode{"*", $1, $3}}
       | EXPR TOK_SLASH EXPR {$$=&arithNode{"/", $1, $3}}
       | EXPR TOK_MOD EXPR {$$=&arithNode{"%", $1, $3}}
       | TOK_MINUS EXPR {$$=&arithNode{"-", &floatLeaf{0}, $2}} %prec UNARY
; 


ARRAY: TOK_LBLOCK WNARG TOK_RBLOCK {$$=&arrNode{$2}}
WNARG: EXPR TOK_COMMA WNARG {x:=[]node{$1}; $$=append(x, $3...)}
       |EXPR  {x:=[]node{$1}; $$=x}
; 

WORD_CONCAT: TOK_WORD TOK_COMMA WORD_CONCAT {$$=append([]node{&strLeaf{$1}}, $3...)}
           | {$$=nil} 
           | TOK_WORD {$$=[]node{&strLeaf{$1}}}
;
//Argument Accumulator
ARGACC: TOK_MINUS TOK_WORD {$$=map[string]interface{}{$2:nil}}
       |TOK_MINUS TOK_WORD ARGACC {$3[$2]=nil;$$=$3}
       |TOK_MINUS TOK_WORD TOK_WORD {$$=map[string]interface{}{$2:$3}}
       //PRINTF TYPE 
       |TOK_MINUS TOK_WORD TOK_LPAREN PRINTF TOK_RPAREN {$$=map[string]interface{}{$2:$4}}
       |TOK_MINUS TOK_WORD TOK_LPAREN PRINTF TOK_RPAREN ARGACC {$6[$2]=$4;$$=$6}
       
       |TOK_MINUS TOK_WORD TOK_WORD ARGACC {$4[$2]=$3;$$=$4}
       |TOK_MINUS TOK_WORD TOK_STR { $$=map[string]interface{}{$2:$3}}
       |TOK_MINUS TOK_WORD TOK_STR ARGACC { $4[$2]=$3;$$=$4}
;

//For printf arguments
WNARG2: TOK_WORD TOK_COMMA WNARG2 {$$=append([]string{$1},$3...)}
       |TOK_WORD {$$=[]string{$1}}
;

PRINTF: TOK_STR TOK_COMMA WNARG2 {$$=map[string]interface{}{$1:$3}}
;


GETOBJS: PATH TOK_COMMA GETOBJS {x:=[]node{$1}; $$=append(x, $3...)}
       | PATH {x:=[]node{$1}; $$=x}
;

OBJ_TYPE: TOK_TENANT | TOK_SITE | TOK_BLDG | TOK_ROOM | TOK_RACK | TOK_DEVICE | TOK_AC | TOK_PANEL |TOK_CABINET 
       | TOK_SENSOR | TOK_CORIDOR | TOK_GROUP | TOK_OBJ_TMPL | TOK_ROOM_TMPL
;

LSOBJ_COMMAND: TOK_LSTEN {$$=0} | TOK_LSSITE {$$=1} | TOK_LSBLDG {$$=2} | TOK_LSROOM {$$=3} | TOK_LSRACK {$$=4}
       | TOK_LSDEV {$$=5} | TOK_LSAC {$$=6} | TOK_LSPANEL {$$=7}
       | TOK_LSCAB {$$=8} | TOK_LSCORRIDOR {$$=9} | TOK_LSSENSOR{$$=10}
;

UI_TOGGLE: TOK_UI_DEBUG{$$="debug"} | TOK_UI_INFOS{$$="infos"} | TOK_UI_WIREFRAME{$$="wireframe"}

//DOCUMENTATION (ie: man pwd)
COMMAND: TOK_LINK{$$="link"} | TOK_UNLINK{$$="unlink"} | TOK_CLR{$$="clear"} | TOK_LS{$$="ls"}
       | TOK_PWD{$$="pwd"} | TOK_PRNT{$$="print"} | TOK_CD{$$="cd"} | TOK_CAM{$$="camera"} 
       | TOK_UI{$$="ui"} | TOK_GET{$$="get"} | TOK_LSENTERPRISE{$$="lsenterprise"}
       | TOK_HIERARCH{$$="hc"} | TOK_TREE{$$="tree"} | TOK_DRAW{$$="draw"} 
       | TOK_IF{$$="if"} | TOK_WHILE{$$="while"} | TOK_FOR{$$="for"} | TOK_UNSET{$$="unset"}
       | TOK_SELECT{$$="select"} | TOK_LSOG{$$="lsog"} | TOK_ENV{$$="env"} 
       | TOK_LSTEN{$$="lsten"} | TOK_LSSITE{$$="lssite"} | TOK_LSBLDG{$$="lsbldg"} | TOK_LSROOM{$$="lsroom"} 
       | TOK_LSRACK{$$="lsrack"} | TOK_LSDEV{$$="lsdev"} | TOK_MINUS{$$="-"} | TOK_TEMPLATE{$$=".template"}
       | TOK_CMDS{$$=".cmds"} | TOK_VAR{$$=".var"} | TOK_PLUS{$$="+"} | TOK_EQUAL{$$="="} 
       | TOK_GREATER{$$=">"} | TOK_DRAWABLE{$$="drawable"}
       | TOK_GETU{$$="getu"} | TOK_GETSLOT{$$="getslot"}
       | TOK_GREP {$$="grep"} | TOK_UNDRAW {$$="undraw"}
;

//Special case here, need to check if word
//is orientation type. A special token doesn't make 
//sense here since the user cant create or delete 
//objects that use the orientation as name
ORIENTATION: TOK_WORD {$$=&strLeaf{$1} }
       | TOK_WORD TOK_MINUS TOK_WORD {$$=&strLeaf{$1+$2+$3}}
       | TOK_WORD TOK_PLUS TOK_WORD {$$=&strLeaf{$1+$2+$3}}
       | TOK_PLUS TOK_WORD {$$=&strLeaf{$1+$2}}
       | TOK_MINUS TOK_WORD {$$=&strLeaf{$1+$2}}
       | TOK_PLUS TOK_WORD TOK_MINUS TOK_WORD {$$=&strLeaf{$1+$2+$3+$4}}
       | TOK_PLUS TOK_WORD TOK_PLUS TOK_WORD  {$$=&strLeaf{$1+$2+$3+$4}}
       | TOK_MINUS TOK_WORD TOK_MINUS TOK_WORD {$$=&strLeaf{$1+$2+$3+$4}}
       | TOK_MINUS TOK_WORD TOK_PLUS TOK_WORD {$$=&strLeaf{$1+$2+$3+$4}}

OCCR:   
        TOK_TENANT TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR_NOQUOTE {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"color":$5}}
              $$=&getOCAttrNode{$3, cmd.TENANT, attributes}
        }
        |TOK_SITE TOK_COL PHYSICAL_PATH TOK_ATTRSPEC ORIENTATION {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"orientation":$5}}
              $$=&getOCAttrNode{$3, cmd.SITE, attributes}
        } 
        |TOK_BLDG TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}}
              $$=&getOCAttrNode{$3, cmd.BLDG, attributes}
        }
        |TOK_ROOM TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC ORIENTATION TOK_ATTRSPEC EXPR_NOQUOTE{
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7, "orientation":$9, "floorUnit":$11}}
              $$=&getOCAttrNode{$3, cmd.ROOM, attributes}
        }
        |TOK_ROOM TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC ORIENTATION {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7, "orientation":$9}}
              $$=&getOCAttrNode{$3, cmd.ROOM, attributes}
        }
        |TOK_ROOM TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR_NOQUOTE {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "template":$7}}
              $$=&getOCAttrNode{$3, cmd.ROOM, attributes}
        }
        |TOK_RACK TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR_NOQUOTE TOK_ATTRSPEC EXPR_NOQUOTE {$$=&createRackNode{$3, [3]node{$5, $7, $9}}}
        |TOK_DEVICE TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR_NOQUOTE TOK_ATTRSPEC EXPR_NOQUOTE {$$=&createDeviceNode{$3, [3]node{$5, $7, nil}}}
        |TOK_DEVICE TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR_NOQUOTE TOK_ATTRSPEC EXPR_NOQUOTE TOK_ATTRSPEC EXPR_NOQUOTE {$$=&createDeviceNode{$3, [3]node{$5, $7, $9}}}
        |TOK_CORIDOR TOK_COL PHYSICAL_PATH TOK_ATTRSPEC TOK_LBRAC EXPR_NOQUOTE TOK_COMMA EXPR_NOQUOTE TOK_RBRAC TOK_ATTRSPEC EXPR_NOQUOTE {
              $$=&createCorridorNode{$3,$6,$8,$11}
        }
        |TOK_GROUP TOK_COL PHYSICAL_PATH TOK_ATTRSPEC TOK_LBRAC GETOBJS TOK_RBRAC {$$=&createGroupNode{$3, $6}}
        |TOK_ORPH TOK_DEVICE TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR_NOQUOTE {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"template":$6}}
              $$=&getOCAttrNode{$4, cmd.STRAY_DEV, attributes}
        }
        |TOK_ORPH TOK_SENSOR TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR_NOQUOTE {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"template":$6}}
              $$=&getOCAttrNode{$4, cmd.STRAYSENSOR, attributes}
        }
       //EasyPost syntax STRAYSENSOR
       |OBJ_TYPE TOK_USE_JSON PHYSICAL_PATH {$$=&easyPostNode{$1, $3}}
;