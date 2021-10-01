#!/usr/bin/env python
import requests,json, os, sys
from dotenv import load_dotenv

PIDS={"tenantID":None, "siteID":None, "buildingID":None,
        "roomID":None, "rackID":None, "deviceID":None}


with open('.env', 'w') as file:
     file.write(json.dumps(PIDS)) # use `json.loads` to do the reverse
