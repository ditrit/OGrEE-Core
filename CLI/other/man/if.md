USAGE: if expr1 {body1} elif expr2 {body2} elif ... else {bodyN}


The  if command evaluates expr1 as an expression.
The value of the expression must be a boolean.
If it is true then body1 is executed. Otherwise expr2 is
evaluated as an expression and if it is true then body2 is executed, and so on.
If no expression evaluates to true then bodyN is executed.
The Else argument is optional.
There may be any number of elif clauses,
including zero.  BodyN may also be omitted as long as else is omitted too.
The  return  value for now is NULL.   

EXAMPLE   

   if 5 < 6 {ls}   
   if 5 < 6 {ls} else {tree}   
   if 5 == 6  {ls} elif 5 == 4 {tree} else {pwd}   