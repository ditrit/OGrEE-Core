OGrEE handles a lot of data of different types and sources to create a datacenter digital twin. This data is organized in a certain structure and the access to it is controlled, as explained below.

## Data Structure
At top level, all of OGrEE's data is divided in 3 **namespaces**: Physical, Logical and Organisational. 

### Physical
It's under Physical that we will find everything that occupies a physical space in the datacenter, such as rooms, racks, servers, sensors, air conditioners, etc. A hierarchy of objects is defined:
```
├── Site
    ├── Building
        ├── Room
            ├── Corridor
            ├── Generic
            ├── Rack
                ├── Device
                    ├── Device [...]
└── Stray Object
```
The root of Physical is one or multiple sites. A site can only have buildings. A building, only rooms. Under a room, we can create racks or generic objects. This last one is highly customizable, can be used to represent ACs, power panels, tables, chairs and much more. A rack can also have multiple devices, as a rack with multiple servers, for example. Each device can always have multiple devices, representing the multiple components (processors, slots, disks, ports) a server can have, for example. Not directly linked to this hierarchy, we can also have stray objects, this is pratical to, for example, temporarily represent a server that is not currently attached to a rack or is being moved.

### Logical
Here is where we find templates and grouping entities as well as virtual objects. 

**Templates** are useful to quickly create physical objects that share the same properties. There are different types of templates: building, room, rack, device and generic template. This last one is used to create physical objects of type Generic. A device template, for example, can define the properties of a specific model of server, making it easy to create those servers as devices in the Physical namespace. More information about templates can be found [here](https://github.com/ditrit/OGrEE-Core/wiki/%F0%9F%93%97-%5BUser-Guide%5D-API-%E2%80%90-JSON-templates-definitions).

**Grouping entities** are tags, layers and groups. Every object from the Physical namespace can have one or more tags associated to it. This can help filtering and search. Groups can be created only under rooms or racks and need a pre-determined list of which of their children while layers can be applied everywhere in the physical namespace and only need a filtering expression. More information about it can be found [here](https://github.com/ditrit/OGrEE-Core/wiki/%F0%9F%93%97-%5BUser-Guide%5D-CLI-%E2%80%90-Language#layers). 

**Virtual Objects** (vobjs) are logical elements in the datacenter. It can be used to represent virtual machines, kubernetes clusters, docker containers, logical volumes, virtual switches, etc. They can have a device or another vobj as parent, or even no parent at all. A vobj can have **vlinks**, those are virtual links, a way to "point" towards other devices. For example, a virtual object can be a network bond that has two vlinks, each pointing to a physical interface in a physical device. 

```
├── Physical device
    ├── Virtual object network bond
        ├── vlink to -> Physical device interface0
        └── vlink to -> Physical device interface1
    ├── Virtual object remote storage
        └── vlink to -> Far away physical device disk
```

### Organisational
This namespace contains domains which are also organized in a hierarchical manner. A domain can have multiple domain children and are used for access control. All objects from the Physical namespace have a domain and can only be seen or managed by a user that has the right permission in such domain. More information below.
```
├── DomainA
    ├── SubDomainA1
        ├── SubDomainA2
├── DomainB
    ├── SubDomainB1
└── DomainC
```

## Access Control (RBAC)
The OGrEE-API implements a Role-based access control (RBAC). Each user has one or multiple domains with a level of permission attached to each domain. The default user created `admin`, automatically created with each new tenant, has all permissions on all domains (changing its password is **highly** recommended). You can create and manage users with [SuperAdmin APP](https://github.com/ditrit/OGrEE-Core/wiki/%F0%9F%93%97-%5BUser-Guide%5D-APP-%E2%80%90-SuperAdmin) in the Users tab under the tenant's 🔍 info page.

**Users permission:** 
One of the following roles should be assigned for each domain assigned to an user:
- _viewer_ : read-only access, can only see the objects from its domain environment, but cannot create/modify/delete them. In http words : only GET requests, no POST/PUT/PATCH/DELETE.
- _user_ : can read and write within its domain environment. All possible http requests whitin its domain.
- _manager_ : same as user + can create domains and users within its domain environment. A _super_ user would be a user with manager role to the root domain "*" (the default `admin` user).

**Domain organization:** 
Domains are organized in a hierarchy. If a user has access to the parent domain, he also has the same level of access to the children domains. However, this also implies an automatic read-only acces (_viewer_ role) to all the parent domains of the user domain, but very limited: can only see the objects' names (hierarchyNames). Examples:
- John has _user_ role in domain **A.B.C**. Concerning objects, John can:

| Object's domain | Read | Create/Modify/Delete | Users/Domains |
|-----------------|------|----------------------|---------------|
| A               | YES (only name)  | NO                   | NO            |
| A.B             | YES (only name)  | NO                   | NO            |
| **A.B.C** (John)    | YES  | YES                  | NO            |
| **A.B.C**.D         | YES  | YES                  | NO            |
| A.B.Z           | NO   | NO                   | NO            |
| A.Y             | NO   | NO                   | NO            |

- Now, if John was _manager_ of domain **A.B.C**:

| Object's domain | Read | Create/Modify/Delete | Users/Domains |
|-----------------|------|----------------------|---------------|
| A               | YES (only name)  | NO                   | NO            |
| A.B             | YES (only name)  | NO                   | NO            |
| **A.B.C** (John)    | YES  | YES                  | YES           |
| **A.B.C**.D         | YES  | YES                  | YES           |
| A.B.Z           | NO   | NO                   | NO            |
| A.Y             | NO   | NO                   | NO            |

- Finally, as _viewer_ of domain **A.B.C**:

| Object's domain | Read | Create/Modify/Delete | Users/Domains |
|-----------------|------|----------------------|---------------|
| A               | YES (only name)  | NO                   | NO            |
| A.B             | YES (only name)  | NO                   | NO            |
| **A.B.C** (John)    | YES  | NO                   | NO            |
| **A.B.C**.D         | YES  | NO                   | NO            |
| A.B.Z           | NO   | NO                   | NO            |
| A.Y             | NO   | NO                   | NO            |

All objects have a single domain. The parent object must always belong to the same domain or a parent domain of the child object. Examples:
- Rack domain = A.B    -> Device domain = A.B.C :heavy_check_mark:
- Rack domain = A.B.C -> Device domain = A.B.C :heavy_check_mark:
- Rack domain = A.B.Z -> Device domain = A.B.C :x: