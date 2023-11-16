USAGE:  > [PATH/TO/OBJECT]    
Sends a command to the Unity Client to focus on desired object. If the path is empty, it will command the Unity Client to 'unfocus' all objects and return user to the root directory, otherwise it will change the current path according to the given argument.      

NOTE:   
    Tenants, Sites, Buildings and Rooms are restricted objects and cannot be focused.


EXAMPLE:   
    
    > DEMO/ALPHA/B/R1/A01/DeviceA   
    >

For more information please refer to:   
https://github.com/ditrit/OGrEE-3D/wiki/CLI-langage#focus-an-object