package services

import (
	"encoding/json"
	"fmt"
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

func (s *svc) GetDMs(body *models.DMEvent) error {
	for _, val := range body.DirectMessageEvents {

		if (strings.Contains(val.Message.Data.Text, "-nem") || strings.Contains(val.Message.Data.Text, "-Nem") || strings.Contains(val.Message.Data.Text, "-NEM")) && val.Message.SenderID != "1215181869567725568" {
			log.Println("Webhook Triggered")

			payload := &models.Message{ID: time.Now().Format("20060102150405"), Message: val.Message.Data.Text, SenderID: val.Message.SenderID}

			if val.Message.Data.Attachment != nil {
				log.Println("has media detected")
				if val.Message.Data.Attachment.Media.Type == "video" || val.Message.Data.Attachment.Media.Type == "animated_gif" {
					mediaID, err := s.uploadVideo(val.Message.Data.Attachment.Media.VideoInfo.Variants[0].URL)
					if err != nil {
						return err
					}

					payload.MediaID = append(payload.MediaID, mediaID)
					payload.Message = strings.ReplaceAll(payload.Message, val.Message.Data.Attachment.Media.URLEntity.URL, "")
				} else {
					mediaID, err := s.uploadMedia(val.Message.Data.Attachment.Media.MediaURLHttps)
					if err != nil {
						log.Print(err)
					}
					payload.MediaID = append(payload.MediaID, mediaID)
					payload.Message = strings.ReplaceAll(payload.Message, val.Message.Data.Attachment.Media.URLEntity.URL, "")
				}

			}
			//	ffmt.Pjson(val.Message.Data)
			data, _ := json.Marshal(payload)
			// if err := s.stan.Publish(config.ChName, data); err != nil {
			// 	log.Print(err)
			// }
			config.Q.Enqueue(data)
		}
	}
	return nil
}

func (s *svc) TweetDMs() {
	var body models.Message
	item := config.Q.Dequeue()
	if err := json.Unmarshal(item, &body); err != nil {
		log.Printf("[SubsToPOST]: error message: %v", err)
		return
	}

	params := &twitter.StatusUpdateParams{Status: body.Message}
	if len(body.MediaID) > 0 {
		params.MediaIds = body.MediaID
	} else {
		params = nil
	}
	//	log.Printf("params '%+v'", params)
	twitt, _, err := s.twtClient.Statuses.Update(body.Message, params)
	if err != nil {
		log.Print(err)
	}
	s.twtClient.DirectMessages.EventsNew(&twitter.DirectMessageEventsNewParams{
		Event: &twitter.DirectMessageEvent{
			Type: "message_create",
			Message: &twitter.DirectMessageEventMessage{
				Target: &twitter.DirectMessageTarget{
					RecipientID: body.SenderID,
				},
				Data: &twitter.DirectMessageData{
					Text: fmt.Sprintf("Menfess-em wes ke tweet nang tweet iku, suwun yo lur https://twitter.com/tbnfess_/status/%v", twitt.ID),
				},
			},
		},
	})

	log.Printf("Dms Tweeted '%+v'", body)
}

// func (s *svc) SubsToTweetDMs() {
// 	var counter int
// 	s.stan.Subscribe(config.ChName, func(msg *stan.Msg) {
// 		log.Printf("Posting tweet")
// 		var body models.Message
// 		if err := json.Unmarshal(msg.Data, &body); err != nil {
// 			log.Printf("[SubsToPOST]: error message: %v", err)
// 			msg.Ack()
// 			return
// 		}

// 		params := &twitter.StatusUpdateParams{Status: body.Message}
// 		if len(body.MediaID) > 0 {
// 			params.MediaIds = body.MediaID
// 		} else {
// 			params = nil
// 		}
// 		//	log.Printf("params '%+v'", params)
// 		_, _, err := s.twtClient.Statuses.Update(body.Message, params)
// 		if err != nil {
// 			log.Print(err)
// 		}
// 		log.Printf("Dms Tweeted '%+v'", body)
// 		counter++
// 		msg.Ack()

// 		if counter == 20 {
// 			log.Printf("Sleeping")
// 			time.Sleep(5 * time.Minute)
// 		}
// 	}, stan.SetManualAckMode(), stan.DurableName("POST-DURABLE"))
// }

func (s *svc) getMedia(url string) ([]byte, error) {
	resp, err := s.OauthClient.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	bodyResp, _ := ioutil.ReadAll(resp.Body)

	return bodyResp, nil
}

func (s *svc) uploadMedia(url string) (int64, error) {

	buf, err := s.getMedia(url)
	if err != nil {
		return 0, err
	}

	encoded := b64.StdEncoding.EncodeToString(buf)
	log.Printf("Media encoded to base64 with size '%v'", len(buf))
	media, err := s.Upload.UploadMedia(encoded)
	if err != nil {
		log.Printf("err %+v\n", err)
		return 0, err
	}
	return media.MediaID, nil
}

func (s *svc) uploadVideo(url string) (int64, error) {
	buf, err := s.getMedia(url)
	if err != nil {
		return 0, err
	}

	chunks := split(buf, 30000)

	var encodedChunks []string
	for _, chunk := range chunks {
		encoded := b64.StdEncoding.EncodeToString(chunk)
		encodedChunks = append(encodedChunks, encoded)
	}

	chunkMed, err := s.Upload.UploadVideoInit(len(buf), "video/mp4")
	if err != nil {
		log.Printf("log %+v\n", err)
		return 0, err
	}

	for i, enc := range encodedChunks {

		err = s.Upload.UploadVideoAppend(chunkMed.MediaIDString, i, enc)
		if err != nil {
			log.Printf("log %+v\n", err)
			return 0, err
		}

	}

	media, err := s.Upload.UploadVideoFinalize(chunkMed.MediaIDString)
	if err != nil {
		log.Printf("log %+v\n", err)
		return 0, err
	}
	return media.MediaID, nil
}

func split(buf []byte, lim int) [][]byte {
	var chunk []byte
	chunks := make([][]byte, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:len(buf)])
	}
	return chunks
}
