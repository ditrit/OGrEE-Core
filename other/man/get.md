USAGE: get [PATH](optional) 
OR get [OBJECT] [ATTRIBUTES]
Retrieves object information from API and displays it's information in JSON format. The alternative syntax performs a search for objects and returns information of the results in JSON format.    

NOTE
If path is not specified then the current path will be used. 
The object specifies which object type. The ATTRIBUTES form the search parameters of the search query. There is no wildcard support at this time.

EXAMPLE   

    get 
    get /Physical/SiteA
    get ../rack01/device-ibm3

    get site color=FFFF name=A
    get device name=ibm
