package http

import (
	"compress/gzip"
	"compress/zlib"
	"dna"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

var DefaulHeader = http.Header{
	"Accept-Encoding": []string{"gzip,deflate"},
	"Accept-Language": []string{"en-US,en"},
	"Cache-Control":   []string{"max-age=0"},
	"Connection":      []string{"keep-alive"},
	"User-Agent":      []string{"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"},
	"Cookie":          []string{""},
}

var client = &http.Client{}

// Get impliments getting site with basic properties.
// Enable gzip, deflat by default to reduce  network data, redirect to new location from response.
// It returns data (String type) and error
// if err is nil then data is "" (empty).
func Get(url dna.String) (*Result, error) {
	req, err := http.NewRequest("GET", url.ToPrimitiveValue(), nil)
	req.Header = DefaulHeader
	// req.Header.Add("Accept-Encoding", "gzip,deflate")
	// req.Header.Add("Accept-Language", "en-US,en")
	// req.Header.Add("Cache-Control", "max-age=0")
	// req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Host", url.ToPrimitiveValue())
	// req.Header.Add("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	// req.Header.Add("Cookie", "")
	// dna.Log(req.Header)
	res, err := client.Do(req)
	if err != nil {
		return new(Result), err
	}

	var data []byte
	var myErr error

	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		var reader io.ReadCloser
		reader, err := gzip.NewReader(res.Body)
		if err != nil {
			return new(Result), err
		}
		data, myErr = ioutil.ReadAll(reader)
		reader.Close()
	case "deflate":
		// Logv("sdsafsd")
		reader, err := zlib.NewReader(res.Body)
		if err != nil {
			return new(Result), err
		}
		data, myErr = ioutil.ReadAll(reader)
		reader.Close()
	default:
		data, myErr = ioutil.ReadAll(res.Body)
	}

	if myErr != nil {
		return new(Result), myErr
	}

	res.Body.Close()
	return NewResult(dna.Int(res.StatusCode), dna.String(res.Status), res.Header, dna.String(data)), nil
}

func Post(url dna.String, bodyStr dna.String) (*Result, error) {
	client := &http.Client{}
	// dna.Log(bodyStr)
	req, err := http.NewRequest("POST", url.String(), strings.NewReader(bodyStr.String()))
	req.Header = DefaulHeader
	req.Header.Add("Host", url.ToPrimitiveValue())
	res, err := client.Do(req)
	if err != nil {
		return new(Result), err
	}

	var data []byte
	var myErr error

	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		var reader io.ReadCloser
		reader, err := gzip.NewReader(res.Body)
		if err != nil {
			return new(Result), err
		}
		data, myErr = ioutil.ReadAll(reader)
		reader.Close()
	case "deflate":
		// Logv("sdsafsd")
		reader, err := zlib.NewReader(res.Body)
		if err != nil {
			return new(Result), err
		}
		data, myErr = ioutil.ReadAll(reader)
		reader.Close()
	default:
		data, myErr = ioutil.ReadAll(res.Body)
	}

	if myErr != nil {
		return new(Result), myErr
	}

	res.Body.Close()
	return NewResult(dna.Int(res.StatusCode), dna.String(res.Status), res.Header, dna.String(data)), nil
}
