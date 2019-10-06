#!/bin/bash

gzip pgg.1
go build

echo "Installing pgg..."
cp pgg /usr/bin/

echo "Installing pgg manual..."
cp pgg.1.gz /usr/share/man/man1/

