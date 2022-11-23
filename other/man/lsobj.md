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

The possible flags that can be supplied at this time are '-r' option for recursive ls, '-s' option for sorting results and '-f' for displaying results with their attributes. By default the ls{obj} commands will not be recursive and does no sorting etc. 

NOTE   
    '-r' flag takes no arguments
    '-s' flag takes a single word argument
    '-f' flag takes a single quote enclosed string argument with attributes separated by ':'

EXAMPLE   

    lsten   
    lsten ../../Physical
    lscabinet
    lspanel
    lsac ../../DEMO -r 