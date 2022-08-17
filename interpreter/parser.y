%{
package main
import (
cmd "cli/controllers"
"path/filepath"
l "cli/logger"
)

var root node 
var _ = l.GetInfoLogger() //Suppresses annoying Dockerfile build error
%}

%union {
  n int
  s string
  f float64
  sarr []string
  ast *ast
  node node
  boolNode boolNode
  numNode numNode
  nodeArr []node
  arr []interface{}
  mapArr []map[int]interface{}
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
       TOK_CREATE TOK_GET TOK_UPDATE TOK_DELETE TOK_SEARCH
       TOK_EQUAL TOK_DOUBLE_EQUAL TOK_CMDFLAG TOK_SLASH TOK_DOUBLE_DOT
       TOK_EXIT TOK_DOC TOK_CD TOK_PWD
       TOK_CLR TOK_GREP TOK_LS TOK_TREE
       TOK_LSOG TOK_LSTEN TOK_LSSITE TOK_LSBLDG
       TOK_LSCAB TOK_LSSENSOR TOK_LSAC TOK_LSPANEL
       TOK_LSCORRIDOR TOK_LSU TOK_LSSLOT TOK_GETU
       TOK_LSROOM TOK_LSRACK TOK_LSDEV
       TOK_ATTRSPEC TOK_GETSLOT
       TOK_COL TOK_SELECT TOK_LBRAC TOK_RBRAC
       TOK_COMMA TOK_CMDS TOK_TEMPLATE TOK_VAR TOK_DEREF
       TOK_SEMICOL TOK_IF TOK_FOR TOK_WHILE
       TOK_ELSE TOK_LBLOCK TOK_RBLOCK
       TOK_LPAREN TOK_RPAREN TOK_OR TOK_AND TOK_IN TOK_PRNT TOK_QUOT
       TOK_NOT TOK_NOT_EQUAL 
       TOK_MULT TOK_GREATER TOK_GREATER_EQUAL TOK_LESS TOK_LESS_EQUAL 
       TOK_THEN TOK_FI TOK_DONE
       TOK_MOD
       TOK_UNSET TOK_ELIF TOK_DO TOK_LEN
       TOK_USE_JSON TOK_PARTIAL TOK_LINK TOK_UNLINK
       TOK_CAM TOK_UI TOK_HIERARCH TOK_DRAWABLE TOK_ENV TOK_ORPH
       TOK_DRAW TOK_SETENV TOK_TRUE TOK_FALSE
       
%type <n> LSOBJ_COMMAND
%type <s> OBJ_TYPE COMMAND
%type <nodeArr> WNARG GETOBJS
%type <node> OCCR PATH PHYSICAL_PATH STRAY_DEV_PATH EXPR CONCAT CONCAT_TERM ARITHEXPR stmnt DEREF st2
%type <boolNode> BOOLEXPR
//%type <mapVoid> EQUAL_LIST

%right TOK_EQUAL TOK_GET TOK_CD TOK_LS TOK_TREE TOK_DRAW TOK_HIERARCH TOK_UNSET TOK_DRAWABLE TOK_SETENV TOK_VAR TOK_CMDS TOK_TEMPLATE TOK_SELECT TOK_LINK TOK_UNLINK TOK_LEN TOK_PRNT TOK_DOC
%left TOK_OR
%left TOK_AND
%left TOK_DOUBLE_EQUAL TOK_NOT_EQUAL
%left TOK_LESS TOK_GREATER TOK_LESS_EQUAL TOK_GREATER_EQUAL
%left TOK_MINUS TOK_PLUS
%left TOK_MULT TOK_SLASH
%left TOK_NOT
//%left concat

%%

start: st2 {root = $1}

st2:    {$$=nil}
       |stmnt {$$=&ast{[]node{$1} }}
       |stmnt TOK_SEMICOL st2 {$$=&ast{[]node{$1, $3}}}
;

stmnt:   TOK_GET PATH {$$=&getObjectNode{$2}}
       //| TOK_GET OBJ_TYPE EQUAL_LIST {$$=&searchObjectsNode{$2, $3}}
       | TOK_EQUAL PHYSICAL_PATH {$$=&selectObjectNode{$2}}
       | TOK_EQUAL TOK_LBRAC GETOBJS TOK_RBRAC {$$=&selectChildrenNode{$3}}
       | PHYSICAL_PATH TOK_COL TOK_WORD TOK_EQUAL EXPR {$$=&updateObjNode{$1, map[string]interface{}{$3:$5}}}
       | PHYSICAL_PATH TOK_COL TOK_WORD TOK_EQUAL EXPR TOK_ATTRSPEC EXPR {$$=&specialUpdateNode{$1, $3, $5, $7}}
       | TOK_CD PATH {$$=&cdNode{$2}}
       | TOK_LS PATH {$$=&lsNode{$2}}
       | TOK_LS {$$=&lsNode{&strLeaf{""}}}
       | LSOBJ_COMMAND PATH {$$=&lsObjNode{$2, $1}}
       | TOK_TREE PATH TOK_INT {$$=&treeNode{$2, $3}}
       | TOK_TREE PATH {$$=&treeNode{$2, 0}}
       | TOK_DRAW PATH {$$=&drawNode{$2, 0}}
       | TOK_DRAW PATH TOK_INT {$$=&drawNode{$2, $3}}
       | TOK_HIERARCH PATH {$$=&hierarchyNode{$2, 1}}
       | TOK_HIERARCH PATH TOK_INT {$$=&hierarchyNode{$2, $3}}
       | TOK_UNSET TOK_MINUS TOK_WORD TOK_WORD {$$=&unsetVarNode{$2+$3, $4}}
       | TOK_DRAWABLE PATH {$$=&isEntityDrawableNode{$2}}
       | TOK_SETENV TOK_WORD TOK_EQUAL EXPR {$$=&setEnvNode{$2, $4}}
       | TOK_PLUS OCCR {$$=$2}
       | TOK_MINUS PATH {$$=&deleteObjNode{$2}}
       | TOK_MINUS TOK_SELECT {$$=&deleteSelectionNode{}}     
       | TOK_VAR TOK_COL TOK_WORD TOK_EQUAL EXPR {$$=&assignNode{$3, $5}}
       | TOK_CMDS TOK_COL EXPR {$$=&loadNode{$3}}
       | TOK_TEMPLATE TOK_COL EXPR {$$=&loadTemplateNode{$3}}
       | TOK_SELECT {$$=&selectNode{}}

       // LINKING
       | TOK_LINK TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR {$$=&linkObjectNode{[]interface{}{$3, $5}}}
       | TOK_LINK TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR {$$=&linkObjectNode{[]interface{}{$3, $5, $7}}}
       | TOK_UNLINK TOK_COL PHYSICAL_PATH {$$=&unlinkObjectNode{[]interface{}{$3}}}
       | TOK_UNLINK TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR {$$=&unlinkObjectNode{[]interface{}{$3,$5}}}
       | TOK_LEN TOK_LPAREN TOK_WORD TOK_RPAREN {$$=&lenNode{$3}}

       // BASH
       | TOK_CLR {$$=&clrNode{}}
       | TOK_GREP {$$=&grepNode{}}
       | TOK_PRNT EXPR {$$=&printNode{$2}}
       | TOK_LSOG {$$=&lsogNode{}}
       | TOK_ENV {$$=&envNode{}}
       | TOK_PWD {$$=&pwdNode{}}
       | TOK_EXIT {$$=&exitNode{}}
       | TOK_DOC COMMAND {$$=&helpNode{$2}}
       | TOK_DOC {$$=&helpNode{""}}
       | TOK_DOC TOK_WORD {$$=&helpNode{$2}}          
;

PHYSICAL_PATH: EXPR {$$=&pathNode{$1, PHYSICAL}};
STRAY_DEV_PATH: EXPR {$$=&pathNode{$1, STRAY_DEV}};
PATH: EXPR {$$=&pathNode{$1, STD}};

EXPR:    TOK_DEREF TOK_WORD TOK_LBLOCK EXPR TOK_RBLOCK {$$=&arrayReferenceNode{$2, $4}}
       | TOK_DEREF TOK_LPAREN TOK_LPAREN ARITHEXPR TOK_RPAREN TOK_RPAREN {$$=$4}
       | TOK_DEREF TOK_LPAREN TOK_LPAREN BOOLEXPR TOK_RPAREN TOK_RPAREN {$$=$4}
       | TOK_INT {$$=&intLeaf{$1}}
       | TOK_FLOAT {$$=&floatLeaf{$1}}
       | TOK_MINUS TOK_INT {$$=&intLeaf{-$2}}
       | TOK_MINUS TOK_FLOAT {$$=&floatLeaf{-$2}}
       | TOK_TRUE {$$=&boolLeaf{true}}
       | TOK_FALSE {$$=&boolLeaf{false}}
       | CONCAT {$$=$1}
       | TOK_LBLOCK WNARG TOK_RBLOCK {$$=&arrNode{$2}}
;

CONCAT:  CONCAT_TERM {$$=$1}
       | CONCAT_TERM CONCAT {$$=&concatNode{[]node{$1, $2}}}
;

CONCAT_TERM: TOK_DEREF TOK_WORD {$$=&symbolReferenceNode{$2}}
       | TOK_DEREF TOK_LBRAC TOK_WORD TOK_RBRAC {$$=&symbolReferenceNode{$3}}
       | TOK_WORD {$$=&strLeaf{$1}}
       | TOK_STR {$$=&strLeaf{$1}}
       | TOK_SLASH {$$=&strLeaf{"/"}}
;

BOOLEXPR: EXPR TOK_OR EXPR {$$=&logicalNode{"||", $1, $3}}
       | EXPR TOK_AND EXPR {$$=&logicalNode{"&&", $1, $3}}
       | EXPR TOK_DOUBLE_EQUAL EXPR {$$=&equalityNode{"==", $1, $3}}
       | EXPR TOK_NOT_EQUAL EXPR {$$=&equalityNode{"!=", $1, $3}}
       | EXPR TOK_LESS EXPR {$$=&comparatorNode{"<", $1, $3}}
       | EXPR TOK_LESS_EQUAL EXPR {$$=&comparatorNode{"<=", $1, $3}}
       | EXPR TOK_GREATER_EQUAL EXPR {$$=&comparatorNode{">=", $1, $3}}
       | EXPR TOK_GREATER EXPR {$$=&comparatorNode{">", $1, $3}}
       | TOK_NOT EXPR {$$=&negateNode{$2}}
       | TOK_TRUE {$$=&boolLeaf{true}}
       | TOK_FALSE {$$=&boolLeaf{false}}
; 

ARITHEXPR: EXPR TOK_PLUS EXPR {$$=&arithNode{"+", $1, $3}}
       | EXPR TOK_MINUS EXPR {$$=&arithNode{"-", $1, $3}}
       | EXPR TOK_MULT EXPR {$$=&arithNode{"*", $1, $3}}
       | EXPR TOK_SLASH EXPR {$$=&arithNode{"/", $1, $3}}
       | EXPR TOK_MOD EXPR {$$=&arithNode{"%", $1, $3}}
       | TOK_MINUS TOK_INT {$$=&intLeaf{-$2}}
       | TOK_MINUS TOK_FLOAT {$$=&floatLeaf{-$2}}
       | TOK_MINUS DEREF {$$=&arithNode{"-", &floatLeaf{0}, $2}}
       | TOK_MINUS TOK_LPAREN EXPR TOK_LPAREN {$$=&arithNode{"-", &floatLeaf{0}, $3}}
       | TOK_INT {$$=&intLeaf{$1}}
       | TOK_FLOAT {$$=&floatLeaf{$1}}

DEREF: TOK_DEREF TOK_LBRAC TOK_WORD TOK_RBRAC {$$=&symbolReferenceNode{$3}}
       | TOK_DEREF TOK_LBRAC TOK_WORD TOK_LBLOCK EXPR TOK_RBLOCK TOK_RBRAC {$$=&arrayReferenceNode{$3, $5}}
;

/* EQUAL_LIST: TOK_WORD TOK_EQUAL EXPR EQUAL_LIST {m := $4; m[$1]=$3; $$=m}
       | TOK_WORD TOK_EQUAL EXPR {m := make(map[string]interface{}); m[$1]=$3; $$=m}
; */

//For making array types
WNARG: EXPR TOK_COMMA WNARG {x:=[]node{$1}; $$=append(x, $3...)}
       |EXPR  {x:=[]node{$1}; $$=x}
; 

GETOBJS: PATH TOK_COMMA GETOBJS {x:=[]node{$1}; $$=append(x, $3...)}
       | PATH {x:=[]node{$1}; $$=x}
;

OBJ_TYPE: TOK_TENANT | TOK_SITE | TOK_BLDG | TOK_ROOM | TOK_RACK | TOK_DEVICE | TOK_AC | TOK_PANEL |TOK_CABINET 
       | TOK_SENSOR | TOK_CORIDOR | TOK_GROUP | TOK_OBJ_TMPL | TOK_ROOM_TMPL
;

LSOBJ_COMMAND: TOK_LSTEN {$$=0} | TOK_LSSITE {$$=1} | TOK_LSBLDG {$$=2} | TOK_LSROOM {$$=3} | TOK_LSRACK {$$=4}
       | TOK_LSDEV {$$=5} | TOK_LSAC {$$=6} | TOK_LSPANEL {$$=7}
       | TOK_LSCAB {$$=9} | TOK_LSCORRIDOR {$$=12} | TOK_LSSENSOR{$$=13}
;

COMMAND: TOK_LINK{$$="link"} | TOK_UNLINK{$$="unlink"} | TOK_CLR{$$="clear"} | TOK_LS{$$="ls"}
       | TOK_PWD{$$="pwd"} | TOK_PRNT{$$="print"} | TOK_CD{$$="cd"} | TOK_CAM{$$="camera"} 
       | TOK_UI{$$="ui"} | TOK_GET{$$="get"}
       | TOK_HIERARCH{$$="hc"} | TOK_TREE{$$="tree"} | TOK_DRAW{$$="draw"} 
       | TOK_IF{$$="if"} | TOK_WHILE{$$="while"} | TOK_FOR{$$="for"} | TOK_UNSET{$$="unset"}
       | TOK_SELECT{$$="select"} | TOK_LSOG{$$="lsog"} | TOK_ENV{$$="env"} 
       | TOK_LSTEN{$$="lsten"} | TOK_LSSITE{$$="lssite"} | TOK_LSBLDG{$$="lsbldg"} | TOK_LSROOM{$$="lsroom"} 
       | TOK_LSRACK{$$="lsrack"} | TOK_LSDEV{$$="lsdev"} | TOK_MINUS{$$="-"} | TOK_TEMPLATE{$$=".template"}
       | TOK_CMDS{$$=".cmds"} | TOK_VAR{$$=".var"} | TOK_PLUS{$$="+"} | TOK_EQUAL{$$="="} 
       | TOK_GREATER{$$=">"} | TOK_DRAWABLE{$$="drawable"}
;

OCCR:   
        TOK_TENANT TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"color":$5}}
              $$=&getOCAttrNode{$3, cmd.TENANT, attributes}
        }
        |TOK_SITE TOK_COL PHYSICAL_PATH TOK_ATTRSPEC TOK_ORIENTATION {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"orientation":&strLeaf{$5}}}
              $$=&getOCAttrNode{$3, cmd.SITE, attributes}
        } 
        |TOK_BLDG TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}}
              $$=&getOCAttrNode{$3, cmd.BLDG, attributes}
        }
        |TOK_ROOM TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC TOK_ORIENTATION TOK_ATTRSPEC EXPR{
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7, "orientation":&strLeaf{$9}, "floorUnit":$11}}
              $$=&getOCAttrNode{$3, cmd.ROOM, attributes}
        }
        |TOK_ROOM TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC TOK_ORIENTATION {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7, "orientation":&strLeaf{$9}}}
              $$=&getOCAttrNode{$3, cmd.ROOM, attributes}
        }
        |TOK_ROOM TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "template":$7}}
              $$=&getOCAttrNode{$3, cmd.ROOM, attributes}
        }
        |TOK_RACK TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR {$$=&createRackNode{$3, [3]node{$5, $7, $9}}}
        |TOK_DEVICE TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR {$$=&createDeviceNode{$3, [3]node{$5, $7, nil}}}
        |TOK_DEVICE TOK_COL PHYSICAL_PATH TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR TOK_ATTRSPEC EXPR {$$=&createDeviceNode{$3, [3]node{$5, $7, $9}}}
        |TOK_CORIDOR TOK_COL PHYSICAL_PATH TOK_ATTRSPEC TOK_LBRAC EXPR TOK_COMMA EXPR TOK_RBRAC TOK_ATTRSPEC EXPR {
              attributes := map[string]interface{}{"leftRack":$6, "rightRack":$8, "temperature":$11}
              $$=&getOCAttrNode{$3, cmd.CORIDOR, attributes}
        }
        |TOK_CORIDOR TOK_COL PHYSICAL_PATH TOK_ATTRSPEC TOK_LBRAC EXPR TOK_COMMA EXPR TOK_COMMA TOK_RBRAC TOK_ATTRSPEC EXPR {
              attributes := map[string]interface{}{"leftRack":$6, "rightRack":$8, "temperature":$12}
              $$=&getOCAttrNode{$3, cmd.CORIDOR, attributes}
        }
        |TOK_GROUP TOK_COL PHYSICAL_PATH TOK_ATTRSPEC TOK_LBRAC GETOBJS TOK_RBRAC {$$=&createGroupNode{$3, $6}}
        |TOK_ORPH PHYSICAL_PATH TOK_DEVICE TOK_COL EXPR TOK_ATTRSPEC EXPR {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"template":$7}}
              $$=&getOCAttrNode{$5, cmd.STRAY_DEV, attributes}
        }
        |TOK_ORPH PHYSICAL_PATH TOK_SENSOR TOK_COL EXPR TOK_ATTRSPEC EXPR {
              attributes := map[string]interface{}{"attributes":map[string]interface{}{"template":$7}}
              $$=&getOCAttrNode{$5, cmd.STRAYSENSOR, attributes}
        }
       //EasyPost syntax STRAYSENSOR
       |OBJ_TYPE TOK_USE_JSON EXPR {$$=&easyPostNode{$1, $3}}
; 


/* HANDLEUI: TOK_UI TOK_DOT TOK_WORD TOK_EQUAL EXPR {$$=&handleUnityNode{[]interface{}{"ui", $3, ($5).(node).execute()}}}
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
          |TOK_GREATER P {$$=&focusNode{$2}} */
;


//Child devices of rack for group 
//Since the OCLI syntax defines no limit
//for the number of devices 
//a NonTerminal state is neccessary
/* CDORFG: TOK_ATTRSPEC TOK_WORDORNUM CDORFG {x:=$2; $$=x+","+$3}
       | {$$=""}
       ; */