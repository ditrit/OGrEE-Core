#!/usr/bin/env python3
import requests,json, os, sys
from pymongo import MongoClient
from bson.objectid import ObjectId
from pprint import pprint


#RETURN FLAG
res = True

#DB SETUP
dbport=27017   
client = MongoClient("localhost", dbport)     
db = client.ogree

#CONSTANTS
PIDS={"tenantID":None, "siteID":None, "buildingID":None,
        "roomID":None, "acID":None, "panelID":None,
        "cabinetID":None, "groupID":None, "corridorID":None,
        "rackID":None, "deviceID":None,
        "room-sensorID":None,"rack-sensorID":None,
        "device-sensorID":None,
        "room-templateID": None, "obj-templateID": None}


#URL & HEADERS
url = "http://localhost:3001/api"
headers = {
  'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySWQiOjAsIkN1c3RvbWVyIjoib2dyZWUifQ.3A0ZI7oSXxhZ8CGhJORYRXIm3MKroo_9LM939dgGHIo',
  'Content-Type': 'application/json'
}

payloadTable = {
    "tenant":{"name": "DEMO","id": None,"parentId": None,"category": "tenant","description": [],"domain": "DEMO","attributes": {  "color": "FFFFFF",  "mainContact": "Ced",  "mainPhone": "0612345678",  "mainEmail": "ced@ogree3D.com"}},
    "site":{"name": "ALPHA","parentId": None,"category": "site","description": [    "This is a demo..."],"domain": "DEMO","attributes": {    "orientation": "NW",    "usableColor": "5BDCFF",    "reservedColor": "AAAAAA",    "technicalColor": "D0FF78",    "address": "1 rue bidule",    "zipcode": "42000",    "city": "Truc",    "country": "FRANCE",    "gps": "[1,2,0]"}},
    "building":{"name": "B","parentId": None,"category": "building","description": [    "Building B"],"domain": "DEMO","attributes": {    "posXY": "{\"x\":-30.0,\"y\":0.0}",    "posXYUnit": "m",    "posZ": "10",    "posZUnit": "m",    "size": "{\"x\":25.0,\"y\":29.399999618530275}",    "sizeUnit": "m",    "height": "0",    "heightUnit": "m",    "nbFloors": "1"}},
    "room":{"name": "R1","parentId": None,"category": "room","description": [    "First room"],"domain": "DEMO","attributes": {    "posXY": "{\"x\":0.0,\"y\":0.0}",    "posXYUnit": "m",    "posZ": "0",    "posZUnit": "m",    "floorUnit":"f",    "template": "demo.R1",    "orientation": "+N+W",    "size": "{\"x\":22.799999237060548,\"y\":19.799999237060548}",    "sizeUnit": "m",    "height": "3",    "heightUnit": "m"}},
    "rack":{"name": "A01","parentId": None,"category": "rack","description": [    "Rack A01",    "The original one",    "-3/-5\\nA0-Z9"],"domain": "DEMO","attributes": {    "posXY": "{\"x\":10.0,\"y\":0.0}",    "posXYUnit": "tile",    "size": "{\"x\":60.0,\"y\":120.0}",    "sizeUnit": "cm",    "posZ": "Some position",    "posZUnit": "cm",    "height": "42",    "heightUnit": "U",    "template": "Some template",    "orientation": "front",    "vendor": "someVendor",    "type": "someType",    "model": "someModel",    "serial": "someSerial"}},
    "device":{"name": "DeviceA","parentId": None,"category": "device","description": [    "Rack A01",    "The original one",    "-3/-5\\nA0-Z9"],"domain": "Device DOMAIN","attributes": {    "posXY": "{\"x\":10.0,\"y\":0.0}",    "posXYUnit": "tile",    "posZ": "{\"x\":10.0,\"y\":0.0}",    "posZUnit": "tile",    "size": "{\"x\":60.0,\"y\":120.0}",    "sizeUnit": "cm",    "height": "42",    "heightUnit": "U",    "template": "Some template",    "orientation": "front",    "vendor": "someVendor",    "type": "someType",    "model": "someModel",    "serial": "someSerial"}},
    "ac":{"name": "TCL2021","id": None,"parentId": None,"category": "ac","description": [    "TCL"],"domain": "AC DOMAIN","attributes": {}},
    "panel":{"name": "PanelA","id": None,"parentId": None,"category": "powerpanel","description": [    "YINGLI"],"domain": "Panel DOMAIN","attributes": {}},

    "cabinet":{  "name": "CabinetA","parentId" : None,"domain":"DEMO"},
    "group":{"name": "GroupA","type" : "rack","contents":  [],"category":"group", "domain":"DEMO", "parentId":None},
    "corridor":{"name": "CorridorA","parentId" : None,"temperature": "warm","category":"corridor", "domain":"DEMO"},
    
    "room-sensor":{"name": "RoomSensorLight","parentId" : None,"category": "SENSOR-R", "type":"room", "domain":"DEMO"},
    "rack-sensor":{"name": "SensorA","parentId" : None,"category": "SENSOR-A", "type":"rack", "domain":"DEMO"},
    "device-sensor":{"name": "DeviceSensorA","parentId" : None,"category": "SENSOR-D", "type":"device", "domain":"DEMO"},

    "room-template":{"slug"          : "RT1","orientation"   : "+N+E","sizeWDHm"      : ["width","depth","height"],"technicalArea" : ["width","depth","height"],"reservedArea"  : ["width","depth","height"],"separators"    : [],"colors"        : [],"tiles"         : [],"rows"        : []},
    "obj-template":{"slug"        : "RACK1T","description" : "Rack Template 1","category"    : "rack","sizeWDHmm"   : ["width","depth","height"],"fbxModel"    : "1","attributes"  : {  "type" : ""},"colors"      : [],"components"  : [],"slots"       : []}

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

def fixID(item):
    if '_id' in item:
        v = item['_id']
        item['id'] = str(v)
        del item['_id']
    
    return item

def writeEnv():
    with open('.localenv', 'w') as file:
        file.write(json.dumps(PIDS)) # use `json.loads` to do the reverse

    

##ITERATE
ID = None
req = None
for i in payloadTable:
    URL = None
    payload = payloadTable[i]
    if (i == "ac" or i == "panel" or 
        i == "corridor" or 
        i == "cabinet" or i == "room-sensor" or 
        i == "group"):
        payload["parentId"] = PIDS["roomID"]

    elif i == "rack-sensor":
        payload["parentId"] = PIDS["rackID"]

    elif i == "device-sensor":
        payload["parentId"] = PIDS["deviceID"]

    elif i == "room-template" or i == "obj-template":
        #DO NOTHING
        payload
    else:
        payload["parentId"] = ID

    if i.find("-sensor") != -1:
        URL = url+"/sensors"
    else:
        URL = url+"/"+i+"s"


    response = requests.request("POST", URL,headers=headers, data=json.dumps(payload))
    verifyCreate(response.status_code, i[0].upper()+i[1:])
    

    if (i == "room-template" or i == "obj-template"):
        ID = response.json()['data']['slug']
        URL = url+"/"+i+"s/"+ID
        req = {"slug": ID}
    else:
        if 'data' not in response.json():
          print("URL: ", URL)
          print(response.json())
        
        ID = response.json()['data']['id']
        req = {"_id":ObjectId(ID)}

    
    j1 = response.json()['data']



    #Need to match with collection name
    #in DB
    x = i.find("-")
    if x != -1:
        if i.find("-sensor") != -1:
            item = db["sensor"].find_one(req)
        else:
            item = db[i[:x]+"_"+i[x+1:]].find_one(req)
    else:
        item = db[i].find_one(req)

    

    verifyOutput(j1, fixID(item))
    PIDS[i+"ID"] = ID
    print()
    print()




#UPDATE THE ID LIST
writeEnv()

#RETURN VALUE
if res != True:
    sys.exit(-1)

#else success
sys.exit()
