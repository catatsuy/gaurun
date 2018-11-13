package fcm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

type Client struct {
	URL        string
	HTTPClient *http.Client
	APIKey     string
}

type Payload struct {
	Message Message `json:"message"`
}

type Message struct {
	Name         string            `json:"name"`
	Data         map[string]string `json:"data,omitempty"`
	Notification Notification      `json:"notification,omitempty"`
	Android      AndroidConfig     `json:"android,omitempty"`
	Webpush      WebpushConfig     `json:"webpush,omitempty"`
	Apns         ApnsConfig        `json:"apns,omitempty"`
	Token        string            `json:"token,omitempty"`
	Topic        string            `json:"topic,omitempty"`
	Condition    string            `json:"condition,omitempty"`
}

type Notification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type AndroidConfig struct {
	CollapseKey string `json:"collapse_key"`
	Priority    string `json:"priority"`
	TTL         string `json:"ttl"`
}

type WebpushConfig struct{}

type ApnsConfig struct{}

func NewClient(urlPartedStr, project, apiKey string) (*Client, error) {
	if len(urlPartedStr) == 0 {
		return nil, fmt.Errorf("missing FCM endpoint URL")
	}

	if len(apiKey) == 0 {
		return nil, fmt.Errorf("missing api key")
	}

	if len(project) == 0 {
		return nil, fmt.Errorf("missing project")
	}

	urlStr := fmt.Sprintf(urlPartedStr, project)
	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse url: %s", urlStr)
	}

	client := &Client{
		URL:        parsedURL.String(),
		HTTPClient: http.DefaultClient,
		APIKey:     apiKey,
	}

	return client, nil
}

func (c *Client) newRequest(ctx context.Context, method string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, c.URL, body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	req.Header.Add("Content-Type", "application/json")

	return req, nil
}

func (c *Client) Send(ctx context.Context, payload *Payload) (*Message, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest(ctx, "POST", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to read res.Body and the status code of the response from FCM was not 200")
		}
		return nil, fmt.Errorf("status code: %d; body: %s", res.StatusCode, b)
	}

	msg := &Message{}

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(msg)

	return msg, err
}
