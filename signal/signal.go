/// Copyright (c) 2018 Pion
/// https://github.com/pion/webrtc

// Package signal contains helpers to exchange the SDP session
// description between examples.
package signal

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// Allows compressing offer/answer to bypass terminal input limits.
const compress = false

// MustReadStdin blocks until input is received from stdin
func MustReadStdin() (string, error) {
	r := bufio.NewReader(os.Stdin)

	var in string
	for {
		var err error
		in, err = r.ReadString('\n')
		if err != io.EOF {
			if err != nil {
				return "", err
			}
		}
		in = strings.TrimSpace(in)
		if len(in) > 0 {
			break
		}
	}

	fmt.Println("")

	return in, nil
}

// Encode encodes the input in base64
// It can optionally zip the input before encoding
func Encode(obj interface{}) (string, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}

	if compress {
		b, err = zip(b)
		if err != nil {
			return "", err
		}
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

// Decode decodes the input from base64
// It can optionally unzip the input after decoding
func Decode(in string, obj interface{}) error {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return err
	}

	if compress {
		b, err = unzip(b)
		if err != nil {
			return err
		}
	}

	err = json.Unmarshal(b, obj)
	if err != nil {
		return err
	}

	return nil
}

func zip(in []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	_, err := gz.Write(in)
	if err != nil {
		return nil, err
	}
	err = gz.Flush()
	if err != nil {
		return nil, err
	}
	err = gz.Close()
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func unzip(in []byte) ([]byte, error) {
	var b bytes.Buffer
	_, err := b.Write(in)
	if err != nil {
		return nil, err
	}
	r, err := gzip.NewReader(&b)
	if err != nil {
		return nil, err
	}
	res, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return res, nil
}
