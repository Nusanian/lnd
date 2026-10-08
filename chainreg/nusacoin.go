package chainreg

import (
	"math/big"
	"time"

	"github.com/btcsuite/btcd/chaincfg/v2"
	"github.com/btcsuite/btcd/chainhash/v2"
	bitcoinWire "github.com/btcsuite/btcd/wire/v2"
	"github.com/lightningnetwork/lnd/keychain"
)

// Nusacoin chain parameters for lnd.
//
// Nusacoin is a Bitcoin Core-based chain (X11 PoW) with SegWit (BIP141),
// CSV (BIP68/112/113) and CLTV (BIP65) active since block 1, which are the
// consensus prerequisites for Lightning. Values below are taken from
// src/chainparams.cpp of the nusacoin codebase.
//
// NOTE: nusacoin uses the *same* genesis block for mainnet, testnet,
// signet and regtest; the networks are distinguished by their P2P magic
// bytes, default ports and bech32 HRPs.

// NusacoinHDCoinType is the BIP44 coin type used by the Nusacoin
// ecosystem (m/44'/662'/... and m/84'/662'/... as documented by the
// Nusacoin web wallet). Test networks use the SLIP-44 testnet coin
// type (1).
const NusacoinHDCoinType uint32 = 662

// Block-time sensitive defaults, scaled from the Bitcoin defaults by the
// block time ratio (10 min Bitcoin vs 6 min Nusacoin) so that the
// *time*-based security properties are preserved.

// DefaultNusacoinTimeLockDelta is the default forwarding CLTV delta for
// Nusacoin channels: 80 Bitcoin blocks (~13.3h) ~= 133 Nusacoin blocks.
const DefaultNusacoinTimeLockDelta = 133

// DefaultNusacoinMaxLocalCSVDelay is the default maximum CSV delay we
// will accept for our own funds: 2016 Bitcoin blocks (2 weeks) ~= 3360
// Nusacoin blocks.
const DefaultNusacoinMaxLocalCSVDelay = 3360

// MinNusacoinRemoteDelay is the minimum CSV delay we require from a
// channel counterparty on unilateral close: 144 Bitcoin blocks (24h)
// ~= 240 Nusacoin blocks.
const MinNusacoinRemoteDelay uint16 = 240

// DefaultNusacoinMinHTLCInMSat / DefaultNusacoinMinHTLCOutMSat mirror the
// Bitcoin defaults; they are value-based, not block-time based.
var (
	DefaultNusacoinMinHTLCInMSat  = DefaultBitcoinMinHTLCInMSat
	DefaultNusacoinMinHTLCOutMSat = DefaultBitcoinMinHTLCOutMSat
	DefaultNusacoinBaseFeeMSat     = DefaultBitcoinBaseFeeMSat
	DefaultNusacoinFeeRate         = DefaultBitcoinFeeRate
)

var (
	// nusacoinGenesisHash is the genesis block hash shared by all
	// Nusacoin networks (mainnet/testnet/signet/regtest).
	nusacoinGenesisHash = newHashFromStrMust(
		"0000017123d6f996589bc2e58bb5502218012ac7f527ab599a3be84c1951c669",
	)

	// nusacoinGenesisMerkleRoot is the merkle root of the shared
	// genesis block.
	nusacoinGenesisMerkleRoot = newHashFromStrMust(
		"660edadda7d259dff20e3d21fd57165417b0ce9bd2d40a28ae437638aa61fd9b",
	)

	// nusacoinGenesisBlock is the shared Nusacoin genesis block:
	// version 1, time 1764547439, bits 0x1e0ffff0, nonce 29271886.
	nusacoinGenesisBlock = bitcoinWire.MsgBlock{
		Header: bitcoinWire.BlockHeader{
			Version:    1,
			PrevBlock:  chainhash.Hash{},
			MerkleRoot: *nusacoinGenesisMerkleRoot,
			Timestamp:  time.Unix(1764547439, 0),
			Bits:       0x1e0ffff0,
			Nonce:      29271886,
		},
	}

	// nusacoinPowLimit is the highest allowed proof of work value,
	// derived from the compact genesis bits 0x1e0ffff0.
	nusacoinPowLimit = new(big.Int).Lsh(
		big.NewInt(0x0ffff0), 8*(0x1e-3),
	)
)

