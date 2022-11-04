Usage: ls[OBJ] [PATH] [FLAG](optional)   
Displays specified object type from given path.

Where OBJ can be:   
ten   
site   
bldg   
room   
rack   
dev   
ac   
cabinet   
corridor   
panel   
sensor      

The possible flags that can be supplied at this time is the '-r' option for recursive ls. By default the ls{obj} commands will not be recursive 

EXAMPLE   

    lsten   
    lsten ../../Physical
    lscabinet
    lspanel
    lsac ../../DEMO -r 