#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#
# Exit on first error, print all commands.
set -ev

# don't rewrite paths for Windows Git Bash users
export MSYS_NO_PATHCONV=1

export CA1_KEY=$(ls crypto-config/peerOrganizations/org1.shareshares.com/ca/ | grep _sk)

docker-compose -f docker-compose.yml down

# docker-compose -> 컨테이터수행 및 net_basic 네트워크 생성
docker-compose -f docker-compose.yml up -d ca.org1.shareshares.com orderer.shareshares.com peer0.org1.shareshares.com cli # couchdb1 couchdb2 couchdb3
docker ps -a
docker network ls
# wait for Hyperledger Fabric to start
# incase of errors when running later commands, issue export FABRIC_START_TIMEOUT=<larger number>
export FABRIC_START_TIMEOUT=10
#echo ${FABRIC_START_TIMEOUT}
sleep ${FABRIC_START_TIMEOUT}

# Create the channel -> mychannel.block cli working dir 복사
docker exec cli peer channel create -o orderer.shareshares.com:7050 -c mychannel -f /etc/hyperledger/configtx/channel.tx
# clie workding dir (/etc/hyperledger/configtx/) mychannel.block

# Join peer0.org1.shareshares.com to the channel.
docker exec -e "CORE_PEER_LOCALMSPID=Org1MSP" -e "CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/msp/users/Admin@org1.shareshares.com/msp" peer0.org1.shareshares.com peer channel join -b /etc/hyperledger/configtx/mychannel.block

sleep 5

# anchor ORG1 mychannel update
docker exec -e "CORE_PEER_LOCALMSPID=Org1MSP" -e "CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/msp/users/Admin@org1.shareshares.com/msp" peer0.org1.shareshares.com peer channel update -f /etc/hyperledger/configtx/Org1MSPanchors.tx -c mychannel -o orderer.shareshares.com:7050