package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path"
	"strings"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl/singlesig"
	"github.com/multiversx/mx-chain-go/vm"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/blockchain"
	"github.com/multiversx/mx-sdk-go/blockchain/cryptoProvider"
	"github.com/multiversx/mx-sdk-go/builders"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/multiversx/mx-sdk-go/examples"
	"github.com/multiversx/mx-sdk-go/interactors"
	"github.com/multiversx/mx-sdk-go/workflows"
)

const keysDir = `/home/jules01/keys`
const idxEGLD = 4
const walletKeyFilename = "wallet.pem"
const validatorsKeysFilename = "all.pem"
const sponsorWalletFilename = "sponsor.pem"
const gateway = examples.TestnetGateway // for local testnet, use "http://127.0.0.1:7950"
const dataByteGasLimit = 1500
const maxTimeoutForTransactionToComplete = time.Minute * 2
const blockTime = time.Second * 6
const stakeGasPerNode = 6000000
const baseStakeGas = 50000000
const makeContractGas = 510000000
const feeString = "0320" // 8.00%

var oneELGD = big.NewInt(1000000000000000000)
var stakeForOneNode = big.NewInt(0).Mul(big.NewInt(2500), oneELGD)
var log = logger.GetOrCreate("manualStaking")
var walletSuite = ed25519.NewEd25519()
var blsSuite = mcl.NewSuiteBLS12()
var walletKeyGen = signing.NewKeyGenerator(walletSuite)
var blsKeyGen = signing.NewKeyGenerator(blsSuite)
var blsSingleSigner = singlesig.NewBlsSigner()

type stakeInfo struct {
	walletKey      *walletKeyAddress
	blsPrivateKeys [][]byte
	blsPublicKeys  []string
	stakeValue     *big.Int
}

type walletKeyAddress struct {
	skBytes       []byte
	address       sdkCore.AddressHandler
	bech32Address string
}

type processStatusProxy interface {
	ProcessTransactionStatus(ctx context.Context, hexTxHash string) (transaction.TxStatus, error)
}

func main() {
	readStakeInfo := readDirStakeInfo()

	sum := big.NewInt(0)
	for _, si := range readStakeInfo {
		sum.Add(sum, si.stakeValue)
	}
	log.Info("read stake info", "num accounts", len(readStakeInfo), "total sum", sum.String())

	proxy := createTestnetProxy()
	sponsorWalletKeyAddress := loadWalletKeyAddress(sponsorWalletFilename)
	account, err := proxy.GetAccount(context.Background(), sponsorWalletKeyAddress.address)
	requireNilErr(err)

	log.Info("sponsor account", "address", sponsorWalletKeyAddress.bech32Address, "balance", account.Balance)

	netConfigs, err := proxy.GetNetworkConfig(context.Background())
	requireNilErr(err)

	for _, si := range readStakeInfo {
		processStakeInfo(si, proxy, sponsorWalletKeyAddress, netConfigs)
	}
}

func loadWalletKeyAddress(filename string) *walletKeyAddress {
	wallet := interactors.NewWallet()
	skBytes, err := wallet.LoadPrivateKeyFromPemFile(filename)
	requireNilErr(err)

	address, err := wallet.GetAddressFromPrivateKey(skBytes)
	requireNilErr(err)

	bech32Address, err := address.AddressAsBech32String()
	requireNilErr(err)

	return &walletKeyAddress{
		skBytes:       skBytes,
		address:       address,
		bech32Address: bech32Address,
	}
}

func readDirStakeInfo() []*stakeInfo {
	entries, err := os.ReadDir(keysDir)
	requireNilErr(err)

	readStakeInfo := make([]*stakeInfo, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		si := parseStakeDir(entry.Name())
		if si == nil {
			continue
		}

		readStakeInfo = append(readStakeInfo, si)
	}

	return readStakeInfo
}

func parseStakeDir(dirName string) *stakeInfo {
	dirPath := path.Join(keysDir, dirName)
	splt := strings.Split(dirName, " ")
	if len(splt) < 5 {
		return nil
	}

	egldString := splt[idxEGLD]
	val, _ := big.NewInt(0).SetString(egldString, 10)
	val.Mul(val, oneELGD)

	walletKey := loadWalletKeyAddress(path.Join(dirPath, walletKeyFilename))

	privateKeysBytes, publicKeys, err := core.LoadAllKeysFromPemFile(path.Join(dirPath, validatorsKeysFilename))
	requireNilErr(err)

	log.Info("loaded data", "path", dirPath, "value", val.String(), "num BLS keys", len(publicKeys))

	return &stakeInfo{
		walletKey:      walletKey,
		blsPrivateKeys: privateKeysBytes,
		blsPublicKeys:  publicKeys,
		stakeValue:     val,
	}
}

