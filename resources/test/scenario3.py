#!/usr/bin/env python3
import os, sys,subprocess

#Exit Flag
res = True

#Check Return Values
def checkRes(val, out, name):
    if val == -1:
        print("Failure!")
        print("Test Name: ", name)
        print(out)
        res = False



#Empty, populate, update then delete all
#which should test the cascade
#relationship enforced by the API
r1,txt1 = subprocess.getstatusoutput("./verifyDBEmpty.py")
checkRes(r1, txt1, "verifyDBEmpty")

r2,txt2 = subprocess.getstatusoutput("./populateMongo.py")
checkRes(r2, txt2, "populateMongo")

r3,txt3 = subprocess.getstatusoutput("./updateMongo.py")
checkRes(r3, txt3, "updateMongo")

r4,txt4 = subprocess.getstatusoutput("./deleteTenantHierarchy.py")
checkRes(r4, txt4, "deleteTenantHierarchy")

r5,txt5 = subprocess.getstatusoutput("./verifyDBEmpty.py")
checkRes(r5, txt5, "verifyDBEmpty")


if res == True:
    print("******************************************************")
    print("SCENARIO-3 CASE-1: 	 SUCCESS ")
    print("******************************************************")
else:
    print("******************************************************")
    print("SCENARIO-3 CASE-1: 	 FAILURE ")
    print("******************************************************")


#Success
sys.exit()