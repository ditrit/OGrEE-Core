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
r2 = os.system("./resources/test/populateDB.py")
r3 = os.system("./resources/test/updateDB.py")
r4 = os.system("./resources/test/deleteTenantHierarchy.py")
r5 = os.system("./resources/test/verifyDBEmpty.py")
checkTestResults([r1,r2,r3,r4, r5])

print("******************************************************")
print("SCENARIO-2 CASE-1: 	 SUCCESS ")
print("******************************************************")


#Empty, populate, update, patch then delete all
#which should test the cascade
#relationship enforced by the API
r1 = os.system("./resources/test/verifyDBEmpty.py")
r2 = os.system("./resources/test/populateDB.py")
r3 = os.system("./resources/test/updateDB.py")
r4 = os.system("./resources/test/patchUpdateDB.py")
r5 = os.system("./resources/test/deleteTenantHierarchy.py")
r6 = os.system("./resources/test/verifyDBEmpty.py")
checkTestResults([r1,r2,r3,r4,r5,r6])

print("******************************************************")
print("SCENARIO-2 CASE-2: 	 SUCCESS ")
print("******************************************************")


#Empty, populate, patch, update then delete all
#which should test the cascade
#relationship enforced by the API
r1 = os.system("./resources/test/verifyDBEmpty.py")
r2 = os.system("./resources/test/populateDB.py")
r3 = os.system("./resources/test/patchUpdateDB.py")
r4 = os.system("./resources/test/updateDB.py")
r5 = os.system("./resources/test/deleteTenantHierarchy.py")
r6 = os.system("./resources/test/verifyDBEmpty.py")
checkTestResults([r1,r2,r3,r4,r5,r6])

print("******************************************************")
print("SCENARIO-2 CASE-3: 	 SUCCESS ")
print("******************************************************")

#Success
sys.exit(1)