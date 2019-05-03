#! /bin/bash

set -e

docker-compose run -d --rm --name etcd --service-ports etcd > /dev/null

sleep 1

make integration-test

docker rm -f etcd > /dev/null
