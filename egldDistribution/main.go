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
		"erd1vqw75zwpnxnpzkkp55a9accndh56qqclyhgu2lddhwv0hhn80uqsfgjs94",
		"erd1xwdm96t8fydx96vj8fzxw24hfxqrnkadzuvgmntve8pyu2e6g6xstsy267",
		"erd14zwcsq4hl70vsw8tcf4pztp9n6gvc6xxn5x4x86frvffk39z22wsr0tqql",
		"erd1jfl0sl9qpg4gem9w4ghumrzggz83w5gd440swplfzm253e70gl3sf78lgj",
		"erd1duy6zkxemcx9mjvpy6n24qhktxndupf4j4edtgykr4rkm8n7v9uq9y0n0k",
		"erd1e9zqe5c9sk6al3zwturhesknuux7qwzku3q07yu95q33njnm7lnqpmcp3f",
		"erd1vvqprkgmzvu8lmcmal8vf7we4qqzj7f6cztsef6val5yqe6xz00s2ajj8p",
		"erd1c8rgsuclp485sgz7x4zxhzsf73fwfwdk88qxaukmt8p7ef32svmshass9v",
		"erd12q6p3h46xuuv3c89uwwmjgxkwxea5y3tvpvml5k335al4cuhyx7qza4k4p",
		"erd163zew5prat9k78k553quypvyyfghyde02573c6vnvngx52u2csys2l7xs2",
		"erd17vs7ysdk0xashhz5dhr6acyx88algzxwd4ef34m5pf3tnapgtplqqzndh4",
		"erd1spm863ywflqdpu3kxjrndgze28gecqtfgx9nvtf5n33j0jr9fjeqdfayw4",
		"erd16flffuf6nkqy8ywz3dguu0hm0hlza4mvrwqejtyd7euu2ja8xalqutan2g",
		"erd1wg38aly7qqcjpcxyg96dgg96nps0e5n28qw993t42gzay6d4pcps4g7f09",
		"erd15nvmqk4eefn6ms7ma9nx7kmn448vvtrrmqvekc28w6uhrxwqq2csrrt25u",
		"erd1eyez5runnrpe5j62c0k9a9tpl3dsq9gg7pfxenhkt6lrte9dmk3shmqnrf",
		"erd19xtlf5ewwklvk546gkpxlpqx2n9utuyfqevujqkre5vg6wee9zrqy4km7c",
		"erd1una2gcwdtv9enxcv4trmyst0tk0usqfe6phzvdsk62ug4se94xvsgk22p3",
		"erd19qm58g7yutfy0nl5mx8m5fku6l5tpka0yuf9ykhautkxkhxaw5hs6dvcrl",
		"erd1rn3sqgvq0xd8fy239cqefsuc394kwha8d90jy97kpujwgxxgsnrs5tk2jw",
		"erd1hz6ewfqn4unnatfmw0ymwwkxvra5jwlr6fztyvaesqazrx2f9e2qm0632h",
		"erd1duj0emxl97m3r4g56jmfxsj04tcr4tl2z9r08grfpgry80dg6slsvym6cx",
		"erd1ls0qxn80xamwusknxl7ztm4pw4vgm6d3a0xpghzpykfddqyrvveqr96pe5",
		"erd1lrmgwmldn6kvvjqh5jypl2cldyquj0eyeqh7hf0eydmlvnlpgmkqg0dhh6",
		"erd1tdnh6fswm05we2efqrzk25rxrqzzfm4l6kj6vceqcaazswrpw8wq0u8tua",
		"erd1dw4yz8y66np395vmeeaw5vqql26wcym8umxqtwuzlyj54h0xs5dqe9d6tu",
		"erd1k5jz3ywendz6vvhfhfn25ktmk5defcw6w0mt7jvr6tzzk0f8yakqm0gp3n",
		"erd1zfr36cjqsn4lkd9f4f0qpyf48pxwe5jswqfjur2a6svhse33wgaspxfnrh",
		"erd1gk6aw8nfflpnw88j7c3hvv40cmc4qquj6z7jqz4xy8a4wcsd4vmszefngf",
		"erd12ezvgxneqn5r7q6xx4stvnhdx7a48ggg5rzxrgntwv7fzq4acdasc7dyl2",
		"erd127jte6gdsaldeyr0m7qsw95dt882xxk8a0f07kxshjq56rzgsqlsswpgfw",
		"erd12nk8jwsfrnp6zdrwsc28nnn54psfcj5rjzqmm8z2xl0xv8g8ra5q5evzw8",
		"erd1lhna5wme8r0e9gxhpqdtsxu26uzrr3gxtsdg0nr7xjtn8wze69xqpffkhx",
		"erd1dn7rfwv5n6lk70zglnwfmfy0l23qgwuclkdjkgzh4hzvjpqnfm6q93zg8d",
		"erd106chdu3wknj8d743ktezsr2f262vvaf3qd88xkn68ysd86dwxqkskhs2qp",
		"erd1hcmk8hx6dyjr7f7u6jdv6sk7hrd8zka4d9p0ukse0dl0sfsyjqsqy585vq",
		"erd1qutw8s65g72227r0zknqssrjgqnnk69j8z4yp9j06946jxdhxxwshvxvz2",
		"erd1tkc62psh0flcj6anm6gt227gqqu7sp4xc3c3cc0fcmgk9ax6vcqs2w8h2s",
		"erd16myes42p5t6xnxkxw8jghg8sypuammz9fc44g4f2u43dj2hd0fusp9286j",
		"erd1evywsr2jq8d989varsvq9sun2rtcx8s4qw9u7cfepz9ql56nxums98kmdw",
		"erd1uv32khhl2r276afpc2zuwfu3qyp4ukgrhu3wf2mu4tcj3edqzclqlntkxx",
		"erd1fdvf829mz67duut6d0pm43afz9grrzq79ksmxjcrk6hldmnp4w5s9xhwyw",
		"erd15pzx949cdvz83j0uve5zt46pme432lf67rsyxu7f99emyv92lg3sgmwfqw",
		"erd1tjygwhw5ylmv3v52ucvhmz0q7r0hafz4cfndjaskss5ahz28l3hqdvxqct",
		"erd15k3vk7jc8ywvx6wxac4zvv8wdzehmy7wmfjc00qh9twzwu47qq0qad8sgd",
		"erd120xuctg972jntplfc2pe9fyykpue4nl6nl3g2nyua8fsu65p7hrs4p2yfu",
		"erd1xh8qwu3lta3zm0mxncxeyttgfwusjueeqavkzchu4yc9myyas6cqh3lydm",
		"erd10gnrje5gpwxqvx5zunlmqr6uyg66r3n4yylrjvgd2g5rf3wdrujqlx2ee4",
		"erd147tf4vjtjyy7g4gdp0v0k5gtvqts8crlfpaldc67s2zwn82aza4qkh7yrt",
		"erd1ragf2tu8rkz3mme3m2yx6f33huy8hkwnknm82eg8s2zdpgy7mwas4d52hm",
		"erd17vhjpyakh7zjpr9f9ac4s0aq4x7fdm0j92kne7w8z8ll2g7vxp9qjsd8z3",
		"erd1hk4rdy55r5u2yycnl29ac26wm0ga9j542ejkna8gdu0jfy7ufsnq0jxs79",
		"erd12xspx8yr9gv25nlemkv3pnsa49w9zutwd60w4xcfln2f8hrm500q5adl8m",
		"erd18r295ezl6nt4cwp52pp80as7m8ddppcll7k7h2ek8sw7h89lsunqkus5ux",
		"erd1msam4dx0x4hrwz6xp06sh4drxvtxg7kq97cx5nf47a8q440ezy8q0jwacs",
		"erd1kv5mkar6fvt6vhqj7evfqr9jnmmlqps3q9dp0t0gr9tcpqupyrsshlnvd0",
		"erd146zgxv6dv2x5d2cangu7r6flw8gv7ck2sjzf84l7lueh6h2lgg5s7ud0g8",
		"erd1th4hlfetl0v9ttkzrd4e6cjck62n5v72yw42a39why82yz7fnw7qmj866h",
		"erd1skd7d5s79jlxpnxvwqczthyt7jnr2xdz9aa26674vag0ekd7uussn2y0ur",
		"erd12ju6ut97nffc54py0090thgd4tk4nw42marta6gzd923r75nhwwqegyjgv",
		"erd1mdvj4jt6hv83pedd080zlvlug2raax3u97dv6y08qnexwm0qkcwqruh0m4",
		"erd12l6sk3ceklpf5jx6atut5mydqh3dqfpt73h2gxqh7zzmqxwx2jwqf5yj8e",
		"erd1lt0zj3783zztc6l7l8kr73levq6phj5da3ldmd5j3j3jua9r300q76d9d5",
		"erd15yhduccpcs8akr8ld8uewx3kvz3ggq2th6sm5nl0y0az6p7w9flsd7kvak",
		"erd1dfpejlst45ltw3tcez5gvj85rla8whgcm6l5fekr5es3j9aldaysmektc5",
		"erd1uvkszrp8x52kssr5j3s527prhmpyknrfxwc2kvu3pfv4ryq0l2as9vpu5q",
		"erd1eca9zakdprzng0y3yme9rjwrmaktqg96eqkddv3d27gs3cpm8rzs2nu5hj",
		"erd1jsl5s3w4fjg68f5ge842g9569rcag4x6jtcfradejuv2w8cqm2psru20em",
		"erd19rh30cq9964an8vj7qnj7gwaus90nv6020vxpu69ramwrn8yr78smteycl",
		"erd1h6p2fvll3jy7pcctxp0cj0vdqlsqhnhdny3qs8yyd4znt3qhz27qtj88ey",
		"erd1dt9wtpcpusrhyrhuuqh2h64k9f2x435l00smfmrkkrm2mjwdvcxqce8r5a",
		"erd1akkpa99huyqrstk7keq2tlz86uu2fpcjdh2adpg39xje9493jy5sw27v3x",
		"erd1npzvy2zehwnemvs5nsnkxfsw50h398tp5nq4lds8ys3fwlwwyryqyp7t3w",
		"erd1vu7vm2dvttpkky8g8mmcd6x8vcz43e6cftfv9yn8muuxsam0crxqfvyzu5",
		"erd1xzhv0naa3zavrtrwe2cpnw0cz6k2pw7jn2nhfnwwegzuqrwq8psqnkc0wk",
		"erd1q9xj4uqrfy9ge6n7lefn24qfa78pqd9q5dlc2fj8yv79smtdf9qqcglc34",
		"erd1687tqy77dfl8g69h3lmnjcdej8z6pkepk7kumhmgwtscxm5s69rsvnf46s",
		"erd1kkcvtdc6j235d9x830hlz8xc4vvz3q63mlhqykw5nq8f2t0tfx7sx4jhhj",
		"erd1lan2yzvzvdkt4adgxza0zsx8e0kd6czdpsjy390dk44kmurwlu6qk9kz2q",
		"erd1dhfjkzpwgnneu4jwdjclf658s24y3zc5mq4tq4uczccnykzkrq8spea0nw",
		"erd1grn30xypva4w2z8lu9n9k8hkglhdhjy49kvz3vn9869kkvecht3qn6gzz0",
		"erd1clrajudw2a3scwtc3zw7m75yx653elf59dhxfchvu2v5qk4fg6kqy9w8u7",
		"erd1j2xujqpl2nhuufyah6e0qv24f8z3lec0837x4fxe2v308lkyetsqz8au7w",
		"erd1wtcwesal6gzqpydd3fk846dzvkrxalj0r5g9zewqn0rvyr27zc3swzjgnd",
		"erd1kc7v0lhqu0sclywkgeg4um8ea5nvch9psf2lf8t96j3w622qss8sav2zl8",
		"erd1xzu4gffwtlp9t63ewsh5zsly2mehjdyyym4vdjy5uwk4xnm96sqq4f4td2",
		"erd14jmchcwn4wu4f8z6804pdy7klgwhp54n4k4a3m2ql50etlfsvyksc8zks3",
		"erd1mlw8enhzaqup3fpegaf7apvssvx60rfgn0ltgd3ehe5hmng76pcq3ts6pf",
		"erd1936f0gzwesk4s3j0nr9xnu8ehau7sz2eaulclll0f8daexmsactsc99v22",
		"erd1rw80cg2c4fyxsrxc2deeqrahe620fwz9nzy249mrqs2gek06qdssu0mm7j",
		// part 2
		//"erd14x04l5sln5t5qudndxqhe7543jwrdl6gp7ph3y5ehqr7pr9kr9sqxr8c0v",
		//"erd1yrvhanh7qnml5hgnkj2smpp45vpn4z2dcl2htpqvdqas5t96cn7qyuapuc",
		//"erd196rr5s6vn0y6xqvqnp82j0xa7ryzwnwl8f2kwph5gx9yycsznvwsxwg9jc",
		//"erd1npzsn9wh3lmznya5uu0rc0t7wsfvplvlse2juv5rc625rvzdgp4q09tm45",
		//"erd19dmypldjghjhd2xrzjt9j9u7euxl5jwch0h88lagrlaervnwqs7q02e6a9",
		//"erd19fxkru7gjd22sssq28lsl4q0rgm23sj993wf60kt55yjrmqdh55shz9d8u",
		//"erd1ytlrvq2pxz49yplqky36pjklpgmtg6ucq626huuvg5r92m0gf5aqjc37hd",
		//"erd1aawv2txs9zq9k73fhc5584txwpmh29j36j29xh8f2vspmy74u2es7fg67p",
		//"erd1nwda23plec9car6plfy24xq06h6nq6u6p7pledurj2xvh2pztw8s9gu4a7",
		//"erd1w8juz2r6z7469twsfn5c3ecvjtvut0pwf8k7yw0tt34y78ljwe2s2tvtef",
		//"erd1xqmn9udryaz3dklpgyjqt9cklmfw700eymjalu34egeh5jpaewgqt8hts0",
		//"erd1jt4uxja70zpdwku7qxl06vq36ww3caz4ngkp5554ysjvxtm992wqzjyppm",
		//"erd1ucfwrdgwe6q8grpwq0zka27h64ukgskqr3va0ef3ms99uzq9wvqs9ujrrn",
		//"erd1rs0tryhjmc6gafefqedv7rwngjjlyswhlyhnsxgf0qvz7mq5sjcs26xrcq",
		//"erd10dadclxznjgcstxp5yadjpxmxuypdw4vjrt5lcx5dqe7uxzkaxjq7d92uq",
		//"erd1pu364rhl0zc4hqtxacer5w8lhrseut3tyhfkvtpc2pdd8kq8p9dsa4j6u9",
		//"erd1rmep0h59yu26qyj95e4yus4mxduzd8ejv4hv69kh3l8dxzkvm90s052aag",
		//"erd1sjvkdd42jj8hvz2c9n396rmtlfkpdfdakr48q20zejshevd8k86sj5znvj",
		//"erd1g2kce728p8em7zx94xp85zl9q7cq0ulkcgjvhe62awmmlp7vq4qqyx7svc",
		//"erd1mn0rg9j3t5qe73qj6ktaclzrtr4xyqfqgc8n57eaa4h8x3ynz0cswtavn0",
		//"erd12x2pd9ww6th67xsw4pewwkeysjlj29efnf05rs35gruw0sz82lwskzfpxn",
		//"erd1t79j33ttyty0alttuvumpty0x0safpfeuhnup35thqgpy4dvkjyqsx7vq4",
		//"erd172r5vn7jceqmmhusr0xehshxmpp0p5k7um67tzted6ga93t7peys8lzvcu",
		//"erd1j6d98q0mx4kugnge70hc92ctvxj3mxyzfhguta2ntynqrzx4xqeql33s8j",
		//"erd1qklrm95gr75mv5a64q5n4ned3yy2v7nhdje6vn56kjvcuvy44v6skd3te7",
		//"erd1egh8894lmv0t2y26stag09mc4gnqhnm4lyalewgh7y3t39tavfmq4kh0pl",
		//"erd1np780tjgufhpnfg3wh5f7t7jtju7tskuhtsrxjw6z245rl0rw2qs6dredc",
		//"erd1vamp3c2m9y09tj6l63r6g83qhdvxudhlxwr6ee6sd362e4z0h3tqrs3czw",
		//"erd15gad3x08y0rd5kdcs55wsghqj906a9d3y89pvuy3zmjzwwlfdr6q4cm539",
		//"erd1ll6c0464re5n8457wmcpt6tpdsd9egy9ngz3jwv4law2cahne6nss5mn9y",
		//"erd1kxde94mzcwld7ndypc5lahgx6kz2hw34aflavjl3ztprhtmajxkqp6wuzq",
		//"erd10y3jzsd2ps8t3c3jv4ges7hpvkf7kg72r8t9fyfrq8ph7vzd0gmsfl8r3m",
		//"erd1gck9cvs6l06vzjp52lg9yzk8pv85lv7l95nuvxkr2vey48rptxdq4xafj7",
		//"erd1ww7wmxvae5nqetcc0u6774n64zugdnjhte365qxm456w9r6rs5nsg9ywdu",
		//"erd1uryswdasmdcmp89fan73wq76t8l2kn8u52gh8892fsd50yv57sls4apn0n",
		//"erd1k3w2cxkm4pc2g30nle7zgk0gmex3tv7jhdrefks5wxv99gp5sx9q7jzlct",
		//"erd1u73phgrlhzqvn6eyznt53308l72mawa2u0gryh5uszl5l0ajk99skvczfa",
		//"erd1uw9muhnne0dddq5rvgsle9lwcxy78fjd0gjplfgu0x4jfukakcxq6lvptj",
		//"erd1ee8mq5u5azaq68qyv59gn9xqfs744kny6gk70t8ytq2u0slagt6q3jy5xv",
		//"erd1kryj7mrpfsfdhp3vr6d5z0kf7r9svpacmkzv275zm3kkkdd9qsjqds9vn2",
		//"erd1pfzvqtv40vrjde0h4dtz7nsjzcw2dpxdmuuu8k6exerfmcutnscsrgzlqw",
		//"erd1jq8sj7g03xk0cpz8nyxxhas9vflgwg5mky2mrxk2hq947xjeggms02x2rr",
		//"erd10yhvk5y79mpj3jpfw0awu47z2hsnngchjcvektaf6ur2mmyuzfkqshglq9",
		//"erd1juw30utjqemqhq5aazy4wd44xkvetcx2pa7px9gehf6ylrhgepksjsrkpn",
		//"erd19fg4uvh9mhgruwmwwjzrqh3vdwk9uyl5m0eq0hf8k9rnmvv7wfxqfx3pfu",
		//"erd10j8u2hfm0j083uxzh87jxt5wrxevx8q703kmc2s3ttcymha4gfjsdcau6a",
		//"erd1k7kqvcwavq9ankczpauyfkltcl8evjkxnrpnd85nssx9g5x7e5xs3eaw3h",
		//"erd195ntkudg0myu6fhkp0yktkcmjj6jeesh9sjmumpsajmwd3y4fpqqw64q8r",
		//"erd1zlwjqj9t29ymjh374fr5uhltj9gjp4m92kcfahsgw70e4vxkfxcs0qfvgm",
		//"erd14nccmactchjefg3zwpcxv6m8pdwl7lga2403d0m258x3a57x4s4qhz0lx6",
		//"erd1k25kkczyqdnmupqrxnl7d7r05wy9jkyh52x4g9mdg6emwx8fs0lq9vxlae",
		//"erd1pjrr55xksdgg62zqrfq8ksdg0djp4qacqh5x7lg6yc5qlgtqczvsl83052",
		//"erd1nhcatx0l4uaa2r22r5g8v0r2q5pd68mf2yessu7jav522pcvn8wqejzwxx",
		//"erd1c0xphraeszu73l44gt54jce4q9wvxr2y9lpyv05l0tcs5wxeqj5q4k87u7",
		//"erd15rxksvp2ru0ejkuq8zgtjxapy4c95kgvflxdrv4qe57tp8cdwuzqt0h2ny",
		//"erd15rj32f8mp5ek3gu8c60zqfn6lzelxvm827jdg4dl9yzf0ftpvjrqfy0l99",
		//"erd1dpald6wq8l2v9yrd4q6fqm25v5vtqkrfxlpezrlw35d458ddpcqsmkljnc",
		//"erd1gztuhxycghmm7u7cwfgdx3hl0ldn60n54pqdl2w5yv4ptyg98jvs0hu6lc",
		//"erd1a6zsa7wpfuc8hpju2lkt33fd4lltn8ackx2mejf3j67r0j38w4kq9hmxz6",
		//"erd1yv9m5wjg83r4x62kz7emzwsqszmpnku3ej6jhrqz9rnrp40am29sqtgudc",
		//"erd1elje8ekfx29f4hdhm5e706ldnyy0pp77f3kg0d8t73rsa0wtladsherthn",
		//"erd1244vg7sm2sx2s75qapzqzjh626l2dn7vke6k0esvv0zw9r9rr07s08chjc",
		//"erd1ywvvm4k63r9mucxtvd47d5trf4n3sq3mv8g026j3fzeevywh5mpsw6nvnh",
		//"erd1k087kdn7a4xg0j7mqhu86avrxhph3sxhev4yt2x8gxw4vxkwaajsacmpv0",
		//"erd17rud5syyk3g5w2ktach0muh64kx42g6l786su2yvmdevg66m49qq9jrlp9",
		//"erd18jhnq2mj3vqstjzzktjll7ykxrhkh6vkacumavu3afyrrjv8p92qj9zvs5",
		//"erd15rjnqdmyfwey82nl8axm3z3v86824st30zqtwps25w8lfe00j5lsfqnfl6",
		//"erd1mwr6tjefzpw7852d90mmdyrtefy42zhqvhkg5lxge5gueryz9pxqvgp3kx",
		//"erd16hcpxffy0u4ln07cycvp4vlz5xz3vhhu2gv6chsjwxycfxty2dsqhy4cts",
		//"erd1vlv5jh20xnn0ccv664344fxevk9jht980chv79p4m2vmwckzzfwqcsstgg",
		//"erd1ekherd9jk6us7g29qvx55c7x8mfpud7f5thzurh49r7jhny0hr0q60ajfj",
		//"erd1vfa59zy8q23l5mequ0hurx02kcql0ydma04x8urnkxxderzfp87sc04ykt",
		//"erd12hs5s0u5fdgp330eeqnua7xnr50xpf5dk7n5pys8aweqjxf3ru6qk29g7n",
		//"erd1j69tpx6h85xfuf88hx9pct57w23a9j7j26jhxf9tp5v6sa7hrfeslmus5q",
		//"erd152u47en7w509r8uf7z7yqnvpnmsk06yqnpmykfcyc0ch4ws9n6hs6cm43q",
	}
)

func main() {
	checkForDuplicates()

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

func checkForDuplicates() {
	m := map[string]struct{}{}
	for _, addr := range walletAddresses {
		_, found := m[addr]
		if found {
			panic(fmt.Sprintf("key %s is defined more than once", addr))
		}
		m[addr] = struct{}{}
	}
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
