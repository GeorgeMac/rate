#! /bin/bash

echo "Running Attack"

cat << EOF | vegeta attack -rate=10/5s -duration=3m > ./result.bin
GET http://localhost:4040
GET http://localhost:4041
EOF

echo "Plotting Results"

cat result.bin | vegeta plot > plot.html

open plot.html
