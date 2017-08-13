#!/bin/bash
WORKING_DIR=`dirname $0`

DOCKER_COMPOSE_FILE=${WORKING_DIR}/int-tests/env/docker-compose.yml
TEST_RESULT=0

echo ======================= Prepare environment =======================
docker-compose --file ${DOCKER_COMPOSE_FILE} down
docker-compose --file ${DOCKER_COMPOSE_FILE} up -d

echo -e "\n======================= Run tests ======================="
go test github.com/nawa/http-ssh-proxy/...
TEST_RESULT=$?

echo -e "\n======================= Destroy environment ======================="
docker-compose --file ${DOCKER_COMPOSE_FILE} down

exit ${TEST_RESULT}