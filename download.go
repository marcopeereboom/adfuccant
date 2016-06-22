package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
)

func downloadToMem(url string, skipVerify bool) ([]byte, error) {
	tr := &http.Transport{}
	if skipVerify {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client := &http.Client{Transport: tr}

	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %v", res.Status)
	}

	return ioutil.ReadAll(res.Body)
}
