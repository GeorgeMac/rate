#! /bin/bash

set -e

docker-compose run -d --rm --name etcd --service-ports etcd

sleep 1

make integration-test

docker rm -f etcd 
