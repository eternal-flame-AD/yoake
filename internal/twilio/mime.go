package twilio

import "mime"

func init() {
	mime.AddExtensionType(".mp3", "audio/mpeg")
}
