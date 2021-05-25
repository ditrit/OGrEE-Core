%{
package main
import ("os"
"cli/cmd"
)
%}

%union {
  //n int
  s string
}

%token <s> TOKEN_WORD
%token <s> TOKEN_ENTITY
%token
       TOKEN_CRUDOP TOKEN_ATTR 
       TOKEN_BASHTYPE TOKEN_EQUAL 
       TOKEN_CMDFLAG TOKEN_SLASH 
       TOKEN_EXIT TOKEN_DOC



%%
start: K      {println("@State start");}
       | Q
       | D
       |L
;

K:      TOKEN_CRUDOP E    {println("@State K");}
       | TOKEN_CRUDOP Z P F
;


E:     TOKEN_ENTITY F
;

F:     TOKEN_ATTR TOKEN_EQUAL TOKEN_WORD F
       | TOKEN_ATTR TOKEN_EQUAL TOKEN_WORD M {println("Taking the M"); println("SUP DUDE: ", $3)}
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

D:    TOKEN_EXIT     {os.Exit(0)}
;

L:     TOKEN_DOC {cmd.Help()}
;

%%