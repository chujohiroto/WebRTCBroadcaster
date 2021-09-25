/// Copyright (c) 2018 Pion
/// https://github.com/pion/webrtc

package signal

import (
	"io/ioutil"
	"net/http"
)

// HTTPSDPServer starts a HTTP Server that consumes SDPs
func HTTPSDPServer(mux *http.ServeMux) chan string {
	sdpChan := make(chan string)
	mux.HandleFunc("/sdp", func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		sdpChan <- string(body)
	})

	return sdpChan
}
