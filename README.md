
<div align="center">

```
███████╗██╗   ██╗██████╗ ███████╗██████╗ ██████╗  █████╗ ██████╗ ███████╗██████╗ ███████╗ █████╗ ██████╗ ███████╗
██╔════╝██║   ██║██╔══██╗██╔════╝██╔══██╗██╔══██╗██╔══██╗██╔══██╗██╔════╝██╔══██╗██╔════╝██╔══██╗██╔══██╗██╔════╝
███████╗██║   ██║██████╔╝█████╗  ██████╔╝██████╔╝███████║██████╔╝█████╗  ██████╔╝█████╗  ███████║██████╔╝███████╗
╚════██║██║   ██║██╔═══╝ ██╔══╝  ██╔══██╗██╔══██╗██╔══██║██╔══██╗██╔══╝  ██╔══██╗██╔══╝  ██╔══██║██╔══██╗╚════██║
███████║╚██████╔╝██║     ███████╗██║  ██║██║  ██║██║  ██║██║  ██║███████╗██║  ██║███████╗██║  ██║██║  ██║███████║
╚══════╝ ╚═════╝ ╚═╝     ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝
```

# 🐻 SuperRareBears

### Battle of Nodes — Guild Wars · Challenge 2: Supernova Surge

[![Network](https://img.shields.io/badge/Network-MultiversX%20Supernova-23A1F1?style=for-the-badge&logo=data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cGF0aCBkPSJNMTIgMkM2LjQ4IDIgMiA2LjQ4IDIgMTJzNC40OCAxMCAxMCAxMCAxMC00LjQ4IDEwLTEwUzE3LjUyIDIgMTIgMnoiIGZpbGw9IndoaXRlIi8+PC9zdmc+)](https://bon.multiversx.com)
[![Chain](https://img.shields.io/badge/Chain%20ID-B-00D4FF?style=for-the-badge)](#)
[![Block Time](https://img.shields.io/badge/Block%20Time-600ms-00FF88?style=for-the-badge)](#)
[![Wallets](https://img.shields.io/badge/Wallets-500-FF6B35?style=for-the-badge)](#)
[![License](https://img.shields.io/badge/License-MIT-purple?style=for-the-badge)](#)

</div>

---

## ⚡ What Is This

This is **SuperRareBears'** transaction toolkit for [Battle of Nodes Guild Wars](https://bon.multiversx.com/guild-wars) — Challenge 2: **Supernova Surge**.

The Supernova upgrade dropped block times from **6 seconds → 600 milliseconds** (10× faster). This toolkit is built to exploit that speed: 500 concurrent wallets, mempool-aware nonce pipelining, and zero wasted gas.

---

## 🏆 Challenge Overview

| Parameter | Value |
|---|---|
| 🌐 Network | Post-Supernova MultiversX shadow fork |
| ⏱ Block time | **600ms** (was 6s in Challenge 1) |
| 💼 Wallets | 500 unique sending wallets |
| 🎯 Milestone | 2,500,000 transactions |
| 🏁 Window A | 16:00 – 16:30 UTC · 2,000 EGLD budget |
| 🏁 Window B | 17:00 – 17:30 UTC · 500 EGLD budget |

### Schedule — March 16, 2026 (UTC)

```
15:45  ██████░░░░  Funds arrive → distribute to 500 wallets
16:00  ████████░░  Window A OPEN — full send
16:30  ░░░░░░░░░░  Window A closes — 30 min break
17:00  ████████░░  Window B OPEN — optimized send
17:30  ░░░░░░░░░░  Challenge complete
```

---

## 🔢 Performance Stats

```
Wallets         :   500
Gas per tx      :   50,000 (minimum — no data field)
Cost per tx     :   0.00005 EGLD
Max txs/wallet  :   80,000  (4 EGLD budget)
Max total txs   :   40,000,000
Estimated peak  :   ~16,500 tx/s
Mempool window  :   96 nonces (node limit: 100)
```

> Tested at **~200 tx/s per 3 wallets** in a 60s live run. Extrapolates to 16,500+ tx/s at full 500-wallet scale.

---

## 🚀 Quick Start

### Prerequisites

```bash
go 1.22+
```

### 1. Generate Wallets

```bash
go run generate_wallets.go
```

Outputs:
- `wallets.json` — 500 private keys (**keep secret, never commit**)
- `addresses.json` — 500 bech32 wallet addresses

### 2. Build

```bash
go build -o bon main.go
```

### 3. Fund Wallets

Run at **15:45 UTC** when the guild leader wallet receives 2,500 EGLD:

```bash
./bon fund
```

Distributes **4 EGLD** to each of the 500 sending wallets directly from `leader.pem`. Handles mempool limits automatically — no babysitting required.

### 4. Spam — Window A

Run at **16:00 UTC**:

```bash
./bon spam <leader_address> 1800
```

### 5. Spam — Window B

Run at **17:00 UTC** (same command):

```bash
./bon spam <leader_address> 1800
```

---

## 🏗 Architecture

```
leader.pem
    │
    ▼  fund (4 EGLD × 500)
┌───────────────────────────────────┐
│  500 Sending Wallets              │
│  ┌──────┐ ┌──────┐ ┌──────┐      │
│  │ W001 │ │ W002 │ │ W003 │ ...  │
│  └──┬───┘ └──┬───┘ └──┬───┘      │
└─────┼────────┼────────┼──────────┘
      │        │        │
      ▼        ▼        ▼
   MoveBalance txs (1 attoEGLD each)
      │        │        │
      └────────┴────────┘
               │
               ▼
        leader address  ← receives tiny amounts back
        fees burned     ← counted for your score
```

### Nonce Pipeline

Each wallet goroutine maintains a sliding window of **96 pending nonces**:

```
confirmed nonce: 1000
send nonce:      1096  ← 96 ahead (mempool window full)
                  │
                  ▼
           poll on-chain nonce
           wait for confirmation
           refill window → repeat
```

This keeps the mempool saturated without triggering rejections.

---

## 📁 Files

| File | Description |
|---|---|
| `main.go` | Core toolkit — `fund` and `spam` commands |
| `generate_wallets.go` | Generates 500 ed25519 keypairs |
| `funder.go` | Original Challenge 1 funder (reference) |
| `addresses.json` | 500 wallet addresses (public — safe to commit) |
| `wallets.json` | ⚠️ Private keys — **gitignored, never commit** |
| `leader.pem` | ⚠️ Guild leader key — **gitignored, never commit** |

---

## 🔗 Links

| Resource | URL |
|---|---|
| 🏟 Guild Wars Portal | [bon.multiversx.com/guild-wars](https://bon.multiversx.com/guild-wars) |
| 🔍 Explorer | [bon-explorer.multiversx.com](https://bon-explorer.multiversx.com) |
| 📡 API | [api.battleofnodes.com](https://api.battleofnodes.com/docs) |
| 💬 Guild Leaders TG | [t.me/BoN_Guild_Leaders](https://t.me/BoN_Guild_Leaders) |
| 🐻 Guild | SuperRareBears |

---

<div align="center">

**SuperRareBears** · Battle of Nodes Guild Wars 2026

*Built for speed. Optimized for Supernova.*

</div>
