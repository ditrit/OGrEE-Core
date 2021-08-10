%{
package main
import (
cmd "cli/controllers"
"strings"
"strconv"
)

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
}


%token <n> TOKEN_NUM
%token <s> TOKEN_WORD
%token <s> TOKEN_TENANT TOKEN_SITE TOKEN_BLDG TOKEN_ROOM
%token <s> TOKEN_RACK TOKEN_DEVICE TOKEN_SUBDEVICE TOKEN_SUBDEVICE1
%token <s> TOKEN_ATTR TOKEN_PLUS TOKEN_OCDEL
%token
       TOKEN_CREATE TOKEN_GET TOKEN_UPDATE
       TOKEN_DELETE TOKEN_SEARCH
       TOKEN_BASHTYPE TOKEN_EQUAL 
       TOKEN_CMDFLAG TOKEN_SLASH 
       TOKEN_EXIT TOKEN_DOC
       TOKEN_CD TOKEN_PWD
       TOKEN_CLR TOKEN_GREP TOKEN_LS TOKEN_TREE
       TOKEN_LSOG TOKEN_LSTEN TOKEN_LSSITE TOKEN_LSBLDG
       TOKEN_LSROOM TOKEN_LSRACK TOKEN_LSDEV
       TOKEN_LSSUBDEV TOKEN_LSSUBDEV1
       TOKEN_OCBLDG TOKEN_OCDEV
       TOKEN_OCRACK TOKEN_OCROOM TOKEN_ATTRSPEC
       TOKEN_OCSITE TOKEN_OCTENANT
       TOKEN_OCSDEV TOKEN_OCSDEV1 TOKEN_OCPSPEC
%type <s> F E P P1 ORIENTN WORDORNUM
%type <s> NT_CREATE NT_DEL NT_GET NT_UPDATE


%%
start: K
       | Q
       | OCLISYNTX
;

K: NT_CREATE     {println("@State start");}
       | NT_GET
       | NT_UPDATE
       | NT_DEL
;

NT_CREATE: TOKEN_CREATE E F {cmd.PostObj(cmd.EntityStrToInt($2),$2, resMap(&$3))/*println("@State NT_CR");*/}
       | TOKEN_CREATE E P F {$$=$4; /*println("Finally: "+$$);*/ cmd.Disp(resMap(&$4)); cmd.PostObj(cmd.EntityStrToInt($2),$2, resMap(&$4)) }
;

NT_GET: TOKEN_GET {println("@State NT_GET"); cmd.GetObject("")}
       | TOKEN_GET P {cmd.GetObject($2)}
       | TOKEN_GET E F {/*cmd.Disp(resMap(&$4)); */cmd.SearchObjects($2, resMap(&$3)) }
;

NT_UPDATE: TOKEN_UPDATE  F {println("@State NT_UPD"); cmd.UpdateObj("", resMap(&$2))}
       | TOKEN_UPDATE P F {$$=$3;/*cmd.Disp(resMap(&$4));*/ cmd.UpdateObj($2, resMap(&$3))}
;

NT_DEL: TOKEN_DELETE P {println("@State NT_DEL");cmd.DeleteObj($2)}
;

E:     TOKEN_TENANT 
       | TOKEN_SITE 
       | TOKEN_BLDG 
       | TOKEN_ROOM 
       | TOKEN_RACK 
       | TOKEN_DEVICE 
       | TOKEN_SUBDEVICE 
       | TOKEN_SUBDEVICE1 
;

ORIENTN: TOKEN_PLUS {$$=$1}
         | TOKEN_OCDEL {$$=$1}
         | {$$=""}
         ;

WORDORNUM: TOKEN_WORD {$$=$1}
           |TOKEN_NUM {x := strconv.Itoa($1);$$=x}
           |ORIENTN TOKEN_WORD ORIENTN TOKEN_WORD {$$=$1+$2+$3+$4}
           ;

F:     TOKEN_ATTR TOKEN_EQUAL WORDORNUM F {$$=string($1+"="+$3+"="+$4); println("So we got: ", $$)}
       | TOKEN_ATTR TOKEN_EQUAL WORDORNUM {$$=$1+"="+$3}
;


P:     P1
       | TOKEN_SLASH P1 {$$="/"+$2}
;

P1:    TOKEN_WORD TOKEN_SLASH P1 {$$=$1+"/"+$3}
       | TOKEN_WORD {$$=$1}
       | {$$=""}
;

