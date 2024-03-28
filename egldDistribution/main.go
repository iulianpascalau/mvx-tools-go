package main

import (
	"context"
	"fmt"
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

const walletFilename = "./erd1q2yzhcy8nwq778v23j7hgdcnsa4pmlwjl0jwr9v86gyff4vr3sdqyyg49s.pem"

var (
	dataByteGasLimit = 1500
	value            = "20000000000000000000000"
	suite            = ed25519.NewEd25519()
	keyGen           = signing.NewKeyGenerator(suite)
	log              = logger.GetOrCreate("unstakeNodesFromLegacy")
	walletAddresses  = []string{
		"erd1q2yzhcy8nwq778v23j7hgdcnsa4pmlwjl0jwr9v86gyff4vr3sdqyyg49s",
		"erd186eyn6x82nn7sh6gdswku69ge9f6hqntny96fdu8etasl0j85h0qq4yeja",
		"erd1rj2k9sum6k32mn6nmnj5rm0cne2afnaflq6afs8zwv9yma4tgqkqk84e3c",
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

	for idx, walletAddress := range walletAddresses {
		generateAndSendMintEgldTx(proxy, walletAddress, ownerAddress, netConfigs, ti, holder, ownerAccount.Nonce+uint64(idx), idx)
		time.Sleep(time.Second)
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

func generateAndSendMintEgldTx(
	proxy interactors.Proxy,
	walletAddress string,
	ownerAddress core.AddressHandler,
	netConfigs *data.NetworkConfig,
	ti workflows.TransactionInteractor,
	holder core.CryptoComponentsHolder,
	nonce uint64,
	index int,
) {
	proxyHandler := proxy.(workflows.ProxyHandler)

	tx, _, err := proxyHandler.GetDefaultTransactionArguments(context.Background(), ownerAddress, netConfigs)
	if err != nil {
		panic(err)
	}

	tx.Receiver = walletAddress
	tx.Value = value
	tx.GasLimit = 50000
	tx.Data = []byte(fmt.Sprintf("🥩 #%d - Battle of Stakes testing campaign", index))
	tx.GasLimit += uint64(dataByteGasLimit * len(tx.Data))
	tx.Nonce = nonce

	err = ti.ApplyUserSignature(holder, &tx)
	if err != nil {
		panic(err)
	}
	ti.AddTransaction(&tx)

	log.Info("generated tx", "nonce", tx.Nonce, "sender", tx.Sender, "receiver", tx.Receiver, "data", string(tx.Data))
}
