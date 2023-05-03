USAGE:  ui . [UnityATTRIBUTE]=[VALUE]    
Sends a command to the Unity Client to adjust the ui by 'UnityATTRIBUTE' according to 'VALUE'    

OTHER USAGE: ui.clearcache
Clears unity cache

For more information please refer to:   
https://github.com/ditrit/OGrEE-3D/wiki/CLI-langage#Manipulate-UI


UI UNITY ATTRIBUTE DESCRIPTIONS

    delay - delay before each command max up to 2 seconds
    wireframe - display objects in wireframe mode
    infos - display infos panel
    debug - display debug panel
    highlight - selects an object for unity to highlight
    hl - same as highlight


UI UNITY ATTRIBUTE VALUE FORMATS

    delay = integer
    wireframe = bool
    infos = bool
    debug = bool
    highlight = path
    hl = path


EXAMPLE

    ui.delay = 99
    ui.wireframe = true
    ui.infos = false
    ui.debug = true
    ui.highlight = /Physical/DEMO/testSite/BuildingB
    ui.hl = /Physical/DEMO/testSite/BuildingB