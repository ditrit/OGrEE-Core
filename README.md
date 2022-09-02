Introduction
------------
This is a Shell interfacing with a RESTful API for data centre management.


Building
------------
This is not yet tested on Windows and macOS   
The ReadMe assumes that you already have the latest version of GO installed and your GO Environment PATHS properly setup  
You must also have python3 setup and installed 
For BSD systems, GO can be installed from the respective ports  
For Linux, consult your respective Distribution docs  

[Otherwise you can follow instructions from the GO site](https://golang.org/doc/install)  

   
  Clone the API repository  
  Execute make. It should automatically retrieve the necessary libraries. If not then execute the commands below 
  ```
  go get github.com/blynn/nex
  go get -u github.com/cznic/goyacc
  ```  

    make


Running
-------------
 - Execute ```./main```

Usage & Notes
-------------
Sometimes the shell can crash with a 'stream error'. This is assumed to be a problem with the API Backend since it occurs on startup, when the Shell queries a lot of data
 
 The following can be done in the shell:
 - cd       : Change directory
 - pwd      : Print working directory
 - ls       : Display contents of directory
 - clear    : Clear the terminal screen
 - exit     : Quit program
 - grep     : Make a search on stream (NOT YET IMPLEMENTED)
 - man      : Quick introduction to the program
 - create   : Create objects
 - gt       : Get Object details
 - update   : Update an object
 - delete   : Delete an object

There is no advanced shell capabilities (scripting etc.) implemented yet.

Anatomy
-------------
The Shell follows the [MVC Architecture](https://en.wikipedia.org/wiki/Model%E2%80%93view%E2%80%93controller). 
The code is divided into isolated components where each component performs a certain set of tasks. The Model manages the communication with the API.
The Controller interacts with the View and Model   
The View is the front end that manages the user input   
<p align="center">
  <img src="https://upload.wikimedia.org/wikipedia/commons/thumb/a/a0/MVC-Process.svg/218px-MVC-Process.svg.png">
</p>


   
This design was chosen to follow the architecture of the API
   
API Files
-------------
   
### Folder Structure   
```  
├─controllers 
├─interpreter  
├─models 
├─readline   
├─other   
└─utils   
```
    

The 'controllers' dir contains controller files 
The 'interpreter' dir contains files to build the lexer and parser 
The 'models' dir contains model files  
The 'readline' dir contains files for managing the readline and terminal
The 'utils' dir contains useful functions for JSON messaging

### Files of Interest in the root directory  
```
- makefile   
- main.go    
```

   
The makefile is a build directive for the shell. You may need to modify it depending on your environment   
main.go is the entry of the Shell  





### Shell Architecture
Diagrams will be added in the future   


