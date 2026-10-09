# Panduan Instalasi lnd-nux (Lightning Network untuk Nusacoin)

Panduan ini menjelaskan cara menjalankan **node Lightning Network Nusacoin** (`lnd-nux`) di **testnet**, langkah demi langkah, untuk operator komunitas yang bukan expert. Semua perintah di bawah ini sudah diverifikasi bekerja di regtest.

> **Catatan:** `lnd-nux` adalah fork LND yang di-porting untuk chain Nusacoin (X11 PoW). Nusacoin **belum** mengaktifkan Taproot — untuk saat ini `lnd-nux` disesuaikan sehingga channel/payment standar (non-taproot) berjalan normal. Taproot masuk roadmap upgrade codebase ke depan. Invoice testnet diawali `lntn`, alamat on-chain diawali `tn1`.

---

## 1. Prasyarat Sistem

- **OS:** Linux 64-bit (Ubuntu/Debian 22.04+ disarankan)
- **RAM:** minimal 2 GB (cukup untuk testnet)
- **Disk:** minimal 10 GB kosong
- **Koneksi internet** yang stabil (node harus online terus agar channel aman)

---

## 2. Mendapatkan Binary

Ada dua program yang dibutuhkan:

| Program | Fungsi | Sumber |
|---------|--------|--------|
| `nusacoind` + `nusacoin-cli` | Backend blockchain (full node Nusacoin) | Build dari https://github.com/TaobotX11/nusacoin, atau minta ke komunitas |
| `lnd` + `lncli` | Node Lightning (`lnd-nux`) | Build dari https://github.com/Nusanian/lnd branch `nusacoin`, atau minta ke komunitas |

### 2a. Build `lnd-nux` dari source (butuh Go)

```bash
# Install Go (cek versi yang diminta file go.mod terlebih dahulu)
# https://go.dev/dl/

git clone https://github.com/Nusanian/lnd lnd-nux
cd lnd-nux
git checkout nusacoin          # PENTING: jangan pakai master
go build -o /usr/local/bin/lnd ./cmd/lnd
go build -o /usr/local/bin/lncli ./cmd/lncli
```

### 2b. Build `nusacoind` dari source (lama: 1–2 jam)

> **Sudah punya `nusacoind`? LEWATI bagian ini.** Kalau di sistem sudah ada `nusacoind` yang jalan (misalnya node yang sudah sync), tidak perlu build ulang — langsung lanjut ke bagian 3 dan pastikan config-nya mengaktifkan RPC + ZMQ seperti contoh di sana. `lnd-nux` tinggal diarahkan ke node yang sudah ada tersebut.

```bash
git clone https://github.com/TaobotX11/nusacoin
cd nusacoin
./autogen.sh
./configure --without-gui --disable-tests --disable-bench \
    --without-sqlite --with-zmq --with-incompatible-bdb
make -j$(nproc)
# Binary ada di: src/nusacoind dan src/nusacoin-cli
sudo cp src/nusacoind src/nusacoin-cli /usr/local/bin/
```

> Butuh ZMQ (`--with-zmq`) — `lnd-nux` **wajib** menerima notifikasi block/transaksi via ZMQ dari `nusacoind`.

---

## 3. Menjalankan `nusacoind` (Testnet)

Buat file config `~/.nusacoin/nusacoin.conf`:

```ini
server=1
daemon=1
txindex=1
rpcuser=nuxrpc
rpcpassword=GANTI_DENGAN_PASSWORD_YANG_KUAT_DAN_ACAK

[test]
rpcport=18332
zmqpubrawblock=tcp://127.0.0.1:28332
zmqpubrawtx=tcp://127.0.0.1:28333
```

Jalankan dan tunggu sinkronisasi selesai:

```bash
nusacoind -daemon
# Cek progres (100% = sudah sync):
nusacoin-cli -rpcuser=nuxrpc -rpcpassword=GANTI_DENGAN_PASSWORD_YANG_KUAT_DAN_ACAK \
    -rpcport=18332 getblockchaininfo | grep verificationprogress
```

