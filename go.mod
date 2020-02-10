module fg-test

go 1.13

require (
	fg-engine v0.0.0-00010101000000-000000000000
	github.com/bmizerany/perks v0.0.0-20141205001514-d9a9656a3a4b // indirect
	github.com/dgryski/go-gk v0.0.0-20140819190930-201884a44051 // indirect
	github.com/golang/protobuf v1.3.2
	github.com/influxdata/tdigest v0.0.0-20181121200506-bf2b5ad3c0a9 // indirect
	github.com/mailru/easyjson v0.7.0 // indirect
	github.com/rakyll/hey v0.1.2
	github.com/streadway/quantile v0.0.0-20150917103942-b0c588724d25 // indirect
	github.com/tsenart/vegeta v12.7.0+incompatible
	golang.org/x/net v0.0.0-20191011234655-491137f69257 // indirect
	golang.org/x/text v0.3.2 // indirect
)

replace fg-engine => ../server/src/fg-engine
