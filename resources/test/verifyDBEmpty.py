#!/usr/bin/env python3
import requests, sys
from pymongo import MongoClient
from bson.objectid import ObjectId
from pprint import pprint

expected = 404

url = "http://localhost:3001/api/"
res = True


#DB SETUP
dbport=27017   
client = MongoClient("localhost", dbport)     
db = client.ogree

#FUNCTIONS
def extractCursor(c):
  arr = [i for i in c]
  return arr

def checkResponse(code, entity):
    global res
    if code == expected:
        collec = entity[:len(entity)-1]
        if collec.find("-") != -1:
          collec = collec.replace("-", "_")
        x = extractCursor(db[collec].find())

        if len(x) == 0:
          print(entity+" Empty: Successful")
        else:
          print("API contradicted the DB!")
          res = False
    else:
        res = False
        print(entity+" Empty: Fail")

entRange=["tenants","sites","buildings","rooms","acs","panels",
  "racks","devices", "panels", "separators", "aisles","tiles",
  "cabinets", "groups", "corridors","room-sensors","rack-sensors",
  "device-sensors","room-templates","obj-templates"]

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
sys.exit()