func processStakeInfo(si *stakeInfo, proxy interactors.Proxy, sponsorWallet *walletKeyAddress, netConfig *data.NetworkConfig) {
	log.Info("")
	log.Info("############### processing for " + si.walletKey.bech32Address + " ###############")
	processMint(si, proxy, sponsorWallet, netConfig)
	processStake(si, proxy, netConfig)
	makeDelegationContract(si, proxy, netConfig)
}

func processMint(si *stakeInfo, proxy interactors.Proxy, sponsorWallet *walletKeyAddress, netConfig *data.NetworkConfig) {
	valueToMint := big.NewInt(0).Add(si.stakeValue, oneELGD)
	log.Info("minting account", "from", sponsorWallet.bech32Address, "to", si.walletKey.bech32Address, "value", valueToMint.String())
	holder, _ := cryptoProvider.NewCryptoComponentsHolder(walletKeyGen, sponsorWallet.skBytes)
	txBuilder, err := builders.NewTxBuilder(cryptoProvider.NewSigner())
	requireNilErr(err)

	ti, err := interactors.NewTransactionInteractor(proxy, txBuilder)
	requireNilErr(err)

	account, err := proxy.GetAccount(context.Background(), sponsorWallet.address)
	requireNilErr(err)

	proxyHandler := proxy.(workflows.ProxyHandler)
	tx, _, err := proxyHandler.GetDefaultTransactionArguments(context.Background(), sponsorWallet.address, netConfig)
	requireNilErr(err)

	tx.Receiver = si.walletKey.bech32Address
	tx.Value = valueToMint.String()
	tx.GasLimit = 50000
	tx.Data = []byte("initial mint")
	tx.GasLimit += uint64(dataByteGasLimit * len(tx.Data))
	tx.Nonce = account.Nonce

	err = ti.ApplyUserSignature(holder, &tx)
	requireNilErr(err)

	ti.AddTransaction(&tx)
	hash, err := ti.SendTransactionsAsBunch(context.Background(), 1)
	requireNilErr(err)

	log.Info("generated & sent tx",
		"hash", hash[0],
		"nonce", tx.Nonce,
		"sender", tx.Sender,
		"receiver", tx.Receiver,
		"data", string(tx.Data))

	waitForTransactionToCompleteSuccessfully(proxy, hash[0])
}

func processStake(si *stakeInfo, proxy interactors.Proxy, netConfig *data.NetworkConfig) {
	log.Info("stake keys", "owner", si.walletKey.bech32Address, "num keys", len(si.blsPublicKeys), "stake value", si.stakeValue.String())
	holder, _ := cryptoProvider.NewCryptoComponentsHolder(walletKeyGen, si.walletKey.skBytes)
	txBuilder, err := builders.NewTxBuilder(cryptoProvider.NewSigner())
	requireNilErr(err)

	ti, err := interactors.NewTransactionInteractor(proxy, txBuilder)
	requireNilErr(err)

	proxyHandler := proxy.(workflows.ProxyHandler)

	var currentTx *transaction.FrontendTransaction
	validatorAddress := data.NewAddressFromBytes(vm.ValidatorSCAddress)
	numStake := 0
	totalStakedValue := big.NewInt(0)

	account, errGet := proxy.GetAccount(context.Background(), si.walletKey.address)
	requireNilErr(errGet)
	nonce := account.Nonce

	for blsIndex := 0; blsIndex < len(si.blsPublicKeys); blsIndex++ {
		if currentTx == nil {
			tx, _, errGetArgs := proxyHandler.GetDefaultTransactionArguments(context.Background(), si.walletKey.address, netConfig)
			requireNilErr(errGetArgs)

			tx.Receiver, _ = validatorAddress.AddressAsBech32String()
			tx.GasLimit = baseStakeGas
			tx.Nonce = nonce

			currentTx = &tx
		}

		decodedSk, errDecode := hex.DecodeString(string(si.blsPrivateKeys[blsIndex]))
		requireNilErr(errDecode)

		blsKey, errConvert := blsKeyGen.PrivateKeyFromByteArray(decodedSk)
		requireNilErr(errConvert)

		hexSig, errSig := blsSingleSigner.Sign(blsKey, si.walletKey.address.AddressBytes())
		requireNilErr(errSig)

		currentTx.Data = append(currentTx.Data, []byte(fmt.Sprintf("@%s@%x", si.blsPublicKeys[blsIndex], hexSig))...)
		currentTx.GasLimit += stakeGasPerNode
		numStake++

		if blsIndex%50 == 0 && blsIndex+1 < len(si.blsPublicKeys) && blsIndex > 0 {
			stakeValue := big.NewInt(0).Mul(big.NewInt(int64(numStake)), stakeForOneNode)
			totalStakedValue.Add(totalStakedValue, stakeValue)
			currentTx.Value = stakeValue.String()
			currentTx.Data = []byte(fmt.Sprintf("stake@%x", big.NewInt(int64(numStake)).Bytes()) + string(currentTx.Data))

			err = ti.ApplyUserSignature(holder, currentTx)
			requireNilErr(err)

			ti.AddTransaction(currentTx)

			log.Info("generated stake tx",
				"nonce", currentTx.Nonce,
				"value", currentTx.Value,
				"gasLimit", currentTx.GasLimit,
				"sender", currentTx.Sender,
				"receiver", currentTx.Receiver,
				"data", string(currentTx.Data))

			currentTx = nil
			nonce++
			numStake = 0
		}
	}

	if currentTx != nil {
		finalStakeValue := big.NewInt(0).Set(si.stakeValue)
		finalStakeValue.Sub(finalStakeValue, totalStakedValue)
		currentTx.Value = finalStakeValue.String()
		currentTx.Data = []byte(fmt.Sprintf("stake@%x", big.NewInt(int64(numStake)).Bytes()) + string(currentTx.Data))

		err = ti.ApplyUserSignature(holder, currentTx)
		requireNilErr(err)

		ti.AddTransaction(currentTx)

		log.Info("generated last stake tx",
			"nonce", currentTx.Nonce,
			"value", currentTx.Value,
			"gasLimit", currentTx.GasLimit,
			"sender", currentTx.Sender,
			"receiver", currentTx.Receiver,
			"data", string(currentTx.Data))

		currentTx = nil
		nonce++
		numStake = 0
	}

	txHashes, err := ti.SendTransactionsAsBunch(context.Background(), 100)
	requireNilErr(err)
	log.Info("sent transactions as bunch", "tx hashes", txHashes)

	waitForTransactionsToCompleteSuccessfully(proxy, txHashes)
}

