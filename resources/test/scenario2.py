#!/usr/bin/env python3
import os, sys, subprocess


#Exit Flag
res = True
c1 = True
c2 = True
c3 = True

#Check Return Values
def checkRes(val, out, name, caseNum):
    global res
    global c1
    global c2
    global c3
    if val == 255:
        print("Failure!")
        print("Test Name: ", name)
        print(out)
        res = False
        if caseNum == 1:
            c1 = False
        if caseNum == 2:
            c2 = False
        if caseNum == 3:
            c3 = False



#Empty, populate, update then delete all
#which should test the cascade
#relationship enforced by the API
r1, txt1 = subprocess.getstatusoutput("./resources/test/verifyDBEmpty.py")
checkRes(r1, txt1, "verifyDBEmpty",1)

r2, txt2 = subprocess.getstatusoutput("./resources/test/populateDB.py")
checkRes(r2, txt2, "populateDB",1)

r3, txt3 = subprocess.getstatusoutput("./resources/test/updateDB.py")
checkRes(r3, txt3, "updateDB",1)

r4, txt4 = subprocess.getstatusoutput("./resources/test/deleteTenantHierarchy.py")
checkRes(r4, txt4, "deleteTenantHierarchy",1)

r5, txt5 = subprocess.getstatusoutput("./resources/test/verifyDBEmpty.py")
checkRes(r5, txt5, "verifyDBEmpty",1)



if c1 == True:
    out = "SUCCESS"
else:
    out = "FAILURE"
print("******************************************************")
print("SCENARIO-2 CASE-1: 	  "+out)
print("******************************************************")


#Empty, populate, update, patch then delete all
#which should test the cascade
#relationship enforced by the API
r1, txt1 = subprocess.getstatusoutput("./resources/test/verifyDBEmpty.py")
checkRes(r1, txt1, "verifyDBEmpty",2)

r2, txt2 = subprocess.getstatusoutput("./resources/test/populateDB.py")
checkRes(r2, txt2, "populateDB",2)

r3, txt3 = subprocess.getstatusoutput("./resources/test/updateDB.py")
checkRes(r3, txt3, "updateDB",2)

r4, txt4 = subprocess.getstatusoutput("./resources/test/patchUpdateDB.py")
checkRes(r4, txt4, "patchUpdateDB",2)

r5, txt5 = subprocess.getstatusoutput("./resources/test/deleteTenantHierarchy.py")
checkRes(r5, txt5, "deleteTenantHierarchy",2)

r6, txt6 = subprocess.getstatusoutput("./resources/test/verifyDBEmpty.py")
checkRes(r6, txt6, "verifyDBEmpty",2)


if c2 == True:
    out = "SUCCESS"
else:
    out = "FAILURE"
print("******************************************************")
print("SCENARIO-2 CASE-2: 	  "+out)
print("******************************************************")


#Empty, populate, patch, update then delete all
#which should test the cascade
#relationship enforced by the API
r1, txt1 = subprocess.getstatusoutput("./resources/test/verifyDBEmpty.py")
checkRes(r1, txt1, "verifyDBEmpty",3)
r2, txt2 = subprocess.getstatusoutput("./resources/test/populateDB.py")
checkRes(r2, txt2, "populateDB",3)
r3, txt3 = subprocess.getstatusoutput("./resources/test/patchUpdateDB.py")
checkRes(r3, txt3, "patchUpdateDB",3)
r4, txt4 = subprocess.getstatusoutput("./resources/test/updateDB.py")
checkRes(r4, txt4, "updateDB",3)
r5, txt5 = subprocess.getstatusoutput("./resources/test/deleteTenantHierarchy.py")
checkRes(r5, txt5, "deleteTenantHierarchy",3)
r6, txt6 = subprocess.getstatusoutput("./resources/test/verifyDBEmpty.py")
checkRes(r6, txt6, "verifyDBEmpty",3)

if c3 == True:
    out = "SUCCESS"
else:
    out = "FAILURE"
print("******************************************************")
print("SCENARIO-2 CASE-3: 	 SUCCESS ")
print("******************************************************")


#Success
sys.exit()