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
    '-f' flag can 2 types of arguments. First type is a single quote enclosed string argument with attributes separated by ':' and the second type is of the form ("Format String",Attribute0,...,AttributeN) like a printf    

    The first type is a simpler version of the second type. The first type will take all the arguments separated by ':' and display them automatically (while inferring the data type for you)   

    The second type is more manual and allows you to precisely specify the formatting just like how a 'printf' function works.    

EXAMPLE   

    lsten   
    lsten ../../Physical
    lscabinet
    lspanel
    lsac ../../DEMO -r 
    lsdev . -s heightUnit -r
    lsrack -r -s heightUnit -f ("Value1:%s\tValue2:%s",attr1,attr2)
    lsdev . -f ("HeightUnit:%d\t\tColor:%x",heightUnit,color) 
    lsten . -f "color:mainContact:mainPhone"
    lssite ../../DEMO -f "zipcode:orientation"