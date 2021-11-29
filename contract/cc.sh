#!/bin/bash

set -x

# 설치 1 -> cli -> peer0.org1.shareshares.com
docker exec cli peer chaincode install -n shareshares -v 1.0 -p github.com/
# linux /home/bstudent/fabric-samples/chaincode/shareshares -> cli /opt/gopath/src/github.com/shareshares

docker exec  cli peer chaincode list --installed # 설치된 체인코드 쿼리 -> ID부여된 설치 체인코드이름 버전

# 배포 peer0.org1.shareshares.com  -> dev-shareshares 인도서 피어 컨테이너가 생성, 커미터 피어 couchdb mychannel_papercontract 테이블이생성
docker exec cli peer chaincode instantiate -n shareshares -v 1.0 -C mychannel -c '{"Args":["portfolio","10000000"]}' -P 'AND ("Org1MSP.member")' #
# 체인코드 같은 이름으로 배포 -> upgrade

sleep 3

# 배포 확인 
docker exec cli peer chaincode list --instantiated -C mychannel

# peer0.org1.shareshares.com invoke -> ws putstate -> block 생성 
docker exec cli peer chaincode invoke -n shareshares -C mychannel -c '{"Args":["putTrading", "2021-11-26", "005930.KS", "72100", "100"]}'  --peerAddresses peer0.org1.shareshares.com:7051 
sleep 3

docker exec cli peer chaincode invoke -n shareshares -C mychannel -c '{"Args":["putTrading", "2021-11-26", "005930.KS", "80000", "-70"]}'  --peerAddresses peer0.org1.shareshares.com:7051 
sleep 3

# 넥소 연료탱크용량 (kg/ℓ)	6.33 / 156.6
# cid, id, amount, date, place, price
docker exec cli peer chaincode invoke -n shareshares -C mychannel -c '{"Args":["getHoldShare","trading"]}'  --peerAddresses peer0.org1.shareshares.com:7051 
sleep 3

docker exec cli peer chaincode invoke -n shareshares -C mychannel -c '{"Args":["getHoldShare","portfolio"]}'  --peerAddresses peer0.org1.shareshares.com:7051 
sleep 3

docker exec cli peer chaincode invoke -n shareshares -C mychannel -c '{"Args":["getHoldShare", "005930.KS"]}'  --peerAddresses peer0.org1.shareshares.com:7051
sleep 3

docker exec cli peer chaincode query -n shareshares -C mychannel -c '{"Args":["getHistoryForShare", "005930.KS"]}'  --peerAddresses peer0.org1.shareshares.com:7051 

# peer0.org1.shareshares.com query -> ws getstate -> block 생성 되지않음


# docker exec cli peer chaincode query -n shareshares -C mychannel -c '{"Args":["history","U101"]}'
# docker exec cli peer chaincode query -n shareshares -C mychannel -c '{"Args":["history","ManagerIDKey"]}'
# docker exec cli peer chaincode query -n shareshares -C mychannel -c '{"Args":["history","TotalSupplyKey"]}'

