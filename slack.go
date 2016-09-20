package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

var message = `{
    "attachments": [
        {
            "fallback": "Required plain-text summary of the attachment.",
            "color": "#36a64f",
            "pretext": "Optional text that appears above the attachment block",
            "author_name": "Bobby Tables",
            "author_link": "http://flickr.com/bobby/",
            "author_icon": "http://flickr.com/icons/bobby.jpg",
            "title": "Slack API Documentation",
            "title_link": "https://api.slack.com/",
            "text": "Optional text that appears within the attachment",
            "fields": [
                {
                    "title": "Priority",
                    "value": "High",
                    "short": false
                }
            ],
            "image_url": "http://my-website.com/path/to/image.jpg",
            "thumb_url": "http://example.com/path/to/thumb.png",
            "footer": "Slack API",
            "footer_icon": "https://platform.slack-edge.com/img/default_application_icon.png",
            "ts": 123456789
        }
    ]
}`

type Message struct {
	Text        string        `json:"text",omitempty`
	Attachments []Attachement `json:"attachments",omitempty`
}

type Attachement struct {
	Title      string   `json:"title"`
	TitleLink  *url.URL `json:"title_link",omitempty`
	Text       string   `json:"text"`
	Fallback   string   `json:"fallback",omitempty`
	Color      string   `json:"color",omitempty`
	Pretext    string   `json:"pretext",omitempty`
	AuthorName string   `json:"author_name",omitempty`
	AuthorLink *url.URL `json:"author_link",omitempty`
	AuthorIcon *url.URL `json:"author_icon",omitempty`
	TS         int      `json:"ts"`
}

func Send(endpoint string) {
	client := &http.Client{}

	v := url.Values{}

	m := Message{
		Text: "This is some text",
		Attachments: []Attachement{
			{
				Title: "This is a test",
				Text:  "This is a test....",
				Color: "#36a64f",
			},
		},
	}

	msg, _ := json.Marshal(&m)

	v.Add("payload", string(msg))

	resp, err := client.PostForm(endpoint, v)
	if err != nil {
		fmt.Printf("Failed to send slack notification %q\n", err)
		if resp != nil {
			b, _ := ioutil.ReadAll(resp.Body)
			fmt.Println(string(b))
		}
		return
	}

	fmt.Printf("Sent message")
}
