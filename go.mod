module github.com/honeytrap

go 1.11

require (
	github.com/AndreasBriese/bbloom v0.0.0-20170702084017-28f7e881ca57
	github.com/BurntSushi/toml v0.3.0
	github.com/Logicalis/asn1 v0.0.0-20160307192209-c9c836c1a3cd
	github.com/Shopify/sarama v1.16.0
	github.com/boltdb/bolt v1.3.1
	github.com/davecgh/go-spew v1.1.0
	github.com/dgraph-io/badger v0.0.0-20180227002726-94594b20babf
	github.com/dgryski/go-farm v0.0.0-20180109070241-2de33835d102
	github.com/dimfeld/httptreemux v3.9.0+incompatible
	github.com/dutchcoders/gobus v0.0.0-20180915095724-ece5a7810d96
	github.com/eapache/go-resiliency v1.0.0
	github.com/eapache/go-xerial-snappy v0.0.0-20160609142408-bb955e01b934
	github.com/eapache/queue v1.1.0
	github.com/elazarl/go-bindata-assetfs v0.0.0-20180223160309-38087fe4dafb
	github.com/fatih/color v1.6.0
	github.com/fuyufjh/splunk-hec-go v0.3.3
	github.com/glycerine/rbuf v0.0.0-20171031012212-54320fe9f6f3
	github.com/go-asn1-ber/asn1-ber v0.0.0-20170511165959-379148ca0225
	github.com/golang/protobuf v0.0.0-20180202184318-bbd03ef6da3a
	github.com/golang/snappy v0.0.0-20170215233205-553a64147049
	github.com/google/gopacket v1.1.14
	github.com/google/netstack v0.0.0
	github.com/gorilla/websocket v1.2.0
	github.com/honeytrap/honeytrap v0.0.0-20190405081451-87794dac6942 // indirect
	github.com/honeytrap/honeytrap-web v0.0.0-20180212153621-02944754979e
	github.com/lxc/go-lxc v0.0.0-20180227230311-2660c429a942
	github.com/mailru/easyjson v0.0.0-20171120080333-32fa128f234d
	github.com/mattn/go-colorable v0.0.9
	github.com/mattn/go-isatty v0.0.3
	github.com/miekg/dns v1.0.4
	github.com/mimoo/StrobeGo v0.0.0-20171206114618-43f0c284a7f9
	github.com/mimoo/disco v0.0.0-20180114190844-15dd4b8476c9
	github.com/op/go-logging v0.0.0-20160211212156-b2cb9fa56473
	github.com/oschwald/maxminddb-golang v1.3.0
	github.com/pierrec/lz4 v0.0.0-20171218195038-2fcda4cb7018
	github.com/pierrec/xxHash v0.1.1
	github.com/pkg/errors v0.8.0
	github.com/pkg/profile v1.2.1
	github.com/rcrowley/go-metrics v0.0.0-20180125231941-8732c616f529
	github.com/rs/xid v0.0.0-20170604230408-02dd45c33376
	github.com/satori/go.uuid v1.2.0
	github.com/songgao/packets v0.0.0-20160404182456-549a10cd4091
	github.com/songgao/water v0.0.0-20180221190335-75f112d19d5a
	github.com/streadway/amqp v0.0.0-20180315184602-8e4aba63da9f
	github.com/vishvananda/netlink v1.0.0
	github.com/vishvananda/netns v0.0.0-20171111001504-be1fbeda1936
	github.com/yuin/gopher-lua v0.0.0-20190206043414-8bfc7677f583 // indirect
	golang.org/x/crypto v0.0.0-20180621125126-a49355c7e3f8
	golang.org/x/net v0.0.0-20180218175443-cbe0f9307d01
	golang.org/x/sys v0.0.0-20190204203706-41f3e6584952
	golang.org/x/time v0.0.0-20170927054726-6dc17368e09b
	gopkg.in/olivere/elastic.v5 v5.0.65
	gopkg.in/urfave/cli.v1 v1.20.0
)

replace github.com/google/netstack => ./vendor/github.com/google/netstack
