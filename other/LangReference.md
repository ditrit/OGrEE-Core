Introduction
------------
The OGREE Language Reference. The OGREE Shell command interpreter interfaces with the API and optionally a Unity viewer. The command interpreter provides a command line interface for data centre management.   

The scripting language is modelled to behave like bash but has many differences. It has syntax similar to ruby/python and bash. Filename globbing (wildcard matching), piping, here documents are not supported. Brace expansion has very limited support. 


Environment 
------------
At any point in time the 'Unity' environment variable may be set to either false or true. This variable indicates to the shell whether or not to inform the Unity viewer of any updates
```
Unity = true
```
This environment variable is automatically set upon startup and it's value depends upon whether or not the shell was able to establish contact with the Unity viewer. The user would be notified of this.   

This environment variable is the only variable available at this time. 

Comments 
------------
Comments use the double slash characters '//' . They continue until the end of the line, this is the only way to input comments and there is no way to indicate multi-line comments. 

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
float  
node
json
bool
array
```


### Strings
Strings are exclusively surrounded by double quotes "", and anything may be placed in them. Single quotes '' are not used and thus are not part of the language. Escape sequences are not yet supported by the language.
Strings can be concatenated:   
```
"my string" + "another string"
```   
Variables in a string will not be dereferenced. To dereference a variable in a string you must concatenate:   
```
"this a custom string" + $x 
```

### Arrays
Arrays are declared as:
```
.var:array=(x, y, z)
```
Index into arrays:
```
$array[1]
```
$array is equivalent to $array[0]   
Arrays are not restricted to holding values of a single data type:   
```
.var:array=("this is a string", 808.9, false, "another string")
```

Arrays can not immediately have their lengths changed. And can only be changed by reassigning the variable.   
Single element arrays are not supported. If a single element array is assigned it will be treated as a variable of the data's respective type   

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
.var:array=(1, 2, 3, 4)
for k in len(array); {$array[$k]=999;commands;} done
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
.cmds:"PATH/TO/YOUR/FILE"
```


Command Substitution
------------
```
.var:a=cd
$a
```

Updating Objects
------------
Unless referring to an actual dir for executing scripts, representing current dir by using a '.' is not possible, instead leave the path empty. 
Objects can be updated in 2 ways
```
update [PATH(optional)] : attribute=myNewValue
PATH:attribute=myNewValue
```
It is possible to update multiple objects via OCLI Syntax. The objects to be edited must be first 'selected' as follows. The paths to each objects must be separated by a comma.
```
={Path/to/object, path/to/another/object, path/to/another/object2}
```
Finally execute
```
selection.myAttribute = someValue
```
If the attribute does not exist, it will be inserted automatically under attributes dictionary. Any string values MUST be surrounded by quotes (" "). Please note that object checking will only occur on the update command.

It is possible to delete an attribute of an object. First a variable must be assigned to an object, then unset can be called. Note: The attribute to be deleted must be surround by quotes (" ") or referred to as a string variable.
```
.var:x=$(gt)
unset $x["myAttribute"]
```

Easy Create & Update Syntax
------------
You can create and update objects much faster by specifying the object via a json file. The JSON file is a raw description of the object (nothing related to Ogree CLI shall be described) and follows the standard that is described on the wiki.   
```
create tenant useJn:path/to/your/json/file
+tn:useJn:path/to/your/json/file
```   
For updating the keyword partial is used to indicate if you want to perform a partial update. A full PUT operation will be performed otherwise   
```
update path/to/object : usJn partial : path/to/json/file
update path/to/object : usJn : path/to/json/file
```

### OCLI Syntax
There is additional syntax for fast object creation:
#### Create a Tenant
```
+tenant:path/to/tenantName@[color]
+tn:path/to/tenantName@[color]
```


#### Create a Site
```
+site:path/to/siteName@[name]@[orientation]  
+si:path/to/siteName@[name]@[orientation]
```
Placeholder values are assigned to the object upon creation so as to satisfy the object requirements:
- usableColor = "DBEDF2"
- reservedColor = "F2F2F2"
- technicalColor = "EBF2DE"
- domain = Tenant's Name

#### Create a Building
[pos] is a Vector2 [x,y] (m,m)
[size] is a Vector3 [width,length,height] (m,m,m)
```
+building:path/to/bldgName@[pos]@[size]  
+bd:path/to/bldgName@[pos]@[size]
```
Placeholder values are assigned to the object upon creation so as to satisfy the object requirements:
- posXYUnit = "m"
- sizeUnit = "m"
- heightUnit = "m"
- height = 0
- domain = Tenant's Name 

