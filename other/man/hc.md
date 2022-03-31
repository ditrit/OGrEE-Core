USAGE: hc [PATH] (optional) [NUM] (optional)   
Displays object's hierarchy in JSON format.   

The PATH parameter specifies path to object. If PATH is not provided then the current path will be used.   


The NUM parameter limits the depth of the hierarchy call. If the depth is not provided then 1 will be used by default which will only retrieve the current object and thus act as a get call.   

EXAMPLE   

    hc DEMO/ALPHA/B 2