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
        "wallID":None, "aisleID":None,"tileID":None, 
        "cabinetID":None, "groupID":None, "corridorID":None,
        "rackID":None, "deviceID":None,
        "room-sensorID":None,"rack-sensorID":None,
        "device-sensorID":None,
        "room-templateID": None, "obj-templateID": None}
        

#URL & HEADERS    
url = "http://localhost:3001/api"
headers = {
  'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySWQiOjY2NDA0NjEyNzM0MjQxOTk2OX0.cB1VkYQLlXCatzMiEWGFfJKKx9h8Vsr2vdlylNMe7hs',
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
  WALL = 8
  CABINET = 9
  AISLE = 10
  TILE = 11 

  CORRIDOR = 12
  ROOMSENSOR = 13 
  RACKSENSOR = 14
  DEVICESENSOR = 15
  ROOMTMPL = 16
  OBJTMPL = 17
  GROUP = 18



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
        print(sorted(j1.items()))
        print("J2")
        print(sorted(j2.items()))

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
      x = getManyFromDBGeneric("rack_sensor", {"parentId":t["id"]})
      if x != None:
        x = extractCursor(x, "rack_sensor")
        t["rack_sensors"] = x['objects']


    if entity == "room":
      #ITER Through all nonhierarchal objs
      i = sToI("AC")
      while i < sToI("RACKSENSOR"):
        ent = iToS(i).lower()
        if ent == "roomsensor":
            ent = "room_sensor"
        x = getManyFromDBGeneric(ent, {"parentId":t["id"]})
        if x != None:
          x = extractCursor(x, ent)
          
            
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
  x = getManyFromDBGeneric("device_sensor", {"parentId":top["id"]})
  if x != None:
    x = extractCursor(x, "device_sensor")
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
    finalURL = url+"/"+objs+"/"+ID+"/all?limit="+str(limitNum)
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





#INIT PID Dict
writeEnv()



#ITERATE 
for i in PIDS:
  print("yo")


  obj = i[:len(i)-2]
  objs = i[:len(i)-2]+"s"

  #SINGLE OBJ GET WITH ID
  response = requests.request("GET", url+"/"+objs+"/"+PIDS[i], headers=headers, data={})
  verifyGet(response.status_code, i[0].upper()+i[1:len(i)-2])

  j2 = getFromDB(obj,PIDS[i])

  if 'data' in response.json():
    j1 = response.json()['data']
  else:
    print("ERROR! Data was not in the response for:", obj)
    j1 = response.json()
  
  verifyOutput(j1, j2)
  response.close()
  print()
  print()


  #GET ALL OBJS
  response = requests.request("GET", url+"/"+objs, headers=headers, data={})
  verifyGet(response.status_code, i[0].upper()+i[1:len(i)-2]+"s")
  j2 = getManyFromDB(obj)

  if 'data' in response.json():
    j1 = response.json()['data']
  else:
    print("ERROR! Data was not in the response for:", obj)
    j1 = response.json()

  #Fix response 
  extractedJ2 = extractCursor(j2, obj)

  verifyOutput(j1, extractedJ2)
  response.close()
  print()
  print()
  #sys.exit(0)


  #OBJECT SEARCH 
  #(only the main attrs. category, domain, name)

  
  #GET EXCEPTIONS
  if (obj == "tenant" or obj == "site" 
      or obj == "building" or obj == "room"):

    subObj = iToS(sToI(obj.upper())+2).lower()
    print("SubOBJ:", subObj)
    if obj == "tenant":
      response = requests.request("GET", url+"/tenants/DEMO/"+subObj+"s", headers=headers, data={})
      locID = "DEMO" # From other scripts, this is the name of tenant
    else:
      finalURL=  url+"/"+objs+"/"+PIDS[i]+"/"+subObj+"s"
      response = requests.request("GET",finalURL, headers=headers, data={})
      locID = PIDS[i]
    
    verifyGet(response.status_code, i[0].upper()+i[1:len(i)-2]+"\'s "+subObj +"s")

    j2 = getEntitiesOfAncestor(sToI(obj.upper()), obj, locID, "")

    if 'data' in response.json():
      j1 = response.json()['data']
    else:
      print("ERROR! Data was not in the response for:", obj)
      j1 = response.json()

    verifyOutput(j1, j2)
    response.close()
    print()
    print()


  #GET OBJ HIERARCHY
  if (obj == "tenant" or obj == "site" 
      or obj == "building" or obj == "room" or obj == "rack"):
    response = requests.request("GET", url+"/"+objs+"/"+PIDS[i]+"/all", headers=headers, data={})
    verifyGet(response.status_code, i[0].upper()+i[1:len(i)-2]+"'s Hierarchy")
    j2 = getEntityHierarchy(obj, sToI(obj.upper()), PIDS[i], 6)

    if 'data' in response.json():
      j1 = response.json()['data']
    else:
      print("ERROR! Data was not in the response for:", obj)
      j1 = response.json()


    verifyOutput(j1, j2)
    response.close()
    print()
    print()


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


    response = requests.request("GET", finalURL, headers=headers, data={})
    verifyGet(response.status_code, i[0].upper()+i[1:len(i)-2]+"s"+" one level lower req")
    j2 = getManyFromDBGeneric(subEnt, {"parentId": PIDS[i]})

    if 'data' in response.json():
      j1 = response.json()['data']
    else:
      print("ERROR! Data was not in the response for:", obj)
      j1 = response.json()

    #Fix response 
    extractedJ2 = extractCursor(j2, subEnt)

    verifyOutput(j1, extractedJ2)
    response.close()
    print()
    print()

    if obj == "room":
      #All associated objs for room
      idx = 6 #AC
      while idx < 14: #Loop until Rack Sensor
        subEnt = iToS(idx).lower()
        if subEnt == "roomsensor":
          subEnt = "room-sensor"
        
        finalURL = url+"/"+objs+"/"+PIDS[i]+"/"+subEnt+"s"
        response = requests.request("GET", finalURL, headers=headers, data={})
        verifyGet(response.status_code, i[0].upper()+i[1:len(i)-2]+"s"+" one level lower req: "+subEnt)
        
        if subEnt == "room-sensor":
          subEnt = "room_sensor"

        j2 = getManyFromDBGeneric(subEnt, {"parentId": PIDS[i]})

        if 'data' in response.json():
          j1 = response.json()['data']
        else:
          print("ERROR! Data was not in the response for:", obj)
          j1 = response.json()

        #Fix response 
        extractedJ2 = extractCursor(j2, subEnt)

        verifyOutput(j1, extractedJ2)
        response.close()
        print()
        print()
        idx += 1

    
  if obj == "rack":
    subEnt = "rack-sensor"
        
    finalURL = url+"/"+objs+"/"+PIDS[i]+"/"+subEnt+"s"
    response = requests.request("GET", finalURL, headers=headers, data={})
    verifyGet(response.status_code, i[0].upper()+i[1:len(i)-2]+"s"+" one level lower req: "+subEnt)
    

    j2 = getManyFromDBGeneric("rack_sensor", {"parentId": PIDS[i]})

    if 'data' in response.json():
      j1 = response.json()['data']
    else:
      print("ERROR! Data was not in the response for:", obj)
      j1 = response.json()

    #Fix response 
    extractedJ2 = extractCursor(j2, subEnt)

    verifyOutput(j1, extractedJ2)
    response.close()
    print()
    print()

  if obj == "device":
    subEnt = "device-sensor"
        
    finalURL = url+"/"+objs+"/"+PIDS[i]+"/"+subEnt+"s"
    response = requests.request("GET", finalURL, headers=headers, data={})
    verifyGet(response.status_code, i[0].upper()+i[1:len(i)-2]+"s"+" one level lower req: "+subEnt)

    j2 = getManyFromDBGeneric("device_sensor", {"parentId": PIDS[i]})

    if 'data' in response.json():
      j1 = response.json()['data']
    else:
      print("ERROR! Data was not in the response for:", obj)
      j1 = response.json()

    #Fix response 
    extractedJ2 = extractCursor(j2, subEnt)

    verifyOutput(j1, extractedJ2)
    response.close()
    print()
    print()



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
  "/api/rooms/"+PIDS["roomID"]+"/walls/Undercover",
  "/api/rooms/"+PIDS["roomID"]+"/aisles/AisleA",
  "/api/rooms/"+PIDS["roomID"]+"/tiles/TileA",
  "/api/rooms/"+PIDS["roomID"]+"/cabinets/CabinetA",
  "/api/rooms/"+PIDS["roomID"]+"/corridors/CorridorA",
  "/api/rooms/"+PIDS["roomID"]+"/room-sensors/RoomSensorLight"
  ]

  rackList = [
    "/api/racks/"+PIDS["rackID"]+"/rack-sensors/SensorA",
    "/api/racks/"+PIDS["rackID"]+"/devices/DeviceA"
  ]

  deviceList = ["/api/devices/"+PIDS["deviceID"]+"/device-sensors/DeviceSensorA"]

  listDict = {"tenantID":tenantList, "siteID":siteList,
              "buildingID":bldgList, "roomID":roomList,
              "rackID":rackList, "deviceID":deviceList
              }
  
  if i in listDict:
    for q in listDict[i]:
      x = q[5:].split("/")
      val = getEntityUsingAncestry(x, obj)
      #print(val)

      #GET OBJ HIERARCHY
      response = requests.request("GET", url+"/"+q[5:], headers=headers, data={})
      verifyGet(response.status_code, i[0].upper()+i[1:len(i)-2]+"'s Ancestry call")
      print("FOR: "+url+"/"+q[5:])

      if 'data' in response.json():
        j1 = response.json()['data']
      else:
        print("ERROR! Data was not in the response for:", obj)
        print("URL:"+url+"/"+q[5:] )
        j1 = response.json()

      #Fix if cursor
      if type(val) != dict:
        entityType = (x[len(x)-1])[:len((x[len(x)-1]))-1]
        val = extractCursor(val, entityType)


      verifyOutput(j1, val)
      response.close()
      print()
      print()
      #sys.exit(0)
      




