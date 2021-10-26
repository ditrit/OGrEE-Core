#!/usr/bin/env python3
import requests,json, os, sys

#RETURN FLAG
res = True

#CONSTANTS
PIDS={"tenantID":None, "siteID":None, "buildingID":None,
        "roomID":None, "ACID":None, "panelID":None,
        "wallID":None, "rackID":None, "deviceID":None,
        "room-templateID": None, "obj-templateID": None}
        
url = "http://localhost:27020/api"
headers = {
  'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySWQiOjY2NDA0NjEyNzM0MjQxOTk2OX0.cB1VkYQLlXCatzMiEWGFfJKKx9h8Vsr2vdlylNMe7hs',
  'Content-Type': 'application/json'
}


#FUNCTIONS
def verifyCreate(code, entity):
  global res
  if code == 201:
    print('Successfully created '+entity)
  else:
    res = False
    print('Failed to create '+entity)

def verifyGet(code,entity):
  global res
  if code == 200:
    print('Successfully got '+entity)
  else:
    res = False
    print('Failed to get '+entity)

def verifyOutput(j1, j2):
    global res
    if sorted(j1.items()) == sorted(j2.items()):
        print("JSON Match: Success")
    else:
        res = False
        print("JSON Match: Fail")
        print("J1")
        print(sorted(j1.items()))
        print("J2")
        print(sorted(j2.items()))


def writeEnv():
    with open('.localenv', 'w') as file:
        file.write(json.dumps(PIDS)) # use `json.loads` to do the reverse

    




#TENANT CREATE & GET
payload="{\n  \"name\": \"DEMO\",\n  \n  \"parentId\": null,\n  \"category\": \"tenant\",\n  \"description\": [],\n  \"domain\": \"DEMO\",\n  \"attributes\": {\n    \"color\": \"FFFFFF\",\n    \"mainContact\": \"Ced\",\n    \"mainPhone\": \"0612345678\",\n    \"mainEmail\": \"ced@ogree3D.com\"\n  }\n}"
response = requests.request("POST", url+"/tenants", headers=headers, data=payload)
verifyCreate(response.status_code, "Tenant")
ID = response.json()['data']['id']
j1=response.json()['data']
tenantID=ID

response = requests.request("GET", url+"/tenants/"+ID, headers=headers, data={})
verifyGet(response.status_code, "Tenant")
j2=response.json()['data']
verifyOutput(j1, j2)
response.close()
print()
print()



#SITE CREATE & GET
payload={
    "name": "ALPHA",
    "parentId": None,
    "category": "site",
    "description": [
        "This is a demo..."
    ],
    "domain": "DEMO",
    "attributes": {
        "orientation": "NW",
        "usableColor": "5BDCFF",
        "reservedColor": "AAAAAA",
        "technicalColor": "D0FF78",
        "address": "1 rue bidule",
        "zipcode": "42000",
        "city": "Truc",
        "country": "FRANCE",
        "gps": "[1,2,0]"
    }
}
payload['parentId'] = tenantID
response = requests.request("POST", url+"/sites", 
          headers=headers, data=json.dumps(payload))
verifyCreate(response.status_code, "Site")
ID = response.json()['data']['id']
j1=response.json()['data']
siteID=ID

response = requests.request("GET", url+"/sites/"+ID, headers=headers, data={})
verifyGet(response.status_code, "Site")
j2=response.json()['data']
verifyOutput(j1, j2)
response.close()
print()
print()



#BLDG CREATE & GET
payload={
    "name": "B",
    "parentId": None,
    "category": "building",
    "description": [
        "Building B"
    ],
    "domain": "DEMO",
    "attributes": {
        "posXY": "{\"x\":-30.0,\"y\":0.0}",
        "posXYUnit": "m",
        "posZ": "10",
        "posZUnit": "m",
        "size": "{\"x\":25.0,\"y\":29.399999618530275}",
        "sizeUnit": "m",
        "height": "0",
        "heightUnit": "m",
        "nbFloors": "1"
    }
}
payload['parentId'] = siteID
response = requests.request("POST", url+"/buildings",
              headers=headers, data=json.dumps(payload))
