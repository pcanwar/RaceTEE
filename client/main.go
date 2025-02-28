package main

import (
	"client/deploy"
	"client/help"
	"client/operation"
	pb "client/proto"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/protobuf/proto"
)

const golangProgPath = "./userpackage/goTest/goTest.go"
const solidityProgPath = "./artifacts/UserContract.json"
const solidityProg2Path = "./artifacts/UserContract2.json"
const ERC20Path = "./artifacts/ERC20.json"
const DEXPath = "./artifacts/DEX.json"
const quickSelectPath = "./artifacts/QuickSelect.json"
const SPAPath = "./artifacts/SecondPriceAuction.json"
const kMeanProgPath = "./userpackage/KMean/KMean.go"
const calProgPath = "./artifacts/Calculate.json"

var mainAccountIndex = 7

var defaultConfig = pb.UserConfig{
	HistoryKeyDiscard: true,
	KeyRotation:       200,
	ACL:               []string{}, // empty ACL means anyone can execute the program
}
var timeInterval int
var blockInterval = 1

// ./client -mode d -case g
func main() {
	// Parse command line arguments
	var userCase string
	var mode string
	var adr1 string
	var adr2 string
	var adr3 string
	var i int
	var userIndex int
	flag.StringVar(&mode, "mode", "c", "User mode: e(executor), d(deployer) or c(combined)")
	flag.StringVar(&userCase, "case", "s", "Use cases: g(golang test), kmean(KMean), s(solidity test), i(interact test), erc20(erc20), dex(dex), quick(quick select), SPA(second price auction), or cal(calculation)")
	flag.StringVar(&adr1, "adr1", "", "Address of the first program")
	flag.StringVar(&adr2, "adr2", "", "Address of the second program")
	flag.StringVar(&adr3, "adr3", "", "Address of the third program")
	flag.IntVar(&i, "i", 2000000, "Execution interval(in microseconds)")
	flag.IntVar(&userIndex, "userIndex", 0, "User index")
	flag.Parse()

	timeInterval = i
	mainAccountIndex += userIndex
	switch mode {
	case "e":
		switch userCase {
		case "g":
			runGolang(common.HexToAddress(adr1))
		case "s":
			runSolidity(common.HexToAddress(adr1))
		case "kmean":
			runKMean(common.HexToAddress(adr1))
		case "i":
			runInteract(common.HexToAddress(adr1), common.HexToAddress(adr2))
		case "erc20":
			runERC20(common.HexToAddress(adr1))
		case "dex":
			tokenAddrs := []common.Address{common.HexToAddress(adr2), common.HexToAddress(adr3)}
			runDEX(common.HexToAddress(adr1), tokenAddrs)
		case "quick":
			runQuickSelect(common.HexToAddress(adr1))
		case "SPA":
			runSPA(common.HexToAddress(adr1), common.HexToAddress(adr2))
		case "cal":
			runCal(common.HexToAddress(adr1))
		}
	case "d":
		switch userCase {
		case "g":
			adr1 = deployGolang().Hex()
		case "kmean":
			adr1 = deployKMean().Hex()
		case "s":
			adr1 = deploySolidity().Hex()
		case "i":
			_adr1, _adr2 := deployInteract()
			adr1 = _adr1.Hex()
			adr2 = _adr2.Hex()
		case "erc20":
			adr1 = deployERC20().Hex()
		case "dex":
			_adr1, tokenAddrs := deployDEX()
			adr1 = _adr1.Hex()
			adr2 = tokenAddrs[0].Hex()
			adr3 = tokenAddrs[1].Hex()
		case "quick":
			adr1 = deployQuickSelect().Hex()
		case "SPA":
			_adr1, _adr2 := deploySPA()
			adr1 = _adr1.Hex()
			adr2 = _adr2.Hex()
		case "cal":
			adr1 = deployCal().Hex()
		}
	case "c":
		switch userCase {
		case "g":
			PRGAddress := deployGolang()
			runGolang(PRGAddress)
		case "kmean":
			PRGAddress := deployKMean()
			runKMean(PRGAddress)
		case "s":
			PRGAddress := deploySolidity()
			runSolidity(PRGAddress)
		case "i":
			PRGAddress, PRGAddress2 := deployInteract()
			runInteract(PRGAddress, PRGAddress2)
		case "erc20":
			PRGAddress := deployERC20()
			runERC20(PRGAddress)
		case "dex":
			PRGAddress, tokenAddrs := deployDEX()
			runDEX(PRGAddress, tokenAddrs)
		case "quick":
			PRGAddress := deployQuickSelect()
			runQuickSelect(PRGAddress)
		case "SPA":
			PRGAddress, tokenAddr := deploySPA()
			runSPA(PRGAddress, tokenAddr)
		case "cal":
			PRGAddress := deployCal()
			runCal(PRGAddress)
		}
	}
}

