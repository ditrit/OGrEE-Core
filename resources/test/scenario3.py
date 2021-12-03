#!/usr/bin/env python3
import os, sys


#Check Return Values
def checkTestResults(resArr):
    for i in resArr:
        if i == -1:
            sys.exit(-1)


#Empty, populate, update then delete all
#which should test the cascade
#relationship enforced by the API
r1 = os.system("./resources/test/verifyDBEmpty.py")
r2 = os.system("./resources/test/populateMongo.py")
r3 = os.system("./resources/test/updateMongo.py")
r4 = os.system("./resources/test/deleteTenantHierarchy.py")
r5 = os.system("./resources/test/verifyDBEmpty.py")
checkTestResults([r1,r2,r3,r4, r5])

print("******************************************************")
print("SCENARIO-3 CASE-1: 	 SUCCESS ")
print("******************************************************")


#Success
sys.exit(1)