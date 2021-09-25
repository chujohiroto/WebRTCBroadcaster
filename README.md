# WebRTCBroadcaster
WebRTCでカメラ映像(H.264)を複数人に配信するソフトウェア　Raspberry Pi 3 4対応

## 機能

## 実行方法
```shell
./WebRTCBroadcaster -dummy -h
Usage of ./WebRTCBroadcaster:
  -dummy
        カメラデバイスを使わず、ダミー映像で配信する
  -height int
        カメラデバイスから取得する解像度の高さ (default 1080)
  -page
        テストで閲覧するためのWebページを表示する (default true)
  -port int
        シグナリングやテストで閲覧するためのWebページを表示するポート (default 8080)
  -webhook string
        認証WebHookのURL
  -width int
        カメラデバイスから取得する解像度の幅 (default 1920)
```

**ダミー映像での実行例**
```shell
./WebRTCBroadcaster -dummy -page
```

**カメラから映像取得、認証Webhookありでの実行例**
```shell
./WebRTCBroadcaster -page -webhook http://localhost:8888/auth
```

## ビルド方法
**Raspberry Piでハードウェアエンコーダを使用する場合は、本体に内臓されている [mmal](https://github.com/raspberrypi/userland/tree/master/interface/mmal) が必要です。**

**ハードウェアエンコーダが必要な場合は、Rasberry Piでビルドしてください。**

### Raspberry Pi
```shell
sudo apt install pkg-config

go build

./WebRTCBroadcaster
```

### Ubuntu
```shell
sudo apt install pkg-config

sudo apt install libx264-dev

go build

./WebRTCBroadcaster
```