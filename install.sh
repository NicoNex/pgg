#!/bin/bash
echo "Installing pgg..."
cp pgg /usr/bin/

echo "Installing pgg manual..."
gzip -c pgg.1 > pgg.1.gz
cp pgg.1.gz /usr/share/man/man1/

echo "Done"
