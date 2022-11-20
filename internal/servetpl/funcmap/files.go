package funcmap

import "net/url"

func FileAccess(path string) string {
	return "https://yoake.yumechi.jp/file_access.html?path=" + url.QueryEscape(path)
}
