// Copyright (c) 2018 Andreas Auernhammer. All rights reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package s3

import (
	gobase64 "encoding/base64"
	"io"
	"strconv"
	gotime "time"

	"github.com/aead/sf"
)

// Type represents a type of an AWS S3 HTTP header.
type Type interface {
	// Random returns a random value of this type.
	Random(rand sf.Random) string

	// Kind returns the kind this type belongs to.
	Kind() Kind
}

// Kind represents all different kinds of AWS S3 HTTP
// header types.
type Kind string

const (
	// Enum is a set of values
	Enum Kind = "enum"
	// Int is a natural number
	Int Kind = "int"
	// Time is a date in time.
	Time Kind = "time"
	// Base64 is a base64-encoded string
	Base64 Kind = "base64"
)

// TypeOf returns the type of the provided header.
// The additional ok value indicates whether the header
// exists.
func TypeOf(header string) (t Type, ok bool) {
	t, ok = headers[header]
	return
}

// TypeOfKind returns a random type of the requested kind.
// The additional ok value indicates whether such a kind exists
// and whether there are any types of that kind.
func TypeOfKind(kind Kind, rand sf.Random) (t Type, ok bool) {
	types, ok := values[kind]
	if !ok || len(types) == 0 {
		return nil, ok
	}
	return TypeOf(types[rand.Int()%len(types)])
}

var ( // check interface compatibility at compile time
	_ Type = (*s3Int)(nil)
	_ Type = (enum)(nil)
	_ Type = (*time)(nil)
	_ Type = (*base64)(nil)
)

var values = map[Kind][]string{}

func init() {
	for k, v := range headers {
		kind := v.Kind()
		values[kind] = append(values[kind], k)
	}
}

const (
	formatISO8601 = "20060102T150405Z"
)

var headers = map[string]Type{
	"Content-Length":                                  s3Int{},
	"Content-MD5":                                     base64{16},
	"Expect":                                          enum([]string{"100-continue"}),
	"Date":                                            time([]string{gotime.RFC1123, gotime.RFC1123Z, formatISO8601}),
	"X-Amz-Content-Sha256":                            base64{32},
	"X-Amz-Date":                                      time([]string{gotime.RFC1123, gotime.RFC1123Z, formatISO8601}),
	"X-Amz-Server-Side-Encryption":                    enum([]string{"AES256", "aws:kms"}),
	"X-Amz-Server-Side-Encryption-Customer-Algorithm": enum([]string{"AES256"}),
	"X-Amz-Server-Side-Encryption-Customer-Key":       base64{32},
	"X-Amz-Server-Side-Encryption-Customer-Key-Md5":   base64{16},
	"X-Amz-Server-Side-Encryption-Context":            base64{-8 * 1024}, // max 8 KB
}

type s3Int struct{}

func (s3Int) Random(rand sf.Random) string { return strconv.Itoa(rand.Int()) }
func (s3Int) Kind() Kind                   { return Int }

type enum []string

func (e enum) Random(rand sf.Random) string { return e[rand.Int()%len(e)] }
func (enum) Kind() Kind                     { return Enum }

type time []string

func (t time) Random(rand sf.Random) string {
	format := t[rand.Int()%len(t)]
	return rand.Date().Format(format)
}
func (time) Kind() Kind { return Time }

type base64 struct{ Length int }

func (b base64) Random(rand sf.Random) string {
	if b.Length < 0 { // choose a random length
		b.Length = rand.Int() % (-b.Length)
	}
	data := make([]byte, b.Length)
	if _, err := io.ReadFull(&rand, data); err != nil {
		panic("ran out of randomness")
	}
	return gobase64.StdEncoding.EncodeToString(data)
}
func (base64) Kind() Kind { return Base64 }