func deployGolang() common.Address {
	code := help.LoadGolangCode(golangProgPath)
	PRGAddress := deploy.DeployProgramme([]byte(code), mainAccountIndex, defaultConfig)
	return common.HexToAddress(PRGAddress)
}

func runGolang(PRGAddress common.Address) {
	go operation.Result(PRGAddress)

	// Create a ticker
	ticker := time.NewTicker(time.Duration(timeInterval) * time.Microsecond)
	defer ticker.Stop()
	// Run an infinite loop
	for {
		select {
		case <-ticker.C:
			input := `{"funcName": "Add", "args": 1}`
			operation.Execute(PRGAddress, mainAccountIndex, []byte(input))
		}
	}
}

func deploySolidity() common.Address {
	code := help.LoadBytecode(solidityProgPath)
	address := deploy.DeployProgramme(code, mainAccountIndex, defaultConfig)
	return common.HexToAddress(address)
}

func runSolidity(PRGAddress common.Address) {
	go operation.Result(PRGAddress)

	// Create a ticker
	ticker := time.NewTicker(time.Duration(timeInterval) * time.Microsecond)
	defer ticker.Stop()

	ParsedUserContractABI := help.LoadABI(solidityProgPath)
	for {
		select {
		case <-ticker.C:
			input, err := ParsedUserContractABI.Pack("add", big.NewInt(1))
			if err != nil {
				panic(err)
			}
			operation.Execute(PRGAddress, mainAccountIndex, input)
		}
	}
}

func deployInteract() (common.Address, common.Address) {
	code := help.LoadBytecode(solidityProgPath)
	PRGAddress := deploy.DeployProgramme(code, mainAccountIndex, defaultConfig)

	// wait for the block to be mined
	time.Sleep(time.Duration(blockInterval) * time.Second)

	code2 := help.LoadBytecode(solidityProg2Path)
	// generate constructor arguments
	prog2ABI := help.LoadABI(solidityProg2Path)
	prog2ConstructorInput, err := prog2ABI.Pack("", common.HexToAddress(PRGAddress))
	if err != nil {
		fmt.Printf("Failed to pack prog2 constructor function call: %v", err)
	}
	PRGAddress2 := deploy.DeployProgramme(append(code2, prog2ConstructorInput...), mainAccountIndex, defaultConfig)

	// wait for the block to be mined
	time.Sleep(time.Duration(blockInterval) * time.Second)

	return common.HexToAddress(PRGAddress), common.HexToAddress(PRGAddress2)
}

func runInteract(PRGAddress common.Address, PRGAddress2 common.Address) {
	go operation.Result(PRGAddress)
	go operation.Result(PRGAddress2)

	ticker := time.NewTicker(time.Duration(timeInterval) * time.Microsecond)
	defer ticker.Stop()

	progABI := help.LoadABI(solidityProgPath)
	prog2ABI := help.LoadABI(solidityProg2Path)
	index := 0
	for {
		select {
		case <-ticker.C:
			if index%2 == 0 {
				input, err := prog2ABI.Pack("add", big.NewInt(2))
				if err != nil {
					panic(err)
				}
				operation.Execute(PRGAddress2, mainAccountIndex, input)
			} else {
				input, err := progABI.Pack("multiply", big.NewInt(1))
				if err != nil {
					panic(err)
				}
				operation.Execute(PRGAddress, mainAccountIndex, input)
			}
			index++
		}
	}
}

func deployERC20() common.Address {
	return _deployERC20("TestToken", "TT")
}

func _deployERC20(name string, symbol string) common.Address {
	code := help.LoadBytecode(ERC20Path)
	// generate constructor arguments
	progABI := help.LoadABI(ERC20Path)
	params := []interface{}{name, symbol, uint8(18), big.NewInt(9000000000000000000)}
	progConstructorInput, err := progABI.Pack("", params...)
	if err != nil {
		fmt.Printf("Failed to pack ERC20 constructor function call: %v", err)
	}
	address := deploy.DeployProgramme(append(code, progConstructorInput...), mainAccountIndex, defaultConfig)
	// wait for the block to be mined
	time.Sleep(time.Duration(blockInterval) * time.Second)
	return common.HexToAddress(address)
}

func runERC20(PRGAddress common.Address) {
	go operation.Result(PRGAddress)

	// Create a ticker
	ticker := time.NewTicker(time.Duration(timeInterval) * time.Microsecond)
	defer ticker.Stop()

	ParsedERC20ABI := help.LoadABI(ERC20Path)
	reciverAddress := common.HexToAddress(help.Accounts[11].Address)
	index := 0
	// Run an infinite loop
	for {
		select {
		case <-ticker.C:
			params := []interface{}{reciverAddress, big.NewInt(100000)}
			input, err := ParsedERC20ABI.Pack("transfer", params...)
			if err != nil {
				panic(err)
			}
			operation.Execute(PRGAddress, mainAccountIndex, input)
			index++
		}

	}
}

