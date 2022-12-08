USAGE:  draw [PATH] (optional) DEPTH (optional) -FLAG (optional) ARGUMENT (optional)   
Retrieves hierarchy of object with depth as limit and sends results to Unity.
If no options are specified then draw executes with    
current path and depth of 0          

If the number of objects to draw exceeds the draw threshold (user defined, 50 by default) then a warning prompt will request the user if this is ok to send. 

You can optionally 'force' a response (useful for scripting) such that the CLI will not ask and this can be done via '-f y' or '-f n' respectively. 'y' means yes and 'n' means no

EXAMPLE   

    draw   
    draw 2  
    draw DEMO_RACK/DeviceA 2
    draw /Physical/SiteA
    draw $x
    draw $x 1 -f y
    draw $x 5 -f n
