#!/usr/bin/env python3
import requests,json, os, sys
from pymongo import MongoClient
from bson.objectid import ObjectId
from pprint import pprint
from enum import Enum


#GLOBALS
URL = ''

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

#ENUMs
class Entity(Enum):
  TENANT = 0
  SITE = 1 
  BUILDING = 2
  ROOM = 3
  RACK = 4
  DEVICE = 5
  AC = 6
  PANEL = 7
  CABINET = 8

  CORRIDOR = 9
  SENSOR = 10
  GROUP = 11
  ROOMTMPL = 12
  OBJTMPL = 13
  



#FUNCTIONS 
def writeEnv():
    with open('.localenv', 'r') as file:
        data = file.read() # use `json.loads` to do the reverse
        x = json.loads(data)
        for i in x:
            PIDS[i] = x[i]

def verifyOutput(j1, j2):
    global res
    if sorted(j1.items()) == sorted(j2.items()):
        print("JSON Match: Success")
    else:
        res = False
        print("JSON Match: Fail")
        print("J1")
        #print(sorted(j1.items()))
        print(json.dumps(j1, indent=4, sort_keys=True))
        print("J2")
        #print(sorted(j2.items()))
        print(json.dumps(j2, indent=4, sort_keys=True))

def verifyGet(code,entity):
  global res
  if code == 200:
    print('Successfully got '+entity)
    return True
  else:
    res = False
    print('Failed to get '+entity)
    return False

def fixID(item):
    if '_id' in item:
        v = item['_id']
        item['id'] = str(v)
        del item['_id']
    
    return item

def getFromDB(entity, ID):
  req = {}

  x = entity.find("-")
  if x != -1:
    entity = entity[:x]+"_"+entity[x+1:]
    
  x = entity.find("template")
  if x != -1:
    req = {"slug":ID}
  else:
    req = {"_id":ObjectId(ID)}
  
  item = db[entity].find_one(req)
  if item != None:
    item = fixID(item)

  return item

def extractCursor(c, entity):
  docs = {}
  arr = []
  for i in c:
    if entity == "tenant":
      i["parentId"] = None
    
    arr.append(fixID(i))

  docs['objects'] = arr
  return docs

def compareNested(j1, j2):
  for i in j2:
    q = json.dumps(j1[i])
    verifyOutput(q, j2[i])

def getManyFromDB(entity):
  x = entity.find("-")
  if x != -1:
    entity = entity[:x]+"_"+entity[x+1:]

  return db[entity].find({})

def getOneFromDBGeneric(entity, req):
  return db[entity].find_one(req)

def getManyFromDBGeneric(entity, req):
  return db[entity].find(req)

def sToI(ent):
    return Entity[ent].value

def iToS(ent):
    return Entity(ent).name

def getEntitiesOfAncestor(ent, entStr, ID, wantedEnt):
  ans = []
  if ent == Entity.TENANT.value:

    t = getOneFromDBGeneric(entStr, {"name":ID})
    if t == None:
      print("Tenant was nil")
      print("ENT: ", entStr)
      print("ENT#:", ent)
      print("Name-ID:",ID)
      
      return None

  else:
    ID = ObjectId(ID)
    t = getFromDB(entStr, ID)
    if t == None:
      print(entStr+" was NIl")
      return None
  

  t = fixID(t)
  v = ent+1
  sub = getManyFromDBGeneric(iToS(v).lower(), {"parentId": str(t["id"])})
  if sub == None:
    print("SUB was nil")
    return None
    
  sub = extractCursor(sub, iToS(ent+1))
  if wantedEnt == "":
    wantedEnt = iToS(ent + 2).lower()
	
  for i in sub['objects']:
    req = {"parentId":str(i["id"])}
    x = getManyFromDBGeneric(wantedEnt, req)
    x = extractCursor(x, wantedEnt)
    ans.append(x)
  
  return ans[0]