Q:     TOKEN_CD TOKEN_WORD TOKEN_CMDFLAG
       |TOKEN_CD P {cmd.CD($2)}
       | TOKEN_LS P {cmd.LS($2)}
       | TOKEN_LSTEN P {cmd.LSOBJECT($2, 0)}
       | TOKEN_LSSITE P {cmd.LSOBJECT($2, 1)}
       | TOKEN_LSBLDG P {cmd.LSOBJECT($2, 2)}
       | TOKEN_LSROOM P {cmd.LSOBJECT($2, 3)}
       | TOKEN_LSRACK P {cmd.LSOBJECT($2, 4)}
       | TOKEN_LSDEV P {cmd.LSOBJECT($2, 5)}
       | TOKEN_LSSUBDEV P {cmd.LSOBJECT($2, 6)}
       | TOKEN_LSSUBDEV1 P {cmd.LSOBJECT($2, 7)}
       | TOKEN_TREE TOKEN_NUM {cmd.Tree("", $2)}
       | TOKEN_TREE P {cmd.Tree($2, 0)}
       | TOKEN_TREE P TOKEN_NUM {cmd.Tree($2, $3)}
       | BASH     {cmd.Execute()}
;

BASH:  TOKEN_CLR
       | TOKEN_GREP {}
       | TOKEN_LSOG {cmd.LSOG()}
       | TOKEN_PWD {cmd.PWD()}
       | TOKEN_EXIT     {cmd.Exit()}
       | TOKEN_DOC {cmd.Help("")}
       | TOKEN_DOC TOKEN_LS {cmd.Help("ls")}
       | TOKEN_DOC TOKEN_CD {cmd.Help("cd")}
       | TOKEN_DOC TOKEN_CREATE {cmd.Help("create")}
       | TOKEN_DOC TOKEN_GET {cmd.Help("gt")}
       | TOKEN_DOC TOKEN_UPDATE {cmd.Help("update")}
       | TOKEN_DOC TOKEN_DELETE {cmd.Help("delete")}
       | TOKEN_DOC TOKEN_WORD {cmd.Help($2)}
       | TOKEN_DOC TOKEN_TREE {cmd.Help("tree")}
       | TOKEN_DOC TOKEN_LSOG {cmd.Help("lsog")}
;

OCLISYNTX:  TOKEN_PLUS OCCR
            |OCDEL
            |OCUPDATE
            |OCGET
            ;


OCCR:   TOKEN_OCTENANT TOKEN_OCPSPEC P TOKEN_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.TENANT,map[string]interface{}{"attributes":map[string]interface{}{"color":$5}} ,rlPtr)}
        |TOKEN_TENANT TOKEN_OCPSPEC P TOKEN_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.TENANT,map[string]interface{}{"attributes":map[string]interface{}{"color":$5}} ,rlPtr)}
        |TOKEN_OCSITE TOKEN_OCPSPEC P TOKEN_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.SITE,map[string]interface{}{"attributes":map[string]interface{}{"orientation":$5}} ,rlPtr)}
        |TOKEN_SITE TOKEN_OCPSPEC P TOKEN_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.SITE,map[string]interface{}{"attributes":map[string]interface{}{"orientation":$5}} ,rlPtr)}
        |TOKEN_OCBLDG TOKEN_OCPSPEC P TOKEN_ATTRSPEC WORDORNUM TOKEN_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.BLDG,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}} ,rlPtr)}
        |TOKEN_BLDG TOKEN_OCPSPEC P TOKEN_ATTRSPEC WORDORNUM TOKEN_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.BLDG,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}} ,rlPtr)}
        |TOKEN_OCROOM TOKEN_OCPSPEC P TOKEN_ATTRSPEC WORDORNUM TOKEN_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.ROOM,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}} ,rlPtr)}
        |TOKEN_ROOM TOKEN_OCPSPEC P TOKEN_ATTRSPEC WORDORNUM TOKEN_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.ROOM,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}} ,rlPtr)}
        |TOKEN_OCRACK TOKEN_OCPSPEC P TOKEN_ATTRSPEC WORDORNUM TOKEN_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.RACK,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}} ,rlPtr)}
        |TOKEN_RACK TOKEN_OCPSPEC P TOKEN_ATTRSPEC WORDORNUM TOKEN_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.RACK,map[string]interface{}{"attributes":map[string]interface{}{"posXY":$5, "size":$7}} ,rlPtr)}
        |TOKEN_OCDEV TOKEN_OCPSPEC P TOKEN_ATTRSPEC WORDORNUM TOKEN_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.DEVICE,map[string]interface{}{"attributes":map[string]interface{}{"slot":$5, "sizeUnit":$7}} ,rlPtr)}
        |TOKEN_DEVICE TOKEN_OCPSPEC P TOKEN_ATTRSPEC WORDORNUM TOKEN_ATTRSPEC WORDORNUM {cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath($3)),cmd.DEVICE,map[string]interface{}{"attributes":map[string]interface{}{"slot":$5, "sizeUnit":$7}} ,rlPtr)}
       ; 
OCDEL:  TOKEN_OCDEL P {cmd.DeleteObj(replaceOCLICurrPath($2))}
;

OCUPDATE: P TOKEN_EQUAL WORDORNUM {println("Attribute Acquired"); newStr := replaceOCLICurrPath($1);  q := strings.LastIndex(newStr,"."); val := newStr[q+1:]+"="+$3; cmd.UpdateObj(newStr[:q], resMap(&val))}
;

OCGET: TOKEN_EQUAL P {cmd.GetObject(replaceOCLICurrPath($2))}
;

%%