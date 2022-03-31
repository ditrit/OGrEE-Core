USAGE: for ((init; condition; increment)); commands; done   
OR: for var in {INTEGER..INTEGER}; commands; done   
OR: for var in expression; commands; done   

For  is  a  looping command, the interpreter will execute init once.   
Then it evaluates condition as a bool expression. It will execute the commands and increment.
Then it loops again until the condition evaluates to false
In the 2nd for loop type. A variable is created
which will iterate between the provided range.
This will repeatedly execute commands until the iterator reaches end of range    
In the 3rd type, an internal iterator is created and will
iterate the range of the variable or expression given.
For now the return value of a for loop is NULL.   
   
EXAMPLE   


    for ((x=0; x < 5; $x=$x+1)); pwd done    

    for x in {2..10}; pwd; tree done   

    for x in 5; pwd; done   

    .var:p=3; for x in $p; pwd; done   

    for x in len(array); pwd; done   

    for x in $ls; pwd; done