def getEntityHierarchy(entity, entLoc, ID, end):

  if entLoc < end:

    #Get Top
    t = getFromDB(entity, ID)
    if t == None:
      print("Unable to get Top Entity: ", entity)
      print("While retrieving hierarchy")
      return None

    t = fixID(t)

    #Retrieve associated nonhierarchal objects
    if entity == "device":
        t = getDeviceHierarchy(t)

			
		

    if entity == "rack":
      x = getManyFromDBGeneric("sensor", {"parentId":t["id"]})
      if x != None:
        x = extractCursor(x, "sensor")
        t["rack_sensors"] = x['objects']

      x = getManyFromDBGeneric("group", {"parentId":t["id"]})
      if x != None:
        x = extractCursor(x, "group")
        t["groups"] = x['objects']


    if entity == "room":
      #ITER Through all nonhierarchal objs
      i = sToI("AC")
      while i < sToI("GROUP")+1:
        ent = iToS(i).lower()
        #if ent == "roomsensor":
        #    ent = "room_sensor"
        x = getManyFromDBGeneric(ent, {"parentId":t["id"]})
        if x != None:
          x = extractCursor(x, ent)
          
          if ent == "sensor":
            t["room_sensors"] = x['objects']
          else:
            t[ent+"s"] = x['objects']

        i += 1
    	
    
    subEnt = iToS(entLoc+1).lower()
    #print("SubEnt: ", subEnt)
    # Get immediate children
    children = getManyFromDBGeneric(subEnt, {"parentId":ID})
    if children == None:
      print("Error while getting children for GetHierarchy")
      print("SubEnt: ", subEnt)
      print("PID: ", ID)
      return None
      
    children = extractCursor(children, subEnt)
    t["children"] = children['objects']

    q = 0
    for i in children['objects']:
      subIdx = iToS(entLoc + 1).lower()
      subID = i["id"]
      x = getEntityHierarchy(subIdx, entLoc+1, subID, end)

      if x != None:
        children['objects'][q] = x
      
      q += 1

    
    return t


def getDeviceHierarchy(top):
  if top == None:
    return top

  #GET DEVICE SENSORS
  x = getManyFromDBGeneric("sensor", {"parentId":top["id"]})
  if x != None:
    x = extractCursor(x, "sensor")
    top["device_sensors"] = x['objects']


  #GET CHILDREN 
  children = getManyFromDBGeneric("device", {"parentId":top["id"]})
  children = extractCursor(children, "device")
  if len(children) < 1:
    print("Unable to get children")
    print("PID: ", top["id"])


  #GET HIERARCHY FOR EACH CHILD
  q = 0
  for i in children['objects']:
    children['objects'][q] = getDeviceHierarchy(i)
    q +=1 

  top['children'] = children['objects']
  return top

def handleRangedHierarchy(URLs, obj, ID):
  finalURL = ""

  #ITER  RANGEDHIERARCHY
  for idx in URLs:
    endIdx = idx.rfind("/")
    end = idx[endIdx+1:]
    end = end[:len(end)-1]
    limit = sToI(end.upper())

    if obj == "tenant":
      finalURL =  url+"/tenants/DEMO/"+idx
    else:
      finalURL =  url+"/"+objs+"/"+ID+"/"+idx

    response = requests.request("GET",finalURL,headers=headers, data={})
    verifyGet(response.status_code, 
                  obj[0].upper()+obj[1:]+"'s Ranged Hierarchy to "+end)
      
      
      
    j2 = getEntityHierarchy(obj, sToI(obj.upper()), ID, limit)


    if 'data' in response.json():
      j1 = response.json()['data']
    else:
      print("ERROR! Data was not in the response for:", obj)
      j1 = response.json()


    verifyOutput(j1, j2)
    response.close()
    print()
    print()

    #TEST LIMIT REQUESTs
    #/api/tenants/{tenant_name}/all?limit={#}
    OG = finalURL

    limitNum = limit - sToI(obj.upper()) +1 
    finalURL = url+"/"+objs+"/"+ID+"/all?limit="+str(limitNum - 1)
    print("OG:",OG)
    print("Limit:",finalURL)
    
    response = requests.request("GET",finalURL,headers=headers, data={})
    verifyGet(response.status_code, 
                  obj[0].upper()+obj[1:]+"'s Ranged Hierarchy to "+end+ " using limit")
      


    if 'data' in response.json():
      j1 = response.json()['data']
    else:
      print("ERROR! Data was not in the response for:", obj)
      j1 = response.json()


    verifyOutput(j1, j2)
    response.close()
    print()
    print()

