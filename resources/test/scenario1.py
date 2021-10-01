#!/usr/bin/env python
import os

#In Docker Container
os.system("/home/main")

#Empty, populate then delete all
#which should test the cascade
#relationship enforced by the API
os.system("./verifyDBEmpty.py")
os.system("./populateDB.py")
os.system("./deleteTenantHierarchy.py")
os.system("./verifyDBEmpty.py")