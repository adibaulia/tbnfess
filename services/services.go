package services

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"tbnfess/config"
	"tbnfess/models"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/nats-io/stan.go"
	"gopkg.in/ffmt.v1"

	b64 "encoding/base64"
)

type (
	dao interface {
	}

	svc struct {
		dao
		twtClient   *twitter.Client
		stan        stan.Conn
		OauthClient *http.Client
		Upload      *anaconda.TwitterApi
	}
)

func New(dao dao, c *config.Connection, n stan.Conn) *svc {
	return &svc{dao, c.TwtClient, n, c.OauthClient, c.Upload}
}

var oke = new(chan *models.Message)

func (s *svc) GetDMs(body *models.DMEvent) error {
	ffmt.Pjson(&body)
	for _, val := range body.DirectMessageEvents {

		if (strings.Contains(val.Message.Data.Text, "-nem") || strings.Contains(val.Message.Data.Text, "-Nem") || strings.Contains(val.Message.Data.Text, "-NEM")) && val.Message.SenderID != "1215181869567725568" {
			log.Printf("DM triggered body '%+v'", val.Message.Data.Text)

			payload := &models.Message{ID: time.Now().Format("20060504030201"), Message: val.Message.Data.Text}

			if val.Message.Data.Attachment != nil {
				mediaID, err := s.uploadMedia(val.Message.Data.Attachment.Media.MediaURLHttps)
				if err != nil {
					log.Print(err)
				}
				payload.MediaID = append(payload.MediaID, mediaID)
				payload.Message = strings.ReplaceAll(payload.Message, val.Message.Data.Attachment.Media.URLEntity.URL, "")

			}
			ffmt.Pjson(val.Message.Data)
			data, _ := json.Marshal(payload)
			if err := s.stan.Publish(config.ChName, data); err != nil {
				log.Print(err)
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

		log.Printf("body '%+v'", body)
		params := &twitter.StatusUpdateParams{Status: body.Message}
		if len(body.MediaID) > 0 {
			params.MediaIds = body.MediaID
		} else {
			params = nil
		}
		log.Printf("params '%+v'", params)
		_, _, err := s.twtClient.Statuses.Update(body.Message, params)
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

func (s *svc) uploadMedia(url string) (int64, error) {

	resp, err := s.OauthClient.Get(url)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()
	bodyResp, _ := ioutil.ReadAll(resp.Body)
	//buf := bytes.NewBuffer(bodyResp)
	imgEnc := b64.StdEncoding.EncodeToString(bodyResp)
	media, err := s.Upload.UploadMedia(imgEnc)
	//log.Printf("media %+v\n", media)
	if err != nil {
		log.Printf("err %+v\n", err)
		return 0, err
	}
	log.Printf("media %+v\n", media)
	return media.MediaID, nil
}
