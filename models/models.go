package models

import "github.com/dghubble/go-twitter/twitter"

type (
	Request struct {
		Tweet     string `json:"tweet,omitempty"`
		Crc_token string `json:"crc_token,omitempty" form:"crc_token" query:"crc_token"`
	}
	DMEvent struct {
		ForUserID           string                       `json:"for_user_id"`
		DirectMessageEvents []twitter.DirectMessageEvent `json:"direct_message_events"`
	}
	DBStruct struct {
		Text string `json:"text"`
	}

	Message struct {
		ID      string  `json:"id"`
		Message string  `json:"message"`
		MediaID []int64 `json:"media_id"`
	}
)