#### Create a Room
[pos] is a Vector2 [x,y] (m,m)
[size] is a Vector3 [width,length,height] (m,m,m) [template] is the name of the room template
[orientation] is more precise than the building's one: you have to specify the direction of the rows/columns: [+/-X][+/-Y]. eg: +E+N or -W+S
[floorUnit] is optional: by default set to "t" (tiles), can also be m (meters) or f (feet)
```
+room:path/to/roomName@[pos]@[size]@[orientation]@[floorUnit]  
+room:path/to/roomName@[pos]@[template]
+ro:path/to/roomName@[pos]@[size]@[orientation]@[floorUnit]  
+ro:path/to/roomName@[pos]@[template]
+room:path/to/roomName@[pos]
+ro:path/to/roomName@[pos]
```
Placeholder values are assigned to the object upon creation so as to satisfy the object requirements:
- floorUnit = "t"
- orientation = "+N+E"
- posXYUnit = "m"
- sizeUnit = "m"
- height = 0
- heightUnit = "m"
- domain = Tenant's Name 

#### Create a Rack
```
+rack:path/to/roomName@[pos]@[size]   
+rack:path/to/roomName@[pos]@[template]   
+rk:path/to/roomName@[pos]@[size]   
+rk:path/to/roomName@[pos]@[template]   
```
Placeholder values are assigned to the object upon creation so as to satisfy the object requirements:
- sizeUnit = "m"
- height = 0
- heightUnit = "m"
- posXYUnit = "t"
- orientation = "front"
- domain = Tenant's Name

#### Create a Device
```
+device:path/to/deviceName@[slot]@[sizeUnit]@[side]
+dv:path/to/deviceName@[slot]@[sizeUnit]
```
Placeholder values are assigned to the object upon creation so as to satisfy the object requirements:
- orientation = "front"
- size = 0
- height = 0
- heightUnit = "mm"
- domain = Tenant's Name

#### Create a Corridor
```
+corridor:path/to/corridorName@r1@r2@[temperature]
+co:path/to/corridorName@r1@r2@[temperature]
```

#### Create a Group
Group must be child of a room or a rack
If the group is a child of a room, it can contain racks and corridors.
If the group is a child of a rack, it can contain devices. c1,c2,...,cN are the short names (eg. A01 instead of tn.si.bd.ro.A01)

```
+group:[name]@{c1,c2,...,cN}
+gr:[name]@{c1,c2,...,cN}
```

Arithmetic 
------------
*,/,%,- operators are supported. In the event that one of the arguments is a float and the other being an int, the int will be 'promoted' to the float type. There is no way to cast a float to an int or vice versa. 


I/O Redirection
------------
As of now, I/O Redirection is not supported yet but will be supported in the future. 


Other Operations
------------
Print
```
print "some string"
```
Variable values can be printed
```
print $x
```
Create Object
```
create tenant /Physical/ : name=myTenant domain=myTenant color=FF
```
Delete Object
```
delete path/to/object
- path/to/object
```

Unity Commands
------------
Commands that manipulate the Unity UI are now possible   

#### Delay commands
You can put delay before each command: up to 2 seconds.
```
ui.delay=[time]
```
#### Enable/Disable wireframe mode
```
ui.wireframe=[true|false]
```

#### Display infos panel
```
ui.infos=[true|false]
```

#### Highlight object
This is a "toggle" command: use it to turn on/off the highlighting of an object.
If given object is hidden in its parent, the parent will be highlighted.   
```
ui.highlight=[name]
ui.hl=[name]
```

#### Move camera
Move the camera to the given point.   
[position] is a Vector3: the new position of the camera   
[rotation] is a Vector2: the rotation of the camera   
```
camera.move=[position]@[rotation]
```

#### Translate camera
Move the camera to the given destination. You can stack several destinations, the camera will move to each point in the given order.   
[position] is a Vector3: the position of the camera's destination   
[rotation] is a Vector2: the rotation of the camera's destination   
```
camera.translate=[position]@[rotation]
```

#### Wait between two translations
You can define a delay between two camera translations.   
[time] is the time to wait in seconds   
```
camera.wait=[time]
```


Misc.
------------
The shell history is stored in a file in the folder  ``` .resources/.history``` of the present Shell executable directory.    
   
The user credentials, API URL, Unity URL info are stored in the file  ``` .resources/.env``` of the present Shell executable directory. You can change the URLs to point to anywhere etc.


Debugging
------------
There are 3 levels of debugging messages output. And are specified as program arguments ```-v=x``` where x is in the range 0 -> 3. Any number above 3 is still valid. When 
```
-v=1       Normal debugging messages output
-v=2       Normal + Lexer messages output
-v=3       Normal + Lexer + Parser messages output
```
If program is executed with no arguments then the default level is 0