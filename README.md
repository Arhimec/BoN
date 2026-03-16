# Battle of Nodes — Supernova Surge

Guild challenge toolkit for [Battle of Nodes](https://bon.multiversx.com/guild-wars) — Challenge 2: Supernova Surge.

## Overview

Scripts for participating in the BoN Guild Wars challenge on the MultiversX post-Supernova shadow fork (600ms block times).

## Files

| File | Description |
|---|---|
| `main.go` | Main toolkit — fund + spam commands |
| `generate_wallets.go` | Generates 500 fresh wallets (`wallets.json` + `addresses.json`) |
| `addresses.json` | 500 sending wallet addresses |
| `funder.go` | Original funder script (Challenge 1 reference) |

> **Note:** `wallets.json` and `leader.pem` are excluded from this repo (private keys).

## Usage

### 1. Generate wallets
```bash
go run generate_wallets.go
# outputs: wallets.json (private keys), addresses.json (addresses)
```

### 2. Fund wallets (run at 15:45 UTC)
```bash
go build -o bon main.go
./bon fund
# Sends 4 EGLD from leader.pem to each of the 500 wallets
```

### 3. Spam transactions (run at 16:00 UTC — Window A)
```bash
./bon spam <leader_address> 1800
# Fires MoveBalance txs from all 500 wallets for 30 minutes
# Value: 1 attoEGLD | Gas: 50,000 (minimum) | Mempool window: 96
```

### 4. Window B (run at 17:00 UTC)
```bash
./bon spam <leader_address> 1800
```

## Technical Details

- **Network:** Post-Supernova MultiversX shadow fork (Chain ID: `B`)
- **Block time:** 600ms
- **Tx type:** MoveBalance (simple transfer)
- **Gas:** 50,000 (minimum, no data field)
- **Cost/tx:** 0.00005 EGLD
- **Mempool limit:** ~100 nonces per sender (we use 96)
- **Concurrency:** 1 goroutine per wallet, pipeline fills up to 96 nonces ahead, polls on-chain nonce when window is full

## Challenge Parameters

| | Window A | Window B |
|---|---|---|
| Time | 16:00–16:30 UTC | 17:00–17:30 UTC |
| Fee budget | 2,000 EGLD | 500 EGLD |
| Max wallets | 500 | 500 |
| Milestone | 2,500,000 txs | — |

## Links

- [BoN Portal](https://bon.multiversx.com/guild-wars)
- [Explorer](https://bon-explorer.multiversx.com)
- [API Docs](https://api.battleofnodes.com/docs)
