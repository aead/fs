// Copyright (c) 2018 Andreas Auernhammer. All rights reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package auth

import (
	"net/http"

	"github.com/aead/sf"
	"github.com/minio/minio-go/pkg/s3signer"
	"github.com/minio/minio-go/pkg/s3utils"
)

// SignerV2 returns a Fuzzer signing the request
// using AWS signature V2.
func SignerV2(accessKey, secretKey string) sf.Fuzzer {
	return &signerV2{AccessKey: accessKey, SecretKey: secretKey}
}

type signerV2 struct {
	AccessKey, SecretKey string
}

func (s *signerV2) Fuzz(req *http.Request) error {
	for k, v := range s3signer.SignV2(*req, s.AccessKey, s.SecretKey, s3utils.IsVirtualHostSupported(*req.URL, "")).Header {
		req.Header[k] = v
	}
	return nil
}
