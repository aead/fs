[![Godoc Reference](https://godoc.org/github.com/aead/sf?status.svg)](https://godoc.org/github.com/aead/sf)
[![Travis CI](https://travis-ci.org/aead/sf.svg?branch=master)](https://travis-ci.org/aead/sf)
[![Go Report Card](https://goreportcard.com/badge/aead/sf)](https://goreportcard.com/report/aead/sf)

# S3 Fuzzing

A prototype for an S3 fuzzing library.

## Get Started

```
package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aead/sf"
	"github.com/aead/sf/headers"
	minio "github.com/minio/minio-go"
)

func main() {
	host, accessKey, secretkey := "localhost:9000", "ACCESS_KEY", "SECRET_KEY"

	s3Client, err := minio.NewV2(host, accessKey, secretkey, true)
	if err != nil {
		fmt.Println("Failed to create S3 client:", err)
        return
	}
	customTrans := http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	rand := sf.NewRandom(time.Now().Unix())
	fuzzers := sf.MultiFuzzer{
		sf.Insert(func() (string, string) {
			length, _ := headers.TypeOf("Content-Length")
			return "Content-Length", length.Random(rand)
		}),
		sf.Filter(func(k string) bool { return strings.HasPrefix(k, "User-Agent") }),
		sf.Logger(nil),
	}

	s3Client.SetCustomTransport(sf.RegisterFuzzer(&customTrans, fuzzers...))

	data := make([]byte, 1024)
	_, err = s3Client.PutObject("bucket", "object", bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{})
	if err != nil {
		fmt.Println(err)
		return
	}
}
```