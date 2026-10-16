# Panduan Instalasi lnd-nux (Lightning Network untuk Nusacoin)

Panduan langkah-demi-langkah menjalankan **node Lightning Network Nusacoin** (`lnd-nux`) di **testnet**. Ditulis untuk operator komunitas tingkat pemula-menengah — tidak perlu jadi expert Bitcoin/Lightning untuk mengikuti.

> **Istilah penting:** satuan terkecil Nusacoin disebut **nusan** (bukan "sats"). 1 NUX = 100.000.000 nusan.

> **Status Taproot:** Nusacoin **belum** mengaktifkan Taproot (masuk roadmap upgrade codebase). Mulai versi `v0.0.0-nux-test3`, `lnd-nux` otomatis memakai alamat SegWit v0 (P2WKH) agar kompatibel — kamu tidak perlu setting apa-apa.

---

## Daftar Isi

1. [Gambaran Besar](#1-gambaran-besar)
2. [Prasyarat](#2-prasyarat)
3. [Download Binary (Cara Cepat)](#3-download-binary-cara-cepat)
4. [Build dari Source (Alternatif)](#4-build-dari-source-alternatif)
5. [Menjalankan nusacoind (Testnet)](#5-menjalankan-nusacoind-testnet)
6. [Konfigurasi lnd-nux](#6-konfigurasi-lnd-nux)
7. [Menjalankan lnd & Membuat Wallet](#7-menjalankan-lnd--membuat-wallet)
8. [Isi Saldo On-Chain](#8-isi-saldo-on-chain)
9. [Connect & Buka Channel](#9-connect--buka-channel)
10. [Wumbo Channels (Channel Besar)](#10-wumbo-channels-channel-besar)
11. [Membayar via Lightning](#11-membayar-via-lightning)
12. [Menutup Channel](#12-menutup-channel)
13. [Node Kedua di Mesin yang Sama (Opsional)](#13-node-kedua-di-mesin-yang-sama-opsional)
14. [Troubleshooting](#14-troubleshooting)
15. [Referensi Cepat](#15-referensi-cepat)

---

## 1. Gambaran Besar

Bayangkan Lightning Network seperti **tabungan bersama** antara dua orang:

- Kamu dan temanmu masing-masing setor dana ke sebuah "brankas bersama" (channel).
- Setelah brankas dibuka, kalian bisa saling bayar **berkali-kali, instan, hampir tanpa fee** — tanpa menunggu block baru.
- Kapan pun, salah satu bisa menutup brankas dan mengambil dananya kembali ke dompet biasa.

Untuk ini kamu butuh dua program:

| Program | Peran | Analogi |
|---------|-------|---------|
| `nusacoind` | Full node blockchain Nusacoin — mencatat semua transaksi on-chain | Buku kas utama |
| `lnd` (lnd-nux) | Node Lightning — mengelola channel & payment kilat | Aplikasi tabungan bersama |

`lnd` tidak menyimpan blockchain sendiri — dia "numpang baca" ke `nusacoind` via RPC dan menerima notifikasi block baru via ZMQ.

---

## 2. Prasyarat

- **OS:** Linux 64-bit (Ubuntu/Debian 22.04+ disarankan)
- **RAM:** minimal 2 GB
- **Disk:** minimal 10 GB kosong
- **Internet** stabil (node Lightning harus online terus — kalau offline lama, channel bisa bermasalah)
- `nusacoind` sudah terinstall (komunitas sudah diedukasi untuk install dari GitHub masing-masing)

---

## 3. Download Binary (Cara Cepat)

Ini cara yang disarankan — tidak perlu build sendiri.

```bash
mkdir -p ~/lnd-nux && cd ~/lnd-nux

# Download binary + checksum (ganti amd64 -> arm64 kalau pakai ARM/Raspberry Pi)
wget https://github.com/Nusanian/lnd/releases/download/v0.0.0-nux-test3/lnd-nux-v0.0.0-nux-test3-linux-amd64.tar.gz
wget https://github.com/Nusanian/lnd/releases/download/v0.0.0-nux-test3/lnd-nux-v0.0.0-nux-test3-linux-amd64.tar.gz.sha256

# Verifikasi checksum (WAJIB — memastikan file tidak korup/rusak)
sha256sum -c lnd-nux-v0.0.0-nux-test3-linux-amd64.tar.gz.sha256
# harus keluar: lnd-nux-v0.0.0-nux-test3-linux-amd64.tar.gz: OK

# Extract (isinya langsung lnd + lncli, tanpa subfolder)
tar -xzf lnd-nux-v0.0.0-nux-test3-linux-amd64.tar.gz

# Cek versi
./lnd --version
# harus keluar: lnd version 0.21.99-beta commit=b9e46ac (atau lebih baru)
```

> **Penting:** jangan rename file `.tar.gz` sebelum verifikasi checksum — nama file harus sama persis seperti di release agar `sha256sum -c` bekerja.

> **Versi minimum:** pakai `v0.0.0-nux-test3` atau lebih baru. Versi ini memuat fix penting: change address otomatis memakai P2WKH (SegWit v0) karena Taproot belum aktif di Nusacoin. Versi lama membuat change P2TR yang tidak bisa di-spend — transaksi funding akan ditolak network secara diam-diam.

---

## 4. Build dari Source (Alternatif)

Hanya kalau kamu mau build sendiri (butuh Go):

```bash
git clone https://github.com/Nusanian/lnd lnd-nux
cd lnd-nux
git checkout nusacoin          # PENTING: jangan pakai master!
go build -o lnd ./cmd/lnd
go build -o lncli ./cmd/lncli
```

---

## 5. Menjalankan nusacoind (Testnet)

`lnd-nux` butuh `nusacoind` yang sudah sync testnet dengan **txindex** dan **ZMQ** aktif.

Buat/edit `~/.nusacoin/nusacoin.conf`:

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

**Yang wajib diperhatikan:**
- `txindex=1` — lnd butuh mencari transaksi lama. (Catatan: `nusacoind` versi ini tidak punya perintah `getindexinfo`, jadi verifikasi via config dan log, bukan via RPC.)
- **Port ZMQ block dan tx HARUS beda** (di contoh: 28332 vs 28333). Kalau disamakan, `lnd` menolak start.
- `[test]` berarti setting ini hanya untuk testnet — mainnet tidak ikut berubah.

Jalankan dan tunggu sync:

```bash
nusacoind -daemon
# Cek progres (1.0 = sudah sync 100%):
nusacoin-cli -rpcuser=nuxrpc -rpcpassword=PASSWORD_KAMU -rpcport=18332 getblockchaininfo | grep verificationprogress
```

> Sync awal testnet bisa makan waktu berjam-jam. `lnd` baru bisa jalan setelah backend 100% sync.

---

## 6. Konfigurasi lnd-nux

Buat file `~/.lnd/lnd.conf`:

```ini
[Application Options]
alias=NamaNodeKamu
color=#00ff00
restlisten=127.0.0.1:8080
rpclisten=127.0.0.1:10009
listen=0.0.0.0:9735

[Nusacoin]
nusacoin.testnet=1
nusacoin.node=bitcoind

[Bitcoind]
bitcoind.rpchost=127.0.0.1:18332
bitcoind.rpcuser=nuxrpc
bitcoind.rpcpass=GANTI_DENGAN_PASSWORD_YANG_KUAT_DAN_ACAK
bitcoind.zmqpubrawblock=tcp://127.0.0.1:28332
bitcoind.zmqpubrawtx=tcp://127.0.0.1:28333

[protocol]
protocol.wumbo-channels=true
```

**Penjelasan per bagian (untuk yang penasaran):**
- `[Application Options]` — `alias` adalah nama tampilan node-mu di jaringan. Tiga port: `10009` (perintah `lncli`), `8080` (API REST), `9735` (komunikasi antar node Lightning).
- `[Nusacoin]` — mengaktifkan chain Nusacoin mode testnet. (Jangan tambahkan `nusacoin.active=1` — itu sintaks lama yang sudah deprecated.)
- `[Bitcoind]` — cara lnd ngobrol ke `nusacoind`. Nama section-nya tetap `[Bitcoind]` walau backend-nya Nusacoin — itu nama modul konektornya, bukan nama koinnya. RPC untuk tanya-jawab, ZMQ untuk dapat kabar block baru secara instan.
- `[protocol]` — mengaktifkan **wumbo channels** (channel besar, lihat bagian 10). **Catatan penting:** opsi yang pakai nama section (seperti `protocol.*` dan `bitcoind.*`) wajib ditulis format **dotted** (`protocol.wumbo-channels`, bukan `wumbo-channels` saja) — kalau tidak, lnd error `unknown option`.

---

## 7. Menjalankan lnd & Membuat Wallet

Jalankan lnd (biarkan jalan — pakai `screen`, `tmux`, atau `nohup` agar tidak mati saat terminal ditutup):

```bash
nohup ~/lnd-nux/lnd > ~/lnd-nux/lnd.log 2>&1 &
```

Di terminal **lain**, buat wallet (hanya sekali, saat pertama kali):

```bash
~/lnd-nux/lncli -n testnet --chain=nusacoin create
```

Ikuti panduannya:
1. Masukkan **password wallet** 2x — catat baik-baik. Lupa password = dana tidak bisa dibuka.
2. Seed mnemonic: jawab `n` untuk bikin seed baru (atau `y` kalau restore dari backup).
3. Cipher seed passphrase: Enter saja (kosongkan) kecuali kamu paham fungsinya.
4. Ketik `YES`. **Simpan 24 kata seed di tempat AMAN, jangan pernah dibagikan ke siapa pun!**

> ⚠️ **Selalu tambahkan `-n testnet --chain=nusacoin`** di setiap perintah `lncli`. Tanpa itu, `lncli` mencari data di folder yang salah dan error.

Setiap kali lnd restart, buka wallet-nya:

```bash
~/lnd-nux/lncli -n testnet --chain=nusacoin unlock
# masukkan password wallet
```

Supaya tidak mengetik panjang tiap kali, tambahkan alias ke `~/.bashrc`:

```bash
alias lncli1='~/lnd-nux/lncli -n testnet --chain=nusacoin'
```

Lalu `source ~/.bashrc`. Mulai sekarang cukup ketik `lncli1 ...`.

---

## 8. Isi Saldo On-Chain

Minta alamat deposit (pakai tipe `p2wkh` — SegWit v0, kompatibel penuh):

```bash
lncli1 newaddress p2wkh
# contoh: tn1q...
```

Kirim **tNUX** (koin testnet, tidak ada nilainya) ke alamat itu, lalu cek:

```bash
lncli1 walletbalance
```

Tunggu minimal **1 konfirmasi** sebelum dipakai. Block Nusacoin ~6 menit.

> **Shortcut (opsional):** difficulty testnet saat ini sangat rendah, jadi kamu bisa mine 1 block sendiri untuk mempercepat konfirmasi:
> ```bash
> nusacoin-cli -testnet generatetoaddress 1 $(nusacoin-cli -testnet getnewaddress)
> ```
> Ini hanya bisa diandalkan selama difficulty testnet masih rendah.

> **Coinbase maturity:** koin hasil mining baru bisa dipakai setelah **10 block** (bukan 100 seperti Bitcoin). Jadi kalau kamu mine sendiri, tunggu 10 block sebelum dananya spendable.

---

## 9. Connect & Buka Channel

**Langkah 1 — Connect (berkenalan):** dapatkan `pubkey@host:port` node tujuan dari komunitas.

```bash
lncli1 connect <pubkey>@<host>:9735
lncli1 listpeers   # verifikasi sudah connect
```

> `connect` hanya berkenalan di jaringan P2P — belum ada dana yang dikunci. `<host>` bisa IP (`203.0.113.10`) atau domain (`node.komunitas.org`).

**Langkah 2 — Open channel (kunci dana):**

```bash
lncli1 openchannel --node_key=<pubkey_tujuan> --local_amt=<jumlah_nusan>
```

Contoh — buka channel 10.000.000 nusan (0,1 NUX):

```bash
lncli1 openchannel --node_key=03abc...def --local_amt=10000000
```

Perintah ini mengembalikan `funding_txid` — ID transaksi yang mengunci dana ke channel. Transaksi ini dibuat **saat** proses open (bukan transaksi yang sudah ada sebelumnya).

**Berapa konfirmasi sampai aktif?** Tergantung ukuran channel (otomatis):

| Ukuran channel | Konfirmasi |
|----------------|------------|
| Kecil (di bawah ~0,016 NUX) | 1 |
| Sedang | 2–3 |
| Besar / wumbo | hingga 6 |

Cek status:

```bash
lncli1 listchannels
# "active": true berarti siap dipakai
```

> **Batas default:** tanpa wumbo, channel maksimal **16.777.215 nusan** (~0,167 NUX). Kalau butuh lebih besar, lihat bagian 10.

---

## 10. Wumbo Channels (Channel Besar)

"Wumbo" = channel jumbo, melebihi batas default. Istilahnya dari Spongebob (artinya "besar") 😄

**Kenapa ada batas?** Di awal Lightning, batas kecil dipasang sebagai pengaman — kalau ada bug, dana yang berisiko tidak terlalu besar. Wumbo adalah cara kedua node sepakat untuk melonggarkan batas itu.

**Cara kerja:** saat open channel, kedua node saling bertukar "feature bit" (tanda "saya support channel besar"). Kalau **keduanya** setuju, limit naik sampai **10 NUX**. Kalau salah satunya tidak setuju, tetap pakai batas default.

**Syarat:** tambahkan ini di `lnd.conf` **kedua** node (sudah termasuk di contoh bagian 6):

```ini
[protocol]
protocol.wumbo-channels=true
```

Lalu restart lnd di kedua sisi.

> Keputusan batas channel untuk **mainnet** (apakah default perlu dinaikkan atau cukup dokumentasikan wumbo) masih didiskusikan — akan diputuskan sebelum mainnet launch.

---

## 11. Membayar via Lightning

**Penerima** bikin invoice:

```bash
lncli1 addinvoice --amt=100000 --memo="kopi"
# catat "payment_request" (diawali lntn...)
```

**Pengirim** membayar:

```bash
lncli1 payinvoice <payment_request>
```

Hasil `SUCCEEDED` = lunas dalam hitungan detik, tanpa menunggu block. Fee antar node biasanya sangat kecil atau nol.

> Prefix invoice testnet Nusacoin adalah `lntn` (`ln` + `t` testnet + `n` nusacoin). Di mainnet nanti menjadi `lnn`.

---

## 12. Menutup Channel

Lihat dulu `channel_point` (format `txid:output_index`):

```bash
lncli1 listchannels
```

**Tutup kooperatif** (disarankan — murah, dana langsung kembali):

```bash
lncli1 closechannel <funding_txid> <output_index>
```

**Force-close** (kalau lawan offline / tidak kooperatif):

```bash
lncli1 closechannel --force <funding_txid> <output_index>
```

> ⚠️ **Force-close mengunci dana sementara** (timelock). Minimal 240 block, maksimal 3360 block. Pada pace desain (6 menit/block) itu ~1 hari sampai ~2 minggu; pada pace nyata saat ini yang lebih lambat, bisa lebih lama. Ini normal sesuai protokol — bukan bug. Jangan force-close kecuali benar-benar perlu.

---

## 13. Node Kedua di Mesin yang Sama (Opsional)

Untuk testing, dua node bisa jalan di 1 VPS (data dan port terpisah):

```bash
# Node 2: lnddir dan port berbeda
nohup ~/lnd-nux/lnd --lnddir=~/.lnd2 --rpcserver=127.0.0.1:10010 > ~/lnd-nux/lnd2.log 2>&1 &
```

Tambahkan alias:

```bash
alias lncli2='~/lnd-nux/lncli --lnddir=~/.lnd2 --rpcserver=127.0.0.1:10010 -n testnet --chain=nusacoin'
```

> Node 2 **wajib** pakai `--rpcserver` eksplisit, karena `lncli` default mencari di port 10009 (milik node 1).

Port P2P node 2 juga harus beda — tambahkan di `~/.lnd2/lnd.conf`:
```ini
[Application Options]
listen=0.0.0.0:9736
rpclisten=127.0.0.1:10010
```

> Kalau membuka banyak channel test dan kena error `Number of pending channels exceed maximum`, tambahkan `--maxpendingchannels=5` saat start lnd (default-nya 1).

---

## 14. Troubleshooting

| Gejala | Penyebab & Solusi |
|--------|-------------------|
| `unable to read macaroon path .../bitcoin/...` | Lupa `--chain=nusacoin`. Selalu sertakan di tiap perintah `lncli` (atau pakai alias). |
| `server is still in the process of starting` | lnd masih sync. Tunggu sampai `getinfo` menunjukkan `"synced_to_chain": true`. |
| `synced_to_chain: false` padahal backend sync | Testnet NUX mining-nya sporadis. lnd menganggap "belum sync" kalau block terakhir lebih tua dari 2 jam. **Solusi:** mine 1 block baru via `generatetoaddress` — langsung jadi `true`. Ini bukan error. |
| lnd tidak mau start | Cek: (1) `nusacoind` sudah 100% sync? (2) Port ZMQ block & tx beda? (3) Lihat log `~/lnd-nux/lnd.log`. |
| `unknown option: wumbo-channels` | Opsi ber-namespace wajib format dotted: `protocol.wumbo-channels=true` (bukan `wumbo-channels=true`). |
| Channel tidak aktif setelah open | Butuh konfirmasi (1–6 block tergantung ukuran, lihat tabel bagian 9). Cek `listchannels`. |
| `funding amount is too large, the max channel size is: 0.16777215 NUX` | Melebihi batas default 16.777.215 nusan. Aktifkan wumbo di kedua node (bagian 10). |
| Payment gagal / `no route` | Pastikan channel `active: true`, peer connect, dan likuiditas cukup di arah yang benar. |
| `Number of pending channels exceed maximum` | Tambahkan `--maxpendingchannels=5` saat start lnd (untuk testing). |
| Funding tx tidak masuk mempool | Pastikan pakai binary `v0.0.0-nux-test3` atau lebih baru (fix P2WKH). Cek dengan `nusacoin-cli -testnet getmempoolentry <txid>`. |

### FAQ

**Q: Node saya saja yang jalan, bisa dipakai?**
A: Daemon dan wallet bisa jalan normal, tapi payment butuh channel, dan channel butuh 2 pihak. Minimal 2 node + 1 channel untuk payment pertama. Kabar baik: 2 node bisa di 1 mesin (bagian 13).

**Q: Apa bedanya `connect` dan `openchannel`?**
A: `connect` = berkenalan di jaringan P2P (belum ada dana). `openchannel` = mengunci dana dan membuka channel pembayaran (dilakukan setelah connect).

**Q: Apakah dana di channel aman kalau node mati?**
A: Ya, selama kamu pegang 24 kata seed. Channel bisa di-force-close dan dana kembali on-chain (dengan timelock, lihat bagian 12).

**Q: Kenapa invoice diawali `lntn` bukan `lnbc`?**
A: `lnbc` untuk Bitcoin. `lntn` = Lightning + testnet + nusacoin. Setiap chain punya prefix sendiri.

---

## 15. Referensi Cepat

| Item | Nilai |
|------|-------|
| Config lnd | `~/.lnd/lnd.conf` |
| Config nusacoind | `~/.nusacoin/nusacoin.conf` |
| Binary | `~/lnd-nux/lnd`, `~/lnd-nux/lncli` |
| Log lnd | `~/lnd-nux/lnd.log` |
| lnd RPC / REST / P2P | 10009 / 8080 / 9735 |
| nusacoind testnet RPC | 18332 |
| nusacoind ZMQ block / tx | 28332 / 28333 (harus beda!) |
| Prefix alamat testnet | `tn1...` |
| Prefix invoice testnet | `lntn...` |
| Satuan terkecil | nusan (1 NUX = 100.000.000 nusan) |
| Block time (desain) | ~6 menit |
| Coinbase maturity | 10 block |
| Max channel default | 16.777.215 nusan |
| Max channel wumbo | 10 NUX (1.000.000.000 nusan) |
| Konfirmasi channel | 1–6 block (skala otomatis) |

---

*Dokumen ini bagian dari proyek **LightningNUX** — branch `nusacoin` di https://github.com/Nusanian/lnd.*
*Terakhir diperbarui: 13 Oktober 2026 (mencakup hasil live test: channel 1 NUX + payment 500.000 nusan sukses di testnet).*