func deployDEX() (common.Address, []common.Address) {
	// deploy the tokens
	TokenA := _deployERC20("TokenA", "A")
	TokenB := _deployERC20("TokenB", "B")

	// deploy the Dex
	code := help.LoadBytecode(DEXPath)
	// generate constructor arguments
	progABI := help.LoadABI(DEXPath)
	progConstructorInput, err := progABI.Pack("", TokenA, TokenB)
	if err != nil {
		fmt.Printf("Failed to pack DEX constructor function call: %v", err)
	}
	address := deploy.DeployProgramme(append(code, progConstructorInput...), mainAccountIndex, defaultConfig)

	// wait for the block to be mined
	time.Sleep(time.Duration(blockInterval) * time.Second)

	return common.HexToAddress(address), []common.Address{TokenA, TokenB}

}

func runDEX(dexAddr common.Address, tokenAddrs []common.Address) {
	go operation.Result(dexAddr)
	go operation.Result(tokenAddrs[0])
	go operation.Result(tokenAddrs[1])

	// Create a ticker
	ticker := time.NewTicker(time.Duration(timeInterval) * time.Microsecond)
	defer ticker.Stop()

	DEXABI := help.LoadABI(DEXPath)
	ParsedERC20ABI := help.LoadABI(ERC20Path)
	index := 0
	// Run an infinite loop
	for {
		select {
		case <-ticker.C:
			if index == 0 {
				// approve A
				approveInput, err := ParsedERC20ABI.Pack("approve", dexAddr, big.NewInt(90000000))
				if err != nil {
					panic(err)
				}
				operation.Execute(tokenAddrs[0], mainAccountIndex, approveInput)

				// approve B
				approveInputB, err := ParsedERC20ABI.Pack("approve", dexAddr, big.NewInt(90000000))
				if err != nil {
					panic(err)
				}
				operation.Execute(tokenAddrs[1], mainAccountIndex, approveInputB)
				// Check the balance of the tokens B
				input, err := ParsedERC20ABI.Pack("balances", dexAddr)
				if err != nil {
					panic(err)
				}
				operation.Execute(tokenAddrs[1], mainAccountIndex, input)
			} else if index == 1 {
				// prepare the liquidity
				params := []interface{}{big.NewInt(40000000), big.NewInt(40000000)}
				input, err := DEXABI.Pack("addLiquidity", params...)
				if err != nil {
					panic(err)
				}
				operation.Execute(dexAddr, mainAccountIndex, input)
			} else {
				// Change A to B
				params := []interface{}{big.NewInt(10)}
				input, err := DEXABI.Pack("swap", params...)
				if err != nil {
					panic(err)
				}
				operation.Execute(dexAddr, mainAccountIndex, input)
			}
			index++
		}
	}
}

func deployQuickSelect() common.Address {
	code := help.LoadBytecode(quickSelectPath)
	address := deploy.DeployProgramme(code, mainAccountIndex, defaultConfig)

	// wait for the block to be mined
	time.Sleep(time.Duration(blockInterval) * time.Second)

	return common.HexToAddress(address)
}

func readRandomNumbers(file string) []*big.Int {
	// read the random numbers from the json file
	fs, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer fs.Close()
	decoder := json.NewDecoder(fs)
	var randomNumbers []*big.Int
	err = decoder.Decode(&randomNumbers)
	if err != nil {
		panic(err)
	}
	return randomNumbers
}

func runQuickSelect(PRGAddress common.Address) {
	go operation.Result(PRGAddress)

	// Create a ticker
	ticker := time.NewTicker(time.Duration(timeInterval) * time.Microsecond)
	defer ticker.Stop()

	QuickSortABI := help.LoadABI(quickSelectPath)
	arr := readRandomNumbers("./testData/random_numbers.json")
	params := []interface{}{arr, big.NewInt(1000)}
	// Run an infinite loop
	index := 0
	for {
		select {
		case <-ticker.C:
			input, err := QuickSortABI.Pack("quickSelect", params...)
			if err != nil {
				panic(err)
			}
			operation.Execute(PRGAddress, mainAccountIndex, input)
			index++
		}
	}
}