func makeDelegationContract(si *stakeInfo, proxy interactors.Proxy, netConfig *data.NetworkConfig) {
	log.Info("make delegation contract", "owner", si.walletKey.bech32Address)
	holder, _ := cryptoProvider.NewCryptoComponentsHolder(walletKeyGen, si.walletKey.skBytes)
	txBuilder, err := builders.NewTxBuilder(cryptoProvider.NewSigner())
	requireNilErr(err)

	ti, err := interactors.NewTransactionInteractor(proxy, txBuilder)
	requireNilErr(err)

	proxyHandler := proxy.(workflows.ProxyHandler)

	tx, _, errGetArgs := proxyHandler.GetDefaultTransactionArguments(context.Background(), si.walletKey.address, netConfig)
	requireNilErr(errGetArgs)

	account, errGet := proxy.GetAccount(context.Background(), si.walletKey.address)
	requireNilErr(errGet)

	delegationManagerAddress := data.NewAddressFromBytes(vm.DelegationManagerSCAddress)

	tx.Nonce = account.Nonce
	tx.GasLimit = makeContractGas
	tx.Value = "0"
	tx.Receiver, _ = delegationManagerAddress.AddressAsBech32String()
	delegationCap := big.NewInt(0).Set(si.stakeValue).Bytes()
	tx.Data = []byte(fmt.Sprintf("makeNewContractFromValidatorData@%x@%s", delegationCap, feeString))

	err = ti.ApplyUserSignature(holder, &tx)
	requireNilErr(err)

	ti.AddTransaction(&tx)
	hash, err := ti.SendTransactionsAsBunch(context.Background(), 1)
	requireNilErr(err)

	log.Info("generated & sent makeNewContractFromValidatorData tx",
		"hash", hash[0],
		"nonce", tx.Nonce,
		"sender", tx.Sender,
		"receiver", tx.Receiver,
		"data", string(tx.Data))

	waitForTransactionToCompleteSuccessfully(proxy, hash[0])
}

func waitForTransactionToCompleteSuccessfully(proxy interactors.Proxy, hexTxHash string) {
	processStatusProxyInstance := proxy.(processStatusProxy)

	ctx, cancelFunc := context.WithTimeout(context.Background(), maxTimeoutForTransactionToComplete)
	defer cancelFunc()

	for {
		status, err := processStatusProxyInstance.ProcessTransactionStatus(ctx, hexTxHash)
		if err != nil {
			log.Error("error getting transaction status", "tx hash", hexTxHash, "error", err)
			panic(err)
		}

		if status == transaction.TxStatusSuccess {
			return
		}
		if status == transaction.TxStatusPending {
			time.Sleep(blockTime)
			continue
		}

		log.Info("transaction failed", "tx hash", hexTxHash)
		panic("transaction failed")
	}
}

func waitForTransactionsToCompleteSuccessfully(proxy interactors.Proxy, hexTxHashes []string) {
	for _, txHash := range hexTxHashes {
		waitForTransactionToCompleteSuccessfully(proxy, txHash)
	}
}

func createTestnetProxy() interactors.Proxy {
	args := blockchain.ArgsProxy{
		ProxyURL:            gateway,
		Client:              nil,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       false,
		CacheExpirationTime: time.Minute,
		EntityType:          sdkCore.Proxy,
	}
	ep, err := blockchain.NewProxy(args)
	if err != nil {
		panic(err)
	}

	return ep
}

func requireNilErr(err error) {
	if err == nil {
		return
	}

	panic(err)
}
