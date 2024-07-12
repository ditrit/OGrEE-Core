# Contents <!-- omit in toc -->

- [Glossary](#glossary)
- [Language Syntax](#language-syntax)
   * [Comments](#comments)
   * [Variables](#variables)
      + [Set a variable](#set-a-variable)
      + [Use a variable](#use-a-variable)
   * [Expressions](#expressions)
      + [Primary expressions](#primary-expressions)
      + [Operators](#operators)
   * [Print](#print)
      + [String formatting](#string-formatting)
   * [Control flow](#control-flow)
      + [Conditions](#conditions)
      + [Loops](#loops)
      + [Aliases](#aliases)
- [Loading Commands](#loading-commands)
   * [Load commands from a text file](#load-commands-from-a-text-file)
   * [Commands over multiple lines](#commands-over-multiple-lines)
   * [Load template from JSON](#load-template-from-json)
- [Object Generic Commands](#object-commands)
   * [Select an object](#select-an-object)
      + [Select child object](#select-child-object)
      + [Select parent object](#select-parent-object)
   * [Focus an object](#focus-an-object)
   * [Get object](#get-object)
      + [Wildcards](#wildcards)
      + [Filters](#filters)
   * [Ls object](#ls-object)
      + [Layers](#layers)
      + [Filters](#filters-1)
   * [Tree](#tree)
   * [Delete object](#delete-object)
   * [Modify object attribute](#modify-object-attribute)
   * [Delete object attribute](#delete-object-attribute)
   * [Link/Unlink object](#linkunlink-object)
- [Object Specific Commands](#object-specific-commands)
   * [Domain](#domain)
   * [Site](#site)
      + [Set colors for zones of all rooms in a datacenter](#set-colors-for-zones-of-all-rooms-in-a-datacenter)
   * [Building](#building)
   * [Room](#room)
      + [Set reserved and technical zones of a room](#set-reserved-and-technical-zones-of-a-room)
      + [Room separators](#room-separators)
      + [Room pillars](#room-pillars)
      + [Interact with Room](#interact-with-room)
   * [Rack](#rack)
      + [Interact with Rack](#interact-with-rack)
   * [Device](#device)
      + [Interact with Device](#interact-with-device)
      + [Add Virtual Config](#add-virtual-config)
   * [Group](#group)
      + [Interact with Group](#interact-with-group)
   * [Corridor](#corridor)
   * [Generic Object](#generic-object)
   * [Virtual Object](#virtual-object)
   * [Tag](#tag)
      + [Apply tags to objects](#apply-tags-to-objects)
   * [Layer](#layer)
      + [Applicability Patterns](#applicability-patterns)
      + [Copy Layer](#copy-layer)
   * [Labels](#labels)
      + [Choose Label](#choose-label)
      + [Modify label's font](#modify-labels-font)
      + [Modify label's background color](#modify-labels-background-color)
- [Manipulate UI](#manipulate-ui)
   * [Delay commands](#delay-commands)
   * [Display infos panel](#display-infos-panel)
   * [Display debug panel](#display-debug-panel)
   * [Highlight object](#highlight-object)
- [Manipulate camera](#manipulate-camera)
   * [Move camera](#move-camera)
   * [Translate camera](#translate-camera)
   * [Wait between two translations](#wait-between-two-translations)

# Glossary

`[name]` is case sensitive. It includes the whole path of the object (for example: `tn/si/bd/ro/rk`)  
`[color]` is a hexadecimal code (*ffffff*)  

# Language Syntax
Just like a programming language, the CLI Language allows the user to use variables, control flow such as for loops and much more, as described below.

## Comments

You can put comments in an .ocli file with the `//` indicator. This can be useful in `.ocli` files, that is, text files with multiple CLI commands (check [Load commands from a text file](#load-commands-from-a-text-file)).

```
// This is a comment
+si:example@ffffff // This is another comment
```

## Variables

### Set a variable

```
.var:[name]=[value]
```

where ```[value]``` can be either:

- a vector ```[val1, ..., valn]```
- the result of a command via ```$(([command]))```
- a string (that can be [formatted](#string-formatting))
- the evaluation of an expression via ```eval [expr]```

```
.var: myvector = [1, 2, 3]
.var: mystring = hello world
.var: mystring = 41 + 1 // mystring will contain the string "41 + 1"
.var: mynumber = eval 41 + 1 // mynumber will contain the number 42
.var: mynumber = 42 // my number will contain the string "42"
                    // even though it can be used as a number in expressions
```

To unset a variable, that is, to completely remove it, use:
```
unset -v [name]
```

### Use a variable

`${[name]}` or `$[name]`  
in the second case, the longest identifier is used as ```[name]```
```
.var:siteNameVar=SITE
+si:$siteNameVar

.var:ROOM=/P/$siteNameVar/BLDG/ROOM1
${ROOM}/Rack:description="Rack of site $siteNameVar"

.var:i=eval (10+10)*10; // i=200
.var:j=eval $i/10-1;    // j=19
```

## Expressions

### Primary expressions

- booleans true / false
- integers
- floats
- string between double quotes
- vectors
- variable dereferencing

They can be used to build more complex expressions through operators.  

### Operators

#### Compute operators

these will only work if both side return a number  
+, -, *, /, \ (integer division), % (modulo)

#### Boolean operators
<, >, <=, >=, ==, !=, || (or), && (and) 

## Print

The ```print``` command prints the given string. The argument can take the same values as for variable assignments.

The ```printf``` command is equivalent to a ```print format```, see next section for details about ```format```.

### String formatting

You can dereference variables inside strings with ```${[name]}``` or ```$[name]```.

```
.var: result = eval 2+3
print 2+3 equals $result // prints "2+3 equals 5"
```

You can evaluate expressions inside strings with ```$(([expr]))```.

```
print 2+3 equals $((2+3)) // prints "2+3 equals 5"
```

For a sprintf-like formatting, you can use the format function, it uses the go fmt.Sprintf function under the hood (see <https://pkg.go.dev/fmt> for the specific syntax).

```
print format("2+3 equals %02d", 2+3) // prints "2+3 equals 05"
```

## Control flow

### Conditions
```
if condition { commands } elif condition { commands } else { commands }
```
`If-elif-else` statements have a similar syntax to Go. It expects an expression to be evaluated to true or false, followed by `{}` contaning the commands to execute if true. Examples:
```
> if 42 > 43 { print toto } elif 42 == 43 { print tata } else { print titi }
titi

// Multiple lines
if $shouldCreateSite == true {  \
  +si:/P/SITE                   \
  /P/SITE:reservedColor=AAAAAA  \
}
```

### Loops
```
for index in start..end { commands }
```
A `for` loop expects the name to give the index followed by a range and then `{}` with the commands. The range must be a start interger number followed by `..` and an end integer number superior to start. Examples: 
```
> for i in 0..3 { .var: i2 = $(($i * $i)) ; print $i^2 = $i2 }
0^2 = 0
1^2 = 1
2^2 = 4
3^2 = 9

// Multiple lines
for i in 0..5 {                                           \
    .var:r=eval 10+$i;                                    \
    .var:x=eval (36+$i*4)/3;                              \
    +rk:/P/SI/BLDG/ROOM/J${r}@[ $x, 52]@[80,120,42]@rear; \
    +rk:/P/SI/BLDG/ROOM/K${r}@[ $x, 55]@[80,120,42]@front \
}
```
Another loop comandavaiable is the `while`.
```
while condition { commands }
```
A while loop expects an expression to be evaluated to true or false, followed by `{}` contaning the commands to execute repeteadly with the expression remains true. Examples: 
```
>.var: i = 0; while $i<4 {print $i^2 = $(($i * $i)); .var: i = eval $i+1 }
0^2 = 0
1^2 = 1
2^2 = 4
3^2 = 9
```

### Aliases
```
alias name { commands }
```

An `alias` can be created to replace a list of commands, that is, to create a function without arguments. It expects the name of the alias followed by `{}` containing the commands it should evoke. Examples:
```
>alias pi2 { .var: i2 = $(($i * $i)) ; print $i^2 = $i2 }
>for i in 0..3 { pi2 }
0^2 = 0
1^2 = 1
2^2 = 4
3^2 = 9
```

To unset a function, that is, remove the alias, use:
```
unset -f [name]
```

# Loading Commands

## Load commands from a text file
```
.cmds:[path]
```

*`[path]` path of the file*

```
.cmds:/path/to/file.ocli
```

By convention, these files carry the extension .ocli. It should include a sequence of commands accepted by the CLI. For example, it can include the creation of variables used in a for loop with commands to create sites, buildings, rooms, etc. It can also include a `.cmds`command to call for another file with more commands. Check some examples [here](https://github.com/ditrit/OGrEE-Core/wiki/📗-%5BUser-Guide%5D-CLI-%E2%80%90-Get-Started#create-with-ocli-files).

## Commands over multiple lines

In .ocli script files, commands are usually separated by line breaks, however it is possible to have commands over multiple lines by using the \ character at the end of one or more consecutive lines, as shown below:

```
for i in 0..5 {                                           \
    .var:r=eval 10+$i;                                    \
    .var:x=eval (36+$i*4)/3;                              \
    +rk:/P/SI/BLDG/ROOM/J${r}@[ $x, 52]@[80,120,42]@rear; \
    +rk:/P/SI/BLDG/ROOM/K${r}@[ $x, 55]@[80,120,42]@front \
}
```

## Load template from JSON

```
.template:[path]
```
*`[path]` path of the json file*

```
.template:/path/to/template.json
```

# Object Commands

The following commands are generic and can be applied to most of the OGrEE objects.

## Select an object

```
=[name]
```
*If `[name]` is empty, go back to root*

```
=/Physical/Site/Building/RoomToSelect
```

### Select child object

Select one or several children of current selected object.  
*`[relativeName]` is the hierarchy name without the selected object part*

```
={[relativeName]}
={[relativeName],[relativeName],...}
```  

### Select parent object

```
=..
```

## Focus an object

*If `[name]` is empty, unfocus all items*

```
>[name]
```

## Get object

The information of an object or a list of objects can be obtained using the get command:

```
get [path]
```

where `[path]` can be either the path of a single object (get rack1) or a list of objects using wildcards (get rack1/*) (see [Wildcards](#wildcards)).

```
get SiteA
get /Physical/SiteB
get *                              // get all objs in current path
get * -f category=rack & height>10 // complex filter
get A*                             // get all objs with name starting by A
get A* category=rack               // simple filter
```

To see all possible options run:

```
man get
```

### Wildcards

The path used for the get command may contain wildcards:

Special Terms | Meaning
------------- | -------
`*`           | matches any sequence of non-path-separators (including 0 characters)
`/**/`        | matches zero or more directories

Any character with a special meaning can be escaped with a backslash (`\`).

For example:

- `get A*` will allow you to obtain all objects whose name begins with A (or is A).
- `get *A` will allow you to obtain all objects whose name finishes with A (or is A).
- `get *A*` will allow you to obtain all objects whose name has an A (or is A).
- `get *` will allow you to obtain all objects which are children of the current object.
- `get path/to/*` will allow you to obtain all objects which are children of `path/to`.

A doublestar (`**`) should appear surrounded by path separators such as `/**/`.
A mid-pattern doublestar (`**`) behaves like bash's globstar option: a pattern
such as `path/to/**` would return the same results as `path/to/*`. The
pattern you're looking for is `path/to/**/*`.

For example:

- `get **/*` will allow to obtain all the objects that are in the descending inheritance of the current object.
- `get **/A*` will allow to obtain all the objects that are in the descending inheritance of the current object whose name begins with A.

Using the -r parameter will automatically add a doublestar to your search being:

- `get -r A*` equivalent to `get **/A*`.
- `get -r *` equivalent to `get **/*`.

By using the -r option you can select the minimum and maximum depth of the search with the -m and -M parameters, while these options do not exist using `/**/` manually.

To see all possible options run:

```
man get
```

### Filters

Filters can be added to the `get` command to get only the objects that meet a certain characteristic.

#### Simple filters

Simple filters are the ones that use the egality operator (`=`). These filters can be used to get only objects that meet certain characteristics. Objects can be filtered by `name`, `slug`, `id`, `category`, `description`, `domain`, `tag` and by any other attributes, such as `size`, `height` and `rotation`.

It is also possible to specify a `startDate` and an `endDate`, to filter objects last modified _since_ and _up to_, respectively, the `startDate` and the `endDate`. Dates should be defined with the format `yyyy-mm-dd`.

Simple filters can be combined with commas (`,`), performing a logical `AND` operation.

```
get [path] tag=[tag_slug]
get [path] category=[category]
get [path] domain=[domain], height=[height]
get [path] startDate=[yyyy-mm-dd], endDate=[yyyy-mm-dd]
get [path] name=[name], category=[category], startDate=[yyyy-mm-dd]
```

#### Complex filters

Complex filters are an extension of the simple ones, and the same functionalities are applied.

Complex filters can be added to the `get` command with the flag `-f`, composing complex boolean expressions with the operators `=`, `!=`, `<`, `<=`, `>`, `>=`, `&` and `|`. Parenthesis can also be used to separate the expressions.

```
get [path] -f tag=[tag_slug]
get [path] -f domain=[domain], height=[height]
get [path] -f height=[height] & category=[category]
get [path] -f (height>=[height]) | rotation!=[rotation]
get [path] -f (name=[name] & rotation!=[rotation]) | size>[size]
get [path] -f category!=[category] & ((height<=[height] & size<[size]) | id=[id])
```

## Ls object

To obtain the children of an object and facilitate navigation over the hierarchy, the `ls` command can be used:

```
ls [path]
```

ls can also be used without [path] to do ls on the current path.
```
ls
ls DEMO_RACK/DeviceA
ls /Physical/SiteA
ls $x                             // using a variable
ls -s height                      // sort by height
ls -a height:size                 // show obj's height and size on ls result 
ls -r #racks                      // recursive, all levels below
ls . height=12                    // simple filter
ls . -f category=rack & height>10 // complex filter
```
To see all possible options run:

```
man ls
```

### Layers

When ls is performed on an object, the corresponding layers are added. These make navigation easier since they group the children of the object according to their characteristics.

The automatic layers are those that are added automatically depending on the entity of the object on which the ls is performed. They will appear only if at least one of the children of the object meets the conditions of the layer. The list of automatic layers added to each entity is described in the following sections.

In addition, custom layers can be created. For this, see [Create a Layer](#create-a-layer).

Layers can be used with the `ls` and `get` commands to list and to get the information of all the objects that meet the conditions of the layer. By default, only direct children of the object to which the layer belongs will be part of the response of these commands. To get results in the object hierarchy use the -r flag:

```
ls -r [layer_name]
```

or

```
get -r [layer_name]
```

#### Room's automatic layers

- \#corridors: children whose category is corridor
- \#groups: children whose category is group
- \#racks: children whose category is rack

#### Rack's automatic layers

- \#groups: children whose category is group
- \#[type]: for each type that the children of the device category have a layer will be created (example: #chassis)

#### Device's automatic layers

- \#[type]: for each type that the children of the device category have a layer will be created (example: #blades)

### Filters

Filters can be added to the `ls` command to get only the children that meet a certain characteristic.

#### Simple filters

Simple filters are the ones that use the egality operator (`=`). These filters can be used to get only children that meet certain characteristics. Objects can be filtered by `name`, `slug`, `id`, `category`, `description`, `domain`, `tag` and by any other attributes, such as `size`, `height` and `rotation`.

It is also possible to specify a `startDate` and an `endDate`, to filter objects last modified _since_ and _up to_, respectively, the `startDate` and the `endDate`. Dates should be defined with the format `yyyy-mm-dd`.

Simple filters can be combined with commas (`,`), performing a logical `AND` operation.

```
ls [path] tag=[tag_slug]
ls [path] category=[category]
ls [path] domain=[domain], height=[height]
ls [path] startDate=[yyyy-mm-dd], endDate=[yyyy-mm-dd]
ls [path] name=[name], category=[category], startDate=[yyyy-mm-dd]
```

#### Complex filters

Complex filters are an extension of the simple ones, and the same functionalities are applied.

Complex filters can be added to the `ls` command with the flag `-f`, composing complex boolean expressions with the operators `=`, `!=`, `<`, `<=`, `>`, `>=`, `&` and `|`. Parenthesis can also be used to separate the expressions.

```
ls [path] -f tag=[tag_slug]
ls [path] -f domain=[domain], height=[height]
ls [path] -f height=[height] & category=[category]
ls [path] -f (height>=[height]) | rotation!=[rotation]
ls [path] -f (name=[name] & rotation!=[rotation]) | size>[size]
ls [path] -f category!=[category] & ((height<=[height] & size<[size]) | id=[id])
```

#### Filter by category

Several commands are provided for running `ls` with category filter without typing them by hand: `lssite`, `lsbuilding`, `lsroom`, `lsrack`, `lsdev`, `lsac`, `lspanel`, `lscabinet`, `lscorridor`.

In addition, each of these commands accepts all the options of the `ls` command and the addition of more filters.

## Tree

To print the hierarchy below a path, at a certain depth :

```
tree // default path will be the current path (.)
tree [path] // default depth will be 1
tree [path] [depth]
```
Examples:
```
tree .
tree . 2
tree DEMO_RACK/DeviceA 2
tree $x 4
tree /Physical/SiteA
```

## Delete object

Works with single or multi selection.

```
-[name]
```
Examples:
```
-.
-BUILDING/ROOM
-selection // delete all objects previously selected
```

## Modify object attribute
```
[name]:[attribute]=[value]
```
To add or modify new attributes use the syntax bellow, giving the name of the object, followed by the attribute and its value. 
It also works with single or multi selection.  
*`[name]` can be `selection` or `_` for modifying selected objects attributes*

```
selection:[attribute]=[value]
_:[attribute]=[value]
```

- Object's domain can be changed recursively for changing all it's children's domains at the same time.

```
[name]:domain=[value]@recursive
```

- Object's description attribute is a string. Use `\n` to represent line-breaks.

```
[name]:description=[value]
// Example:
/P/SI/BLDG/R1/RACK:description="My Rack\nNew Servers\nWith GPU A4000"
```

- Object's clearance are vector6, they define how much gap (in mm) has to be left on each side of the object:

```
[name]:clearance=[front, rear, left, right, top, bottom]
// Example:
/P/SI/BLDG/R1/RACK:clearance=[800,500,0,0,0,0]
```

## Delete object attribute
```
-[name]:[attribute]
```
Use this command to remove not requirde attributes of a given object.

```
-/P/site/building:mycustomattr
```

## Link/Unlink object
Unlink an object transforms the object in a stray. In other words, it moves the object from the OGrEE hierarchy (no longer has a parent) and changes its type to stray object.

```
unlink [path/to/object]
``` 

Link an object reattaches the object to the OGrEE hierarchy, giving it a parent. It is possible to also set or modify attributes of the object by adding one or more `@attributeName=attributeValue` to the command. 

```
link [path/to/stray-object]@[path/to/new/parent]
link [path/to/stray-object]@[path/to/new/parent]@attributeName=attributeValue
```

Examples:

```
unlink /Physical/site/bldg/room/rack/device
link /Physical/Stray@/Physical/site/bldg/room/rack
link /Physical/Stray@/Physical/site/bldg/room/rack@slots=[slot1,slot2]@orientation=front
```

# Object Specific Commands

Each object entity has its own create command and may have some special commands to allow interaction.

## Domain

```
+domain:[name]@[color]
+do:[name]@[color]
```
To create a domain, a name and color should be provided. Should be in the  `/Organisation/Domain` (`/O/Domain` on short version) path or include it in the domain's name. 
Domains can have a hierarchy, that is, a domain can have a parent domain.  
*`[color]` should be a 6 digit HEX Value (ie 00000A)*

```
+domain:Newdomain@ff00ee
+do:/O/Domain/Newdomain/Newsubdomain@000AAA
```

## Site

```
+site:[name]  
+si:[name]
```
Sites have no parent, only a name is needed to create it.

```
+site:siteA // current path: /Physical
+si:/P/siteB
```

### Set colors for zones of all rooms in a datacenter

Add or modify the following color attributes of site:
```
[site]:usableColor=[color]
[site]:reservedColor=[color]
[site]:technicalColor=[color]
```  

## Building

```
+building:[name]@[pos]@[rotation]@[size]  
+building:[name]@[pos]@[rotation]@[template]  
+bd:[name]@[pos]@[rotation]@[size]
+bd:[name]@[pos]@[rotation]@[template]
```
Building must be child of a Site.  
*`[pos]` is a Vector2 [x,y] (m,m)  
`[rotation]` is the rotation of the building around its lower left corner, in degree  
`[size]` is a Vector3 [width,length,height] (m,m,m)  
`[template]` is the name (slug) of the building template*

```
+building:/P/siteA/BldgA@[5,5]@49.1@[300,300,300]
+bd:BldgA@[5,5]@-27.89@BldgTemplateA 
```

## Room

```
+room:[name]@[pos]@[rotation]@[size]@[axisOrientation]@[floorUnit]  
+room:[name]@[pos]@[rotation]@[template]
+ro:[name]@[pos]@[rotation]@[size]@[axisOrientation]@[floorUnit]  
+ro:[name]@[pos]@[rotation]@[template]
```
Room must be child of a building.  
Its name will be displayed in the center of the room in its local coordinates system.  
*`[pos]` is a Vector2 [x,y] (m,m)  
`[rotation]` is the rotation of the building around its lower left corner, in degree  
`[size]` is a Vector3 [width,length,height] (m,m,m)  
`[axisOrientation]` defines the orientation of the rows and columns. It can be any combination of [+/-]x[+/-]y. eg: +x+y or -x+y  
`[template]` is the name of the room template  
`[floorUnit]` is optional: by default set to "t" (tiles), can also be m (meters) or f (feet)*

```
+ro:/P/siteA/BldgA/R1@[0,0]@-36.202@[22.8,19.8,2]@+N+W@t
+room:/P/siteA/BldgA/R1@[0,0]@-36.202@RoomTemplateA
```


### Set reserved and technical zones of a room

```
[room]:areas=[reserved]@[technical]  
```
Enables tiles edges display.  
You can modify areas only if the room has no racks in it.  
**Technical** area : typically a restricted zone where power panels and AC systems are installed. separated from "IT space" with either a wall or a wire mesh  
**Reserved** area : some tiles around the room that must be kept free to move racks and walk (usually 2 or 3 tiles)  

*`[reserved]` is a vector4: [front,back,right,left] (tile,tile,tile,tile)  
`[technical]` is a vector4: [front,back,right,left] (tile,tile,tile,tile)*

```
/P/SI/BLDG/ROOM:areas=[2,2,2,2]@[3,1,1,1] 
```

### Room separators

Separators (wired or plain walls) can be added inside rooms. To do it, use:

```
[room]:separators+=[name]@[startPos]@[endPos]@[type]
```

Where:  
*`[name]` is an identifier for the separator  
`[startPos]` is a vector2: [x,y] (m,m)  
`[endPos]` is a vector2: [x,y] (m,m)  
`[type]` is the type of wall: wireframe or plain*  

It will add the given separator to `[room].attributes["separators"]`, which is a list of all its separators.

```
/P/SI/BLDG/ROOM:separators+=sep1@[1.2,10.2]@[1.2,14.2]@wireframe
```

Separators can be removed using:

```
[room]:separators-=[name]
```

Where:  
*`[name]` is the identifier of the separator to be removed  

### Room pillars

Pillars can be added inside rooms. To do it, use:

```
[room]:pillars+=[name]@[centerXY]@[sizeXY]@[rotation]
```

Where:  
*`[name]` is an identifier for the pillar  
`[centerXY]` is a vector2: [x,y] (m,m)  
`[sizeXY]` is a vector2: [x,y] (m,m)  
`[rotation]` is the angle of the pillar, in degrees*  

It will add the given pillar to `[room].attributes["pillars"]`, which is a list of all its pillars.

```
/P/SI/BLDG/ROOM:pillars+=pillar1@[4.22,3.85]@[0.25,0.25]@0
```

Pillars can be removed using:

```
[room]:pillars-=[name]
```

Where:  
*`[name]` is the identifier of the pillar to be removed

### Interact with Room

The same way you can modify object's attributes, you can interact with them through specific commands.

- Display or hide tiles name

```
[name]:tilesName=[true|false]
```

- Display or hide colors and textures

```
[name]:tilesColor=[true|false]
```

## Rack

```
+rack:[name]@[pos]@[unit]@[rotation]@[size]
+rack:[name]@[pos]@[unit]@[rotation]@[template]
+rk:[name]@[pos]@[unit]@[rotation]@[size]
+rk:[name]@[pos]@[unit]@[rotation]@[template]
```
Rack must be child of a room.  
`[pos]` is a Vector2 [x,y] (tile,tile) or a Vector3 [x,y,z] (tile,tile,cm) if the rack is wall mounted. It can be decimal or fraction. Can also be negative  
`[unit]` is t(tiles), m(meters) or f(feet)  
`[rotation]` is a Vector3 of angles or one of following keywords :  
  "front":  [0, 0, 180]  
  "rear":   [0, 0, 0]  
  "left":   [0, 90, 0]  
  "right":  [0, -90, 0]  
  "top":    [90, 0, 0]  
  "bottom": [-90, 0, 0]  
`[size]` is a Vector3 [width,length,height] (cm,cm,u)  
`[template]` is the name of the rack template

```
+rack:/P/siteA/BldgA/R1/A01@[1,2]@t@[0,0,180]@[60,120,42]
+rk:A01@[9,1]@t@[60,120,45]@BldgTemplate // current path /P/siteA/BldgA
```

### Interact with Rack

- Display or hide rack's box. This will also affect its label

```
[name]:alpha=[true|false]
```

- Display or hide rack's U helpers to simply identify objects in a rack.

```
[name]:U=[true|false]
```  

- Display or hide rack's slots

```
[name]:slots=[true|false]
```  

- Display or hide rack's local coordinate system

```
[name]:localCS=[true|false]
```  

## Device

If the parent rack doesn't have slots:

```
+device:[name]@[posU]@[sizeU]
+device:[name]@[posU]@[template]
```

If the parent rack has slots:

```
+device:[name]@[slot]@[sizeU]
+device:[name]@[slot]@[template]
+device:[name]@[slot]@[sizeU]@[invertOffset]
+device:[name]@[slot]@[template]@[invertOffset]
```  
A chassis is a *parent* device racked at a defined U position. 
All other devices (blades / components like processor, memory, adapters, disks...) have to be declared with a parent's slot and a template.

```
+device:[name]@[slot]@[template]
+device:[name]@[slot]@[template]@[invertOffset]
+device:[name]@[slot]@[template]@[invertOffset]@[side]
+dv:[name]@[slot]@[template]
+dv:[name]@[slot]@[template]@[invertOffset]
+dv:[name]@[slot]@[template]@[invertOffset]@[side]
```  
 
*`[posU]` is the position in U in a rack  
`[sizeU]` is the height in U in a rack  
`[slot]` is a square brackets list [] with the names of slots in which you want to place the device separated by a comma. Example: [slot1, slot2, slot3]. A shorter version with `..` can be used for a single range of slots: [slot1..slot3]. If no template is given, only one slot can be provided in the list.  
`[template]` is the name of the device template  
`[invertOffset]` is a boolean that tells the 3D client to invert the default offset for positioning the device in its slot (false by default, if not provided)  
`[side]` is from which side you can see the device if not "fullsize". This value is for overriding the one defined in the template. It can be front | rear | frontflipped | rearflipped*  
```
+dv:/P/siteA/BldgA/R1/A01/chassis@12@10
+dv:/P/siteA/BldgA/R1/A01/devA@[SlotA,SlotB]@10
+dv:/P/siteA/BldgA/R1/A01/devB@[SlotC]@DevTemplate
+dv:/P/siteA/BldgA/R1/A01/devB@[SlotC]@DevTemplate@true@front
```  

### Interact with Device

- Display or hide device's box. This will also affect its label

```
[name]:alpha=[true|false]
```

- Display or hide device's slots

```
[name]:slots=[true|false]
```

- Display or hide device's local coordinate system

```
[name]:localCS=[true|false]
```

### Add Virtual Config

A `virtual_config` can be added to a device to add information related to its virtual setup. It can be used to link a device to a virtual object using the `clusterId`. A simple use case is to add a `virtual_config` to a server to state that this server device is a "node" (`type`) of my cluster represented by a virtual object with id "my-proxmox-cluster" (`clusterId`).

```
[name]:virtual_config=[type]
[name]:virtual_config=[type]@[clusterId]
[name]:virtual_config=[type]@[clusterId]@[role]
```

*`[type]` is a string to describe the type of this device in its virtual setup  
`[clusterId]` is the ID of an existing virtual object to which this device will be linked  
`[role]` is a string to describe the role of this device in its virtual setup*  

```
/P/siteA/BldgA/R1/A01/server:virtual_config=node
/P/siteA/BldgA/R1/A01/server:virtual_config=node@my-proxmox-cluster
/P/siteA/BldgA/R1/A01/server:virtual_config=node@my-proxmox-cluster@proxmox
```  

## Group

```
+group:[name]@{c1,c2,...,cN}
+gr:[name]@{c1,c2,...,cN}
```
Group must be child of a room or a rack
A group is represented as a single box in the 3D client, containing all given children.

- If the group is a child of a room, it can contain racks and corridors.
- If the group is a child of a rack, it can contain devices.

`c1,c2,...,cN` are the short names (eg. A01 instead of /P/siteA/BldgA/R1/A01)

```
+gr:/P/siteA/BldgA/R1/GR1@{A01,A02,A03} // group child of room, contains racks
```

### Interact with Group

- Display or hide contained objects

```
[name]:content=[true|false]
```

## Corridor

```
+corridor:[name]@[pos]@[unit]@[rotation]@[size]@[temperature]
+co:[name]@[pos]@[unit]@[rotation]@[size]@[temperature]
```
Corridor must be child of a room.

`[pos]` is a Vector2 [x,y] (tile,tile) or a Vector3 [x,y,z] (tile,tile,cm) if the corridor is wall mounted. It can be decimal or fraction. Can also be negative  
`[unit]` is t(tiles), m(meters) or f(feet)  
`[rotation]` is a Vector3 of angles or one of following keywords :  
  "front":  [0, 0, 180]  
  "rear":   [0, 0, 0]  
  "left":   [0, 90, 0]  
  "right":  [0, -90, 0]  
  "top":    [90, 0, 0]  
  "bottom": [-90, 0, 0]  
`[size]` is a Vector3 [width,length,height] (cm,cm,cm)  
`[temperature]` is cold or warm.

```
+co:/P/siteA/BldgA/R1/CO1@[0,2]@t@[0,0,0]@[180,120,200]@warm
+co:/P/siteA/BldgA/R1/CO2@[3,2]@m@rear@[3*60,2*60,200]@cold
```

## Generic Object

Generic objects allow you to model any type of object that is not of the previous classes (tables, cabinets, doors, etc).

They must be child of a room.

To create them, use one of the following options:

```
+generic:[name]@[pos]@[unit]@[rotation]@[size]@[shape]@[type]
+generic:[name]@[pos]@[unit]@[rotation]@[template]
+ge:[name]@[pos]@[unit]@[rotation]@[size]@[shape]@[type]
+ge:[name]@[pos]@[unit]@[rotation]@[template]
```

Where:

- `[pos]` is a Vector3 [x,y,z] or a Vector2 [x,y] if z is 0. Each value can be decimal (1, 1.2, etc.) or fraction (1/2, 2/3, etc.). Can also be negative (-1, -1.2, -1/2).
- `[unit]` is the unit of the position [pos]. It can be: `t` (tiles), `m` (meters) or `f` (feet).
- `[rotation]` is a Vector3 of angles or one of following keywords:

  "front":  [0, 0, 180]  
  "rear":   [0, 0, 0]  
  "left":   [0, 90, 0]  
  "right":  [0, -90, 0]  
  "top":    [90, 0, 0]  
  "bottom": [-90, 0, 0]  
- `[size]` is a Vector3 [width,length,height] . All values are in cm.
- `[shape]` is a string defining the shape of the object. It can be: `cube`, `sphere` or `cylinder`.
- `[type]` is a string defining the type of the object. No predefined values.
- `[template]` is the name of the rack template

Examples:

```
+ge:/P/SI/BLDG/ROOM/BOX@[0,6,10]@t@[0,0,90]@[10,10,10]@cube@box
+ge:/P/SI/BLDG/ROOM/CHAIR@[5,5]@t@front@chair // with template
```

## Virtual Object

Virtual objects allow you to model any type of logical element that does not occupy a physical space (VMs, kubernetes clusters, docker containers, logical volumes, virtual switches, etc.).

It may not have a parent. If it does have a parent, it must be a device or another virtual object.

To create them, use one of the following options:

```
+vobj:[name]@[type]
+vobj:[name]@[type]@[vlinks]
+vobj:[name]@[type]@[vlinks]@[role]
```

Where:

- `[type]` is a string defining the type of the object.
- `[vlinks]` is an array with zero or more device ou virtual object IDs.
- `[role]` is a string defining the role of the object.

Examples:

```
+vobj:/P/SI/BLDG/ROOM/RACK/DEVICE/VM@vm
+vobj:/P/SI/BLDG/ROOM/RACK/DEVICE/LOGICALDISK@storage@[SI/BLDG/ROOM/RACK/DEVICE/PHYSICALDISK]
```

To add or remove vlinks to an existing virtual object use the following commands:
```
LOGICALDISK:vlinks+=SI/BLDG/ROOM/RACK/DEVICE/PHYSICALDISK
LOGICALDISK:vlinks-=SI/BLDG/ROOM/RACK/DEVICE/PHYSICALDISK
```



## Tag

Tags are identified by a slug. In addition, they have a color, a description and an image (optional). To create a tag, use:

```
+tag:[slug]@[color]
```

The description will initially be defined the same as the slug, but can be modified (see [Modify object's attribute](#modify-objects-attribute)). Image can only be added or modified through the APP.

After the tag is created, it can be seen in /Logical/Tags. The command `get /Logical/Tags/[slug]` can be used to get the tag information. In this get response, the field image contains a value that can be used to download the image from the API (check the endpoint /api/images in the API documentation).

```
+tag:gpu_servers@00ff00
```

### Apply tags to objects

Any object can be taged. When getting an object, it will contain a list of tags, example:

```
$ get /Physical/BASIC
{
    "id": "BASIC",
    ...
    "tags": [
        "demo"
    ]
}
```

To add a tag to an object use:

```
[name]:tags+=[tag_slug]
```

Where tag_slug is the slug of an existing tag, which can be found in /Logical/Tags. The tag _must_ be previously created (check [Create Tag](https://github.com/ditrit/OGrEE-Core/wiki/%F0%9F%93%97-%5BUser-Guide%5D-CLI-%E2%80%90-Language#create-a-tag)).

To remove a tag from an object use:

```
[name]:tags-=[tag_slug]
```

## Layer

Layers are identified by a slug. In addition, they have an applicability and the filters they apply. To create a layer, use:

```
+layer:[slug]@[applicability]@[filter]
```

The applicability is the path in which the layer should be added when running the `ls` command. Patterns can be used in the applicability (see [Applicability Patterns](#applicability-patterns)).

Layers can have simple filters in the format `field=value` or complex ones, composed of boolean expressions with the operators `=`, `!=`, `<`, `<=`, `>`, `>=`, `&` and `|`; parenthesis can also be used to separate the complex expressions. A first filter should be given to to create the layer.

```
+layer:Aobjs@/P/site/bldg/room@name=A* // all objs of room starting by A
// only racks that does not start by A:
+layer:RacksNotA@/P/site/bldg/room@category=racks & name!=A* 
```

To add more filters, simple or complex ones, edit the layer using the following syntax:

```
[layer_path]:filter+=[filter]
```

This action will add an `AND` operation between the new filter and the existing layer filter.

Examples:
```
[layer_path]:filter+=name=[name]
[layer_path]:filter+=(name=[name] & height<[height]) | domain=[domain]
```

Where [layer_path] is `/Logical/Layers/[slug]` (or only `[slug]` if the current path is /Logical/Layers).

To redefine the filter of a layer, edit using the following syntax:

```
[layer_path]:filter=[filter]
```

For the layer to filter the children whose category is device. When adding filters on different attributes, all must be fulfilled for a child to be part of the layer.

After the layer is created, it can be seen in /Logical/Layers. The command `get /Logical/Layers/[slug]` can be used to get the layer information.

### Applicability Patterns

The following special terms are supported in the patterns:

Special Terms | Meaning
------------- | -------
`/*`           | matches anything in the following directory
`/**/`        | matches zero or more directories
`?`           | matches any single character
`[class]`     | matches any single character against a class of characters (see [Character classes](#character-classes))
`{alt1,...}`  | matches a sequence of characters if one of the comma-separated alternatives matches

Any character with a special meaning can be escaped with a backslash (`\`).

A doublestar (`**`) should appear surrounded by path separators such as `/**/`.
A mid-pattern doublestar (`**`) behaves like bash's globstar option: a pattern
such as `path/to/**` would return the same results as `path/to/*`. The
pattern you're looking is probably `path/to/**/*`.

### Copy Layer

Currently it is only possible to copy **layers**. To copy an object use:

```
cp [source] [dest]
```

where `[source]` is the path of the object to be copied (currently only objects in /Logical/Layers are accepted) and `[dest]` is the destination path or slug of the destination layer.

#### Character Classes

Character classes support the following:

Class      | Meaning
---------- | -------
`[abc]`    | matches any single character within the set
`[a-z]`    | matches any single character in the range
`[^class]` | matches any single character which does *not* match the class
`[!class]` | same as `^`: negates the class

#### Examples

```
// layer available at all levels under /P/si/bldg 
// e.g. at /P/si/bldg/room/rack/device
+layer:[slug]@/P/si/bldg/**/*@[filter] 

// layer available only at levels directly under /P/si/bldg 
// e.g. at /P/si/bldg/room but not at /P/si/bldg/room/rack
+layer:[slug]@/P/si/bldg/*@[filter] 

// layer available only at room level with id starting as /P/site/bldg/RoomA
// e.g. at /P/site/bldg/RoomA1 and /P/site/bldg/RoomA2
+layer:[slug]@/P/si/bldg/RoomA?@[filter] 

// layer available only at levels /P/site/bldg/RoomA and /P/site/bldg/RoomB
// e.g. not at /P/site/bldg/RoomC
+layer:[slug]@/P/si/bldg/Room[AB]@[filter] 
```

## Labels

Some objects have a label displayed in there 3D model: racks, devices, rack groups and corridors.  
The default label is the object's name.

### Choose Label

You can change the label by a string or with a chosen attribute:  
*`#[attribute]` is one of the attribute of the object.*
*Use `\n` to insert line-breaks.* 

```
[name]:label=#[attribute]
[name]:label=[string]
```

Examples:
```
[name]:label=#id
[name]:label=This is a rack
[name]:label=My name is #name\nMy id is #id
```

### Modify label's font

You can make the font bold, italic or change its color.

```
[name]:labelFont=bold           //will toggle bold
[name]:labelFont=italic         //will toggle bold
[name]:labelFont=color@[color]  
```

### Modify label's background color

You can change the label's background color when it is hovering over the object.

```
[name]:labelBackground=[color]
```

# Manipulate UI

## Delay commands

You can put delay before each command: up to 2 seconds.

```
ui.delay=[time in seconds]
// Example: 
ui.delay=0.5 // 500ms
```  

## Display infos panel

```
ui.infos=[true|false]
```  

## Display debug panel

```
ui.debug=[true|false]
```  

## Highlight object

This is a "toggle" command: use it to turn on/off the highlighting of an object.  
If given object is hidden in its parent, the parent will be highlighted.

```
ui.highlight=[name]
ui.hl=[name]
// Example:
ui.hl=/P/SI/BLDG/R1/RACK
```

# Manipulate camera

## Move camera

```
camera.move=[position]@[rotation]
```  
Move the camera to the given point.  
*`[position]` is a Vector3: the new position of the camera  
`[rotation]` is a Vector2: the rotation of the camera*

```
camera.move=[-20.2;-71.98,21.32]@[37,0]
```  

## Translate camera

```
camera.translate=[position]@[rotation]
```  
Move the camera to the given destination. You can stack several destinations, the camera will move to each point in the given order.  
*`[position]` is a Vector3: the position of the camera's destination  
`[rotation]` is a Vector2: the rotation of the camera's destination*

```
camera.translate=[-17,15.5,22]@[78,-90]
```  

## Wait between two translations

```
camera.wait=[time]
```
You can define a delay between two camera translations.  
*`[time]` is the time to wait in seconds*

```
camera.wait=5 // 5s
```

