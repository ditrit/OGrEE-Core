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

 ### Variable types
 ```
 string
int 
node
json
bool
array
```

### Arrays
Arrays are declared as:
```
.var:array=(x y z)
```
Index into arrays:
```
$array[1]
```
$array is equivalent to $array[0]

### Modifying Nodes
Nodes cannot be created manually and are obtained as a result of a command.
Node attributes can be modified using the following syntax:
```
.var:x=$gt
$x[ATTRIBUTE]="someValue"
```
Where ATTRIBUTE is an attribute and "someValue" must be in quotes


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

### Function Return Types
```
gt          -> node
gt (search) -> []node
create      -> node
delete      -> bool
update      -> json //containing only the changed entries
ls          -> []node
cd          -> string
print       -> string
pwd         -> string
.cmds       -> string
selection   -> []string
tree        -> null
lsog        -> null
man         -> null
```

### Assigning function return values
```
.var:x=$(ls)
```

Comparators
------------
Comparisons exclusively work between variables of the same type. **NOTE**
That almost all members of a node data type are string

Loops
------------
Loops are of varying types in Bash, there is limited support for the 'dynamic type' (such as iterating over a result of a command, range etc.) of loops that are found in bash. Those dynamic loop types are still in progress. OGREE supports FOR and WHILE loops

For Loops:
```
for var in {INTEGER..INTEGER}; {commands;} done
for ((init; condition; increment)); {commands;} done
for var in expression; {commands;} done
```

While Loop:
```
while (expression) {commands;} done
```

Range / Dynamic:
```
for var in $(command) do {commands;} done
```

### Special Case
Iterating through array variables is not possible using the range loop.
```
.var:array=(1 2 3 4)
for k in len(array); {commands;} done
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


Command Substitution
------------
```
.var:a=cd
$a
```

Debugging
------------
There are 3 levels of debugging messages output. And are specified as program arguments ```-v=x``` where x is in the range 0 -> 3. Any number above 3 is still valid. When 
```
-v=1       Normal debugging messages output
-v=2       Normal + Lexer messages output
-v=3       Normal + Lexer + Parser messages output
```
If program is executed with no arguments then the default level is 0