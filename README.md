Introduction
------------
This is a Shell interfacing with a RESTful API for data centre management.


Building
------------   

  ### NOTE If Building On Windows
  The current method to build is to use ```wsl``` and with this we can invoke ```make win``` to generate a windows build.

  To install wsl please follow the [instructions here](https://learn.microsoft.com/en-us/windows/wsl/install)
  
  Launch wsl and navigate to the directory containing the CLI and continue to follow the instructions below. 

  If you want to generate a windows binary you must execute ```make win```

### Build Instructions
The ReadMe assumes that you already have the latest version of GO installed and your GO Environment PATHS properly setup  
You must also have python3 installed and properly setup  
For BSD systems, GO can be installed from the respective ports  
For Linux, consult your respective Distribution docs  

[Otherwise you can follow instructions from the GO site](https://golang.org/doc/install)  

   
  Clone the CLI repository  
  Execute make. It should automatically retrieve the necessary libraries. If not then execute the commands below 
  ```
  go get github.com/blynn/nex
  go get -u github.com/cznic/goyacc
  ```  

    make




Running
-------------
You must first have a ```.env``` file to be included in the same directory as the executable. If not included the CLI will exit on startup with an error message about this file.   

You can view an example ```.env``` file here: https://ogree.ditrit.io/htmls/clienv.html  

 - Execute ```./main```

If this is the first running the Shell, you will be greeted with a sign up prompt to input a user email and password. 

DO NOT SHARE YOUR ```.env``` file since it contains your credentials 

Usage & Notes
-------------
Please read the more comprehensive and updated how to use guide here: https://ogree.ditrit.io/htmls/programming.html   

Sometimes the shell can crash with a 'stream error'. This is assumed to be a problem with the communication library, when the Shell queries a lot of data
 
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


   
This design was chosen to follow the architecture of the CLI
   
CLI Files
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


