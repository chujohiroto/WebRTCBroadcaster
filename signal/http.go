/// Copyright (c) 2018 Pion
/// https://github.com/pion/webrtc

package signal

import (
	"WebRTCBroadcaster/webhook"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

/// {
///		"sdp_offer": "base64",
///		"authnMetadata": {}
///	}
type sdpRequest struct {
	SDPOffer      string       `json:"sdp_offer"`
	AuthnMetadata *interface{} `json:"authnMetadata,omitempty"`
}

// HTTPSDPServer starts a HTTP Server that consumes SDPs
func HTTPSDPServer(mux *http.ServeMux) (chan string, chan string) {
	sdpChan := make(chan string)
	answerChan := make(chan string)

	mux.HandleFunc("/sdp", func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		req := sdpRequest{}
		err := json.Unmarshal(body, req)
		if err != nil {
			return
		}

		resp, err := webhook.AuthnWebhook(req.AuthnMetadata)
		if err != nil {
			log.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if resp.Allowed != nil && *resp.Allowed == true {
			sdpChan <- req.SDPOffer

			// Answerがくるまで待機
			answer := <-answerChan

			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, answer)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	})

	return sdpChan, answerChan
}
