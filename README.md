# OGrEE-Core
## How to quickly deploy the API

Run ```docker compose -f deploy/docker/docker-compose.yml up```

## Config
To use the frontend (CLI, 3D, APP), a ```config.toml``` file must be created at the root of the repo.

Example :
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
APIKEY = ""
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
