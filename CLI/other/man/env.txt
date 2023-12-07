USAGE:  env [VARIABLE](OPTIONAL) = [VALUE](OPTIONAL)   
Displays and manages environment variables.     



VARIABLES
- Unity   
Indicates to the shell whether or not to inform the Unity viewer of any updates
The value is automatically set upon startup and depends on whether or not the shell was able to establish contact with the Unity viewer. The user would be notified of this.   

- Filter   
Indicates to the shell whether or not to inform the Unity viewer of certain attributes of object updates.     
This variable is set to false by default and must be enabled by the user manually.
 


EXAMPLE   

   env
   env Unity = true
   env Filter = false
