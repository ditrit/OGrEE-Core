#!/usr/bin/env python3
import os, sys,subprocess

#Exit Flag
res = True

#Check Return Values
def checkRes(val, out, name):
    global res
    if val == 255:
        print("Failure!")
        print("Test Name: ", name)
        print(out)
        res = False

def printScenarioResult(num):
    global res
    if res == False:
        out = "FAILURE"
    else:
        out = "SUCCESS"
    
    print("******************************************************")
    print("SCENARIO-4 CASE-"+str(num)+": 	  "+out)
    print("******************************************************")


#Populate, get, manually delete each, verify DB is empty
r1,txt1 = subprocess.getstatusoutput("./resources/test/verifyDBEmpty.py")
checkRes(r1, txt1, "verifyDBEmpty")

r2,txt2 = subprocess.getstatusoutput("./resources/test/populateMongo.py")
checkRes(r2, txt2, "populateMongo")

r3,txt3 = subprocess.getstatusoutput("./resources/test/manualDelete.py")
checkRes(r3, txt3, "manualDelete")

r5,txt5 = subprocess.getstatusoutput("./resources/test/verifyDBEmpty.py")
checkRes(r5, txt5, "verifyDBEmpty")

printScenarioResult(1)


#Success
sys.exit()