#!/bin/bash

echo "Building pgg..."
go build

echo "Installing pgg..."
sudo cp pgg /usr/bin/

echo "Installing pgg manual..."
gzip -c pgg.1 > pgg.1.gz
sudo cp pgg.1.gz /usr/share/man/man1/

echo "Done!"

