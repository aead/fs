// Copyright (c) 2018 Andreas Auernhammer. All rights reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

// Package auth provides authenticating and signing functionality
// for S3 requests.
package auth

import (
	"net/http"

	"github.com/aead/sf"
	"github.com/minio/minio-go/pkg/s3signer"
	"github.com/minio/minio-go/pkg/s3utils"
)

// SignerV4 returns a Fuzzer signing the request
// using AWS signature V4.
func SignerV4(accessKey, secretKey, token string) sf.Fuzzer {
	return &signerV2{AccessKey: accessKey, SecretKey: secretKey, SessionToken: token}
}

type signerV4 struct {
	AccessKey, SecretKey, SessionToken string
}

func (s *signerV4) Fuzz(req *http.Request) error {
	region := s3utils.GetRegionFromURL(*req.URL)
	if region == "" {
		region = "us-east-1" // default to us-east-1
	}
	for k, v := range s3signer.SignV4(*req, s.AccessKey, s.SecretKey, s.SessionToken, region).Header {
		req.Header[k] = v
	}
	return nil
}
