USAGE: lsslot [PATH_TO_RACK] (optional)    
Displays devices of a rack in a sorted ascending order according to their 'slot' name in a rack. If the path given does not refer to a rack then the command returns an error. If no argument is given, then the current path will be used.   

NOTE
The command will place objects that have invalid slot values at the end of the list.


EXAMPLE   

    lsslot   
    lsslot DEMO_RACK/