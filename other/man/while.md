USAGE: while (test) body done   
Execute command(s) repeatedly as long as a condition is met   

The  while command evaluates test as a boolean.   
If it is true  value  then body  is  executed.   
Once body has been executed then test is evaluated again,
and the process repeats until eventually test evaluates to a false boolean value.
The while command for now returns NULL.   

EXAMPLE   

    .var:x=0
    while ($x < 5) pwd; $x=$x+1; done 