USAGE:  unset [VAR/FUNC NAME] [OPTIONS]   
Deletes function or variable   

There is also an alternative usage for deleting an attribute of an object:   
```
unset ($[VAR NAME][ATTRIBUTE])
```   
Note that the attribute must be in quotes!
OPTIONS

-v    Deletes variable
-f    Deletes function

EXAMPLE   

    unset -v myVariable
    unset -f myFunc
    