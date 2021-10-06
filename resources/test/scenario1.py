#!/usr/bin/env python3
import os, sys

#Empty, populate then delete all
#which should test the cascade
#relationship enforced by the API
r1 = os.system("./resources/test/verifyDBEmpty.py")
r2 = os.system("./resources/test/populateDB.py")
r3 = os.system("./resources/test/deleteTenantHierarchy.py")
r4 = os.system("./resources/test/verifyDBEmpty.py")

#Check Return Values
if r1 == -1 or r2 == -1 or r3 == -1 or r4 == -1:
    sys.exit(-1)

#Success
sys.exit(1)