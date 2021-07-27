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
  //n int
  s string
}

%token <s> TOKEN_WORD
%token <s> TOKEN_TENANT TOKEN_SITE TOKEN_BLDG TOKEN_ROOM
%token <s> TOKEN_RACK TOKEN_DEVICE TOKEN_SUBDEVICE TOKEN_SUBDEVICE1
%token <s> TOKEN_ATTR
%token
       TOKEN_CREATE TOKEN_GET TOKEN_UPDATE
       TOKEN_ATTR TOKEN_DELETE
       TOKEN_BASHTYPE TOKEN_EQUAL 
       TOKEN_CMDFLAG TOKEN_SLASH 
       TOKEN_EXIT TOKEN_DOC
       TOKEN_CD TOKEN_PWD
       TOKEN_CLR TOKEN_GREP TOKEN_LS
%type <s> F E P
%type <s> K NT_CREATE NT_DEL NT_GET NT_UPDATE


%%
start: K      {println("@State start");}
       | Q
;

K: NT_CREATE
       | NT_GET
       | NT_UPDATE
       | NT_DEL
;

NT_CREATE: TOKEN_CREATE E F {cmd.PostObj($2, "", resMap(&$3))/*println("@State NT_CR");*/}
       | TOKEN_CREATE E P F {$$=$4; /*println("Finally: "+$$);*/ cmd.Disp(resMap(&$4)); cmd.PostObj($2,$3, resMap(&$4)) }
;

NT_GET: TOKEN_GET {println("@State NT_GET"); cmd.GetObject("")}
       | TOKEN_GET P {cmd.GetObject($2)}
       | TOKEN_GET E P F {$$=$4;/*cmd.Disp(resMap(&$4)); cmd.SearchObjects($2, resMap(&$4))*/ }
;

NT_UPDATE: TOKEN_UPDATE E F {println("@State NT_UPD");}
       | TOKEN_UPDATE E P F {$$=$4;cmd.Disp(resMap(&$4)); cmd.UpdateObj($2, resMap(&$4))}
;

NT_DEL: TOKEN_DELETE E F {println("@State NT_DEL");cmd.DeleteObj($2, resMap(&$3))}
       | TOKEN_DELETE E P F {$$=$4;cmd.Disp(resMap(&$4)); cmd.DeleteObj($2,resMap(&$4) )}
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
       | TOKEN_ATTR TOKEN_EQUAL TOKEN_WORD {$$=$1+"="+$3; println("Taking the M"); 
       println("SUP DUDE: ", $3);}
;


P: TOKEN_WORD TOKEN_SLASH P {$$=$1+"/"+$3}
       | TOKEN_WORD {$$=$1}
       | TOKEN_WORD TOKEN_SLASH {$$=$1}
;



Q:     TOKEN_CD TOKEN_WORD TOKEN_CMDFLAG
       |TOKEN_CD TOKEN_WORD {cmd.CD($2)}
       |TOKEN_CD P {cmd.CD($2)}
       | TOKEN_LS P {cmd.LS($2)}
       | TOKEN_LS TOKEN_WORD TOKEN_CMDFLAG
       | BASH TOKEN_WORD
       | BASH     {cmd.Execute()}
;

BASH:  TOKEN_CD {cmd.CD("")}
       | TOKEN_CLR
       | TOKEN_GREP {}
       | TOKEN_LS {cmd.LS("")}
       | TOKEN_LS P {cmd.LS($2)}
       | TOKEN_PWD {cmd.PWD()}
       | TOKEN_EXIT     {cmd.Exit()}
       | TOKEN_DOC {cmd.Help()}
;


%%