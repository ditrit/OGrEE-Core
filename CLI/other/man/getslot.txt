USAGE: getslot [PATH_TO_RACK] [SLOT]   
Retrieves object located in a rack at slot [SLOT] location. If the path given does not refer to a rack then the command returns an error. If the value provided for the [SLOT] parameter does not exist in the rack then nothing will be returned.        

NOTE
If path is not specified then the current path will be used.  


EXAMPLE   

    getslot L01
    getslot RACKA/ X01
    getslot X05
