#!/bin/bash

echo start building
echo task 1
env GOOS=linux GOARCH=amd64 go build -mod=vendor -o fast_writer main.go
scp fast_writer root@xxx:/xxx/fast_writer


rm -f fast_writer
echo build finished%