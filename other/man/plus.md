USAGE: + [OCLIENTITY]:[PATH]@[OCLIOPTIONS]   
Shorthand syntax for creating objects   

Each entity type has a specific OCLIOPTIONS   
When properly executed object will be created.   
The required attributes for each object is found:    
https://github.com/ditrit/OGREE-3D/wiki/How-it-works#ogreeobject-class   

USAGES   

+tn:PATH/TENANT_NAME@TENANT_NAME@COLOR   
+tenant:PATH/TENANT_NAME@TENANT_NAME@COLOR   
User must specify the path, TENANT_NAME and COLOR   

Where COLOR refers to a 6 digit Hex value. If the    
value is less than 6 digits then zeros will be prepended.   
A string or numerical Hex value maybe given here



+si:PATH/SITE_NAME   
+site:PATH/SITE_NAME   
User must specify the path and SITE_NAME   


+bd:PATH/BLDG_NAME@POSITION@ROTATION@SIZE   
+building:PATH/BLDG_NAME@POSITION@ROTATION@SIZE 
User must specify the path, BLDG_NAME, SIZE, ROTATION   

Where the POSITION (posXY attribute) must be a 2 element array/vector of coordinates (ie [1,2])   

Where the SIZE (size attribute) must be a 3 element array/vector (ie [1,2,3])

Where the ROTATION (rotation attribute) must be a numerical value (ie 45)

+bd:PATH/BLDG_NAME@POSITION@ROTATION@TEMPLATE
+building:PATH/BLDG_NAME@POSITION@ROTATION@TEMPLATE
User must specify the path, BLDG_NAME, POSITION, ROTATION,TEMPLATE

Where the POSITION (posXY attribute) must be a 2 element array/vector of coordinates (ie [1,2])   

Where the ROTATION (rotation attribute) must be a numerical value (ie 45)

Where TEMPLATE refers to the bldg template name (which must be already existing)


+ro:PATH/ROOM_NAME@POSITION@ROTATION@SIZE@AXISORIENTATION@FLOORUNIT   
+room:PATH/ROOM_NAME@POSITION@ROTATION@SIZE@AXISORIENTATION@FLOORUNIT    
User must specify the path, ROOM_NAME, ROTATION, POSITION, SIZE, ORIENTATION and FLOORUNIT


Where POSITION (posXY attribute) must be a 2 element array/vector of coordinates (ie [1,2]) 

Where SIZE is a 3 numerical element array/vector (ie [1,2,3])

Where ROTATION must be a numerical value (ie 36) 

Where AXISORIENTATION refers to the cardinal directions and can only be of the following format: {[+/-][N/E/W/S][+/-][N/E/W/S]} (ie +N-E)

Where FLOORUNIT refers to the measurement unit for the floor which can only be: {f,m,t}  


+ro:PATH/ROOM_NAME@POSITION@ROTATION@SIZE@AXISORIENTATION   
+room:PATH/ROOM_NAME@POSITION@ROTATION@SIZE@AXISORIENTATION    
User must specify the path, ROOM_NAME, ROTATION, POSITION, SIZE, ORIENTATION 


Where POSITION (posXY attribute) must be a 2 element array/vector of coordinates (ie [1,2]) 

Where SIZE is a 3 numerical element array/vector (ie [1,2,3])

Where ROTATION must be a numerical value (ie 36) 

Where AXISORIENTATION refers to the cardinal directions and can only be of the following format: {[+/-][N/E/W/S][+/-][N/E/W/S]} (ie +N-E)


+ro:PATH/ROOM_NAME@POSITION@ROTATION@TEMPLATE   
+room:PATH/ROOM_NAME@POSITION@ROTATION@TEMPLATE    
User must specify the path, ROOM_NAME,ROTATION, POSITION and TEMPLATE


Where POSITION (posXY attribute) must be a 2 element array/vector of coordinates (ie [1,2]) 

Where ROTATION must be a numerical value (ie 36) 

Where TEMPLATE refers to the room template name (which must be already existing)


+rk:PATH/RACK_NAME@POSITION@SIZE@ORIENTATION   
+rack:PATH/RACK_NAME@POSITION@SIZE@ORIENTATION   
User must specify the path, RACK_NAME, POSITION and SIZE and ORIENTATION 

Where POSITION (posXY attribute) must be a 2 or a 3 element array/vector of coordinates (ie [1,2]) 

Where SIZE is a 3 numerical element array/vector (ie [1,2,3])

Where ORIENTATION is a string and can only be of the following values: {front,rear,left,right}


+rk:PATH/RACK_NAME@POSITION@TEMPLATE@ORIENTATION   
+rack:PATH/RACK_NAME@POSITION@TEMPLATE@ORIENTATION   
User must specify the path, RACK_NAME, POSITION and SIZE and ORIENTATION and TEMPLATE   

Where POSITION (posXY attribute) must be a 2 or a 3 element array/vector of coordinates (ie [1,2]) 

Where TEMPLATE refers to the rack template name (which must be already existing)

