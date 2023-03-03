USAGE: getslot [PATH_TO_RACK](optional) , [SLOT]   
Retrieves object located in a rack at slot [SLOT] location. If the path given does not refer to a rack or if the slot is empty then the command returns an error. If an invalid value is provided for the [SLOT] parameter then nothing will be returned.        

NOTE
If path is not specified then the current path will be used.  


EXAMPLE   

    getslot ,L01
    getslot RACKA/ , X01
    getslot , X05
