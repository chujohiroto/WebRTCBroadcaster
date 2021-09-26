# WebRTCBroadcaster
WebRTCでカメラ映像(H.264)を複数人に配信するソフトウェアです。Raspberry Pi 3, 4のH.264 ハードウェアエンコーダに対応。

Windows Subsystem for Linuxでのテスト映像での動作は確認済み。未確認ですがWindows, Macでもおそらく動作します。

複数人で一つのカメラ映像を低遅延で閲覧することに特化したシンプルな設計です。

## 機能
- カメラデバイスから映像を配信
- テスト映像を配信
- HTTPで動く簡易的なWebRTCシグナリングサーバー
- 簡易閲覧ページ
- 認証Webhook
- JPEGスクリーンショットを取得できるHTTP API
- H264に限定したエンコード

**認証Webhookに関しては時雨堂様のAyameを参考にしました。**

**一部利用しているコードに関してはライセンス表記させていただいています。この場を借りてお礼申し上げます。**

**WebhookのAPIフォーマットについては異なります。また開発元は異なりますので時雨堂様に問い合わせしないようにお願いします。**

## 実行方法
```shell
 ./WebRTCBroadcaster -h
Usage of ./WebRTCBroadcaster:
  -api
        画像、動画取得APIを有効にする (default true)
  -bitrate int
        エンコードする映像ビットレート、帯域 (default 1000000) = 1Mbps
  -dummy
        カメラデバイスを使わず、ダミー映像で配信する
  -framerate float
        フレームレート (default 30)
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

**1080p 60FPS 5Mbpsで配信したい場合の実行例**
```shell
./WebRTCBroadcaster -page -framerate 60 -bitrate 5000000
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

## 認証WebhookのAPIフォーマット
記述中です...

## 対応予定の機能
- 録画HTTP API
- カスタムできる閲覧ページのテンプレート
- 簡易なWeb SDK
- 閲覧状況を確認できるダッシュボード、HTTP API
- Rasberry Pi 3, 4のビルド済みバイナリの提供