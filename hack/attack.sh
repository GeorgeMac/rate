#! /bin/bash

docker run -it --rm --net rate_rate-cluster \
  -v "`pwd`:/data" \
  rate-attack sh -c 'echo "GET http://rate-one:4040" | vegeta attack -rate=10/5s -duration=3m > /data/result.bin'
