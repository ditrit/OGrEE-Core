The CLI is the entry to the API and the 3D client. Check the README from the CLI folder on how to build the CLI.

## Configuration
To use the CLI as well as the 3D client, a `config.toml` file must be created at the root of the repo. If you are not cloning the repo and just executing the CLI binary, use the `--conf-path` launch argument to pass the config file location. If no file is provided, default values will be used just as defined by the example `config.toml` file below:

```
[OGrEE-CLI]
Verbose = "ERROR"
APIURL = "http://127.0.0.1:3001"
UnityURL = "127.0.0.1:5500"
UnityTimeout = "10ms"
HistPath = "./.history"
Script = ""
Drawable = ["all"]
DrawLimit = 100
Updates = ["all"]
User = ""
Variables = [
    {Name = "ROOT", Value = "path_to_root"},
    {Name = "ROOT2", Value = "$ROOT/path_to_root2"},
]

[OGrEE-CLI.DrawableJson]
tenant = "./other/drawTemplates/tenant.json"

[OGrEE-3D]
verbose = true
fullscreen = false
cachePath = "C:/"
cacheLimitMo = 100
# Port used to receive messages     from OGrEE-CLI. 5500 by default
cliPort = 5500
# Value clamped from 0 to 100
alphaOnInteract = 50

[OGrEE-3D.textures]
# Textures loaded in the 3D client at startup, in "name = url" format
perf22 = "https://raw.githubusercontent.com/ditrit/OGREE-3D/master/Assets/Resources/Textures/TilePerf22.png"
perf29 = "https://raw.githubusercontent.com/ditrit/OGREE-3D/master/Assets/Resources/Textures/TilePerf29.png"

[OGrEE-3D.colors]
# Colors can be defines with hexadecimal codes or html colors
selection = "#21FF00"
edit = "#C900FF"
focus = "#FF9F00"
highlight = "#00D5FF"

usableZone = "#DBEDF2"
reservedZone = "#F2F2F2"
technicalZone = "#EBF2DE"

[OGrEE-3D.temperature]
# Minimum and maximum values for temperatures in Celsius and Fahrenheit to define the range of the temperature gradients
minC = 0
maxC = 100
minF = 32
maxF = 212

# Define a custom gradient by defining up to 8 colors in rgba format (rgb from 0 to 255, a from 0 to 100)
useCustomGradient = true
customTemperatureGradient = [
    [0,0,255,0],
    [255,0,0,100],
    [255,255,0,50]
]
```

## Launch arguments
These arguments override the respective values found in the config.toml file.

* --conf_path (or -c) : Specify the location of the config.toml file
* --verbose (or -v) : Indicates level of debugging messages.The levels are in ascending order:{NONE,ERROR,WARNING,INFO,DEBUG}. (default "ERROR")
* --unity_url (or -u) : Specify the Unity URL
* --api_url (or -a) : Specify API URL
* --history_path (or -h) : Specify location of the .history file
* --file (or -f) : Interpret an OCLI script file

## Create your first objects

After launching and logging in the CLI, we land in a prompt where we can start typing commands. The CLI is organized like a filesystem but here, instead of files, we have OGrEE objects disposed in a hierarchy with OGrEE namespaces as root. More information about namespaces and the OGrEE hierarchy can be found in the [Basic Concepts](https://github.com/ditrit/OGrEE-Core/wiki/%F0%9F%93%97-%5BUser-Guide%5D-Basic-Concepts) page. To navigate the tree of objects, Unix inspired commands can be used, such as `ls` to see the objects of the current level and `cd` to change level. Too see all possible commands, see the CLI Language page.

Let's create some OGrEE components! First, we will create a site named `siteA`:
```
+site:/P/siteA
```

The `/P/` is a prefix to the complete path and is the same as `/Physical/`. We can also go to Physical and create the site by only giving the name:
```
cd /Physical
+site:siteA
```
Now, let's create a building named `blgdA` under our new site:

```
+bd:/P/siteA/blgdA@[0,0]@-90@[25,29.4,1]
```
Here, we also add some attributes with `@` followed by the attribute's value in the order and format described in the [Create Building](https://github.com/ditrit/OGrEE-Core/wiki/%F0%9F%93%97-%5BUser-Guide%5D-CLI-%E2%80%90-Language#create-a-building) command. 
Next, we will create a room named `R1`:
```
+ro:/P/siteA/blgdA/R1@[0,0]@0@[22.8,19.8,0.5]@+x+y
```
Instead of giving attributes, another option for the [Create Room](https://github.com/ditrit/OGrEE-Core/wiki/%F0%9F%93%97-%5BUser-Guide%5D-CLI-%E2%80%90-Language#create-a-room) command is to, first, create a room [template](https://github.com/ditrit/OGrEE-Core/wiki/%F0%9F%93%97-%5BUser-Guide%5D-API-%E2%80%90-JSON-templates-definitions) and then create a room passing the template as an attribute. This will automatically apply the template values to the new room.
Inside the room, we will now create a two racks, `A01` and `A02`:
```
+rk:/P/siteA/blgdA/R1/A01@[1,2]@t@[0,0,180]@[60,120,42]
+rk:/P/siteA/blgdA/R1/A02@[2,2]@t@[0,0,180]@[60,120,42]
```
The [Create Rack](https://github.com/ditrit/OGrEE-Core/wiki/%F0%9F%93%97-%5BUser-Guide%5D-CLI-%E2%80%90-Language#create-a-rack) (above) as well as the [Crete Device](https://github.com/ditrit/OGrEE-Core/wiki/%F0%9F%93%97-%5BUser-Guide%5D-CLI-%E2%80%90-Language#create-a-device) (below) command can also use templates. Let's create a simple device:
```
+dv:/P/siteA/blgdA/R1/A01/DeviceA@1@3
```
We have created a complete hierarchy of OGrEE objects! We can check it out with the [Tree](https://github.com/ditrit/OGrEE-Core/wiki/%F0%9F%93%97-%5BUser-Guide%5D-CLI-%E2%80%90-Language#tree) command:
```
> tree /Physical 5
/Physical
├── Stray
└── siteA
    └── blgdA
        └── R1
            ├── A01
            │   └── DeviceA
            └── A02
```
If you have a local 3D client running and connected to your CLI, you can then see your recently created objects in 3D with the draw command:
```
draw /Physical/siteA 5
```