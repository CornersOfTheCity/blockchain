#!/bin/bash
rm blockchain
rm *.db
rm *.dat

go build -o blockchain *.go
./blockchain