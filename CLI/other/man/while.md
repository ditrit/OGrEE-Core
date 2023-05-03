USAGE: while (condition) {body}    
Execute command(s) repeatedly as long as a condition is met   

The  while command evaluates condition as a boolean.   
If it is true value then body is executed.   
Once body has been executed then condition is evaluated again,
and the process repeats until eventually condition evaluates to a false boolean value.
The while command for now returns NULL.   


NOTE   

    There is no increment syntax supported here yet. So you must increment using the variable declaration syntax (.var:x=$x+1)

    
EXAMPLE   

    .var:x=0
    while ($x < 5) {pwd; .var:x=$x+1;} 