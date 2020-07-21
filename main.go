package main

import (
	"fmt"
	"tbnfess/api"
	"tbnfess/config"
	"tbnfess/services"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	c := config.GetInstance()

	// sc, err := stan.Connect(config.CLUSTERID, config.HOSTNAME, stan.NatsConn(c.Nats))
	// if err != nil {
	// 	log.Printf("Error Connecting to stan server, %+v", err)
	// 	return
	// }

	svc := services.New(nil, c, nil)
	api.Router(e, svc)
	// svc.SubsToTweetDMs()

	// c.Nats.SetDisconnectHandler(func(nc *nats.Conn) {
	// 	log.Println("Disconnected from nats server.")
	// })
	// c.Nats.SetReconnectHandler(func(nc *nats.Conn) {
	// 	log.Println("Restarting service nats connection.")

	// 	sc, err := stan.Connect(config.CLUSTERID, config.HOSTNAME, stan.NatsConn(c.Nats))
	// 	if err != nil {
	// 		log.Printf("Error Connecting to stan server, %+v", err)
	// 		return
	// 	}

	// 	svc := services.New(nil, c, sc)
	// 	svc.SubsToTweetDMs()
	// })

	//defer c.Nats.Close()

	go func() {
		for {
			if config.Q.Len() > 0 {
				svc.TweetDMs()
			}
		}
	}()

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%v", config.PORT)))
}
