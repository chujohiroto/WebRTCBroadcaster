package camera

import (
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/x264"
	_ "github.com/pion/mediadevices/pkg/driver/camera"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/webrtc/v3"
)

func GetCameraVideoTrack(width, height int) (*mediadevices.VideoTrack, *webrtc.API) {
	x264Params, _ := x264.NewParams()
	x264Params.Preset = x264.PresetMedium
	x264Params.BitRate = 1_000_000 // 1mbps

	codecSelector := mediadevices.NewCodecSelector(
		mediadevices.WithVideoEncoders(&x264Params),
	)

	mediaEngine := webrtc.MediaEngine{}
	codecSelector.Populate(&mediaEngine)
	api := webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine))

	cameraMediaStream, _ := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(c *mediadevices.MediaTrackConstraints) {
			c.Width = prop.Int(1920)
			c.Height = prop.Int(1080)
		},
		Codec: codecSelector,
	})

	if len(cameraMediaStream.GetVideoTracks()) == 0 {
		return nil, nil
	}

	track := cameraMediaStream.GetVideoTracks()[0]
	videoTrack := track.(*mediadevices.VideoTrack)

	return videoTrack, api
}
