package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"log"
	"net/http"
	"tbnfess/config"
	"tbnfess/models"

	"github.com/labstack/echo/v4"
)

type (
	svc interface {
		GetDMs(body *models.DMEvent) error
	}
	r struct {
		svc
	}
)

func Router(e *echo.Echo, s svc) {
	r := r{s}
	e.POST("/dev/webhooks", r.webhookEvent)
	e.GET("/dev/webhooks", CRC)
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"oke": "mantab",
		})
	})

}

func (r *r) webhookEvent(c echo.Context) error {
	body := new(models.DMEvent)
	if err := c.Bind(body); err != nil {
		log.Print("ERROR", err)
		return err
	}
	if err := r.svc.GetDMs(body); err != nil {
		log.Print("ERROR", err)
		return err
	}
	return c.JSON(http.StatusOK, nil)
}

//CRC Check from twitter Api
func CRC(c echo.Context) error {
	var body = make(map[string]interface{})
	c.Bind(&body)
	log.Printf("CRC check from twitter API with response callback ")

	secret := []byte(config.CONSUMER_KEY_SECRET)
	message := []byte(c.QueryParam("crc_token"))

	hash := hmac.New(sha256.New, secret)
	hash.Write(message)

	// to base64
	token := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	resp := map[string]string{
		"response_token": "sha256=" + token,
	}
	log.Printf("CRC check from twitter API with response callback '%+v'", resp)
	return c.JSON(http.StatusOK, resp)
}
