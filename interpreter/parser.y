%{
package main
import (
cmd "cli/controllers"
"strings"
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
%}

%union {
  n int
  s string
}

%type <n> TOKEN_NUM
%token <n> TOKEN_NUM
%token <s> TOKEN_WORD
%token <s> TOKEN_TENANT TOKEN_SITE TOKEN_BLDG TOKEN_ROOM
%token <s> TOKEN_RACK TOKEN_DEVICE TOKEN_SUBDEVICE TOKEN_SUBDEVICE1
%token <s> TOKEN_ATTR
%token
       TOKEN_CREATE TOKEN_GET TOKEN_UPDATE
       TOKEN_ATTR TOKEN_DELETE TOKEN_SEARCH
       TOKEN_BASHTYPE TOKEN_EQUAL 
       TOKEN_CMDFLAG TOKEN_SLASH 
       TOKEN_EXIT TOKEN_DOC
       TOKEN_CD TOKEN_PWD
       TOKEN_CLR TOKEN_GREP TOKEN_LS TOKEN_TREE
       TOKEN_LSOG
%type <s> F E P P1
%type <s> NT_CREATE NT_DEL NT_GET NT_UPDATE


%%
start: NT_CREATE     {println("@State start");}
       | NT_GET
       | NT_UPDATE
       | NT_DEL
       | Q
;


NT_CREATE: TOKEN_CREATE E F {cmd.PostObj($2, "", resMap(&$3))/*println("@State NT_CR");*/}
       | TOKEN_CREATE E P F {$$=$4; /*println("Finally: "+$$);*/ cmd.Disp(resMap(&$4)); cmd.PostObj($2,$3, resMap(&$4)) }
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

F:     TOKEN_ATTR TOKEN_EQUAL TOKEN_WORD F {$$=string($1+"="+$3+"="+$4); println("So we got: ", $$)}
       | TOKEN_ATTR TOKEN_EQUAL TOKEN_WORD {$$=$1+"="+$3}
;


P:     P1
       | TOKEN_SLASH P1 {$$="/"+$2}
;

P1:    TOKEN_WORD TOKEN_SLASH P1 {$$=$1+"/"+$3}
       | TOKEN_WORD {$$=$1}
       | {$$=""}
;

Q:     TOKEN_CD TOKEN_WORD TOKEN_CMDFLAG
       |TOKEN_CD TOKEN_WORD {cmd.CD($2)}
       |TOKEN_CD P {cmd.CD($2)}
       | TOKEN_LS P {cmd.LS($2)}
       | TOKEN_LS TOKEN_WORD TOKEN_CMDFLAG
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


%%