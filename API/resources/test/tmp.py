#!/usr/bin/env python3
import requests,json, os, sys

url = "http://localhost:3001/api"
headers = {
  'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySWQiOjY2NDA0NjEyNzM0MjQxOTk2OX0.cB1VkYQLlXCatzMiEWGFfJKKx9h8Vsr2vdlylNMe7hs',
  'Content-Type': 'application/json'
}

rackIDS=["6218f9a509f588b5bd74bed5", "6218f99209f588b5bd74bed4"]

payload={
    "name": None,
    "parentId": "{{RackID}}",
    "category": "device",
    "description": [
        "Rack A01",
        "The original one",
        "-3/-5\\nA0-Z9"
    ],
    "domain": "Device DOMAIN",
    "attributes": {
        "posXY": "{\"x\":10.0,\"y\":0.0}",
        "posXYUnit": "tile",
        "posZ": "{\"x\":10.0,\"y\":0.0}",
        "posZUnit": "tile",
        "size": "{\"x\":60.0,\"y\":120.0}",
        "sizeUnit": "cm",
        "height": "42",
        "heightUnit": "U",
        "heightU": "U",
        "template": "Some template",
        "orientation": "front",
        "vendor": "someVendor",
        "type": "someType",
        "model": "someModel",
        "serial": "someSerial"
    }
}

x = 0
for i in range(len(rackIDS)):
    payload["parentId"] = rackIDS[i] #Setup ParentID for 1000 Devices
    while x < 25000:
        payload["name"] = str(x)
        response = requests.request("POST", url+"/devices", headers=headers, data=json.dumps(payload))
        if response.status_code != 201:
            print(response.json())
            sys.exit
        x+=1
    
    x=0
    print(i)

print('DONE!')