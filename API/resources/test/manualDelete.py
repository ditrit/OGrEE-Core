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

#FUNCTIONS
def writeEnv():
    with open('.localenv', 'r') as file:
        data = file.read() # use `json.loads` to do the reverse
        x = json.loads(data)
        for i in x:
            PIDS[i] = x[i]

def extractCursor(c):
  arr = [x for x in c ]
  return arr

def verifyDelete(code,entity, ID):
  global res
  x = None
  if code == 204:
    idx = entity.find("-")
    if idx != -1:
      entity = entity.replace("-", "_")

    if entity == "room_template" or entity == "obj_template":
      x = extractCursor(db[entity].find({"slug":ID}))
    else:
      x = extractCursor(db[entity].find({"_id":ObjectId(ID)}))
    

    if len(x) == 0:
      print('Successfully Deleted '+entity)
      print()
    else:
      print("API contradicted the DB!")
      print()
      res = False

    return True
  else:
    res = False
    print('Failed to Delete '+entity)
    print('Status Code: ', code)
    print()
    return False


#SETUP
writeEnv()


#ITER
for idx in reversed(PIDS.items()):
  if (idx[0].find("-sensorID" ) != -1): 
      #Sensors require different URL
      URL = url+"/sensors/"+PIDS[idx[0]]
  else:
      URL = url+"/"+idx[0][:len(idx[0])-2]+"s/"+PIDS[idx[0]]

  response = requests.request("DELETE",URL,headers=headers, data={})
  verifyDelete(response.status_code, idx[0][:len(idx[0])-2], PIDS[idx[0]] )



#RETURN VALUE
if res == False:
  sys.exit(-1)
