package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	API_URL     = "https://api.battleofnodes.com"
	GATEWAY_URL = "https://gateway.battleofnodes.com"
	ChainID     = "B"

	LeaderPem     = "leader.pem"
	AddressesFile = "addresses.json"
	WalletsFile   = "wallets.json"

	// Funder
	FundGasPrice = 1000000000

	// Spammer
	SpamValue    = "1"
	SpamGasLimit = 50000
	SpamGasPrice = 1000000000
	SpamBatch    = 100 // txs per POST

	// Mempool window: how many nonces ahead we send before waiting
	// BoN node seems to accept ~100 pending nonces per sender
	MempoolWindow = 96
)

type Tx struct {
	Nonce     uint64 `json:"nonce"`
	Value     string `json:"value"`
	Receiver  string `json:"receiver"`
	Sender    string `json:"sender"`
	GasPrice  uint64 `json:"gasPrice"`
	GasLimit  uint64 `json:"gasLimit"`
	ChainID   string `json:"chainID"`
	Version   uint32 `json:"version"`
	Signature string `json:"signature"`
}

func signTx(privKey ed25519.PrivateKey, sender, receiver, value string, nonce, gasPrice, gasLimit uint64) *Tx {
	raw := fmt.Sprintf(
		`{"nonce":%d,"value":"%s","receiver":"%s","sender":"%s","gasPrice":%d,"gasLimit":%d,"chainID":"%s","version":1}`,
		nonce, value, receiver, sender, gasPrice, gasLimit, ChainID,
	)
	sig := ed25519.Sign(privKey, []byte(raw))
	return &Tx{
		Nonce: nonce, Value: value, Receiver: receiver, Sender: sender,
		GasPrice: gasPrice, GasLimit: gasLimit, ChainID: ChainID,
		Version: 1, Signature: hex.EncodeToString(sig),
	}
}

func getAccountNonce(client *http.Client, address string) (uint64, error) {
	resp, err := client.Get(fmt.Sprintf("%s/accounts/%s", API_URL, address))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	if n, ok := res["nonce"]; ok {
		return uint64(n.(float64)), nil
	}
	return 0, fmt.Errorf("no nonce in response")
}

