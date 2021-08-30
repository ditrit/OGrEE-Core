Introduction
------------
The OGREE Language Reference. The scripting language is modelled to behave like bash but has some differences.


Variables 
------------
 In OGREE, all variables by default are defined as global, even if declared inside a function. There is no support yet for local variables.

 Variables shall be declared as follows:
 ```
 .var:myvar=xyz
 ```

 Variables can be unset via the unset command:
 ```
 unset -v myvar
 ```
 Variables can have their values changed in 2 ways:
 ```
 .var:myvar=123
 $myvar=123
 ```

Functions
------------
Functions have only one way of declaration and just like bash, they are not executed unless exclusively invoked. 
Function declaration:
```
myfunc() {command1;command2;...}
```
Use:
```
myfun
```
Unlike bash, semicolons must be added to the end of each command if a block has more than 1 command. Functions can also be unset using the unset -f command:
```
unset -f myfunc
```
Because all variables are global, functions do not support parameters


Comparators
------------
For now comparisons only work between INTEGER type variables

Loops
------------
Loops are of varying types in Bash, there is limited support for the 'dynamic type' (such as iterating over a result of a command, range etc.) of loops that are found in bash. Those dynamic loop types are still in progress. OGREE supports FOR and WHILE loops

For Loops:
```
for var in {INTEGER..INTEGER}; {commands;} done
for ((init; condition; increment)); {commands;} done
```


Execution Control
------------
If statements are of 3 types:
```
if [condition] then {} fi
if [condition] then {} else {} fi
if [condition] then {} elif [condition] then {} else {} fi
```

Scripts
------------
Scripts can be loaded. The commands follow the OGREE language specification, with the exception of multi line commands such as functions and loops. Multi Line commands must have a '\\' before each newline. The last line shall not have the '\\'. The file extension does not matter, for now the only way to invoke a script is to launch the OGREE shell and:
```
.cmds:[PATH/TO/YOUR/FILE]
```