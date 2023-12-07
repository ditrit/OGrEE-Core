USAGE: getu [PATH_TO_RACK] [NUM]   
Retrieves object located in a rack at 'U' [NUM] location. If the path given does not refer to a rack then the command returns an error. A positive value must be provided for the [NUM] parameter. If there is nothing present at the given height, nothing will be returned.         


EXAMPLE   

    getu . 0
    getu RACKA/ 40
    getu . 19
