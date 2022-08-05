#!/usr/bin/env python3
import requests,json, os, sys

#NOTES
'''
The nested entities under room
are updated before room because the PUT will 
erase the entire Room JSON thus losing
the nested entities.
'''

#GLOBALS
URL = ''

#RETURN FLAG
res = True

#CONSTANTS
PIDS={"tenantID":None, "siteID":None, "buildingID":None,
        "roomID":None, "acID":None, "panelID":None,
        "cabinetID":None, "groupID":None, "corridorID":None,
        "room-sensorID":None,"rack-sensorID":None,
        "device-sensorID":None,
        "rackID":None, "deviceID":None,
        "room-templateID": None, "obj-templateID": None}
        
url = "http://localhost:3001/api"
headers = {
  'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySWQiOjAsIkN1c3RvbWVyIjoib2dyZWUifQ.3A0ZI7oSXxhZ8CGhJORYRXIm3MKroo_9LM939dgGHIo',
  'Content-Type': 'application/json'
}


#FUNCTIONS
def verifyUpdate(code, entity):
  global res
  if code == 200:
    print('Successfully updated '+entity)
  else:
    res = False
    print('Failed to update '+entity)
    print('Status Code:', code)

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
    with open('.localenv', 'r') as file:
        data = file.read() # use `json.loads` to do the reverse
        x = json.loads(data)
        for i in x:
            PIDS[i] = x[i]


#INIT PID Dict
writeEnv()



payloadTable = {
    "tenant":{"attributes": {  "another":"one"},"someUsefulAttr":"customValue"},
    "site":{"name": "SiteA999","domain": "DERELICT","attributes": {    "orientation": "NW",    "another":"one"},"someUsefulAttr":"customValue"},
    "building":{"name": "Abandoned-Building","description": [    "Wassup yo",    "Guess whos back"],"domain": "999","attributes": {    "size":"88"}},
    "ac":{"name": "AquaSky01","category": "ac","description": [    "SPEED"],"domain": "AC DOMAIN","attributes": {    "new": "IDX"}},
    "panel":{"name": "Power_Panel_99","category": "SOLARpanel","description": [    "GRENOBLE DERELICT"],"domain": "PANEL DOMAIN","attributes": {    "new": "IDX"}},
    "cabinet":{"name": "TAKEUSDOWN","customAttr":"customValue"},
    "group":{"name": "Group909","racks":  ["SomeRackHere"], "another":"ONE"},
    "corridor":{"name": "Corridor808","temperature": "30C"},
    "room":{"name": "PatchedRoom","category": "999","description": [    "999"],"domain": "999","attributes": {    "posXY": "999",    "patch":"customValue"}, "patch":"customValue"},
    "rack":{"name": "Abandoned-Rack","category": "rack","description": [    "99",    "999",    "9999"],"domain": "Abandoned Rack","attributes": {"posXY": "999","posXYUnit": "tile","size": "99","sizeUnit": "cm","height": "999","heightUnit": "U","template": "","orientation": "front","vendor": "999","type": "999","model": "999","serial": "999"},"someAttr":"customValue"},
    "device":{"name": "Abandoned-Device","category": "999","description": [    "Rack A01",    "The original one",    "-3/-5\\nA0-Z9"],"domain": "99","attributes": {    "posXY": "99",    "posXYUnit": "tile",    "size": "99",    "sizeUnit": "cm",    "height": "99",    "heightUnit": "U",    "template": "",    "orientation": "front",    "vendor": "99",    "type": "99",    "model": "99",    "serial": "99",    "someATTR":"customValue"},"another":"one"},
    "device-sensor":{"name": "Corridor808","temperature": "30C"},
    "rack-sensor":{"name": "Corridor808","temperature": "30C"},
    "room-sensor":{"name": "Corridor808","temperature": "30C"},
    "room-template":{"slug"          : "HOTTESTDNB","orientation"   : "+N+W","another":"ONE"},
    "obj-template":{"slug"        : "RACK2000","description" : "Rack Template 2000","category"    : "rack","sizeWDHmm"   : [],"fbxModel"    : "1","attributes"  : {  "type" : "somevale",  "another":"one"},"colors"      : [],"components"  : [],"slots"       : []}
    }



#START
ID = None
roomID = PIDS["roomID"]
#ITERATE
for i in payloadTable:
    URL = None
    payload = payloadTable[i]

    if i.find("-sensor") != -1:
      URL = url+"/sensors/"+PIDS[i+"ID"]
    else:
      URL = url+"/"+i+"s/"+PIDS[i+"ID"]


    response = requests.request("PATCH", URL,headers=headers, data=json.dumps(payload))
    verifyUpdate(response.status_code, i[0].upper()+i[1:])

    if (i == "room-template" or i == "obj-template"):
        ID = response.json()['data']['slug']
        URL = url+"/"+i+"s/"+ID
    else:
        ID = response.json()['data']['id']

    
    j1 = response.json()['data']
    

    
    response = requests.request("GET", URL, headers=headers, data={})
    verifyGet(response.status_code, i[0].upper()+i[1:])
    j2=response.json()['data']
    verifyOutput(j1, j2)
    response.close()
    print()
    print()


#Just need to update the Template slugs 
#because they were modified in the update process
PIDS["room-templateID"] = "HOTTESTDNB"
PIDS["obj-templateID"] = "RACK2000"
with open('.localenv', 'w') as file:
        file.write(json.dumps(PIDS)) # use `json.loads` to do the reverse