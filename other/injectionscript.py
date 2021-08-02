#!/usr/bin/env python
import os
m = ""
with open(os.getcwd()+"/y.go", "r") as f:
    contents = f.readlines()
    for i in contents:
        m+=i
    
x = m.find("yylex.Error(msg)")

m=m.replace("yylex.Error(msg)", 
    "println(\"OGREE: Unrecognised command!\")\ncmd.WarningLogger.Println(\"Unknown Command\")\t\t\t/*yylex.Error(msg)*/")

    #k = f.readline()
    #while k:
    #    if k.find("yylex.Error(msg)") != -1:
    #        print('TARGET ACQUIRED')
    #    else:
    #        print('Faggot')


#contents.insert(445, 'println("OGREE: Unrecognised command!")\n')

with open(os.getcwd()+"/y.go", "w") as f:
    #contents = "".join(contents)
    f.write(m)