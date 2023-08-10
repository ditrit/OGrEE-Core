USAGE: get [PATH] [DEVICES](optional)
Retrieves object information from API and displays it's information in JSON format.     

NOTE
If path is not specified then the current path will be used.

EXAMPLE   

    get 
    get /Physical/SiteA
    get ../rack01/device-ibm3

OPTION DEVICES:
Retrives devices informations links to an object
It will bind object's name with device's group_name by default
DEVICES take multiple arguments to filter devices

EXAMPLE

get /Physical/SiteA/BuildingA/RoomA/Rack01/device-ibm server -> Get all devices where group_name match device-ibm's name
get /Physical/SiteA/BuildingA/RoomA/Rack01/device-ibm%serialNumber server%serial_number -> Binding attributes serialNumber to serial_number
get /Physical/SiteA/BuildingA/RoomA/Rack01/device-ibm server._name=test -> Filter results by device's name

