package main

import (
	"WebRTCBroadcaster/signal"
	"fmt"
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/x264"
	//_ "github.com/pion/mediadevices/pkg/driver/camera"
	 _ "github.com/pion/mediadevices/pkg/driver/videotest"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/webrtc/v3"
	"image/jpeg"
	"os"
)

func main() {
	track := getCameraVideoTrack()

	offerChan := startHTTPSDPServer

	for {
		newPeerSDP := <- offerChan()

		connection := onConnect(newPeerSDP)

		connection.AddTrack(track)
	}
}

func getCameraVideoTrack() *mediadevices.VideoTrack {
	x264Params, _ := x264.NewParams()
	x264Params.Preset = x264.PresetMedium
	x264Params.BitRate = 1_000_000 // 1mbps

	codecSelector := mediadevices.NewCodecSelector(
		mediadevices.WithVideoEncoders(&x264Params),
	)

	cameraMediaStream, _ := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(c *mediadevices.MediaTrackConstraints) {
			c.Width = prop.Int(1920)
			c.Height = prop.Int(1080)
		},
		Codec: codecSelector,

	})

	track := cameraMediaStream.GetVideoTracks()[0]
	videoTrack := track.(*mediadevices.VideoTrack)

	return videoTrack
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