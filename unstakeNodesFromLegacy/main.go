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
		"6d2ff6b4f75d793428c213af838b33b0f12775dd06a80cbe6e32105ccf0933f9d71fd66781eb102b8a3457248b1951089a5ce35789e0dfc446e1b9db31df9524e2873ea4f27ad246c11ad71701c1075a6474418604be04d9cf587c65df857e8b",
		"91f0c73e0353f1cdcb7d4907af7705399d761307f3dc30c2c3b242cb8d05d464d9c239edddafd6e890d873debf36a5096a79eb70b245dd2d1174544e72532fc96aff49049fd7148f60d33dffc0f510a47ca2ee6ed13e313341d3d2ed970d9416",
		"167429aa3bd30c6ae775594ab6aac4115842527b9160e6806f9d6774bf8e270537326584caf1e6b5013f1d81d24ba60d31f7e5fc7acddfe3ce7ba58f149c3f41773120855a8aafaf61bd64f68e18d8f4d94c953d01b14c04120eb44cdd76a002",
		"120ae67ca7149aa3e65954efa79bc275296f643f492df1f506ce26b10f11a5976ad8b3d8345148d760d663046edd050a99f32147a517f456a6f1986f098882d764ec49e32e268bb11e00abb01a27df6119e9a59593f6ec5193aaed5e2b023104",
		"210f0aa753925b5c4a722c7651e4936250d79e2d0f9741881ba1e7d0f9dd9b02631dcba03d9bdff2e87f1ff35ed0930da1b8391fa3ddcba8e13fb14cbc250815f9d957a4e7e0565afb9792e79058b4a30f26b63c2b06afc6e4caaa0f63adf18c",
		"42ac3a28a4e01843dafea2da7abc316cccc68c41328c2a0d5e2f5621bbb569a4520cf4b2ed5f1034ed71960f9feab506d6088ef7baffe2c9265f5d5ae3829a19692b051ea30c2fbdccde9b9a8d3cf8053d728003306e5b1008b104ae66c4f501",
		"f4d566d1828f30b2bd02f9e68affb70c21f548913efb0792100fb64c304774e4eb85a2bb704a9afaf8acc5f7fc32410099a28e3b4c349334eae6ddf60462ef984f2c32717cc57c90e4baf02e79b5cb2d62588cc1bacb2730c9280585d0fc050b",
		"2392887685fde193d4a52debfa9f9dff3989c97354390d93ede030f315614671cc7387f32c13527792d5bffe96dd9b1814ffa782ba7807e2b2dc5072cf16e47157c50eca02f16b0116cc91186da87cd5f3f7f775d3fdbb5333a761b9b4b24987",
		"8ad30a6e79fb764262a58c93785b853f60d28832f6de391e5a626040829aa406d050eed9291c2ccf4bc573f0999ddb093e63cb55a8c3f68e3c6dfca29b1a1faa46e944d9493853455792bf4c879ff3e6f7a7232be848034b94d92cd324611f99",
		"12f7d2fe149c5c00da3edb576f6883b1dbb17896bdf91080511cd4973de6f1bc4a438d1055f871bf4992eeac2c845a0d98f609a167765af018ebb066d717d306d9c39752aada92b789c39a6bb362697d62603936671cc4d4d050782e36ab6987",
		"fd55e9c30664e436d4d5d0033a62a2f132f39a55505a720c279ff5caa09d4a78c20d0e359be3260e7cd1c7c82f946300126869bbbf258fa8b1a13f380013f3a1481e219e7f31011124ce6855e806aa4362e8eb5cedb91c0997f642550a62ff80",
		"8365bc2034e274d2f745619db7a15632f38b93ccde21c039eb724d2934bcd6c7c28d486d1f3085337eef298444359e0c01e93f651b63cecc98ad6599d8ea6d9af920bfb12e8af44efb5f4a09e35e5970a9d541d6e121d12a453d872174f7ba19",
		"59465559fa112c3b00b385cfc151bd0eed631cb2ab763c3269a353ae732d06322e622938d2a03f76555c40b1833efd00fc91a1a4ccd5878c2d62434e6fbe8f0736cfd711bca49487a453daeece61437fca2a25dceb552025f72d39dce0ac520e",
		"4de9ea57db89fe8fff38bb951e44d3079fe6ef0afaecc2e310aa31267fba9fe00dd1a1c8c8f1fb3a5fe1cd781e53c60ee0f50cb3dd8a2cccd87dd3edc2d47bcc00f9f7db5c7d4a1c477a85464f92d1225e8767496b32e321a5f68b744c2b7817",
		"2517429521cb7e1b0a2ff3f2b4dc46227f4508077ea12cac18eca246a525a57f92f53da3dc1f29374d319df77011b319089a7d49c06979d9f73bfd9836ff4488cbb355ec3c843e3c43b84daed2709fee4a79a9a15bd0a9e668dbcd79011a7882",
		"0244df2ba8f1b026716bd66ff9e574e5ce727c0809fcc9b5a0288be8548979faf68279a7e5dd7d4113e7753397dffe05d9a34de3d105dd62388fdc346a8a177c8789602ea76c659b16b05ebdb4f336521a0289810de785e0f197adeafa36248c",
		"c61648660c715365703d77328fda26b2a1407d9dbb5ea372bd32611c32a85d1f447c4b3f2ef89fd2f517181510b68c0a44f198330485086e58fc41b99351c40dca5974aeb1f4dd79d9b5cd5cd4d4c3948da49cbb0fbb39a74770eed049314014",
		"af84cc87b3058927293cc1d44f2b250e7a1c85441a20540d099d3b2256e01639b6b37eab9c6d24b175fbb7571a5c1d170d9f72d7cdaeb4ad089118df5ceff4a0c675bf27e9bcb510b0978664c811bbee7dc76378a56975d4317ccfa355f76107",
		"13be7cfe593ab3290dacd3e6607b372812eae0cb7cc9a5a3f2f78bb2be217d709fce8fdf05caa0e8e01bd2481c8b970ec0b5b3633f9cbe01a2555fd7728e0c9e3d766825cde5af62d4b65e14bedfc6d72a7d8fd8d5b79bdef71baec6dfce7588",
		"5f670ceff48b8385871f51a668cc9832619bd26768b48dbf9fe0d6c0c11d88515ffa944158d9769f1493a5d732fc0a0503b3c5b541f2d79036d19b5d06a74ba9b51c6924200bb8b6add99652ce177847e29e1b8efe073f89a1a67d3c27594e93",
		"54d3f4df1d6a734e66898c7d8f4e15571e0abf1c54b68a8bb06dde243d3c792502a3eefa0a1a75cd57fe38d1da442f023749a671d220ed35057d6b57239b0abc01ba11bd1af44310fa6636be4d19154fca53fbaecb01d9421a74f57b362b9694",
		"e288f537a7fb3ffce3c6f533f54baaaa88ae269c2c0b45eff39dc486386daf922914a598ddaa76128fabefd1d637e211a82f1b4a0ab4dd68ae6794dbec1ccd5555af7111c436f4744b7e75bc624909f43f1519ed7db4f8368221a55d6943e990",
		"24c33203c8e377a282a1b18bdcef1f2e430781ccc06df6cb5d7d1d9eb40736a50d36be8d91db06c9ba9bb372e8139d0f6fdcb92b0db13ffbcbb603d2b23ed065634070a6a1d299a43af0acaceea5c46852570badecfcf31b6340ba0209231319",
		"412c23ac3f6cb214d04e95f2a306c6450f0c104ad7ddf530b072539f922555c6d0c83cb02b38b055b66167e1f4d7c3135df7a8aba8e5213385390fc2d1df1f356420245d83bd5d94b3eeefcb42d4d0e0a9437fd07af8480b562f2c3befd58295",
		"bc867da379607655140f1408195888602dc4bc04d9921c38561770badc7abd564a9d16dc3b70c180eb4c6e1d580db5074c345be59e8f6d78f1a216b853124de3c85dabf1845add2decc4ea7ff9de421749494c0a13f7fd296977f7128944f716",
		"94a2db1972ce35c3f7757c60e2f1cceb0c97ca53fa8400aa30f6737a3449e9c3cbc35cd290f811b363e33a7b45051f09d999116932a4363e15ed6831d0b58114d1639f94ee90d19fbc8759807884233fd94c83a651a6363ff938853d87f4b495",
		"c1f19998feb73365d543e9481c6b0d2d388a0a26dbb3424d397d363a9670353b1282efec80fa71f5ebf60805fadfe50755417248d8005d522ac20223d220b76cf4977bd5292f56f3debb58bd1a1bf33a473368919590779584e129d47c47390a",
		"6551d0e3d3c753a62e4ab4eb20224585a9975a51762cd1e416f7ab0bd792b07c4a52f0689f7e13302b76e05a95d8ef0e834386e3bfe1794a22a7fc69da75bafed843e748233e6a318dc6ca85f413cdf9af858ef5cc8e3f5e8f1817516fb16e05",
		"b704afee5d0d0612a3485079aa948130a53ea7f83f787d9433b4dddae7d88c15955d9c6431d99dd6e6361b6c6d383f05987a3dbb5c6c11bd897b71a46b1af97a8235001507950bbae882379a378c7312d1130beece6e2918cc0eb4c28e5df787",
		"cb4f6eb1f8f7b4e3c6a57344a5f82f4f214d10168946af210e133413d4a1ab27fd7fd4f53749bcab195d0c412f16ca08a390559acd6b6330be05dc43304cc65846e45736d1c870ca73015a6482d6473c29987c633dd53400d0f9e2b4e4c9f697",
		"8f43f6e9572fcf0d57b8c7bea76d6db3e3c5327259dd5d61077b4b58c1d81727edfbb56865bcd9c73380f7ae4fbdf7103f570f481631e1a6ff1e94e4cb5791df4303055a754e4ffd0bb78a9434165082066d8519981ee80324071e1a23543a90",
		"4923cab0faf83499039fd8ee6be3c827786ed6e34bc6d54e7dba5b0f19d103ae1b6a49750a0fbcd0b4b9deb5e1bcf101973e055205c9f6e74093a333a3f6d4b25bd4a039a39530f5862a71b617725fdcfe8be60a47d2b025d65a77d62d45f60e",
		"f518ac8da349f616c2d76ee9c919a4797c10cb66620430abf1a3decf0fe129ba056c8c93c81e6852469ad6ee5e42fb07e9a138ea4a833bddb485324d9441a9652c65c175e1e5ea7d146d19905a82b7377401c6cbe02b534fbd7fad8238fb8f15",
		"68849ccedddd608c4e2591463000cc0d250772f502283419301b2639973f51e70c1bbc4da210d98a49b93450cc584908c6a147bf075e4a1702e652ce9760293dd39f07e5952d44c75726e8935351add0a0fb904542997d8edd0c1ad3c6abda8e",
		"b1e5c324f47736f12c6a97ecc279d360accea96fa2e81025cc4a1ef820b840f9ff28a6be8f43fbfbd60726df0e910f08e2a7566dcba9b5a16451012bcd577bb21bfb9db9a8907dab35daf30da26f8f8f222faf061cb2bcacc66f2825eaf5cb82",
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

	tx.Receiver = scAddress                    // send to delegation SC
	tx.Value = "0"                             // 0 EGLD
	tx.GasLimit = 300000000                    // 300 million gas units and 100 for million unbond
	tx.Data = []byte("unStakeNodes@" + blsKey) // "unBondNodes@"
	tx.Nonce = nonce

	err = ti.ApplyUserSignature(holder, &tx)
	if err != nil {
		panic(err)
	}
	ti.AddTransaction(&tx)

	log.Info("generated tx", "nonce", tx.Nonce, "sender", tx.Sender, "receiver", tx.Receiver, "data", string(tx.Data))
}
