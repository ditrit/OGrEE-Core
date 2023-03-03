USAGE: unlink: [PATH/TO/DEVICE] @ [PATH/TO/STRAY-DEVICE-PARENT] (Optional)    
Moves a device from the OGREE hierarchy to be stray-device.    
		
NOTE   
Complete paths must be provided. At this time it is preferable to enclose the paths in quotes. An empty path for the destination is optional thus can be left empty         

EXAMPLE   

    unlink: /Physical/demo/buildg/room/rackA/deviceA @ /Physical/Stray/myDevice    
    unlink: "/Physical/demo/buildg/room/rackA/deviceA" @ 