func deploySPA() (common.Address, common.Address) {
	// Deploy the ERC20 token for the auction.
	auctionToken := _deployERC20("AuctionToken", "AT")

	// Pack constructor arguments and deploy the SPA contract.
	code := help.LoadBytecode(SPAPath)
	spaABI := help.LoadABI(SPAPath)
	biddingTime := big.NewInt(12000000000) // set longer bidding time
	constructorInput, err := spaABI.Pack("", auctionToken, biddingTime)
	if err != nil {
		fmt.Printf("Failed to pack SPA constructor: %v\n", err)
	}
	deployedAddr := deploy.DeployProgramme(append(code, constructorInput...), mainAccountIndex, defaultConfig)

	// wait for the block to be mined
	time.Sleep(time.Duration(blockInterval) * time.Second)

	return common.HexToAddress(deployedAddr), auctionToken
}

func runSPA(prgAddress common.Address, tokenAddr common.Address) {
	go operation.Result(prgAddress)

	// approve the SPA contract to transfer tokens
	spaABI := help.LoadABI(SPAPath)
	erc20ABI := help.LoadABI(ERC20Path)
	approveData, err := erc20ABI.Pack("approve", prgAddress, big.NewInt(1000000000))
	if err != nil {
		panic(err)
	}
	operation.Execute(tokenAddr, mainAccountIndex, approveData)

	// wait for the block to be mined
	time.Sleep(time.Duration(blockInterval) * time.Second)

	ticker := time.NewTicker(time.Duration(timeInterval) * time.Microsecond)
	defer ticker.Stop()

	index := 0
	for {
		select {
		case <-ticker.C:
			// place a bid
			bidAmount := big.NewInt(int64((index + 1) * 10))
			bidData, err := spaABI.Pack("placeBid", bidAmount)
			if err != nil {
				panic(err)
			}
			operation.Execute(prgAddress, mainAccountIndex, bidData)
			index++
		}
	}
}

func deployKMean() common.Address {
	code := help.LoadGolangCode(kMeanProgPath)
	PRGAddress := deploy.DeployProgramme([]byte(code), mainAccountIndex, defaultConfig)

	// wait for the block to be mined
	time.Sleep(time.Duration(blockInterval) * time.Second)

	return common.HexToAddress(PRGAddress)
}

// generatePointsString converts a slice of float64 slices (points)
// into an expression string like: "[]Point{{10.0, 20.0, 30.0}, {...}, ...}".
func generatePointsString(points [][]float64) string {
	var sb strings.Builder
	sb.WriteString("[]Point{")
	for i, pt := range points {
		sb.WriteString("{")
		for j, v := range pt {
			sb.WriteString(fmt.Sprintf("%v", v))
			if j < len(pt)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteString("}")
		if i < len(points)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("}")
	return sb.String()
}

func runKMean(PRGAddress common.Address) {
	go operation.Result(PRGAddress)

	// Generate a random dataset: 1000 points, each with 50 dimensions in range [0, 100).
	src := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(src)
	numPoints := 1000
	dim := 5
	data := make([][]float64, numPoints)
	for i := 0; i < numPoints; i++ {
		point := make([]float64, dim)
		for j := 0; j < dim; j++ {
			point[j] = rng.Float64() * 100
		}
		data[i] = point
	}

	// Dynamically generate the data expression.
	dataStr := generatePointsString(data)
	// Number of clusters.
	k := 10
	// Maximum iterations.
	maxIter := 100

	// Construct the args expression without extra quotes.
	argsExpr := fmt.Sprintf("%s, %d, %d", dataStr, k, maxIter)
	_input := pb.GolangInput{
		FuncName: "KMeans",
		Args:     []byte(argsExpr),
	}
	input, err := proto.Marshal(&_input)
	if err != nil {
		panic(err)
	}

	// Create a ticker to trigger periodic execution.
	ticker := time.NewTicker(time.Duration(timeInterval) * time.Microsecond)
	defer ticker.Stop()

	// Periodically execute the operation.
	i := 0
	for {
		select {
		case <-ticker.C:
			operation.Execute(PRGAddress, mainAccountIndex, []byte(input))
			i++
		}
	}
}

func deployCal() common.Address {
	// deploy the program
	code := help.LoadBytecode(calProgPath)
	PRGAddress := deploy.DeployProgramme([]byte(code), mainAccountIndex, defaultConfig)

	// wait for the block to be mined
	time.Sleep(time.Duration(blockInterval) * time.Second)

	return common.HexToAddress(PRGAddress)
}

func runCal(PRGAddress common.Address) {
	go operation.Result(PRGAddress)

	// Create a ticker
	ticker := time.NewTicker(time.Duration(timeInterval) * time.Microsecond)
	defer ticker.Stop()

	CalABI := help.LoadABI(calProgPath)
	index := 1
	for {
		select {
		case <-ticker.C:
			// This block will be executed periodically
			params := []interface{}{big.NewInt(100000)}
			input, err := CalABI.Pack("cal", params...)
			if err != nil {
				panic(err)
			}
			operation.Execute(PRGAddress, mainAccountIndex, input)
			index++
		}
	}
}
