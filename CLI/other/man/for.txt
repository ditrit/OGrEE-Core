USAGE: for var in INTEGER..INTEGER {commands;}    
    
   
For  is  a  looping command, the interpreter will execute init once.   
Then it evaluates condition as a bool expression. It will execute the commands and increment. Then it repeats again until the condition evaluates to false

   
NOTE   

    No ranged loops constructs are supported at this time
    

EXAMPLE   

    
    for x in 2..10 {pwd; tree; print $x}   
   