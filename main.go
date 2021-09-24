package main

import (
	"WebRTCBroadcaster/camera"
	"WebRTCBroadcaster/dummy"
	"WebRTCBroadcaster/signal"
	"flag"
	"fmt"
	"github.com/pion/mediadevices"
	"github.com/pion/webrtc/v3"
	"image/jpeg"
	"os"
)

func main() {
	// Args
	isDummy := flag.Bool("dummy", false, "カメラデバイスを使わず、ダミー映像で配信する")
	width := flag.Int("width", 1080, "カメラデバイスから取得する解像度の幅")
	height := flag.Int("height", 1920, "カメラデバイスから取得する解像度の高さ")

	flag.Parse()

	var track *mediadevices.VideoTrack
	if *isDummy {
		track = dummy.GetCameraVideoTrack(*width, *height)
	} else {
		track = camera.GetCameraVideoTrack(*width, *height)
	}

	offerChan := startHTTPSDPServer

	for {
		newPeerSDP := <- offerChan()

		connection := onConnect(newPeerSDP)

		connection.AddTrack(track)
	}
}

func startHTTPSDPServer() chan string{
	sdpChan := signal.HTTPSDPServer()

	return sdpChan
}

func onConnect(sdp string) *webrtc.PeerConnection  {
	offer := webrtc.SessionDescription{}
	signal.Decode(sdp, &offer)

	peerConnectionConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	peerConnection, err := webrtc.NewPeerConnection(peerConnectionConfig)
	if err != nil {
		panic(err)
	}
	defer func() {
		if cErr := peerConnection.Close(); cErr != nil {
			fmt.Printf("cannot close peerConnection: %v\n", cErr)
		}
	}()

	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	<-gatherComplete

	return peerConnection
}

// GetCameraFrame 現在のカメラの1フレームを取得する
func GetCameraFrame(videoTrack *mediadevices.VideoTrack) {
	videoReader := videoTrack.NewReader(false)
	frame, release, _ := videoReader.Read()
	defer release()

	output, _ := os.Create("frame.jpg")
	jpeg.Encode(output, frame, nil)
}