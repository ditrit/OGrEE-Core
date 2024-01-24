# Contents <!-- omit in toc -->

- [Building template](#building-template)
- [Room template](#room-template)
  - [Separators definition](#separators-definition)
  - [Pillars definition](#pillars-definition)
  - [Tiles definition](#tiles-definition)
  - [Rows definition](#rows-definition)
- [Object template](#object-template)
  - [Components/slots definition](#componentsslots-definition)
- [Colors definition](#colors-definition)

# Building template

**This template is still in progress.**
```json
{
  "slug"     : "",
  "category" : "building",
  "sizeWDHm" : [0,0,0],
  "vertices" : [],
  "center"   : [0,0]
}
```
Type       | Attribute       | Comment
-----------|-----------------|----------
string     | slug            | Name used for using the template, without spaces or points.
string     | category        | Defines which template it is. "building" is mandatory here.
float[3]   | sizeWDHm        | Size of the room in meter : [width,depth,height].
float[2][] | vertices        | Coordinates of the corners of the building in m.
float[2]   | center          | The coordinates [x,y] in m of the center of the building. Used to display the building's name.

# Room template
```json
{
  "slug"            : "",
  "category"        : "room",
  "axisOrientation" : "",
  "sizeWDHm"        : [0,0,0],
  "floorUnit"       : "",
  "technicalArea"   : [0,0,0,0],
  "reservedArea"    : [0,0,0,0],
  "separators"      : {},
  "pillars"         : {},
  "colors"          : [],
  "tiles"           : [],
  "rows"            : [],
  "vertices"        : [],
  "tileAngle"       : 0,
  "center"          : [0,0]
}
```  
Type       | Attribute       | Comment
-----------|-----------------|----------
string                     | slug            | Name used for using the template, without spaces or points.
string                     | category        | Defines which template it is. "room" is mandatory here.
string                     | axisOrientation | List of possible orientations [here](https://github.com/ditrit/OGREE-3D/wiki/How-it-works#Orientations).
float[3]                   | sizeWDHm        | Size of the room in meter : [width,depth,height].
string                     | floorUnit       | Unit used for placing objects in the room. List of possible units [here](https://github.com/ditrit/OGREE-3D/wiki/How-it-works#Floor-units).
int[4]                     | technicalArea   | Tiles used for the technical area : [front,back,right,left].
int[4]                     | reservedArea    | Tiles used for the reserved area : [front,back,right,left].
Dictionary<string, struct> | separators      | Wire mesh or plain walls ([See below](#Separators-definition)).
Dictionary<string, struct> | pillars         | Pillars in the room ([See below](#Pillars-definition)).
struct[]                   | colors          | Custom colors used only in the current json template. ([See below](#Colors-definition))
struct[]                   | tiles           | Tiles definition. ([See below](#Tiles-definition))
struct[]                   | rows            | Not implemented yet. ([See below](#Aisles-definition))
float[2][]                 | vertices        | Coordinates of the corners of the room in cm.
float                      | tileAngle       | Rotation angle of all tiles.
float[2]                   | center          | The coordinates [x,y] in cm of the center of the building. Used to display the building's name.



## Separators definition
```json
{
  "sepname": {
    "startPosXYm" : [0,0],
    "endPosXYm"   : [0,0],
    "type"        : "wireframe|plain"
  }
}
```
Type     | Attribute   | Comment
---------|-------------|----------
float[2] | startPosXYm | Starting point of the wall, [x, y] in meters.
float[2] | endPosXYm   | Ending point of the wall, [x, y] in meters.
string   | type        | "wireframe" or "plain": how is it displayed in the 3D Client

## Pillars definition
```json
{
  "pillarname": {
    "centerXY" : [0,0],
    "sizeXY"   : [0,0],
    "rotation" : 0
  }
}
```
Type     | Attribute   | Comment
---------|-------------|----------
float[2] | centerXY    | Position of the center of the pillar, [x, y] in meters.
float[2] | sizeXY      | Width of the pillar, [x, y] in meters.
float    | rotation    | Angle of the pillar in the room.

## Tiles definition
```json
{
  "location" : "",
  "name"     : "",
  "label"    : "",
  "texture"  : "",
  "color"    : ""
}
```
Type   | Attribute | Comment
-------|-----------|----------
string | location  | Tile location in "X/Y" format. Can be negative to custom technical tiles.
string | name      | Name of the tile (not used, but can be helpful for human reading).
string | label     | Label displayed on the tile
string | texture   | Texture to apply on the tile. Textures are loaded in the [config file](https://github.com/ditrit/OGREE-3D/wiki/How-to-use#Config-file).
string | color     | Hexadecimal code or `@[definedColor]`.

## Rows definition
```json
{
  "name"        : "",
  "locationY"   : "",
  "orientation" : ""
}
```
Type   | Attribute   | Comment
-------|-------------|----------
string | name        | -
string | locationY   | -
string | orientation | -

# Object template
```json
{
  "slug"        : "",
  "description" : "",
  "category"    : "rack|device|generic",
  "sizeWDHmm"   : [0,0,0],
  "fbxModel"    : "",
  "attributes"  : {
    "type" : ""
  },
  "colors"      : [],
  "components"  : [],
  "slots"       : [],
  "clearance"   : [0, 0, 0, 0, 0, 0]
}
```
Type                       | Attribute   | Comment
---------------------------|-------------|----------
string                     | slug        | Name used for using the template, lowercase without spaces.
string                     | description | *description1* of the OgreeObject.
string                     | category    | "rack" or "device"
float[3]                   | sizeWDHmm   | [width,depth,height] in mm.
string                     | fbxModel    | Can be left blank. If completed, the client will download the 3D model with the TriLib plugin. The url has to be the direct link to a ".fbx" file.
Dictrionary<string,string> | attributes  | Custom attributes used to define the object, in the `"key":"value"` format.  
string                     | type        | **Not used for racks** Type of the device/generic object (type attribute).
struct[]                   | colors      | Custom colors used only in the current json template. ([See below](#Colors-definition))
struct[]                   | components  | Sub-parts of the object, with OgreeObject properties. ([See below](#Componentsslots-definition))
struct[]                   | slots       | Slots to place devices. ([See below](#Componentsslots-definition))
float[6]                   | clearance   | [front, rear, left, right, top, bottom] in mm.

## Components/slots definition
```json
{
  "location"   : "",
  "type"       : "",
  "elemOrient" : "",
  "elemPos"    : [0,0,0],
  "elemSize"   : [0,0,0],
  "labelPos"   : "",
  "color"      : ""
  "attributes" : {
    "factor" : ""
  }
}
```
Type                      | Attribute  | Comment
--------------------------|------------|----------
string                    | location   | name of the component/slot.
string                    | type       | *deviceType* for device's slot.
string                    | elemOrient | **Mandatory for rack slot.** "horizontal" or "vertical".
float[3]                  | elemPos    | [x,y,z]
float[3]                  | elemSize   | [width,depth,height]
string                    | labelPos   | "front", "rear", "frontrear", "top", "right", "left" or "none"
string                    | color      | Hexadecimal code or `@[definedColor]`.
Dictionary<string,string> | attributes | Custom attributes for the component / Attributes for slot.
string                    | factor     | **Blank for rack template.** Form factor of the device's slot.

# Colors definition
```json
{
  "name"  : "",
  "value" : ""
}
```
Type   | Attribute | Comment
-------|-----------|----------
string | name      | Name of the color to define
string | value     | Hexadecimal code of the color to define