verifyCreate(response.status_code, "Building")
ID = response.json()['data']['id']
j1=response.json()['data']
buildingID=ID

response = requests.request("GET", url+"/buildings/"+ID, headers=headers, data={})
verifyGet(response.status_code, "Building")
j2=response.json()['data']
verifyOutput(j1, j2)
response.close()
print()
print()



#ROOM CREATE & GET
payload={
    "name": "R1",
    "parentId": None,
    "category": "room",
    "description": [
        "First room"
    ],
    "domain": "DEMO",
    "attributes": {
        "posXY": "{\"x\":0.0,\"y\":0.0}",
        "posXYUnit": "m",
        "posZ": "0",
        "posZUnit": "m",
        "template": "demo.R1",
        "orientation": "+N+W",
        "size": "{\"x\":22.799999237060548,\"y\":19.799999237060548}",
        "sizeUnit": "m",
        "height": "3",
        "heightUnit": "m"
    }
}
payload['parentId'] = buildingID
response = requests.request("POST", url+"/rooms",
              headers=headers, data=json.dumps(payload))
verifyCreate(response.status_code, "Room")
ID = response.json()['data']['id']
j1=response.json()['data']
roomID=ID

response = requests.request("GET", url+"/rooms/"+ID, headers=headers, data={})
verifyGet(response.status_code, "Room")
j2=response.json()['data']
verifyOutput(j1, j2)
print()
print()


#AC CREATE & GET
payload={
    "name": "TCL 2021",
    "id": None,
    "parentId": None,
    "category": "ac",
    "description": [
        "TCL"
    ],
    "domain": "AC DOMAIN",
    "attributes": {
    }
}
payload['parentId'] = roomID
response = requests.request("POST", url+"/acs",
              headers=headers, data=json.dumps(payload))
verifyCreate(response.status_code, "AC")
ID = response.json()['data']['id']
j1=response.json()['data']
ACID=ID

response = requests.request("GET", url+"/rooms/"+roomID+"/acs/"+ID, headers=headers, data={})
verifyGet(response.status_code, "AC")
j2=response.json()['data']
verifyOutput(j1, j2)
response.close()
print()
print()


#POWERPANEL CREATE & GET
payload={
    "name": "Panel A",
    "id": None,
    "parentId": None,
    "category": "powerpanel",
    "description": [
        "YINGLI"
    ],
    "domain": "Panel DOMAIN",
    "attributes": {
    }
}
payload['parentId'] = roomID
response = requests.request("POST", url+"/panels",
              headers=headers, data=json.dumps(payload))
verifyCreate(response.status_code, "Panel")
ID = response.json()['data']['id']
j1=response.json()['data']
panelID=ID

response = requests.request("GET", url+"/rooms/"+roomID+"/panels/"+ID, headers=headers, data={})
verifyGet(response.status_code, "Panel")
j2=response.json()['data']
verifyOutput(j1, j2)
response.close()
print()
print()


#WALL CREATE & GET
payload={
    "name": "Undercover",
    "id": None,
    "parentId": None,
    "category": "wall",
    "description": [
        "2008"
    ],
    "domain": "Wall DOMAIN",
    "attributes": {
    }
}
payload['parentId'] = roomID
response = requests.request("POST", url+"/walls",
              headers=headers, data=json.dumps(payload))
verifyCreate(response.status_code, "Wall")
ID = response.json()['data']['id']
j1=response.json()['data']
wallID=ID

response = requests.request("GET", url+"/rooms/"+roomID+"/walls/"+ID, headers=headers, data={})
verifyGet(response.status_code, "Wall")
j2=response.json()['data']
verifyOutput(j1, j2)
response.close()
print()
print()




