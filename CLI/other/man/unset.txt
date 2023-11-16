USAGE:  unset [OPTIONS] [VAR/FUNC NAME]     
Deletes function or variable or attribute of an object   

There is an alternative usage for deleting an attribute of an object or an element in an attribute:   
```
unset [PATH/TO/OBJECT]:[ATTRIBUTE]
unset [PATH/TO/OBJECT]:[ATTRIBUTE][INDEX]
```   
Note that the attribute must be in quotes!
OPTIONS

-v    Deletes variable
-f    Deletes function

EXAMPLE   

    unset -v myVariable
    unset -f myFunc
    unset path/to/room:attribute1
    unset path/to/object:description[3]
    