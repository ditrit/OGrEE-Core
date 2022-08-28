USAGE: getu [PATH_TO_RACK](optional) [NUM](optional)   
Retrieves object located in a rack at 'U' [NUM] location. If the path given does not refer to a rack then the command returns an error. If an invalid value is provided for the [NUM] parameter then nothing will be returned.        

NOTE
If path is not specified then the current path will be used. If the [NUM] parameter is not specified then a value of 0 will be used.  


EXAMPLE   

    getu 
    getu RACKA/ 40
    getu 19