def getEntityUsingAncestry(url, obj):
  q = 2
  resp = None
  if obj == "tenant":
    resp = getOneFromDBGeneric("tenant", {"name":url[1]})
  else:
    c = url[0][:len(url[0])-1]
    idx = c.find("-")
    if idx != -1:
      c = c.replace("-", "_")
    resp = getOneFromDBGeneric(c, {"_id":ObjectId(url[1])})
  
  
  resp = fixID(resp)
  req = None
  while q < len(url):
    c = url[q][:len(url[q])-1]
    if c.find("-") != -1:
      c = c.replace("-", "_")

    if q+1 < len(url):
      req = {"parentId":resp["id"], "name":url[q+1]}
      r1 = getOneFromDBGeneric(c, req)
    else:
      req = {"parentId":resp["id"]}
      r1 = getManyFromDBGeneric(c, req)
      r1 = extractCursor(r1, c)['objects']


    resp = r1
    resp = fixID(resp)
    q += 2
  
  return resp

def apiCall(URL, entity, iterator, verifyingStr, ID, reqType, wantedEnt):
  j2 = None
  print("OUR URL:", URL)
  response = requests.request("GET", URL, headers=headers, data={})
  verifyGet(response.status_code, verifyingStr)

  if reqType == "single":
    j2 = getFromDB(entity,PIDS[iterator])

  if reqType == "all":
    j2 = getManyFromDB(entity)
    j2 = extractCursor(j2, entity)

  if reqType == "hierarchy":
    j2 = getEntityHierarchy(obj, sToI(obj.upper()), PIDS[i], 6)

  if reqType == "mainGetException":
    j2 = getEntitiesOfAncestor(sToI(entity.upper()), entity, ID, "")

  if reqType == "1levelLower":
    subEnt = iToS(sToI(entity.upper())+1).lower()
    j2 = getManyFromDBGeneric(subEnt, {"parentId": PIDS[iterator]})
    j2 = extractCursor(j2, subEnt)

  if reqType == "1levelLower+":
    if wantedEnt.find("-") != -1:
      wantedEnt = wantedEnt.replace("-", "_")
    
    print("WantedENT:", wantedEnt)
    print("PID:", PIDS[iterator])
    print("ITER:", iterator)
    j2 = getManyFromDBGeneric(wantedEnt, {"parentId": PIDS[iterator]})
    j2 = extractCursor(j2, wantedEnt)
    
  if reqType == "named&subEnt":
    x = wantedEnt
    val = getEntityUsingAncestry(x, entity)
    #Fix if cursor
    if type(val) != dict:
      entityType = (x[len(x)-1])[:len((x[len(x)-1]))-1]
      val = extractCursor(val, entityType)
    
    j2 = val

  if reqType == "finalCases":
    j2 = getEntitiesOfAncestor(sToI(entity.upper()), entity,ID, wantedEnt)
    print()


  if 'data' in response.json():
    j1 = response.json()['data']
  else:
    print("ERROR! Data was not in the response for:", obj)
    j1 = response.json()

  verifyOutput(j1, j2)
  response.close()
  print()
  print()



#INIT PID Dict
writeEnv()



