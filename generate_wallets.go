package main

import (
        "crypto/ed25519"
        "crypto/rand"
        "encoding/hex"
        "encoding/json"
        "fmt"
        "log"
        "os"

        // Updated to the new core SDK path
        "github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
)

const (
        NumWallets    = 500
        WalletsFile   = "wallets.json"
        AddressesFile = "addresses.json"
)

func main() {
        fmt.Printf("Generating %d fresh MultiversX wallets...\n", NumWallets)

        var privateKeysHex []string
        var addressesBech32 []string

        // Updated to the new Sirius v1.5.0 initialization syntax
        converter, err := pubkeyConverter.NewBech32PubkeyConverter(32, "erd")
        if err != nil {
                log.Fatalf("Failed to initialize pubkey converter: %v", err)
        }

        for i := 0; i < NumWallets; i++ {
                // 1. Generate 32 bytes of secure random entropy (the Ed25519 seed)
                seed := make([]byte, 32)
                if _, err := rand.Read(seed); err != nil {
                        log.Fatalf("Failed to generate random seed: %v", err)
                }

                // 2. Derive the Ed25519 keypair from the seed
                privateKey := ed25519.NewKeyFromSeed(seed)

                // The public key is the last 32 bytes of the 64-byte private key
                publicKey := privateKey[32:] 

                // 3. Format to MultiversX standards
                privKeyHex := hex.EncodeToString(seed) 
                bech32Address, err := converter.Encode(publicKey)
                if err != nil {
                        log.Fatalf("Failed to encode address: %v", err)
                }

                privateKeysHex = append(privateKeysHex, privKeyHex)
                addressesBech32 = append(addressesBech32, bech32Address)
        }

        // 4. Save wallets.json (Private Keys for the Spammer)
        saveJSON(WalletsFile, privateKeysHex)

        // 5. Save addresses.json (Public Addresses for the Funder)
        saveJSON(AddressesFile, addressesBech32)

        fmt.Printf("✅ Successfully generated %d wallets.\n", NumWallets)
        fmt.Printf("🔒 Private keys saved to: %s (DO NOT SHARE THIS FILE)\n", WalletsFile)
        fmt.Printf("📬 Public addresses saved to: %s\n", AddressesFile)
}

func saveJSON(filename string, data interface{}) {
        file, err := os.Create(filename)
        if err != nil {
                log.Fatalf("Failed to create %s: %v", filename, err)
        }
        defer file.Close()

        encoder := json.NewEncoder(file)
        encoder.SetIndent("", "  ")
        if err := encoder.Encode(data); err != nil {
                log.Fatalf("Failed to write JSON to %s: %v", filename, err)
        }
}