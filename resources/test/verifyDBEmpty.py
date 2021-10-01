#!/usr/bin/env python
import requests
expected = 404

url = "http://localhost:3001/api/"

def checkResponse(code, entity):
    if code == expected:
        print(entity+" Empty: Successful")
    else:
        print(entity+" Empty: Fail")

entRange=["tenants","sites","buildings","rooms","racks","devices","subdevices","subdevice1s"]
payload={}
headers = {
  'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySWQiOjY2NDA0NjEyNzM0MjQxOTk2OX0.cB1VkYQLlXCatzMiEWGFfJKKx9h8Vsr2vdlylNMe7hs'
}

for i in entRange:
    response = requests.request("GET", url+i, headers=headers, data=payload)
    checkResponse(response.status_code, i)

