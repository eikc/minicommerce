package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	stripe "github.com/stripe/stripe-go"
	"google.golang.org/appengine"
)

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type Attachment struct {
	Fallback   *string   `json:"fallback"`
	Color      *string   `json:"color"`
	PreText    *string   `json:"pretext"`
	AuthorName *string   `json:"author_name"`
	AuthorLink *string   `json:"author_link"`
	AuthorIcon *string   `json:"author_icon"`
	Title      *string   `json:"title"`
	TitleLink  *string   `json:"title_link"`
	Text       *string   `json:"text"`
	ImageUrl   *string   `json:"image_url"`
	Fields     []*Field  `json:"fields"`
	Footer     *string   `json:"footer"`
	FooterIcon *string   `json:"footer_icon"`
	Timestamp  *int64    `json:"ts"`
	MarkdownIn *[]string `json:"mrkdwn_in"`
}

type Payload struct {
	Parse       string       `json:"parse,omitempty"`
	Username    string       `json:"username,omitempty"`
	IconUrl     string       `json:"icon_url,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Channel     string       `json:"channel,omitempty"`
	Text        string       `json:"text,omitempty"`
	LinkNames   string       `json:"link_names,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	UnfurlLinks bool         `json:"unfurl_links,omitempty"`
	UnfurlMedia bool         `json:"unfurl_media,omitempty"`
	Markdown    bool         `json:"mrkdwn,omitempty"`
}

func (attachment *Attachment) addField(field Field) *Attachment {
	attachment.Fields = append(attachment.Fields, &field)
	return attachment
}

func slackLogging(httpClient *http.Client, title, text, status, color string) {
	channel := "#logging"
	url := "https://hooks.slack.com/services/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"

	if appengine.IsDevAppServer() {
		channel = "Eikster"
		url = "https://hooks.slack.com/services/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"
	}

	attachment1 := Attachment{}
	attachment1.addField(Field{Title: "Title", Value: title})
	attachment1.addField(Field{Title: "Status", Value: status})
	attachment1.addField(Field{Title: "Extra info", Value: text})
	attachment1.AuthorIcon = stripe.String(":gopher_dance:")
	attachment1.Color = stripe.String(color)

	payload := Payload{
		Username:    "robot",
		Channel:     channel,
		IconEmoji:   ":gopher_dance:",
		Attachments: []Attachment{attachment1},
	}

	json, _ := json.Marshal(payload)
	reader := bytes.NewReader(json)

	_, err := httpClient.Post(url, "application/json", reader)
	if err != nil {
		fmt.Println("slack error occured: ", err)
	}
}
