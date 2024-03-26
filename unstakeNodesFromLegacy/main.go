package main

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/blockchain"
	"github.com/multiversx/mx-sdk-go/blockchain/cryptoProvider"
	"github.com/multiversx/mx-sdk-go/builders"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/multiversx/mx-sdk-go/examples"
	"github.com/multiversx/mx-sdk-go/interactors"
	"github.com/multiversx/mx-sdk-go/workflows"
)

const walletFilename = "./legacyDelegationOwner.pem"
const scAddress = "erd1qqqqqqqqqqqqqpgq97wezxw6l7lgg7k9rxvycrz66vn92ksh2tssxwf7ep"

var (
	suite   = ed25519.NewEd25519()
	keyGen  = signing.NewKeyGenerator(suite)
	log     = logger.GetOrCreate("unstakeNodesFromLegacy")
	blsKeys = []string{
		"48e10d35eee149ca16cd2c1cbb56a22d8c2d6179566f606a87766d68848e440db8a1276661b5664301c7d76625a49c0de25eb5f6695e9097c50ec63697e4ad29c29dba39e7a753fb10f4682eb21356c001d0c9e8085b251c4e2809035e117215",
		"bbbcd9426b45ee0d1cee8455c8d5099da4b460a1616223934dc7ac8b484d8353b051c9025c2585a33070292d6d67680d679e76e79c5ddf70deeba6e429b1fb341d0e037a184d3176875dcd37dd90ccf7dddaa4594a88ae98e7c64dd56b79358c",
	}
)

func main() {
	proxy := createTestnetProxy()

	wallet := interactors.NewWallet()
	skBytes, err := wallet.LoadPrivateKeyFromPemFile(walletFilename)
	if err != nil {
		panic(err)
	}

	// Generate address from private key
	ownerAddress, err := wallet.GetAddressFromPrivateKey(skBytes)
	if err != nil {
		log.Error("unable to load the address from the private key", "error", err)
		return
	}

	holder, _ := cryptoProvider.NewCryptoComponentsHolder(keyGen, skBytes)
	txBuilder, err := builders.NewTxBuilder(cryptoProvider.NewSigner())

	// netConfigs can be used multiple times (for example when sending multiple transactions) as to improve the
	// responsiveness of the system
	netConfigs, err := proxy.GetNetworkConfig(context.Background())
	if err != nil {
		panic(err)
	}

	ti, err := interactors.NewTransactionInteractor(proxy, txBuilder)
	if err != nil {
		log.Error("error creating transaction interactor", "error", err)
		return
	}

	ownerAccount, err := proxy.GetAccount(context.Background(), ownerAddress)
	if err != nil {
		panic(err)
	}

	for idx, blsKey := range blsKeys {
		generateAndSendUnstakeTx(proxy, blsKey, ownerAddress, netConfigs, ti, holder, ownerAccount.Nonce+uint64(idx))
	}

	hashes, err := ti.SendTransactionsAsBunch(context.Background(), 100)
	if err != nil {
		panic(err)
	}

	log.Info("transactions sent", "hashes", hashes)
}

func createTestnetProxy() interactors.Proxy {
	args := blockchain.ArgsProxy{
		ProxyURL:            examples.TestnetGateway,
		Client:              nil,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       false,
		CacheExpirationTime: time.Minute,
		EntityType:          core.Proxy,
	}
	ep, err := blockchain.NewProxy(args)
	if err != nil {
		panic(err)
	}

	return ep
}

func generateAndSendUnstakeTx(
	proxy interactors.Proxy,
	blsKey string,
	ownerAddress core.AddressHandler,
	netConfigs *data.NetworkConfig,
	ti workflows.TransactionInteractor,
	holder core.CryptoComponentsHolder,
	nonce uint64,
) {
	proxyHandler := proxy.(workflows.ProxyHandler)

	tx, _, err := proxyHandler.GetDefaultTransactionArguments(context.Background(), ownerAddress, netConfigs)
	if err != nil {
		panic(err)
	}

	tx.Receiver = scAddress // send to delegation SC
	tx.Value = "0"          // 0 EGLD
	tx.GasLimit = 300000000 // 300 million gas units
	tx.Data = []byte("unStakeNodes@" + blsKey)
	tx.Nonce = nonce

	err = ti.ApplyUserSignature(holder, &tx)
	if err != nil {
		panic(err)
	}
	ti.AddTransaction(&tx)

	log.Info("generated tx", "nonce", tx.Nonce, "sender", tx.Sender, "receiver", tx.Receiver, "data", string(tx.Data))
}
