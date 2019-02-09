#!/bin/bash

docker build -t nextload-core -f env/docker/Dockerfile .
docker-compose -f env/docker/docker-compose.yml up --build