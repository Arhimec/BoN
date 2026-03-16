package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	API_URL       = "https://api.battleofnodes.com"
	GATEWAY_URL   = "https://gateway.battleofnodes.com"
	ChainID       = "B"
	LeaderPem     = "leader.pem"
	AddressesFile = "addresses.json"
	FundAmount    = "4000000000000000000" // 4 EGLD
	BatchSize     = 50
	GasLimit      = 70000
	GasPrice      = 1000000000
)

type PayloadTx struct {
	Nonce uint64 `json:"nonce"`; Value string `json:"value"`; Receiver string `json:"receiver"`; Sender string `json:"sender"`
	GasPrice uint64 `json:"gasPrice"`; GasLimit uint64 `json:"gasLimit"`; Data string `json:"data"`
	ChainID string `json:"chainID"`; Version uint32 `json:"version"`; Signature string `json:"signature"`
}

func main() {
	leaderAddress := "erd10mcwua04j5r9ujny9y3pza32kmeeq9vqs8eq2je26r27yh28ynfqhxmahk"
	
	pemBytes, _ := ioutil.ReadFile(LeaderPem)
	block, _ := pem.Decode(pemBytes)
	privKey := ed25519.NewKeyFromSeed(block.Bytes[:32])

	resp, _ := http.Get(fmt.Sprintf("%s/address/%s", API_URL, leaderAddress))
	var accRes struct { Data struct { Account struct { Nonce uint64 `json:"nonce"` } `json:"account"` } `json:"data"` }
	json.NewDecoder(resp.Body).Decode(&accRes)
	nonce := accRes.Data.Account.Nonce

	addrBytes, _ := ioutil.ReadFile(AddressesFile)
	var addresses []string
	json.Unmarshal(addrBytes, &addresses)

	dataB64 := base64.StdEncoding.EncodeToString([]byte("supernova-funding"))
	client := &http.Client{Timeout: 10 * time.Second}

	for i := 0; i < len(addresses); i += BatchSize {
		end := i + BatchSize
		if end > len(addresses) { end = len(addresses) }
		
		var batch []*PayloadTx
		for _, addr := range addresses[i:end] {
			raw := fmt.Sprintf(`{"nonce":%d,"value":"%s","receiver":"%s","sender":"%s","gasPrice":%d,"gasLimit":%d,"data":"%s","chainID":"%s","version":1}`,
				nonce, FundAmount, addr, leaderAddress, GasPrice, GasLimit, dataB64, ChainID)
			sig := ed25519.Sign(privKey, []byte(raw))
			batch = append(batch, &PayloadTx{
				Nonce: nonce, Value: FundAmount, Receiver: addr, Sender: leaderAddress,
				GasPrice: GasPrice, GasLimit: GasLimit, Data: dataB64, ChainID: ChainID,
				Version: 1, Signature: hex.EncodeToString(sig),
			})
			nonce++
		}
		
		body, _ := json.Marshal(batch)
		req, _ := http.NewRequest("POST", GATEWAY_URL+"/transaction/send-multiple", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		res, _ := client.Do(req)
		resBody, _ := ioutil.ReadAll(res.Body)
		fmt.Printf("Batch %d Sent. API Response: %s\n", i/BatchSize, string(resBody))
		res.Body.Close()
		time.Sleep(500 * time.Millisecond)
	}
}