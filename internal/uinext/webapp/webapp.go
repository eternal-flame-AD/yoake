package webapp

type IWebApp struct {
	BasePath string

	TrimaImgBase string
}

var Singleton IWebApp