func newHashFromStrMust(s string) *chainhash.Hash {
	h, err := chainhash.NewHashFromStr(s)
	if err != nil {
		panic(err)
	}
	return h
}

// nusacoinDeployments marks CSV and SegWit as always active (both are
// consensus-active since block 1 on every Nusacoin network).
var nusacoinDeployments = [chaincfg.DefinedDeployments]chaincfg.ConsensusDeployment{
	chaincfg.DeploymentTestDummy: {
		BitNumber: 28,
		DeploymentStarter: chaincfg.NewMedianTimeDeploymentStarter(
			time.Time{},
		),
		DeploymentEnder: chaincfg.NewMedianTimeDeploymentEnder(
			time.Time{},
		),
	},
	chaincfg.DeploymentCSV: {
		BitNumber:          0,
		AlwaysActiveHeight: 1,
	},
	chaincfg.DeploymentSegwit: {
		BitNumber:          1,
		AlwaysActiveHeight: 1,
	},
}

// NusacoinMainNetParams contains parameters specific to the Nusacoin
// main network.
var NusacoinMainNetParams = BitcoinNetParams{
	Params: &chaincfg.Params{
		Name:        "mainnet",
		Net:         bitcoinWire.BitcoinNet(0x4153554e), // "NUSA"
		DefaultPort: "28573",
		DNSSeeds:    []chaincfg.DNSSeed{},

		GenesisBlock: &nusacoinGenesisBlock,
		GenesisHash:  nusacoinGenesisHash,
		PowLimit:     nusacoinPowLimit,
		PowLimitBits: 0x1e0ffff0,

		BIP0034Height: 1,
		BIP0065Height: 1,
		BIP0066Height: 1,

		CoinbaseMaturity:         100,
		SubsidyReductionInterval: 300000,
		TargetTimespan:           time.Hour * 24 * 14, // 3360 * 360s
		TargetTimePerBlock:       time.Minute * 6,
		RetargetAdjustmentFactor: 4,

		RuleChangeActivationThreshold: 3192, // 95% of 3360
		MinerConfirmationWindow:       3360,
		Deployments:                   nusacoinDeployments,

		RelayNonStdTxs: false,

		Bech32HRPSegwit: "nu",

		PubKeyHashAddrID: 0x35, // addresses start with 'N'
		ScriptHashAddrID: 0x4b, // addresses start with 'X'
		PrivateKeyID:     0x57,

		HDPrivateKeyID: [4]byte{0x04, 0x88, 0xb2, 0x1e},
		HDPublicKeyID:  [4]byte{0x04, 0x88, 0xad, 0xe4},
		HDCoinType:     NusacoinHDCoinType,
	},
	RPCPort:  "28832",
	CoinType: NusacoinHDCoinType,
}

// NusacoinTestNetParams contains parameters specific to the Nusacoin
// test network.
var NusacoinTestNetParams = BitcoinNetParams{
	Params: &chaincfg.Params{
		Name:        "testnet",
		Net:         bitcoinWire.BitcoinNet(0x0709110b), // 0b110907
		DefaultPort: "38573",
		DNSSeeds:    []chaincfg.DNSSeed{},

		GenesisBlock: &nusacoinGenesisBlock,
		GenesisHash:  nusacoinGenesisHash,
		PowLimit:     nusacoinPowLimit,
		PowLimitBits: 0x1e0ffff0,

		BIP0034Height: 1,
		BIP0065Height: 1,
		BIP0066Height: 1,

		CoinbaseMaturity:         100,
		SubsidyReductionInterval: 300000,
		TargetTimespan:           time.Hour * 24 * 14,
		TargetTimePerBlock:       time.Minute * 6,
		RetargetAdjustmentFactor: 4,

		RuleChangeActivationThreshold: 3192,
		MinerConfirmationWindow:       3360,
		Deployments:                   nusacoinDeployments,

		RelayNonStdTxs: true,

		Bech32HRPSegwit: "tn",

		PubKeyHashAddrID: 111, // 0x6f
		ScriptHashAddrID: 196, // 0xc4
		PrivateKeyID:     239, // 0xef

		HDPrivateKeyID: [4]byte{0x04, 0x35, 0x83, 0x94},
		HDPublicKeyID:  [4]byte{0x04, 0x35, 0x87, 0xcf},
		HDCoinType:     keychain.CoinTypeTestnet,
	},
	RPCPort:  "18332",
	CoinType: keychain.CoinTypeTestnet,
}

