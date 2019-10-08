#!/bin/bash

gzip -c pgg.1 > pgg.1.gz

echo "Installing pgg..."
cp pgg /usr/bin/

echo "Installing pgg manual..."
cp pgg.1.gz /usr/share/man/man1/

