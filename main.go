package main

import (
	"log"
	"tbnfess/api"
	"tbnfess/config"
	"tbnfess/services"

	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
)

func main() {
	e := echo.New()
	c := config.GetInstance()

	sc, err := stan.Connect(config.CLUSTERID, config.HOSTNAME, stan.NatsConn(c.Nats))
	if err != nil {
		log.Printf("Error Connecting to stan server, %+v", err)
		return
	}

	svc := services.New(nil, c.TwtClient, sc)
	api.Router(e, svc)
	svc.SubsToTweetDMs()

	c.Nats.SetDisconnectHandler(func(nc *nats.Conn) {
		log.Println("Disconnected from nats server.")
	})
	c.Nats.SetReconnectHandler(func(nc *nats.Conn) {
		log.Println("Restarting service nats connection.")

		sc, err := stan.Connect(config.CLUSTERID, config.HOSTNAME, stan.NatsConn(c.Nats))
		if err != nil {
			log.Printf("Error Connecting to stan server, %+v", err)
			return
		}

		svc := services.New(nil, c.TwtClient, sc)
		svc.SubsToTweetDMs()
	})

	defer c.Nats.Close()

	e.Logger.Fatal(e.Start(":3033"))
}
