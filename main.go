package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	bscNodeURL       = "https://bsc-dataseed.binance.org/"
	contractAddress  = "0xb2ea51BAa12C461327d12A2069d47b30e680b69D"
	walletAddress    = "0x248Dd3836E2A8B56279C04addC2D11F3c2497836"
	nftBalanceMethod = "balanceOf(address)"
	// API port for server
	apiPort = ":8080"
)

var contractABI = `[{"constant":true,"inputs":[{"name":"owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]`

func getNFTBalance(client *ethclient.Client, contractABI abi.ABI, contractAddr common.Address, walletAddr string) (*big.Int, error) {
	callData, err := contractABI.Pack(nftBalanceMethod, common.HexToAddress(walletAddr))
	if err != nil {
		return nil, fmt.Errorf("Failed to pack contract data: %v", err)
	}

	result, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contractAddr,
		Data: callData,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to call contract: %v", err)
	}

	var balance *big.Int
	err = contractABI.UnpackIntoInterface(&balance, nftBalanceMethod, result)
	if err != nil {
		return nil, fmt.Errorf("Failed to unpack result: %v", err)
	}

	return balance, nil
}

func nftBalanceHandler(w http.ResponseWriter, r *http.Request) {
	client, err := ethclient.Dial(bscNodeURL)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum client: %v", err)
	}

	contractABIParsed, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		log.Fatalf("Failed to load contract ABI: %v", err)
	}

	contractAddr := common.HexToAddress(contractAddress)

	balance, err := getNFTBalance(client, contractABIParsed, contractAddr, walletAddress)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting NFT balance: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"balance": balance.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/api/get-nft-balance", nftBalanceHandler)

	fmt.Println("Server is running on port", apiPort)
	log.Fatal(http.ListenAndServe(apiPort, nil))
}
