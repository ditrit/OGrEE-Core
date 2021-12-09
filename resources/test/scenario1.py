#!/usr/bin/env python3
import os, sys, subprocess

#Exit Flag
res = True

#Check Return Values
def checkRes(val, out, name):
    global res
    if val == -1:
        print("Failure!")
        print("Test Name: ", name)
        print(out)
        res = False


#Empty, populate then delete all
#which should test the cascade
#relationship enforced by the API
r1, txt1 = subprocess.getstatusoutput("./resources/test/verifyDBEmpty.py")
r2, txt2 = subprocess.getstatusoutput("./resources/test/populateDB.py")
r3, txt3 = subprocess.getstatusoutput("./resources/test/deleteTenantHierarchy.py")
r4, txt4 = subprocess.getstatusoutput("./resources/test/verifyDBEmpty.py")


q = [(r1, txt1,"verifyDBEmpty"),
    (r2, txt2, "populateDB"),
    (r3, txt3, "deleteTenantHierarchy"),
    (r4, txt4, "verifyDBEmpty")
    ]


#Check Return Values
for i in q:
    checkRes(i[0], i[1], i[2])


#Print Result
out = ""
if res == True:
    out = "SUCCESS"
else:
    out = "FAILURE"

print("******************************************************")
print("SCENARIO-1 CASE-1: 	  "+out)
print("******************************************************")

#Success
sys.exit()