#FINAL EXCEPTION CASES
bldEList = [
  "/api/buildings/"+PIDS["buildingID"]+"/acs",
  "/api/buildings/"+PIDS["buildingID"]+"/corridors",
  "/api/buildings/"+PIDS["buildingID"]+"/cabinets",
  "/api/buildings/"+PIDS["buildingID"]+"/tiles",
  "/api/buildings/"+PIDS["buildingID"]+"/aisles",
  "/api/buildings/"+PIDS["buildingID"]+"/panels",
  "/api/buildings/"+PIDS["buildingID"]+"/walls",
  "/api/buildings/"+PIDS["buildingID"]+"/room-sensors"
]

for h in bldEList:
  localarr = h.split("/")
  lastone = localarr[len(localarr)-1]
  eType = lastone[:len(lastone)-1]
  if eType.find("-") != -1:
    eType=eType.replace("-", "_")
  someval= getEntitiesOfAncestor(sToI("BUILDING"), "building",PIDS["buildingID"], eType)

  response = requests.request("GET", url+"/"+h[5:], headers=headers, data={})
  b = verifyGet(response.status_code, "building call for room\'s "+eType)
  if b != True:
    print("Failure @URL: ")
    print(url+"/"+h[5:])


  if 'data' in response.json():
    j1 = response.json()['data']
  else:
    print("ERROR! Data was not in the response for:", "building")
    print("URL:"+url+"/"+h[5:] )
    j1 = response.json()



  verifyOutput(j1, someval)
  response.close()
  print()
  print()


