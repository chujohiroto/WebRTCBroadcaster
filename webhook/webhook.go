package webhook

import (
	"WebRTCBroadcaster/config"
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type httpResponse struct {
	Status string      `json:"status"`
	Proto  string      `json:"proto"`
	Header http.Header `json:"header"`
	Body   string      `json:"body"`
}

// JSON HTTP Request をするだけのラッパー
func postRequest(u string, body interface{}) (*http.Response, error) {
	reqJSON, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		u,
		bytes.NewBuffer(reqJSON),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	timeout := time.Duration(5) * time.Second

	client := &http.Client{Timeout: timeout}
	return client.Do(req)
}

var (
	errAuthnWebhook                     = errors.New("AuthnWebhookError")
	errAuthnWebhookResponse             = errors.New("AuthnWebhookResponseError")
	errAuthnWebhookUnexpectedStatusCode = errors.New("AuthnWebhookUnexpectedStatusCode")
	errAuthnWebhookReject               = errors.New("AuthnWebhookReject")
)

type authnWebhookRequest struct {
	AuthnMetadata *interface{} `json:"authnMetadata,omitempty"`
}

type authnWebhookResponse struct {
	Allowed       *bool        `json:"allowed"`
	Reason        *string      `json:"reason"`
	AuthzMetadata *interface{} `json:"authzMetadata"`
}

func AuthnWebhook(req *interface{}) (*authnWebhookResponse, error) {
	if config.AuthnWebhookURL == "" {
		var allowed = true
		authnWebhookResponse := &authnWebhookResponse{Allowed: &allowed}
		return authnWebhookResponse, nil
	}

	resp, err := postRequest(config.AuthnWebhookURL, req)
	if err != nil {
		log.Println("AuthnWebhookError body: \n" + err.Error())
		return nil, errAuthnWebhook
	}

	// http://ikawaha.hateblo.jp/entry/2015/06/07/074155
	defer resp.Body.Close()

	log.Println("Auth Webhook Request")

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("AuthnWebhookResponseError body: \n" + string(body))

		return nil, err
	}

	// ログ出力用
	httpResponse := &httpResponse{
		Status: resp.Status,
		Proto:  resp.Proto,
		Header: resp.Header,
		Body:   string(body),
	}

	// 200 以外で返ってきたときはエラーとする
	if resp.StatusCode != 200 {
		log.Println("AuthnWebhookUnexpectedStatusCode HTTP Status: " + httpResponse.Status)
		return nil, errAuthnWebhookUnexpectedStatusCode
	}

	log.Println("Auth Webhook Response HTTP Status: " + httpResponse.Status)

	authnWebhookResponse := authnWebhookResponse{}
	if err := json.Unmarshal(body, &authnWebhookResponse); err != nil {
		log.Println("AuthnWebhookResponseError " + err.Error())
		return nil, errAuthnWebhookResponse
	}

	return &authnWebhookResponse, nil
}
