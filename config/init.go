package config

import (
	"log"
	"os"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/nats-io/nats.go"
)

var (
	instance            *Connection
	CONSUMER_KEY_SECRET = os.Getenv("CONSUMER_SECRET_KEY")
	CONSUMER_KEY        = os.Getenv("CONSUMER_KEY")
	ACCESS_TOKEN        = os.Getenv("ACCESS_TOKEN")
	ACCESS_SECRET       = os.Getenv("ACCESS_SECRET")
	//HOSTNAME hold hostname
	HOSTNAME, _ = os.Hostname()
	// NATSURL nats streaming server url
	NATSURL string = os.Getenv("NATSURL")
	// CLUSTERID cluster id for nats server
	CLUSTERID string = os.Getenv("NATSCLUSTER")

	ChName string = "ANTRIAN"
)

type (
	Connection struct {
		TwtClient *twitter.Client
		Nats      *nats.Conn
	}
)

func init() {
	if CONSUMER_KEY == "" || CONSUMER_KEY_SECRET == "" || ACCESS_TOKEN == "" || ACCESS_SECRET == "" {
		log.Fatalf("Required Env not found")
	}
	config := oauth1.NewConfig(CONSUMER_KEY, CONSUMER_KEY_SECRET)
	token := oauth1.NewToken(ACCESS_TOKEN, ACCESS_SECRET)
	httpClient := config.Client(oauth1.NoContext, token)
	// Twitter client
	TwtClient := twitter.NewClient(httpClient)

	opts := nats.GetDefaultOptions()
	opts.Url = NATSURL
	opts.MaxPingsOut = 10000
	opts.MaxReconnect = 10000
	opts.ReconnectWait = 30 * time.Second
	opts.PingInterval = 5 * time.Second
	opts.Verbose = true
	opts.AllowReconnect = true

	natsCon, err := opts.Connect()
	if err != nil {
		panic(err)
	}
	instance = &Connection{
		TwtClient: TwtClient,
		Nats:      natsCon,
	}

}

func GetInstance() *Connection {
	return instance
}