> Sinkronisasi awal testnet bisa memakan waktu lama (jam). `lnd-nux` baru bisa jalan setelah backend selesai sync.

---

## 4. Konfigurasi `lnd-nux` (Testnet)

Buat file `~/.lnd/lnd.conf`:

```ini
[Application Options]
alias=NamaNodeKamu
color=#00ff00
restlisten=127.0.0.1:8080
rpclisten=127.0.0.1:10009
listen=0.0.0.0:9735

[Nusacoin]
nusacoin.active=1
nusacoin.testnet=1
nusacoin.node=bitcoind

[Bitcoind]
bitcoind.rpchost=127.0.0.1:18332
bitcoind.rpcuser=nuxrpc
bitcoind.rpcpass=GANTI_DENGAN_PASSWORD_YANG_KUAT_DAN_ACAK
bitcoind.zmqpubrawblock=tcp://127.0.0.1:28332
bitcoind.zmqpubrawtx=tcp://127.0.0.1:28333
```

**Penjelasan singkat tiap bagian:**
- `[Application Options]` — nama tampilan node kamu (`alias`), dan port layanan lnd (10009 = RPC, 8080 = REST, 9735 = P2P antar node Lightning).
- `[Nusacoin]` — mengaktifkan chain Nusacoin di jaringan **testnet**.
- `[Bitcoind]` — cara lnd berbicara ke `nusacoind`: RPC untuk query + **ZMQ untuk notifikasi block** (keduanya wajib, dan port ZMQ block & transaksi harus beda).

> Nama section-nya tetap `[Bitcoind]` walaupun backend-nya `nusacoind` — itu nama modul konektornya, bukan nama coin-nya.

---

## 5. Membuat Wallet

Jalankan lnd (biarkan berjalan di terminal / pakai `screen` / `systemd`):

```bash
lnd
```

Di terminal **lain**, buat wallet baru (hanya sekali, saat pertama jalan):

```bash
lncli -n testnet --chain=nusacoin create
```

Ikuti panduannya:
1. Masukkan **password wallet** (2x) — catat baik-baik, hilang = dana hilang.
2. Saat ditanya seed mnemonic: jawab `n` untuk membuat seed baru (atau `y` kalau restore).
3. Lewati cipher seed passphrase (Enter saja) kecuali kamu paham fungsinya.
4. Ketik `YES` untuk konfirmasi. **Simpan 24 kata seed di tempat aman, jangan di-share ke siapa pun!**

> ⚠️ **Selalu tambahkan `--chain=nusacoin`** di setiap perintah `lncli`. Tanpa itu, `lncli` mencari file di folder `bitcoin` dan akan error `unable to read macaroon path`.

Buka wallet yang sudah ada (setiap kali lnd restart):

```bash
lncli -n testnet --chain=nusacoin unlock
# masukkan password wallet
```

---

## 6. Cek Status & Sinkronisasi

```bash
lncli -n testnet --chain=nusacoin getinfo
```

Yang perlu diperhatikan:
- `"synced_to_chain": true` → node siap dipakai.
- `"block_height"` → harus sama dengan tinggi block testnet terkini.
- `"identity_pubkey"` → identitas node kamu (dibagikan saat connect).

Kalau muncul `server is still in the process of starting`, tunggu — lnd masih sinkronisasi.

---

## 7. Isi Saldo On-Chain (Testnet NUX)

Minta alamat deposit:

```bash
lncli -n testnet --chain=nusacoin newaddress p2wkh
# contoh hasil: tn1q...
```

Kirim **tNUX** (koin testnet, tidak bernilai) ke alamat itu dari faucet/komunitas, lalu cek:

```bash
lncli -n testnet --chain=nusacoin walletbalance
```

Tunggu minimal **1 konfirmasi** sebelum dipakai buka channel. Block Nusacoin ~6 menit, jadi bersabarlah.