#RACK CREATE & GET
payload={
    "name": "A01",
    "parentId": None,
    "category": "rack",
    "description": [
        "Rack A01",
        "The original one",
        "-3/-5\\nA0-Z9"
    ],
    "domain": "DEMO",
    "attributes": {
        "posXY": "{\"x\":10.0,\"y\":0.0}",
        "posXYUnit": "tile",
        "size": "{\"x\":60.0,\"y\":120.0}",
        "sizeUnit": "cm",
        "posZ": "Some position",
        "posZUnit": "cm",
        "height": "42",
        "heightUnit": "U",
        "template": "Some template",
        "orientation": "front",
        "vendor": "someVendor",
        "type": "someType",
        "model": "someModel",
        "serial": "someSerial"
    }
}
payload['parentId'] = roomID
response = requests.request("POST", url+"/racks",
              headers=headers, data=json.dumps(payload))
verifyCreate(response.status_code, "Rack")
ID = response.json()['data']['id']
j1=response.json()['data']
rackID=ID

response = requests.request("GET", url+"/racks/"+ID, headers=headers, data={})
verifyGet(response.status_code, "Rack")
j2=response.json()['data']
verifyOutput(j1, j2)
response.close()
print()
print()



#DEVICE CREATE & GET
payload={
    "name": "DeviceA",
    "parentId": None,
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
        "template": "Some template",
        "orientation": "front",
        "vendor": "someVendor",
        "type": "someType",
        "model": "someModel",
        "serial": "someSerial"
    }
}
payload['parentId'] = rackID
response = requests.request("POST", url+"/devices",
              headers=headers, data=json.dumps(payload))
verifyCreate(response.status_code, "Device")
ID = response.json()['data']['id']
j1=response.json()['data']
deviceID=ID

response = requests.request("GET", url+"/devices/"+ID, headers=headers, data={})
verifyGet(response.status_code, "Device")
j2=response.json()['data']
verifyOutput(j1, j2)
response.close()
print()
print()


#ROOM-TEMPLATE CREATE & GET
payload={
  "slug"          : "RT1",
  "orientation"   : "+N+E",
  "sizeWDHm"      : ["width","depth","height"],
  "technicalArea" : ["width","depth","height"],
  "reservedArea"  : ["width","depth","height"],
  "separators"    : [
  ],
  "colors"        : [
  ],
  "tiles"         : [
  ],
  "aisles"        : [
  ]
}
response = requests.request("POST", url+"/room-templates",
              headers=headers, data=json.dumps(payload))
verifyCreate(response.status_code, "Room-Template")
ID = response.json()['data']['slug']
j1=response.json()['data']
roomTemplateID=ID

response = requests.request("GET", url+"/room-templates/"+ID, headers=headers, data={})
verifyGet(response.status_code, "Room-Template")
j2=response.json()['data']
verifyOutput(j1, j2)
response.close()
print()
print()



#RACK-TEMPLATE CREATE & GET
payload={
  "slug"        : "RACK1T",
  "description" : "Rack Template 1",
  "category"    : "rack",
  "sizeWDHmm"   : ["width","depth","height"],
  "fbxModel"    : "1",
  "attributes"  : {
    "type" : ""
  },
  "colors"      : [
  ],
  "components"  : [
  ],
  "slots"       : [
  ]
}
response = requests.request("POST", url+"/obj-templates",
              headers=headers, data=json.dumps(payload))
verifyCreate(response.status_code, "Obj-Template")
ID = response.json()['data']['slug']
j1=response.json()['data']
objTemplateID=ID

response = requests.request("GET", url+"/obj-templates/"+ID, headers=headers, data={})
verifyGet(response.status_code, "Obj-Template")
j2=response.json()['data']
verifyOutput(j1, j2)
response.close()
print()
print()


#UPDATE THE ID LIST
PIDS['tenantID'] = tenantID
PIDS['siteID'] = siteID
PIDS['buildingID'] = buildingID
PIDS['roomID'] = roomID
PIDS['ACID'] = ACID
PIDS['panelID'] = panelID
PIDS['wallID'] = wallID
PIDS['rackID'] = rackID
PIDS['deviceID'] = deviceID
PIDS['room-templateID'] = roomTemplateID
PIDS['obj-templateID'] = objTemplateID
writeEnv()


#RETURN VALUE
if res != True:
    sys.exit(-1)

#else success
sys.exit(1)