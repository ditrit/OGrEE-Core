USAGE: link: [PATH/TO/STRAY-DEVICE] @ [PATH/TO/RACK/OR/DEVICE] @ Slot (Optional)    
Moves a stray-device to a parent in the OGREE hierarchy.    
		
NOTE   
Complete paths must be provided. At this time it is preferable to enclose the paths in quotes. If the parent is a device then the parent slot must be specified.        

EXAMPLE   

    link: /Physical/Stray/myDevice @ /Physical/demo/buildg/room/rackA    
    link: "/Physical/Stray/myDevice" @ "/Physical/demo/buildg/room/rackA/deviceA" @ 1 