#ITERATE 
for i in PIDS:
  obj = None
  objs = None
  print("yo")

  if i.find("-sensor") != -1:
    obj = "sensor"
    objs = "sensors"
  else:
    obj = i[:len(i)-2]
    objs = i[:len(i)-2]+"s"

  #SINGLE OBJ GET WITH ID
  apiCall(url+"/"+objs+"/"+PIDS[i], obj, i,obj[0].upper()+obj[1:] , None, "single", None)


  #GET ALL OBJS
  apiCall(url+"/"+objs, obj, i, objs, None, "all", None)

  #OBJECT SEARCH 
  #(only the main attrs. category, domain, name)

  
  #GET EXCEPTIONS
  if (obj == "tenant" or obj == "site" 
      or obj == "building" or obj == "room"):

    subObj = iToS(sToI(obj.upper())+2).lower()
    print("SubOBJ:", subObj)
    if obj == "tenant":
      apiCall(url+"/tenants/DEMO/"+subObj+"s", obj, i,
       i[0].upper()+i[1:len(i)-2]+"\'s "+subObj +"s", "DEMO", "mainGetException", None)
    else:
      apiCall(url+"/"+objs+"/"+PIDS[i]+"/"+subObj+"s", obj, i,
       i[0].upper()+i[1:len(i)-2]+"\'s "+subObj +"s", PIDS[i], "mainGetException", None)

  
  #GET OBJ HIERARCHY
  if (obj == "tenant" or obj == "site" 
      or obj == "building" or obj == "room" or obj == "rack"):

    apiCall(url+"/"+objs+"/"+PIDS[i]+"/all", obj, i,
                 obj[0]+obj[1:]+"'s Hierarchy", None, "hierarchy", None)

  
  #RANGED HIERARCHY
  if obj == "tenant":
    tenantRangedURLs = [
                        "all/sites/buildings",
                        "all/sites/buildings/rooms",
                        "all/sites/buildings/rooms/racks",
                        "all/sites/buildings/rooms/racks/devices"
                        ]

    handleRangedHierarchy(tenantRangedURLs, obj, PIDS[i])
  

  if obj == "site":
    siteRangedURLs = ["all/buildings/rooms",
                      "all/buildings/rooms/racks",
                      "all/buildings/rooms/racks/devices"
                      ]
    
    handleRangedHierarchy(siteRangedURLs, obj, PIDS[i])

    
  if obj == "building":
    bldgRangedURLs = ["all/rooms/racks",
                      "all/rooms/racks/devices"
                      ]

    handleRangedHierarchy(bldgRangedURLs, obj, PIDS[i])


  if obj == "room":
    handleRangedHierarchy(["/all/racks/devices"], obj, PIDS[i])



  #We cannot do query testing since searching is broken for now

  #ONE LEVEL LOWER REQUESTS
  if (obj == "tenant" or obj == "site" 
      or obj == "building" or obj == "room" or obj == "rack"):
    finalURL = ""
    subEnt = iToS(sToI(obj.upper())+1).lower()
    if obj == "tenant":
      finalURL = url+"/"+objs+"/DEMO/sites"
    else:
      finalURL = url+"/"+objs+"/"+PIDS[i]+"/"+subEnt+"s"

    apiCall(finalURL, obj, i, 
    obj[0].upper()+obj[1:]+"s"+" one level lower req", None, "1levelLower", None)

    
    if obj == "room":
      #All associated objs for room
      idx = 6 #AC
      while idx < sToI("SENSOR")+1: #Loop until Rack Sensor
        subEnt = iToS(idx).lower()
        #if subEnt == "roomsensor":
        #  subEnt = "room-sensor"
        
        finalURL = url+"/"+objs+"/"+PIDS[i]+"/"+subEnt+"s"
        print(finalURL)
        apiCall(finalURL, obj, i,
        objs[0].upper()+objs[1:]+" one level lower req: "+subEnt,
        None,"1levelLower+", subEnt)
        idx += 1

    
  if obj == "rack":
    subEnt = "sensor"
        
    finalURL = url+"/"+objs+"/"+PIDS[i]+"/"+subEnt+"s"
    print("JUSTb41LevelLower+", finalURL)
    apiCall(finalURL, obj, i,
    objs[0].upper()+objs[1:]+" one level lower req: "+subEnt,
    None, "1levelLower+", subEnt)

  if obj == "device":
    subEnt = "sensor"

    finalURL = url+"/"+objs+"/"+PIDS[i]+"/"+subEnt+"s"
    print("JUSTb41LevelLower+", finalURL)
    apiCall(finalURL, obj, i,
    objs[0].upper()+objs[1:]+" one level lower req: "+subEnt,
    None, "1levelLower+", subEnt)

        

  
  #NAMED & SUBENTITY SECTION
  tenantList = [
  "/api/tenants/DEMO/sites/ALPHA/buildings/B/rooms/R1/racks/A01/devices/DeviceA",
  "/api/tenants/DEMO/sites/ALPHA/buildings/B/rooms/R1/racks/A01/devices",
  "/api/tenants/DEMO/sites/ALPHA/buildings/B/rooms/R1/racks/A01",
  "/api/tenants/DEMO/sites/ALPHA/buildings/B/rooms/R1/racks",
  "/api/tenants/DEMO/sites/ALPHA/buildings/B/rooms/R1",
  "/api/tenants/DEMO/sites/ALPHA/buildings/B/rooms",
  "/api/tenants/DEMO/sites/ALPHA/buildings/B",
  "/api/tenants/DEMO/sites/ALPHA/buildings",
  "/api/tenants/DEMO/sites/ALPHA"
  ]

  siteList = [
  "/api/sites/"+PIDS["siteID"]+"/buildings/B/rooms/R1/racks/A01/devices/DeviceA",
  "/api/sites/"+PIDS["siteID"]+"/buildings/B/rooms/R1/racks/A01/devices",
  "/api/sites/"+PIDS["siteID"]+"/buildings/B/rooms/R1/racks/A01",
  "/api/sites/"+PIDS["siteID"]+"/buildings/B/rooms/R1/racks",
  "/api/sites/"+PIDS["siteID"]+"/buildings/B/rooms/R1",
  "/api/sites/"+PIDS["siteID"]+"/buildings/B/rooms",
  "/api/sites/"+PIDS["siteID"]+"/buildings/B"
  ]

  bldgList = [
  "/api/buildings/"+PIDS["buildingID"]+"/rooms/R1/racks/A01/devices/DeviceA",
  "/api/buildings/"+PIDS["buildingID"]+"/rooms/R1/racks/A01/devices",
  "/api/buildings/"+PIDS["buildingID"]+"/rooms/R1/racks/A01",
  "/api/buildings/"+PIDS["buildingID"]+"/rooms/R1/racks",
  "/api/buildings/"+PIDS["buildingID"]+"/rooms/R1"
  ]

  roomList = [
    "/api/rooms/"+PIDS["roomID"]+"/racks/A01/devices/DeviceA",
  "/api/rooms/"+PIDS["roomID"]+"/racks/A01/devices",
  "/api/rooms/"+PIDS["roomID"]+"/racks/A01",
  "/api/rooms/"+PIDS["roomID"]+"/acs/TCL2021",
  "/api/rooms/"+PIDS["roomID"]+"/panels/PanelA",
  "/api/rooms/"+PIDS["roomID"]+"/cabinets/CabinetA",
  "/api/rooms/"+PIDS["roomID"]+"/corridors/CorridorA",
  "/api/rooms/"+PIDS["roomID"]+"/sensors/RoomSensorLight"
  ]

  rackList = [
    "/api/racks/"+PIDS["rackID"]+"/sensors/SensorA",
    "/api/racks/"+PIDS["rackID"]+"/devices/DeviceA"
  ]

  deviceList = ["/api/devices/"+PIDS["deviceID"]+"/sensors/DeviceSensorA"]

  listDict = {"tenantID":tenantList, "siteID":siteList,
              "buildingID":bldgList, "roomID":roomList,
              "rackID":rackList, "deviceID":deviceList
              }
  
  if i in listDict:
    for q in listDict[i]:

      #GET OBJ HIERARCHY
      x = q[5:].split("/")
      apiCall(url+"/"+q[5:], obj, i,
       objs[0].upper()+objs[1:]+"'s Ancestry call", None, "named&subEnt", x)

      


