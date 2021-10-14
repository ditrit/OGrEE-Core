#!/usr/bin/env python3
import requests, sys
expected = 404

url = "http://localhost:27020/api/"
res = True

def checkResponse(code, entity):
    global res
    if code == expected:
        print(entity+" Empty: Successful")
    else:
        res = False
        print(entity+" Empty: Fail")

entRange=["tenants","sites","buildings","rooms","racks","devices",
            "subdevices","subdevice1s","room-templates",
            "rack-templates","device-templates"]
payload={}
headers = {
  'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySWQiOjY2NDA0NjEyNzM0MjQxOTk2OX0.cB1VkYQLlXCatzMiEWGFfJKKx9h8Vsr2vdlylNMe7hs'
}

for i in entRange:
    response = requests.request("GET", url+i, headers=headers, data=payload)
    checkResponse(response.status_code, i)


#RETURN VALUE
if res != True:
    sys.exit(-1)

#else success
sys.exit(1)