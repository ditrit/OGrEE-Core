USAGE:  unset [OPTIONS] [VAR/FUNC NAME]     
Deletes function or variable or attribute of an object   

There is an alternative usage for deleting an attribute of an object:   
```
unset [PATH/TO/OBJECT]:[ATTRIBUTE]
```   
Note that the attribute must be in quotes!
OPTIONS

-v    Deletes variable
-f    Deletes function

EXAMPLE   

    unset -v myVariable
    unset -f myFunc
    unset path/to/room:attribute1
    