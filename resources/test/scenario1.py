#!/usr/bin/env python3
import os

#Empty, populate then delete all
#which should test the cascade
#relationship enforced by the API
os.system("./resources/test/verifyDBEmpty.py")
os.system("./resources/test/populateDB.py")
os.system("./resources/test/deleteTenantHierarchy.py")
os.system("./resources/test/verifyDBEmpty.py")