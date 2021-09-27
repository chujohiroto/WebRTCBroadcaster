package dummy

import (
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/mmal"
	// "github.com/pion/mediadevices/pkg/codec/x264"
	_ "github.com/pion/mediadevices/pkg/driver/videotest"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/webrtc/v3"
	"log"
)

func GetCameraVideoTrack(width, height, bitrate int, framerate float64) (*mediadevices.VideoTrack, *webrtc.API) {
	mmalParams, _ := mmal.NewParams()
	mmalParams.BitRate = bitrate

	codecSelector := mediadevices.NewCodecSelector(
		mediadevices.WithVideoEncoders(&mmalParams),
	)

	mediaEngine := webrtc.MediaEngine{}
	codecSelector.Populate(&mediaEngine)
	api := webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine))

	cameraMediaStream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(c *mediadevices.MediaTrackConstraints) {
			c.Width = prop.Int(width)
			c.Height = prop.Int(height)
			c.FrameRate = prop.Float(framerate)
		},
		Codec: codecSelector,
	})

	if err != nil {
		log.Println(err.Error())
		return nil, nil
	}

	if len(cameraMediaStream.GetVideoTracks()) == 0 {
		return nil, nil
	}

	track := cameraMediaStream.GetVideoTracks()[0]
	videoTrack := track.(*mediadevices.VideoTrack)

	return videoTrack, api
}
