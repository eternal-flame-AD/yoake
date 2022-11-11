package funcmap

import (
	"io/ioutil"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

const (
	ResponseTypeHTML         = "html"
	ResponseTypeStrippedHTML = "html_stripped"
	ResponseTypeText         = "text"
)

func HttpRequest(method string, URL string, selector string, responseType string) (data interface{}, err error) {
	if method == "" {
		method = http.MethodGet
	}
	if responseType == "" {
		responseType = ResponseTypeHTML
	}
	req, err := http.NewRequest(method, URL, nil)
	if err != nil {
		return nil, err
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if responseType == ResponseTypeHTML || responseType == ResponseTypeStrippedHTML {
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return nil, err
		}
		result := doc.Contents()
		if selector != "" {
			result = result.Find(selector)
		}
		if responseType == ResponseTypeStrippedHTML {
			return result.Text(), nil
		}
		return result.Html()
	}
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return string(response), nil
}