> **Shortcut (opsional):** karena difficulty testnet saat ini sangat rendah, kamu bisa mempercepat dengan menambang 1 block sendiri (block yang kamu mine akan mengonfirmasi transaksi yang masih antre di mempool):
> ```bash
> nusacoin-cli -testnet generatetoaddress 1 "$(nusacoin-cli -testnet getnewaddress)"
> ```
> Catatan: ini hanya bisa diandalkan selama difficulty testnet masih sangat rendah. Kalau perintahnya tidak menghasilkan block, kembali ke cara normal — tunggu konfirmasi dari jaringan.

---

## 8. Connect ke Node Komunitas

"Node" di sini maksudnya **node lnd di mesin lain** — misalnya VPS anggota komunitas lain yang juga menjalankan `lnd-nux`.

> **Bedakan dua langkah ini:**
> - `connect` = berkenalan di jaringan P2P Lightning (saling tukar info jaringan/gossip). Belum ada dana, belum ada channel.
> - `openchannel` (bagian 9) = baru mengunci dana dan membuka channel pembayaran. Ini langkah setelah connect.

Dapatkan `pubkey@host:port` node tujuan dari komunitas, lalu:

```bash
lncli -n testnet --chain=nusacoin connect <pubkey>@<host>:9735
lncli -n testnet --chain=nusacoin listpeers   # verifikasi
```

Agar node kamu bisa ditemukan orang lain, pastikan port **9735** terbuka di firewall/VPS dan `listen=0.0.0.0:9735` sudah diset.

> **Tentang `<host>`:** bisa berupa **alamat IP** (`203.0.113.10`) maupun **nama domain** seperti website (`node.komunitas.org`) — lnd akan me-resolve-nya via DNS. Format lengkapnya `<pubkey>@<host>:<port>`, contoh: `03a1b2...c3@node.komunitas.org:9735`.
>
> **Pasang domain sendiri:** cukup buat DNS **A record** (nama `node` → IP VPS) di pengaturan DNS domain kamu — tidak perlu ada website di domain itu, dan tidak ada setting khusus Lightning. **Tidak ada kaitannya dengan *DNS Seed*:** DNS seed dipakai base layer (`nusacoind`) untuk mencari peer P2P saat pertama start; domain di sini sekadar nama yang mudah diingat untuk node Lightning-mu.

> **Tentang `<pubkey>`:** ini adalah *identity pubkey* node Lightning — "nomor telepon" node di jaringan (66 karakter hex, contoh: `03a1b2...`). Pubkey **dibuat otomatis saat wallet dibuat (langkah 5)** dan bisa dilihat kapan saja via `getinfo` → `"identity_pubkey"` (langkah 6). Pubkey diturunkan dari seed wallet, jadi seed yang sama = pubkey yang sama — itulah kenapa backup 24 kata juga mem-backup identitas node-mu. Untuk `connect`, yang dipakai adalah pubkey **node lawan** (minta ke komunitas); pubkey-mu sendiri dibagikan ke orang lain supaya mereka bisa connect ke kamu.

---

## 9. Membuka Channel

```bash
lncli -n testnet --chain=nusacoin openchannel \
    --node_key=<pubkey_tujuan> \
    --local_amt=<jumlah_dalam_satoshi>
```

Contoh: buka channel 1.000.000 satoshi (0,01 tNUX):

```bash
lncli -n testnet --chain=nusacoin openchannel \
    --node_key=03abc...def \
    --local_amt=1000000
```

- Channel butuh **3 konfirmasi** (~18 menit di testnet) sebelum aktif.
- Cek status: `lncli -n testnet --chain=nusacoin listchannels` → `"active": true` berarti siap dipakai.
- Biaya buka channel = fee transaksi on-chain biasa.

---

## 10. Membayar via Invoice (Lightning Payment)

**Penerima** membuat invoice:

