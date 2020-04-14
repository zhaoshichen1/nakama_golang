#!/bin/sh
mkdir -p dist
go build -buildmode=plugin -o ./dist/libnakama.so  
