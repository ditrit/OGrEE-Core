#!/usr/bin/env python3
import os
m = ""
with open(os.getcwd()+"/y.go", "r") as f:
    contents = f.readlines()
    for i in contents:
        m+=i
    
x = m.find("yylex.Error(msg)")

goCode='\t\t\tif cmd.State.ScriptCalled == true {\n\t\t\t\tprintln("File:",cmd.GetScriptPath() )\n\t\t\t\tprintln("Line Number:", cmd.GetLineNumber())\n\t\t\t}'

m=m.replace("yylex.Error(msg)", 
    "println(\"OGREE: Unrecognised command!\")\n"+goCode+"\n\t\t\tcmd.WarningLogger.Println(\"Unknown Command\")\t\t\t/*yylex.Error(msg)*/")

    #k = f.readline()
    #while k:
    #    if k.find("yylex.Error(msg)") != -1:
    #        print('TARGET ACQUIRED')
    #    else:
    #        print('TARGET NOT ACQ')


#contents.insert(445, 'println("OGREE: Unrecognised command!")\n')

with open(os.getcwd()+"/y.go", "w") as f:
    #contents = "".join(contents)
    f.write(m)