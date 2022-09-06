USAGE: lsu [PATH_TO_RACK] (optional)    
Displays devices of a rack in a sorted ascending order according to their 'heightU' position in a rack. If the path given does not refer to a rack then the command returns an error. If no argument is given, then the current path will be used.   


The command will place objects that are not numerical or otherwise invalid heightU values at the end of the list.


EXAMPLE   

    lsu   
    lsu DEMO_RACK/