#FINAL EXCEPTION CASES
bldEList = [
  "/api/buildings/"+PIDS["buildingID"]+"/acs",
  "/api/buildings/"+PIDS["buildingID"]+"/corridors",
  "/api/buildings/"+PIDS["buildingID"]+"/cabinets",
  "/api/buildings/"+PIDS["buildingID"]+"/panels",
  "/api/buildings/"+PIDS["buildingID"]+"/sensors"
]

for h in bldEList:
  localarr = h.split("/")
  lastone = localarr[len(localarr)-1]
  eType = lastone[:len(lastone)-1]
  if eType.find("-") != -1:
    eType=eType.replace("-", "_")

  apiCall(url+"/"+h[5:], "building", None,"building call for room\'s "+eType, PIDS["buildingID"], "finalCases", eType )


#Room Exception
#roomExceptionURL= "/api/rooms/"+PIDS["roomID"]+"/sensors"
#apiCall(url+"/"+roomExceptionURL[5:], "room", None,"Room call for rack\'s sensors",PIDS["roomID"], "finalCases", "sensor" )



#Rack Exception
#rackExceptionURL = "/api/racks/"+PIDS["rackID"]+"/sensors"
#apiCall(url+"/"+rackExceptionURL[5:], "rack", None, "Rack call for devices\'s sensors", PIDS["rackID"], "finalCases", "sensor")



if res == True:
  print("Test Script Successful")
else:
  print("Test Script Failure")

