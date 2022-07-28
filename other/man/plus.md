USAGE: + [OCLIENTITY]:[PATH]@[OCLIOPTIONS]   
Shorthand syntax for creating objects   

Each entity type has a specific OCLIOPTIONS   
When properly executed, the user will be prompted   
to enter the attributes of the object in question.   
Once the sufficient necessary options have been   
entered. The object will be created.   
The required attributes for each object is found:    
https://github.com/ditrit/OGREE-3D/wiki/How-it-works#ogreeobject-class   

USAGES     

+si:PATH/SITE_NAME@ORIENTATION   
+site:PATH/SITE_NAME@ORIENTATION   
User must specify the path, SITE_NAME and ORIENTATION   


+bd:PATH/BLDG_NAME@POSITION@SIZE   
+building:PATH/BLDG_NAME@POSITION@SIZE   
User must specify the path, BLDG_NAME, POSITION and SIZE   


+ro:PATH/ROOM_NAME@POSITION@SIZE@ORIENTATION   
+room:PATH/ROOM_NAME@POSITION@SIZE@ORIENTATION    
User must specify the path, ROOM_NAME, POSITION and SIZE and ORIENTATION   

+ro:PATH/ROOM_NAME@POSITION@SIZE@ORIENTATION@FLOORUNIT   
+room:PATH/ROOM_NAME@POSITION@SIZE@ORIENTATION@FLOORUNIT    
User must specify the path, ROOM_NAME, POSITION and SIZE and ORIENTATION and FLOORUNIT

+ro:PATH/ROOM_NAME@POSITION@TEMPLATE   
+room:PATH/ROOM_NAME@POSITION@TEMPLATE    
User must specify the path, ROOM_NAME, POSITION and TEMPLATE


+rk:PATH/RACK_NAME@POSITION@SIZE@ORIENTATION   
+rack:PATH/RACK_NAME@POSITION@SIZE@ORIENTATION   
User must specify the path, RACK_NAME, POSITION and SIZE and ORIENTATION   

+rk:PATH/RACK_NAME@POSITION@SIZE@ORIENTATION@TEMPLATE   
+rack:PATH/RACK_NAME@POSITION@SIZE@ORIENTATION@TEMPLATE   
User must specify the path, RACK_NAME, POSITION and SIZE and ORIENTATION and TEMPLATE   


+dv:PATH/DEVICE_NAME@POSU@SIZEUNIT   
+device:PATH/DEVICE_NAME@POSU@SIZEUNIT   
User must specify the path, DEVICE_NAME, POSU and SIZEUNIT   

+dv:PATH/DEVICE_NAME@SLOT@SIZEUNIT   
+device:PATH/DEVICE_NAME@SLOT@SIZEUNIT   
User must specify the path, DEVICE_NAME, SLOT and SIZEUNIT   

+dv:PATH/DEVICE_NAME@SLOT@TEMPLATE   
+device:PATH/DEVICE_NAME@SLOT@TEMPLATE   
User must specify the path, DEVICE_NAME, SLOT and TEMPLATE   

+dv:PATH/DEVICE_NAME@SLOT@TEMPLATE@SIDE   
+device:PATH/DEVICE_NAME@SLOT@TEMPLATE@SIDE   
User must specify the path, DEVICE_NAME, SLOT and TEMPLATE and SIDE  


+co:PATH/ROOM_NAME@CORRIDOR_NAME@LEFT_RACK@RIGHT_RACK@TEMPERATURE   
+corridor:PATH/ROOM_NAME@CORRIDOR_NAME@LEFT_RACK@RIGHT_RACK@TEMPERATURE   
User must specify the path, ROOM_NAME, CORRIDOR_NAME, LEFT_RACK, RIGHT_RACK and TEMPERATURE   


+gr:PATH/ROOM_NAME@RACK0@...@RACKN   
+group:PATH/ROOM_NAME@RACK0@...@RACKN   
User must specify the path, ROOM_NAME, and all RACKs   


+wa:PATH/ROOM_NAME@WALL_NAME@POSITION1@POSITION2   
+wall:PATH/ROOM_NAME@WALL_NAME@POSITION1@POSITION2   
User must specify the path, ROOM_NAME, WALL_NAME POSITION 1 and 2   