// NusacoinRegTestNetParams contains parameters specific to the Nusacoin
// regression test network. Note the bech32 HRP is "nirt" (not "tn").
var NusacoinRegTestNetParams = BitcoinNetParams{
	Params: &chaincfg.Params{
		Name:        "regtest",
		Net:         bitcoinWire.BitcoinNet(0xdab5bffa), // fabfb5da
		DefaultPort: "38673",
		DNSSeeds:    []chaincfg.DNSSeed{},

		GenesisBlock: &nusacoinGenesisBlock,
		GenesisHash:  nusacoinGenesisHash,
		PowLimit:     nusacoinPowLimit,
		PowLimitBits: 0x1e0ffff0,

		// No retargeting on regtest; blocks are generated on demand.
		PoWNoRetargeting: true,

		BIP0034Height: 1,
		BIP0065Height: 1,
		BIP0066Height: 1,

		CoinbaseMaturity:         100,
		SubsidyReductionInterval: 300000,
		TargetTimespan:           time.Hour * 24 * 14,
		TargetTimePerBlock:       time.Minute * 6,
		RetargetAdjustmentFactor: 4,

		RuleChangeActivationThreshold: 108, // 75% of 144
		MinerConfirmationWindow:       144,
		Deployments:                   nusacoinDeployments,

		RelayNonStdTxs: true,

		Bech32HRPSegwit: "nirt",

		PubKeyHashAddrID: 111,
		ScriptHashAddrID: 196,
		PrivateKeyID:     239,

		HDPrivateKeyID: [4]byte{0x04, 0x35, 0x83, 0x94},
		HDPublicKeyID:  [4]byte{0x04, 0x35, 0x87, 0xcf},
		HDCoinType:     keychain.CoinTypeTestnet,
	},
	RPCPort:  "18443",
	CoinType: keychain.CoinTypeTestnet,
}

// NusacoinSigNetParams contains parameters specific to the Nusacoin
// signet network.
//
// TODO: nusacoin derives its signet P2P magic from the signet challenge
// at runtime; the magic below is a placeholder and must be replaced with
// the value of the target signet deployment before use.
var NusacoinSigNetParams = BitcoinNetParams{
	Params: &chaincfg.Params{
		Name:        "signet",
		Net:         bitcoinWire.BitcoinNet(0x2c4d7b9a), // placeholder
		DefaultPort: "38773",
		DNSSeeds:    []chaincfg.DNSSeed{},

		GenesisBlock: &nusacoinGenesisBlock,
		GenesisHash:  nusacoinGenesisHash,
		PowLimit:     nusacoinPowLimit,
		PowLimitBits: 0x1e0ffff0,

		BIP0034Height: 1,
		BIP0065Height: 1,
		BIP0066Height: 1,

		CoinbaseMaturity:         100,
		SubsidyReductionInterval: 300000,
		TargetTimespan:           time.Hour * 24 * 14,
		TargetTimePerBlock:       time.Minute * 6,
		RetargetAdjustmentFactor: 4,

		RuleChangeActivationThreshold: 3192,
		MinerConfirmationWindow:       3360,
		Deployments:                   nusacoinDeployments,

		RelayNonStdTxs: true,

		Bech32HRPSegwit: "tn",

		PubKeyHashAddrID: 111,
		ScriptHashAddrID: 196,
		PrivateKeyID:     239,

		HDPrivateKeyID: [4]byte{0x04, 0x35, 0x83, 0x94},
		HDPublicKeyID:  [4]byte{0x04, 0x35, 0x87, 0xcf},
		HDCoinType:     keychain.CoinTypeTestnet,
	},
	RPCPort:  "38332",
	CoinType: keychain.CoinTypeTestnet,
}

// IsNusacoin returns true if the given params are Nusacoin parameters.
func IsNusacoin(params *BitcoinNetParams) bool {
	return params == &NusacoinMainNetParams ||
		params == &NusacoinTestNetParams ||
		params == &NusacoinRegTestNetParams ||
		params == &NusacoinSigNetParams
}