```bash
lncli -n testnet --chain=nusacoin addinvoice --amt=1000 --memo="kopi"
# catat "payment_request" yang diawali lntn...
```

**Pengirim** membayar:

```bash
lncli -n testnet --chain=nusacoin payinvoice --force <payment_request>
```

Hasil `SUCCEEDED` = pembayaran lunas dalam hitungan detik, tanpa menunggu block.

---

## 11. Menutup Channel

Lihat dulu `channel_point` (format `txid:output_index`):

```bash
lncli -n testnet --chain=nusacoin listchannels
```

**Tutup kooperatif** (disarankan — murah & dana langsung kembali):

```bash
lncli -n testnet --chain=nusacoin closechannel <funding_txid> <output_index>
```

**Tutup paksa / force-close** (kalau lawan offline atau tidak kooperatif):

```bash
lncli -n testnet --chain=nusacoin closechannel --force <funding_txid> <output_index>
```

> Catatan force-close: dana kamu dikunci oleh timelock (CSV delay) dan baru bisa dipakai setelah beberapa puluh block. Ini normal dan sesuai protokol Lightning — bukan bug.

---

## 12. Troubleshooting

| Gejala | Penyebab & Solusi |
|--------|-------------------|
| `lncli` error `unable to read macaroon path .../bitcoin/...` | Lupa `--chain=nusacoin`. Selalu sertakan di tiap perintah `lncli`. |
| `server is still in the process of starting` | lnd masih sync. Cek `getinfo` berkala sampai `synced_to_chain: true`. |
| lnd tidak mau start / macet di awal | Pastikan `nusacoind` sudah fully-synced dan ZMQ aktif. Cek log lnd (`~/.lnd/logs/`). |
| `node backend does not support taproot` | Binary `lnd` terlalu lama — pakai build terbaru dari branch `nusacoin`. |
| Channel tidak aktif setelah open | Butuh 3 konfirmasi (~18 menit). Cek `listchannels` → `active: false` + `pending`. |
| Payment gagal / `no route` | Pastikan channel `active: true`, ada peer yang connect, dan likuiditas cukup di arah yang benar. |
| `Fee estimation failed` di `nusacoin-cli` | Normal di awal chain — set manual: `nusacoin-cli settxfee 0.00001`. |

### FAQ singkat

**Q: Kalau baru node saya saja yang jalan, apakah bisa dipakai?**
A: Daemon dan wallet-nya bisa jalan normal, tapi payment tidak bisa — payment butuh channel, dan channel butuh 2 pihak. Minimal 2 node + 1 channel untuk payment pertama terjadi. Kabar baik: 2 node bisa jalan di 1 mesin yang sama (datadir dan port P2P/RPC yang berbeda) untuk testing.

**Q: Apa bedanya `connect` dan `openchannel`?**
A: `connect` hanya berkenalan di jaringan P2P (belum ada dana). `openchannel` yang mengunci dana dan membuka channel pembayaran — dilakukan setelah connect.

---

## 13. File & Port Penting (Referensi Cepat)

| Item | Lokasi / Nilai |
|------|----------------|
| Config lnd | `~/.lnd/lnd.conf` |
| Config nusacoind | `~/.nusacoin/nusacoin.conf` |
| Macaroon lnd (testnet) | `~/.lnd/data/chain/nusacoin/testnet/admin.macaroon` |
| TLS cert lnd | `~/.lnd/tls.cert` |
| lnd RPC / REST / P2P | 10009 / 8080 / 9735 |
| nusacoind testnet RPC | 18332 |
| nusacoind ZMQ block / tx | 28332 / 28333 |
| Prefix alamat testnet | `tn1...` (bech32) |
| Prefix invoice testnet | `lntn...` |
| Block time | ~6 menit |

---

*Dokumen ini bagian dari proyek LightningNUX — branch `nusacoin` di https://github.com/Nusanian/lnd. Testnet & signet Nusacoin dikelola komunitas.*
