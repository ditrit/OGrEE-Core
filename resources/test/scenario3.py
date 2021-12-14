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
    print("SCENARIO-3 CASE-"+str(num)+": 	  "+out)
    print("******************************************************")


#Empty, populate, update then delete all
#which should test the cascade
#relationship enforced by the API
r1,txt1 = subprocess.getstatusoutput("./resources/test/verifyDBEmpty.py")
checkRes(r1, txt1, "verifyDBEmpty")

r2,txt2 = subprocess.getstatusoutput("./resources/test/populateMongo.py")
checkRes(r2, txt2, "populateMongo")

r3,txt3 = subprocess.getstatusoutput("./resources/test/updateMongo.py")
checkRes(r3, txt3, "updateMongo")

r4,txt4 = subprocess.getstatusoutput("./resources/test/deleteTenantHierarchy.py")
checkRes(r4, txt4, "deleteTenantHierarchy")

r5,txt5 = subprocess.getstatusoutput("./resources/test/verifyDBEmpty.py")
checkRes(r5, txt5, "verifyDBEmpty")

printScenarioResult(1)



#Scenario 2
r1,txt1 = subprocess.getstatusoutput("./resources/test/verifyDBEmpty.py")
checkRes(r1, txt1, "verifyDBEmpty")

r2,txt2 = subprocess.getstatusoutput("./resources/test/populateMongo.py")
checkRes(r2, txt2, "populateMongo")

r3,txt3 = subprocess.getstatusoutput("./resources/test/get.py")
checkRes(r3, txt3, "get")

r4,txt4 = subprocess.getstatusoutput("./resources/test/deleteTenantHierarchy.py")
checkRes(r4, txt4, "deleteTenantHierarchy")

r5,txt5 = subprocess.getstatusoutput("./resources/test/verifyDBEmpty.py")
checkRes(r5, txt5, "verifyDBEmpty")

printScenarioResult(2)

#Success
sys.exit()