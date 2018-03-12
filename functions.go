// Copyright (c) 2018 Andreas Auernhammer. All rights reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package sf

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// Filter is a fuzzer filtering HTTP header keys.
// All keys that let Filter evaluate to true are
// removed from the request.
type Filter func(string) bool

// Fuzz filter out all keys from the HTTP request headers
// which let the fuzzer evaluate to true.
func (f Filter) Fuzz(req *http.Request) error {
	for k := range req.Header {
		if f(k) {
			req.Header.Del(k)
		}
	}
	return nil
}

// And combines two filters such that the returned filter
// returns true iff both filters return true.
func (f Filter) And(fn Filter) Filter { return func(k string) bool { return f(k) && fn(k) } }

// Not inverts a filter such that the returned filter
// returns true if the original filter returned false
// and vice versa.
func (f Filter) Not() Filter { return func(k string) bool { return !f(k) } }

// Or combines two filters such that the returned filter
// returns true if one of the original filters returns true.
func (f Filter) Or(fn Filter) Filter { return func(k string) bool { return f(k) || fn(k) } }

// Map is a fuzzer transforming HTTP headers.
// The map function is applied to every header key-value pair
// which is than replaced by the transformed key-value-pair.
type Map func(string, string) (string, string)

// Fuzz transforms all key-value pairs from the HTTP request
// headers using the map fuzzer.
func (m Map) Fuzz(req *http.Request) error {
	for k := range req.Header {
		v := req.Header.Get(k)
		req.Header.Del(k)
		req.Header.Set(m(k, v))
	}
	return nil
}

// Insert is a fuzzer inserting a key-value pair
// into the HTTP headers. If the key already exists
// than Insert replaces the old value with the new
// value.
type Insert func() (string, string)

// Fuzz inserts a key-value pair into the HTTP request
// headers.
func (f Insert) Fuzz(req *http.Request) error {
	req.Header.Set(f())
	return nil
}

// Logger returns a new Fuzzer which writes the
// HTTP request URL and the headers to dst.
// If dst is nil Logger uses STDOUT as default.
func Logger(dst io.Writer) Fuzzer {
	if dst == nil {
		dst = os.Stdout
	}
	return logger{dst: dst}
}

type logger struct {
	dst io.Writer
}

func (l logger) Fuzz(req *http.Request) error {
	if _, err := fmt.Fprintln(l.dst, req.URL); err != nil {
		return err
	}
	for k, v := range req.Header {
		if _, err := fmt.Fprintln(l.dst, k, v); err != nil {
			return err
		}
	}
	return nil
}