Where ORIENTATION is a string and can only be of the following values: {front,rear,left,right}


+dv:PATH/DEVICE_NAME@SLOT_OR_POSU@SIZEUNIT   
+device:PATH/DEVICE_NAME@SLOT_OR_POSU@SIZEUNIT   
User must specify the path, DEVICE_NAME, SLOT_OR_POSU and SIZEUNIT  

Where SLOT_OR_POSU is the name of the slot in which you want to place the device or the position (posU, position in U) where you want to place the device 

Where SIZEUNIT is the height (in U) in a rack and is a numerical value


+dv:PATH/DEVICE_NAME@SLOT_OR_POSU@TEMPLATE   
+device:PATH/DEVICE_NAME@SLOT_OR_POSU@TEMPLATE   
User must specify the path, DEVICE_NAME, SLOT_OR_POSU and TEMPLATE

Where SLOT_OR_POSU is the name of the slot in which you want to place the device or the position (posU, position in U) where you want to place the device 

Where TEMPLATE refers to the device template name (which must be already existing)


+dv:PATH/DEVICE_NAME@SLOT_OR_POSU@TEMPLATE@SIDE   
+device:PATH/DEVICE_NAME@SLOT_OR_POSU@TEMPLATE@SIDE   
User must specify the path, DEVICE_NAME, SLOT_OR_POSU and TEMPLATE and SIDE  

Where SLOT_OR_POSU is the name of the slot in which you want to place the device or the position (posU, position in U) where you want to place the device 

Where TEMPLATE refers to the device template name (which must be already existing)

Where SIDE refers which side the device is visible if not "fullsize" and can only be of the following values: {rear,frontflipped,front,rearflipped}


+co:PATH/ROOM_NAME@CORRIDOR_NAME@{LEFT_RACK,RIGHT_RACK}@TEMPERATURE   
+corridor:PATH/ROOM_NAME@CORRIDOR_NAME@{LEFT_RACK,RIGHT_RACK}@TEMPERATURE   
User must specify the path, ROOM_NAME, CORRIDOR_NAME, LEFT_RACK, RIGHT_RACK and TEMPERATURE   

Where LEFT_RACK refers to the left most rack of the corridor (which must be existing already) 

Where RIGHT_RACK refers to the right most rack of the corridor (which must be existing already) 

Where TEMPERATURE can only be one of 2 values: {warm|cold}


+gr:PATH/ROOM_NAME@{RACK0,...,RACKN}   
+group:PATH/ROOM_NAME@{RACK0,...,RACKN}   
User must specify the path, ROOM_NAME, and all RACKs   

Where RACK0 is the first rack (which must be already existing)

Where ... is any number of racks (which must be already existing) that you would like to add with each separated by '@'

Where RACKN is the last rack (which must be already existing) that you would like to add 


+ orphan sensor: NAME/OF/SENSOR @ TEMPLATE
User must specify the path, and TEMPLATE  

Where TEMPLATE refers to the sensor template name (which must be already existing)


+ orphan device: NAME/OF/DEVICE @ TEMPLATE
User must specify the path, and TEMPLATE 

Where TEMPLATE refers to the device template name (which must be already existing)


EXAMPLES   

+tn:CED@ced666
+tenant:MARCUS@42ff42

+si:CED.BETA
+site:CED/BETA


+bd:CED/BETA/A@[5,5]@49.1@[300,300,300]
+bd:CED/BETA/A@[5,5]@-27.89@BldgTemplateA
+building:CED/BETA/A@[5,5]@49.1@[300,300,300]
+building:CED/BETA/A@[5,5]@-27.89@BldgTemplateA


+ro:CED/BETA/A/R1@[0,0]@-36.202@[22.8,19.8,2]@+N+W@t
+ro:CED/BETA/A/R1@[0,0]@-36.202@[22.8,19.8,2]@+N+W
+room:CED/BETA/A/R1@[0,0]@-36.202@RoomTemplateA


+rk:CED.BETA.A.R1.A01@[9,1,99]@[60,120,42]@front
+rk:CED.BETA.A.R1.A01@[9,1]@[60,120,42]@front
+rk:CED/BETA/A/R1/A01@[9,1]@RackTemplateA@right


+dv:CED.BETA.A.R1.A01.chT@12@10
+dv:CED.BETA.A.R1.A01.chT@SlotA@10

+dv:CED.BETA.A.R1.A01.chT@12@ibm-ns1200
+dv:CED.BETA.A.R1.A01.chT@Slot5@ibm-ns1200

+dv:CED.BETA.A.R1.A01.chT@15@ibm-ns1200
+dv:CED.BETA.A.R1.A01.chT@Slot5@ibm-ns1200@frontflipped

+co:CED.BETA.A.R1.CorridorD@{A01,A09}@warm 

+gr:CED.BETA.A.R1.GroupG@{A01,A02,A03,A04}

+ orphan sensor: StraySensorDEMO @ SensorTemplateA

+ orphan device: StrayDevDEMO @ DEVICE1T