package main

import (
	"WebRTCBroadcaster/camera"
	"WebRTCBroadcaster/dummy"
	"WebRTCBroadcaster/signal"
	"context"
	"flag"
	"fmt"
	"github.com/pion/mediadevices"
	"github.com/pion/webrtc/v3"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	oss "os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	// Args
	isViewPage := flag.Bool("page", true, "テストで閲覧するためのWebページを表示する")
	webPort := flag.Int("webport", 8080, "シグナリングやテストで閲覧するためのWebページを表示するポート")
	isDummy := flag.Bool("dummy", false, "カメラデバイスを使わず、ダミー映像で配信する")
	width := flag.Int("width", 1920, "カメラデバイスから取得する解像度の幅")
	height := flag.Int("height", 1080, "カメラデバイスから取得する解像度の高さ")

	flag.Parse()

	var track *mediadevices.VideoTrack
	if *isDummy {
		track = dummy.GetCameraVideoTrack(*width, *height)
		log.Println("ダミー映像を取得")
	} else {
		track = camera.GetCameraVideoTrack(*width, *height)
		log.Println("カメラデバイスから映像を取得")
	}

	// HTTPのハンドル周り定義
	mux := http.NewServeMux()
	offerChan, answerChan := startHTTPSDPServer(mux)

	// もしテストページを起動するなら、staticファイル追加
	if *isViewPage {
		srv := http.FileServer(http.Dir("html"))
		mux.Handle("/", srv)
	}

	// HTTPサーバー起動
	go func() {
		log.Println("Testで閲覧するためのHTTP Serverを起動")

		err := http.ListenAndServe(":"+strconv.Itoa(*webPort), mux)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		for {
			newPeerSDP := <-offerChan

			newPeerSDP = strings.Replace(newPeerSDP, "\"", "", -1)

			offer, err := sdpDecode(newPeerSDP)
			if err != nil {
				log.Println(err.Error())
				answerChan <- ""
				continue
			}

			log.Println("New SDF Offer\n" + offer.SDP)

			connection, err := onConnect(offer, answerChan)

			if err != nil {
				log.Println(err.Error())
				connection.Close()
				continue
			}

			track.OnEnded(func(err error) {
				fmt.Printf("Track (ID: %s) ended with error: %v\n", track.ID(), err)
			})

			_, err = connection.AddTransceiverFromTrack(track,
				webrtc.RtpTransceiverInit{
					Direction: webrtc.RTPTransceiverDirectionSendonly,
				},
			)

			if err != nil {
				log.Println(err.Error())
				connection.Close()
				continue
			}

			log.Println("Connected")
		}
	}()

	quit := make(chan os.Signal, 1)
	oss.Notify(quit, syscall.SIGTERM, os.Interrupt)
	log.Printf("SIGNAL %d received, then shutting down...\n", <-quit)

	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown処理を書く コンテキストを渡した場合、5秒以内に終了しない場合は処理がキャンセルされる
	err := track.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Println("Server shutdown")

}

func startHTTPSDPServer(mux *http.ServeMux) (chan string, chan string) {
	sdpChan, answerChan := signal.HTTPSDPServer(mux)

	log.Println("SDPを受け付けるHTTP Serverを起動")

	return sdpChan, answerChan
}

func onConnect(offer *webrtc.SessionDescription, answerChan chan string) (*webrtc.PeerConnection, error) {
	peerConnectionConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	peerConnection, err := webrtc.NewPeerConnection(peerConnectionConfig)
	if err != nil {
		return nil, err
	}

	err = peerConnection.SetRemoteDescription(*offer)
	if err != nil {
		return nil, err
	}

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	answerBody, err := signal.Encode(answer)
	if err != nil {
		return nil, err
	}

	answerChan <- answerBody

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		return nil, err
	}

	<-gatherComplete

	return peerConnection, nil
}

func sdpDecode(sdp string) (*webrtc.SessionDescription, error) {
	offer := webrtc.SessionDescription{}
	err := signal.Decode(sdp, &offer)

	if err != nil {
		return nil, err
	}

	return &offer, nil
}

// GetCameraFrame 現在のカメラの1フレームを取得する
func GetCameraFrame(videoTrack *mediadevices.VideoTrack) {
	videoReader := videoTrack.NewReader(false)
	frame, release, _ := videoReader.Read()
	defer release()

	output, _ := os.Create("frame.jpg")
	jpeg.Encode(output, frame, nil)
}