func sendBatch(client *http.Client, txs []*Tx) (int, error) {
	body, _ := json.Marshal(txs)
	req, _ := http.NewRequest("POST", GATEWAY_URL+"/transaction/send-multiple", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	var r struct {
		Data struct {
			NumOfSentTxs int `json:"numOfSentTxs"`
		} `json:"data"`
	}
	json.NewDecoder(res.Body).Decode(&r)
	return r.Data.NumOfSentTxs, nil
}

func loadLeaderKey() (ed25519.PrivateKey, string, error) {
	pemBytes, err := ioutil.ReadFile(LeaderPem)
	if err != nil {
		return nil, "", err
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, "", fmt.Errorf("failed to decode PEM block")
	}
	keyBytes, err := hex.DecodeString(string(block.Bytes))
	if err != nil {
		return nil, "", fmt.Errorf("hex decode PEM: %v", err)
	}
	privKey := ed25519.NewKeyFromSeed(keyBytes[:32])
	addr := ""
	t := block.Type
	if len(t) > 16 {
		addr = t[len("PRIVATE KEY for "):]
	}
	return privKey, addr, nil
}

// ============================================================
// MODE: fund
// ============================================================
func runFund(fundAmount string) {
	leaderPriv, leaderAddr, err := loadLeaderKey()
	if err != nil {
		fmt.Println("Error loading leader PEM:", err)
		os.Exit(1)
	}
	fmt.Printf("Leader: %s\n", leaderAddr)
	fmt.Printf("Fund amount per wallet: %s attoEGLD (%.4f EGLD)\n", fundAmount, func() float64 {
		v, _ := strconv.ParseFloat(fundAmount, 64)
		return v / 1e18
	}())

	addrData, _ := ioutil.ReadFile(AddressesFile)
	var addresses []string
	json.Unmarshal(addrData, &addresses)
	fmt.Printf("Distributing to %d wallets...\n", len(addresses))

	client := &http.Client{Timeout: 15 * time.Second}
	nonce, _ := getAccountNonce(client, leaderAddr)
	fmt.Printf("Starting nonce: %d\n\n", nonce)

	batchSize := 50
	i := 0
	for i < len(addresses) {
		end := i + batchSize
		if end > len(addresses) {
			end = len(addresses)
		}

		var batch []*Tx
		startNonce := nonce
		for _, addr := range addresses[i:end] {
			raw := fmt.Sprintf(
				`{"nonce":%d,"value":"%s","receiver":"%s","sender":"%s","gasPrice":%d,"gasLimit":%d,"data":"c3VwZXJub3ZhLWZ1bmRpbmc=","chainID":"B","version":1}`,
				nonce, fundAmount, addr, leaderAddr, FundGasPrice, 86000,
			)
			sig := ed25519.Sign(leaderPriv, []byte(raw))
			batch = append(batch, &Tx{
				Nonce: nonce, Value: fundAmount, Receiver: addr, Sender: leaderAddr,
				GasPrice: FundGasPrice, GasLimit: 86000, ChainID: ChainID,
				Version: 1, Signature: hex.EncodeToString(sig),
			})
			nonce++
		}

		sent, err := sendBatch(client, batch)
		if err != nil {
			fmt.Printf("  Batch %d error: %v — retrying\n", i/batchSize, err)
			nonce = startNonce
			time.Sleep(2 * time.Second)
			continue
		}
		fmt.Printf("Batch %d (%d–%d): accepted %d/%d\n", i/batchSize, i, end-1, sent, len(batch))

		if sent < len(batch) {
			needed := startNonce + uint64(sent)
			fmt.Printf("  Mempool full — waiting for on-chain nonce %d...\n", needed)
			for {
				time.Sleep(2 * time.Second)
				onChain, _ := getAccountNonce(client, leaderAddr)
				fmt.Printf("  On-chain: %d\n", onChain)
				if onChain >= needed {
					nonce = onChain
					break
				}
			}
			i += sent
			continue
		}

		// Wait for chain confirmation before next batch
		for {
			time.Sleep(800 * time.Millisecond)
			onChain, _ := getAccountNonce(client, leaderAddr)
			if onChain >= nonce {
				break
			}
		}
		i += len(batch)
	}
	fmt.Println("\n✅ All wallets funded!")
}

// ============================================================
// MODE: spam
// ============================================================
func runSpam(receiver string, durationSec int) {
	keyData, _ := ioutil.ReadFile(WalletsFile)
	var privKeysHex []string
	json.Unmarshal(keyData, &privKeysHex)

	addrData, _ := ioutil.ReadFile(AddressesFile)
	var addresses []string
	json.Unmarshal(addrData, &addresses)

	type Wallet struct {
		Address string
		PrivKey ed25519.PrivateKey
	}
	wallets := make([]Wallet, len(privKeysHex))
	for i, hexKey := range privKeysHex {
		seed, _ := hex.DecodeString(hexKey)
		wallets[i] = Wallet{Address: addresses[i], PrivKey: ed25519.NewKeyFromSeed(seed)}
	}

	fmt.Printf("🚀 Spamming with %d wallets → %s\n", len(wallets), receiver)
	fmt.Printf("   Duration: %ds | Gas: %d | Mempool window: %d\n\n", durationSec, SpamGasLimit, MempoolWindow)

	deadline := time.Now().Add(time.Duration(durationSec) * time.Second)
	var totalSent int64

	// Load all nonces concurrently
	fmt.Print("Loading nonces... ")
	nonces := make([]uint64, len(wallets))
	sem := make(chan struct{}, 50)
	var wg0 sync.WaitGroup
	var mu sync.Mutex
	nonceClient := &http.Client{Timeout: 10 * time.Second}
	for i := range wallets {
		wg0.Add(1)
		sem <- struct{}{}
		go func(idx int) {
			defer wg0.Done()
			defer func() { <-sem }()
			n, _ := getAccountNonce(nonceClient, wallets[idx].Address)
			mu.Lock()
			nonces[idx] = n
			mu.Unlock()
		}(i)
	}
	wg0.Wait()
	fmt.Println("done.")
	fmt.Printf("Starting spam at %s UTC...\n\n", time.Now().UTC().Format("15:04:05"))

	// Progress reporter
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		start := time.Now()
		prev := int64(0)
		for t := range ticker.C {
			if t.After(deadline) {
				return
			}
			elapsed := time.Since(start).Seconds()
			cur := atomic.LoadInt64(&totalSent)
			rate := float64(cur-prev) / 5.0
			prev = cur
			remaining := time.Until(deadline).Seconds()
			fmt.Printf("[%3.0fs | %3.0fs left] txs=%-8d rate=%.0f tx/s\n",
				elapsed, remaining, cur, rate)
		}
	}()

	// One goroutine per wallet
	var wg sync.WaitGroup
	for i, w := range wallets {
		wg.Add(1)
		go func(wallet Wallet, startNonce uint64) {
			defer wg.Done()
			client := &http.Client{Timeout: 8 * time.Second}
			
			// Track two pointers:
			// sendNonce = next nonce to sign and send
			// confirmedNonce = last known on-chain nonce (fetched periodically)
			confirmedNonce := startNonce
			sendNonce := startNonce

			for time.Now().Before(deadline) {
				// How many are in flight?
				inFlight := sendNonce - confirmedNonce

				if inFlight >= MempoolWindow {
					// Mempool window full — poll on-chain nonce
					n, err := getAccountNonce(client, wallet.Address)
					if err == nil {
						confirmedNonce = n
					}
					if sendNonce-confirmedNonce >= MempoolWindow {
						// Still full, short wait then retry
						time.Sleep(300 * time.Millisecond)
						continue
					}
				}

				// Fill a batch up to window limit
				var batch []*Tx
				for len(batch) < SpamBatch &&
					sendNonce-confirmedNonce < MempoolWindow &&
					time.Now().Before(deadline) {
					tx := signTx(wallet.PrivKey, wallet.Address, receiver, SpamValue, sendNonce, SpamGasPrice, SpamGasLimit)
					batch = append(batch, tx)
					sendNonce++
				}

				if len(batch) == 0 {
					continue
				}

				sent, err := sendBatch(client, batch)
				if err != nil {
					// On error back off sendNonce
					sendNonce -= uint64(len(batch) - sent)
					time.Sleep(200 * time.Millisecond)
					continue
				}
				atomic.AddInt64(&totalSent, int64(sent))
				if sent < len(batch) {
					// Partial — back up unsent nonces
					sendNonce -= uint64(len(batch) - sent)
				}
			}
		}(w, nonces[i])
	}

	wg.Wait()
	fmt.Printf("\n✅ Done! Total transactions sent: %d\n", atomic.LoadInt64(&totalSent))
}

// ============================================================
// main
// ============================================================
func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "fund":
		amount := "4000000000000000000"
		if len(os.Args) >= 3 {
			amount = os.Args[2]
		}
		runFund(amount)
	case "spam":
		if len(os.Args) < 4 {
			fmt.Println("Usage: bon spam <receiver_address> <duration_seconds>")
			os.Exit(1)
		}
		receiver := os.Args[2]
		secs, _ := strconv.Atoi(os.Args[3])
		runSpam(receiver, secs)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Battle of Nodes — Supernova Surge toolkit

Usage:
  bon fund [amount_attoEGLD]   — distribute funds to 500 wallets
  bon spam <receiver> <secs>   — spam MoveBalance txs from all wallets

Examples:
  bon fund
  bon spam erd10mcwua04j5r9ujny9y3pza32kmeeq9vqs8eq2je26r27yh28ynfqhxmahk 1800`)
}
