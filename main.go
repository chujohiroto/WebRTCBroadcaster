package main

import (
	"WebRTCBroadcaster/camera"
	"WebRTCBroadcaster/config"
	"WebRTCBroadcaster/dummy"
	"WebRTCBroadcaster/signal"
	"context"
	"flag"
	"fmt"
	"github.com/pion/mediadevices"
	"github.com/pion/webrtc/v3"
	"image/jpeg"
	"io"
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
	port := flag.Int("port", 8080, "シグナリングやテストで閲覧するためのWebページを表示するポート")
	isDummy := flag.Bool("dummy", false, "カメラデバイスを使わず、ダミー映像で配信する")
	width := flag.Int("width", 1920, "カメラデバイスから取得する解像度の幅")
	height := flag.Int("height", 1080, "カメラデバイスから取得する解像度の高さ")
	framerate := flag.Float64("framerate", 30, "フレームレート")
	config.AuthnWebhookURL = flag.String("webhook", "", "認証WebHookのURL")
	isAPI := flag.Bool("api", true, "画像、動画取得APIを有効にする")

	flag.Parse()

	var track *mediadevices.VideoTrack
	var api *webrtc.API

	if *isDummy {
		track, api = dummy.GetCameraVideoTrack(*width, *height, *framerate)
		log.Println("ダミー映像を取得")
	} else {
		track, api = camera.GetCameraVideoTrack(*width, *height, *framerate)
		log.Println("カメラデバイスから映像を取得")
	}

	if track == nil {
		panic("Get Camera Video Track Nil")
	}

	// HTTPのハンドル周り定義
	mux := http.NewServeMux()
	offerChan, answerChan := startHTTPSDPServer(mux)

	// もしテストページを起動するなら、staticファイル追加
	if *isViewPage {
		srv := http.FileServer(http.Dir("html"))
		mux.Handle("/", srv)
		log.Println("テストページを起動")
	}

	// 画像、動画取得API
	if *isAPI {
		mux.HandleFunc("/api/photo", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "image/jpeg")
			GetCameraFrame(track, w)
			log.Println("API Call Get PhotoImage " + r.RemoteAddr)
		})

		mux.HandleFunc("/api/movie", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "video/mp4")
			GetCameraMovie(track, w)
			log.Println("API Call Get PhotoImage " + r.RemoteAddr)
		})

	}

	// HTTPサーバー起動
	go func() {
		log.Println("HTTP Serverを起動　Port:" + strconv.Itoa(*port))

		err := http.ListenAndServe(":"+strconv.Itoa(*port), mux)
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

			log.Println("New SDF Offer" /* offer.SDP */)

			connection, err := onConnect(offer, track, answerChan, api)
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

	return sdpChan, answerChan
}

func onConnect(offer *webrtc.SessionDescription, track *mediadevices.VideoTrack, answerChan chan string, api *webrtc.API) (*webrtc.PeerConnection, error) {
	// Connection生成
	peerConnectionConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	peerConnection, err := api.NewPeerConnection(peerConnectionConfig)
	if err != nil {
		return nil, err
	}

	// Video周り設定
	track.OnEnded(func(err error) {
		fmt.Printf("Track (ID: %s) ended with error: %v\n", track.ID(), err)
	})

	_, err = peerConnection.AddTransceiverFromTrack(track,
		webrtc.RtpTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionSendonly,
		},
	)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	// Offerを登録
	err = peerConnection.SetRemoteDescription(*offer)
	if err != nil {
		return nil, err
	}

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		log.Printf("Peer Connection State has changed: %s\n", s.String())

		if s == webrtc.PeerConnectionStateFailed {
			peerConnection.Close()
		}
	})

	// Gatheringが完了するまで待機
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		return nil, err
	}

	<-gatherComplete

	// Send Answer
	answerBody, err := signal.Encode(answer)
	if err != nil {
		return nil, err
	}
	answerChan <- answerBody

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
func GetCameraFrame(videoTrack *mediadevices.VideoTrack, w io.Writer) {
	videoReader := videoTrack.NewReader(false)
	frame, release, _ := videoReader.Read()
	defer release()

	jpeg.Encode(w, frame, nil)
}

// GetCameraMovie 指定された秒数分映像を取得して、エンコードして出力
func GetCameraMovie(videoTrack *mediadevices.VideoTrack, w io.Writer) {
	return
}