roomExceptionURL= "/api/rooms/"+PIDS["roomID"]+"/rack-sensors"
someval= getEntitiesOfAncestor(sToI("ROOM"), "room",PIDS["roomID"], "rack_sensor")

response = requests.request("GET", url+"/"+roomExceptionURL[5:], headers=headers, data={})
b = verifyGet(response.status_code, "Room call for rack\'s sensors")
if b != True:
  print("Failure @URL: ")
  print(url+"/"+roomExceptionURL[5:])


  if 'data' in response.json():
    j1 = response.json()['data']
  else:
    print("ERROR! Data was not in the response for:", "room")
    print("URL:"+url+"/"+roomExceptionURL[5:] )
    j1 = response.json()



  verifyOutput(j1, someval)
  response.close()
  print()
  print()



rackExceptionURL = "/api/racks/"+PIDS["rackID"]+"/device-sensors"
someval= getEntitiesOfAncestor(sToI("ROOM"), "room",PIDS["roomID"], "rack_sensor")

response = requests.request("GET", url+"/"+rackExceptionURL[5:], headers=headers, data={})
b = verifyGet(response.status_code, "Rack call for devices\'s sensors")
if b != True:
  print("Failure @URL: ")
  print(url+"/"+rackExceptionURL[5:])


  if 'data' in response.json():
    j1 = response.json()['data']
  else:
    print("ERROR! Data was not in the response for:", "rack")
    print("URL:"+url+"/"+rackExceptionURL[5:] )
    j1 = response.json()



  verifyOutput(j1, someval)
  response.close()
  print()
  print()



if res == True:
  print("Test Script Successful")
else:
  print("Test Script Failure")

