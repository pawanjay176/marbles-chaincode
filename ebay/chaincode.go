package main

import (
	"errors"
	"fmt"
	"strconv"
	"encoding/json"
	"strings"
	"time"
	"github.com/hyperledger/fabric/core/chaincode/shim"

)

type SimpleChaincode struct {
}

var itemIndexStr = "_itemindex"
var openTradesStr = "_opentrades"

type Item struct{
	Id string `json:"id"`
	Name string `json:"name"`
	Owner string `json:"owner"`
	Price string `json:"price"`
	Purchase_date string `json:"purchase_date"`
	Warranty_validity string `json:"warranty_validity"`
	Review string `json:"review"`
}


// main. Given function. No changes

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	var Aval int
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState("abc", []byte(strconv.Itoa(Aval)))				//making a test var "abc", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}

	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)								//marshal an emtpy array of strings to clear the index
	err = stub.PutState(itemIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}
	
	return nil, nil

}

// ============================================================================================================================
// Run - Our entry point for Invocations - [LEGACY] obc-peer 4/25/2016
// ============================================================================================================================
func (t *SimpleChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}


// ============================================================================================================================
// Run - Our entry point
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)

	// Handle different functions
	if function == "init" {													//initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "delete" {										//deletes an entity from its state
		return t.Delete(stub, args)
	} else if function == "write" {											//writes a value to the chaincode state
		return t.Write(stub, args)
	} else if function == "init_item" {									//create a new marble
		return t.init_item(stub, args)
	} else if function == "set_user" {										//change owner of a marble
		return t.set_user(stub, args)
	}

	fmt.Println("run did not find func: " + function)						//error

	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" {													//read a variable
		return t.read(stub, args)
	} 
	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query")
}



// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) read(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil													//send it onward
}

// ============================================================================================================================
// Delete - remove a key/value pair from state
// ============================================================================================================================
func (t *SimpleChaincode) Delete(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	
	name := args[0]
	err := stub.DelState(name)													//remove the key from chaincode state
	if err != nil {
		return nil, errors.New("Failed to delete state")
	}

	//get the item index
	itemsAsBytes, err := stub.GetState(itemIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get item index")
	}
	var itemIndex []string
	json.Unmarshal(itemsAsBytes, &itemIndex)								//un stringify it aka JSON.parse()
	
	//remove item from index
	for i,val := range itemIndex{
		fmt.Println(strconv.Itoa(i) + " - looking at " + val + " for " + name)
		if val == name{															//find the correct item
			fmt.Println("found item")
			itemIndex = append(itemIndex[:i], itemIndex[i+1:]...)			//remove it
			for x:= range itemIndex{											//debug prints...
				fmt.Println(string(x) + " - " + itemIndex[x])
			}
			break
		}
	}
	jsonAsBytes, _ := json.Marshal(itemIndex)									//save new index
	err = stub.PutState(itemIndexStr, jsonAsBytes)
	return nil, nil
}

// ============================================================================================================================
// Write - write variable into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) Write(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var name, value string // Entities
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
	}

	name = args[0]															//rename for funsies
	value = args[1]
	err = stub.PutState(name, []byte(value))								//write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ============================================================================================================================
// Init item and store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) init_item(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error

	//   0       1       2       3          4
	// id,    name     owner    price    warrantee
	if len(args) <= 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	fmt.Println("- start init marble")
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return nil, errors.New("3rd argument must be a non-empty string")
	}

	id := strings.ToLower(args[0]) //string
	name := strings.ToLower(args[1]) //string
	owner := strings.ToLower(args[2]) // string
	price := strings.ToLower(args[3])
	purchase_date := time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond)) //unix epoch int64
	warranty := strings.ToLower(args[4])

		//check if marble already exists
	marbleAsBytes, err := stub.GetState(id)
	if err != nil {
		return nil, errors.New("Failed to get marble name")
	}
	res := Item{}
	json.Unmarshal(marbleAsBytes, &res)
	if res.Id == id{
		fmt.Println("This marble arleady exists: " + name)
		fmt.Println(res);
		return nil, errors.New("This marble arleady exists")				//all stop a marble by this name exists
	}
	
	str := `{"id": "` + id + `", "name": "` + name + `", "owner": "` + owner + `", "price": "` + price + `", "purchase_date": "` + strconv.FormatInt(purchase_date, 10) + `", "warranty_validity": "` + warranty + `", "review": "`  +`"}`
	// str := `{"name": "` + name + `", "color": "` + color + `", "size": ` + strconv.Itoa(size) + `, "user": "` + user + `"}`
	err = stub.PutState(id, []byte(str))								//store item with id as key
	if err != nil {
		return nil, err
	}
		
	//get the marble index
	itemAsBytes, err := stub.GetState(itemIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get marble index")
	}
	var itemIndex []string
	json.Unmarshal(itemAsBytes, &itemIndex)							//un stringify it aka JSON.parse()
	
	//append
	itemIndex = append(itemIndex, id)								//add item name to index list
	fmt.Println("! item index: ", itemIndex)
	jsonAsBytes, _ := json.Marshal(itemIndex)
	err = stub.PutState(itemIndexStr, jsonAsBytes)						//store name of item

	fmt.Println("- end init marble")
	return nil, nil
}

// ============================================================================================================================
// Set User Permission on Marble
// ============================================================================================================================
func (t *SimpleChaincode) set_user(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error
	
	//   0       1           2 
	// id       newOwner   newPrice
	if len(args) < 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}
	
	fmt.Println("- start set user")
	fmt.Println(args[0] + " - " + args[1])
	itemAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get thing")
	}
	res := Item{}
	json.Unmarshal(itemAsBytes, &res)										//un stringify it aka JSON.parse()
	res.Owner = args[1]
	newPrice, err := strconv.Atoi(args[2])
	res.Price = newPrice														//change the user
	
	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(args[0], jsonAsBytes)								//rewrite the marble with id as key
	if err != nil {
		return nil, err
	}
	
	fmt.Println("- end set user")
	return nil, nil
}

// func (t *SimpleChaincode) repair_item(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
// 	var err error
// 	//   0   1   
// 	//  id   review

// 	itemAsBytes, err := stub.GetState(args[0])
// 	if err != nil {
// 		return nil, errors.New("Failed to get thing")
// 	}
// 	res := Item{}
// 	json.Unmarshal(itemAsBytes, &res)										//un stringify it aka JSON.parse()													
// 	var newReview string
// 	newReview = res.Review+"^"+args[2]
// 	res.Review = newReview
// 	// str := `{"id": "` + res[0] + `", "name": "` + res[1] + `", "owner": ` + res[2] + `, "price": "` + strconv.Itoa(res[3]) + `, "purchase_date": "` + strconv.FormatInt(res[4], 10) + `, "warrantee": "` + strconv.FormatInt(res[5], 10) + `"}`
// 	jsonAsBytes, _ := json.Marshal(res)
// 	err = stub.PutState(args[0], jsonAsBytes)								//rewrite the marble with id as key
// 	if err != nil {
// 		return nil, err
// 	}
	
// 	fmt.Println("- end set user")
// 	return nil, nil



// }