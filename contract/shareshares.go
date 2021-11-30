package main

import ("encoding/json"
	"fmt"
	"strconv"
	"bytes"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

type SmartContract struct{
}

type Portfolio struct {
	Cash int `json:"Cash"`
	HoldShare []HoldShare `json:"HoldShare"`
	Count int `json:"Count"`
}

type HoldShare struct {

	StartDate	string `json:"StartDate"`
	EndDate		string `json:"EndDate"`
	ShareCode 	string `json:"ShareCode"`
	Amount  	int `json:"Amount"`
	ChangeAmount int `json:"ChangeAmount"`
	Avg_price 	int `json:"Avg_price"` 	
	RecentPrice int `json:"RecentPrice"`	
}

//임시(데이터베이스를 쓸 줄 몰라서....)
type Trading struct {
	Date string `json:"Date"`
	ShareCode string `json:"ShareCode"`
	Price int `json:"Price"`
	Weight float64 `json:"Weight"`
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {

	args := APIstub.GetStringArgs()
	if len(args) != 2 {
		return shim.Error("Invalid Arguments. PortfolioName and Cash asset are required")
	}

	portfolio := Portfolio{} 
	portfolio.Count = 0
	portfolio.Cash, _ = strconv.Atoi(args[1])
	portfolioAsBytes, _ := json.Marshal(portfolio)

	APIstub.PutState(args[0], portfolioAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {


	function, args := APIstub.GetFunctionAndParameters()

	if function == "putTrading" {
		return s.putTrading(APIstub, args)
	} else if function == "getHoldShare" {
		return s.getHoldShare(APIstub, args)
	} else if function == "getHistoryForShare" {
		return s.getHistoryForShare(APIstub, args)
	} 

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) putTrading(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	portfolio := Portfolio{}
	portfolioAsBytes, _ := APIstub.GetState("portfolio")
	json.Unmarshal(portfolioAsBytes, &portfolio)

	holdshare := HoldShare{}

	k := 0
	var i int

	for i = 0; i < portfolio.Count; i++ {

		if portfolio.HoldShare[i].ShareCode == args[1]{
			ChangePortfolio(APIstub, args, i, &portfolio, &holdshare)	
		} else {
			k++
		}
	}
	if k == i {
		CreatePortfolio(APIstub, args, &portfolio, &holdshare)
	}

	newTrading(APIstub, args, &portfolio)

	portfolioAsBytes, _ = json.Marshal(portfolio)
	APIstub.PutState("portfolio", portfolioAsBytes)
	
	holdshareAsBytes, _ := json.Marshal(holdshare)
	key := args[1]
	APIstub.PutState(key, holdshareAsBytes)

	return shim.Success(nil)
}

func  (s *SmartContract) getHoldShare(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect arguments. Expecting a key")
	}

	value, err := APIstub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to get Hold share with error")
	}
	if value == nil {
		return shim.Error("HoldShare not found")
	}
	return shim.Success(value)
}

func (s *SmartContract) getHistoryForShare(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	ShareName := args[0]

	fmt.Printf("- start getHistoryForShare: %s\n", ShareName)

	resultsIterator, err := stub.GetHistoryForKey(ShareName)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the marble
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON marble)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryForShare returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func ChangePortfolio(APIstub shim.ChaincodeStubInterface, args []string, i int, portfolio *Portfolio, holdshare *HoldShare) {

	holdshareAsBytes, _ := APIstub.GetState(args[1])
	json.Unmarshal(holdshareAsBytes, &holdshare)

	holdshare.EndDate = args[0]
	holdshare.RecentPrice, _ = strconv.Atoi(args[2])
	holdshare.ChangeAmount, _ = strconv.Atoi(args[3])
	holdshare.Amount = holdshare.Amount + holdshare.ChangeAmount

	if holdshare.Amount == 0 {
				
		portfolio.HoldShare = remove(portfolio.HoldShare, i)
		portfolio.Count--
		portfolio.Cash -= holdshare.ChangeAmount * holdshare.RecentPrice

	}else{
		holdshare.Avg_price = ((holdshare.Amount - holdshare.ChangeAmount) * holdshare.Avg_price + holdshare.RecentPrice * holdshare.ChangeAmount) / holdshare.Amount
		portfolio.HoldShare = remove(portfolio.HoldShare, i)
		portfolio.HoldShare = append(portfolio.HoldShare, *holdshare)
		portfolio.Cash -= holdshare.ChangeAmount * holdshare.RecentPrice
	}
}

func CreatePortfolio(APIstub shim.ChaincodeStubInterface, args []string, portfolio *Portfolio, holdshare *HoldShare) {

	holdshare.StartDate = args[0]
	holdshare.ShareCode = args[1]
	holdshare.Amount, _ = strconv.Atoi(args[3])
	holdshare.Avg_price, _ = strconv.Atoi(args[2])
	holdshare.RecentPrice, _ = strconv.Atoi(args[2])
	holdshare.ChangeAmount, _ = strconv.Atoi(args[3])

	portfolio.HoldShare = append(portfolio.HoldShare, *holdshare)
	portfolio.Count++
	portfolio.Cash -= holdshare.ChangeAmount * holdshare.RecentPrice
}

func getAssetSum(portfolio *Portfolio) int {

	AssetSum := portfolio.Cash

	for i := 0; i < portfolio.Count; i++ {
		AssetSum = AssetSum + portfolio.HoldShare[i].Avg_price * portfolio.HoldShare[i].Amount 
	}

	return AssetSum
}

func newTrading(APIstub shim.ChaincodeStubInterface, args []string, portfolio *Portfolio){

	AssetSum := getAssetSum(portfolio)
	trading := Trading{}

	trading.Date = args[0]
	trading.ShareCode = args[1]

	price, _  := strconv.Atoi(args[2])
	amount, _ := strconv.Atoi(args[3])

	trading.Price = price
	trading.Weight = float64(price) * float64(amount) / float64(AssetSum)

	tradinghistoryAsBytes, _ := json.Marshal(trading)
	APIstub.PutState("trading", tradinghistoryAsBytes)
}

func remove(slice []HoldShare, s int) []HoldShare {
    return append(slice[:s], slice[s+1:]...)
}


func main(){

	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}

}