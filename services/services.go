package services

import (
	"encoding/json"
	"log"
	"strings"
	"tbnfess/config"
	"tbnfess/models"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/nats-io/stan.go"
)

type (
	dao interface {
	}

	svc struct {
		dao
		twtClient *twitter.Client
		stan      stan.Conn
	}
)

func New(dao dao, t *twitter.Client, n stan.Conn) *svc {
	return &svc{dao, t, n}
}

var oke = new(chan *models.Message)

func (s *svc) GetDMs(body *models.DMEvent) error {
	for _, val := range body.DirectMessageEvents {

		if (strings.Contains(val.Message.Data.Text, "-nem") || strings.Contains(val.Message.Data.Text, "-Nem") || strings.Contains(val.Message.Data.Text, "-NEM")) && val.Message.SenderID != "1215181869567725568" {
			log.Printf("DM triggered body '%+v'", val.Message.Data.Text)
			payload := &models.Message{ID: time.Now().Format("20060504030201"), Message: val.Message.Data.Text}
			data, _ := json.Marshal(payload)
			if err := s.stan.Publish(config.ChName, data); err != nil {
				log.Fatal(err)
			}

		}
	}
	return nil
}

func (s *svc) SubsToTweetDMs() {
	var counter int
	s.stan.Subscribe(config.ChName, func(msg *stan.Msg) {
		var body models.Message
		if err := json.Unmarshal(msg.Data, &body); err != nil {
			log.Printf("[SubsToPOST]: error message: %v", err)
			msg.Ack()
			return
		}

		_, _, err := s.twtClient.Statuses.Update(body.Message, nil)
		if err != nil {
			log.Print(err)
		}
		log.Printf("Dms Tweeted '%v'", body.Message)
		counter++
		msg.Ack()

		if counter == 20 {
			time.Sleep(5 * time.Minute)
		}
	}, stan.SetManualAckMode(), stan.DurableName("POST-DURABLE"))

}
