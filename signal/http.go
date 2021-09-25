/// Copyright (c) 2018 Pion
/// https://github.com/pion/webrtc

package signal

import (
	"io"
	"io/ioutil"
	"net/http"
)

// HTTPSDPServer starts a HTTP Server that consumes SDPs
func HTTPSDPServer(mux *http.ServeMux) (chan string, chan string) {
	sdpChan := make(chan string)
	answerChan := make(chan string)

	mux.HandleFunc("/sdp", func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		sdpChan <- string(body)

		answer := <-answerChan

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, answer)
	})

	return sdpChan, answerChan
}
