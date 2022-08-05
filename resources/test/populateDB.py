#!/usr/bin/env python3
import requests,json, os, sys

#RETURN FLAG
res = True

#CONSTANTS
PIDS={"tenantID":None, "siteID":None, "buildingID":None,
        "roomID":None, "acID":None, "panelID":None,
        "cabinetID":None, "groupID":None, "corridorID":None,
        "rackID":None, "deviceID":None,
        "room-sensorID":None,"rack-sensorID":None,
        "device-sensorID":None,
        "room-templateID": None, "obj-templateID": None}
        
url = "http://localhost:3001/api"
headers = {
  'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySWQiOjAsIkN1c3RvbWVyIjoib2dyZWUifQ.3A0ZI7oSXxhZ8CGhJORYRXIm3MKroo_9LM939dgGHIo',
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
        "floorUnit":"f",
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
    "name": "TCL2021",
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
acID=ID

response = requests.request("GET", url+"/acs/"+ID, headers=headers, data={})
verifyGet(response.status_code, "AC")
j2=response.json()['data']
verifyOutput(j1, j2)
response.close()
print()
print()


#POWERPANEL CREATE & GET
payload={
    "name": "PanelA",
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

response = requests.request("GET", url+"/panels/"+ID, headers=headers, data={})
verifyGet(response.status_code, "Panel")
j2=response.json()['data']
verifyOutput(j1, j2)
response.close()
print()
print()


#CABINET CREATE & GET
payload={
    "name": "CabinetA",
    "domain": "DEMO",
    "category": "cabinet",
  "parentId" : None
}
payload['parentId'] = roomID
response = requests.request("POST", url+"/cabinets",
              headers=headers, data=json.dumps(payload))
verifyCreate(response.status_code, "Cabinet")
ID = response.json()['data']['id']
j1=response.json()['data']
cabinetID=ID

response = requests.request("GET", url+"/cabinets/"+ID, headers=headers, data={})
verifyGet(response.status_code, "Cabinet")
j2=response.json()['data']
verifyOutput(j1, j2)
response.close()
print()
print()



#CORRIDOR CREATE & GET
payload={
    "name": "CorridorA",
    "parentId" : None,
    "domain": "DEMO",
    "category": "corridor",
    "temperature": "warm"
}
payload['parentId'] = roomID
response = requests.request("POST", url+"/corridors",
              headers=headers, data=json.dumps(payload))
verifyCreate(response.status_code, "Corridor")
ID = response.json()['data']['id']
j1=response.json()['data']
corridorID=ID

response = requests.request("GET", url+"/corridors/"+ID, headers=headers, data={})
verifyGet(response.status_code, "Corridor")
j2=response.json()['data']
verifyOutput(j1, j2)
response.close()
print()
print()


#ROOM-SENSOR CREATE & GET
payload={
    "name": "RoomSensorLight",
    "parentId" : None,
    "category": "SENSOR-R",
    "domain": "DEMO",
    "type":"room"
}
payload['parentId'] = roomID
response = requests.request("POST", url+"/sensors",
              headers=headers, data=json.dumps(payload))
verifyCreate(response.status_code, "Room-sensor")
ID = response.json()['data']['id']
j1=response.json()['data']
roomsensorID=ID

response = requests.request("GET", url+"/sensors/"+ID, headers=headers, data={})
verifyGet(response.status_code, "Room-sensor")
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




#RACK-SENSOR CREATE & GET
payload={
    "name": "SensorA",
    "parentId" : None,
    "category": "SENSOR-A",
    "domain": "DEMO",
    "type": "rack"
}
payload['parentId'] = rackID
response = requests.request("POST", url+"/sensors",
              headers=headers, data=json.dumps(payload))
verifyCreate(response.status_code, "Rack-sensor")
ID = response.json()['data']['id']
j1=response.json()['data']
racksensorID=ID

response = requests.request("GET", url+"/sensors/"+ID, headers=headers, data={})
verifyGet(response.status_code, "Rack-sensor")
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


#DEVICE-SENSOR CREATE & GET
payload={
    "name": "DeviceSensorA",
    "parentId" : None,
    "category": "SENSOR-D",
    "domain":"DEMO",
    "type":"device"
}
payload['parentId'] = deviceID
response = requests.request("POST", url+"/sensors",
              headers=headers, data=json.dumps(payload))
verifyCreate(response.status_code, "Device-sensor")
ID = response.json()['data']['id']
j1=response.json()['data']
devicesensorID=ID

response = requests.request("GET", url+"/sensors/"+ID, headers=headers, data={})
verifyGet(response.status_code, "Device-sensor")
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
  "rows"        : [
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



#GROUP CREATE & GET
payload={
    "name": "GroupA",
    "type" : "rack",
    "parentId": None,
    "contents":  []
}
payload['parentId'] = roomID
response = requests.request("POST", url+"/groups",
              headers=headers, data=json.dumps(payload))
verifyCreate(response.status_code, "Group")
ID = response.json()['data']['id']
j1=response.json()['data']
groupID=ID

response = requests.request("GET", url+"/groups/"+ID, headers=headers, data={})
verifyGet(response.status_code, "Group")
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
PIDS['acID'] = acID
PIDS['panelID'] = panelID
PIDS['cabinetID'] = cabinetID
PIDS['groupID'] = groupID
PIDS['corridorID'] = corridorID
PIDS['room-sensorID'] = roomsensorID
PIDS['rack-sensorID'] = racksensorID
PIDS['device-sensorID'] = devicesensorID

PIDS['rackID'] = rackID
PIDS['deviceID'] = deviceID
PIDS['room-templateID'] = roomTemplateID
PIDS['obj-templateID'] = objTemplateID
writeEnv()


#RETURN VALUE
if res != True:
    sys.exit(-1)

#else success
sys.exit()