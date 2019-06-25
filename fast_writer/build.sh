#!/bin/bash

echo start building
echo task 1
env GOOS=linux GOARCH=amd64 go build -mod=vendor -o fast_query main.go
scp fast_query root@xxx:/xxx/fast_query


rm -f fast_query
echo build finished%