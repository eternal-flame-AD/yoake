package funcmap

import "fmt"

func ThemeColor(key string) (string, error) {
	if c, ok := nipponColors[key]; ok {
		return c, nil
	}
	return "", fmt.Errorf("no such color %s", key)
}
