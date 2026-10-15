# Runbook: Node lnd-nux Mainnet (VPS)

Hasil praktik 15 Okt 2026 — 13 langkah dari nol sampai wallet mainnet jadi.
VPS: vps3278893 (sudah jalan nusacoind mainnet+testnet, 2 node lnd testnet).

## Prasyarat

- Binary `lnd`/`lncli` dari release `Nusanian/lnd` (cek tag terbaru)
- `nusacoind` mainnet sudah sync
- Port bebas: P2P 9737, RPC 10011, REST 8082 (sesuaikan bila dipakai)

## Langkah-langkah

### 1. Inspeksi config testnet (jadikan template)
```bash
grep -v "^#" ~/.lnd/lnd.conf | grep -v "^$"
ss -tln | grep -E ":9735|:9736|:10009|:10010|:8080|:8081"
```

### 2. Cek struktur nusacoin.conf + sync mainnet
```bash
grep -n "^\[.*\]" ~/.nusacoin/nusacoin.conf
nusacoin-cli -rpcport=8332 getblockchaininfo | grep -E '"chain"|blocks|headers'
```

### 3. Cek section [main] + proses berjalan
```bash
sed -n '4,19p' ~/.nusacoin/nusacoin.conf   # sesuaikan nomor baris
ps aux | grep nusacoind | grep -v grep
```

### 4. Betulkan ZMQ mainnet (WAJIB: rawblock & rawtx port berbeda)
```bash
cp ~/.nusacoin/nusacoin.conf ~/.nusacoin/nusacoin.conf.bak.$(date +%Y%m%d)
# di section [main]: zmqpubrawtx -> port beda (mis. 8524)
sed -i '4,19s|zmqpubrawtx=tcp://127.0.0.1:8523|zmqpubrawtx=tcp://127.0.0.1:8524|' ~/.nusacoin/nusacoin.conf
```

### 5. Restart nusacoind mainnet (manual, pastikan benar-benar mati)
```bash
nusacoin-cli -rpcport=8332 stop
while pgrep -f "nusacoind -daemon -txindex" > /dev/null; do echo "menunggu mati..."; sleep 5; done; echo "mati OK"
nusacoind -daemon -txindex
sleep 10; nusacoin-cli -rpcport=8332 getblockcount
ss -tln | grep -E ":8523|:8524"   # kedua port ZMQ harus listen
```

### 6. Buat lnd.conf mainnet
```bash
mkdir -p ~/.lnd-mainnet
RPCPASS=$(grep "^bitcoind.rpcpass" ~/.lnd/lnd.conf | cut -d= -f2)
cat > ~/.lnd-mainnet/lnd.conf << EOF
[Application Options]
alias=NUX-Mainnet-1
color=#0000ff
restlisten=127.0.0.1:8082
rpclisten=127.0.0.1:10011
listen=0.0.0.0:9737

[Nusacoin]
nusacoin.active=1
nusacoin.mainnet=1
nusacoin.node=bitcoind

[Bitcoind]
bitcoind.rpchost=127.0.0.1:8332
bitcoind.rpcuser=nusataobot
bitcoind.rpcpass=$RPCPASS
bitcoind.zmqpubrawblock=tcp://127.0.0.1:8523
bitcoind.zmqpubrawtx=tcp://127.0.0.1:8524

[protocol]
protocol.wumbo-channels=true
EOF
```

### 7. Start daemon lnd mainnet
```bash
cd /root/lnd-nux
nohup ./lnd --lnddir=/root/.lnd-mainnet > /root/.lnd-mainnet/lnd.log 2>&1 &
sleep 8; tail -5 /root/.lnd-mainnet/lnd.log
```

### 8. Buat wallet (interaktif)
```bash
/root/lnd-nux/lncli --lnddir=/root/.lnd-mainnet --rpcserver=127.0.0.1:10011 --network=mainnet --chain=nusacoin create
```
Password 2x → `n` (seed baru) → **catat 24 kata seed di tempat aman**.

### 9. Verifikasi
```bash
/root/lnd-nux/lncli --lnddir=/root/.lnd-mainnet --rpcserver=127.0.0.1:10011 --network=mainnet --chain=nusacoin getinfo | grep -E "identity_pubkey|synced_to_chain|block_height"
```

## Troubleshooting (kasus nyata 15 Okt 2026)

| Gejala | Penyebab | Solusi |
|---|---|---|
| `Cannot obtain a lock on data directory` | Proses lama belum mati sempurna | Loop `pgrep` sampai benar-benar hilang (langkah 5) |
| `failed to neuter master extended key: unknown hd private extended key bytes` | **Bug**: `HDPrivateKeyID`/`HDPublicKeyID` mainnet tertukar di `chainreg/nusacoin.go` | Fix kode → rebuild → binary baru (sudah fix di `v0.0.0-nux-test4`+) |
| `wallet already exists` setelah gagal create | Wallet setengah-jadi tertulis ke disk | Bila wallet masih kosong: stop daemon, `rm -rf ~/.lnd-mainnet/data`, ulangi dari langkah 7 |
| `InitWallet: got no unlock message before timeout` | Tidak selesaikan prompt interaktif tepat waktu | Ulangi `lncli create`, siapkan password & catat seed dengan cepat |
| BtcWallet tolak sync | Chain tip > 2 jam (block NUX lambat) | Tunggu block baru masuk, lalu restart lnd |

## Catatan

- Seed & password wallet = tanggung jawab pemegang node. Jangan simpan di chat/log.
- Binary rilis: https://github.com/Nusanian/lnd/releases (tag `v*-nux*`)
- Selalu verifikasi sha256 sebelum extract binary baru.
