%{
package main
import (
"cli/cmd"
"cli/utils"
)

/*func resMap(x *string) map[string]string {
       
}*/
%}

%union {
  //n int
  s string
}

%token <s> TOKEN_WORD
%token <s> TOKEN_ENTITY
%token <s> TOKEN_ATTR
%token
       TOKEN_CRUDOP TOKEN_ATTR 
       TOKEN_BASHTYPE TOKEN_EQUAL 
       TOKEN_CMDFLAG TOKEN_SLASH 
       TOKEN_EXIT TOKEN_DOC
%type <s> F 
%type <s> K 
/*func clrRes(x map[string]string) {
	for k := range x {
		delete(x,k)
	}
}*/


%%
start: K      {println("@State start");}
       | Q
       | D
       |L
;

K:      TOKEN_CRUDOP E    {println("@State K");}
       | TOKEN_CRUDOP Z P F {$$=$4; println("Finally: "+$$)}
;


E:     TOKEN_ENTITY F
;

F:     TOKEN_ATTR TOKEN_EQUAL TOKEN_WORD F {$$=string($1+"#"+$3+","+$4); println("So we got: ", $$)}
       | TOKEN_ATTR TOKEN_EQUAL TOKEN_WORD M {$$=$1+"#"+$3; println("Taking the M"); 
       println("SUP DUDE: ", $3);}
;

M: 
;

Z: TOKEN_ENTITY
;

P: TOKEN_WORD TOKEN_SLASH P
       | TOKEN_WORD
;


Q:    B
;

B:     TOKEN_BASHTYPE TOKEN_WORD TOKEN_CMDFLAG
       | TOKEN_BASHTYPE TOKEN_WORD
       | TOKEN_BASHTYPE     {cmd.Execute()}
;

D:    TOKEN_EXIT     {utils.Exit()}
;

L:     TOKEN_DOC {cmd.Help()}
;

%%