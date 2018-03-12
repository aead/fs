// Copyright (c) 2018 Andreas Auernhammer. All rights reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package sf

import (
	"net/http"
)

// Fuzzer is the generic interface for adding
// S3 fuzzing functionality. A fuzzer takes an
// HTTP request and modifies the request in an
// arbitrary way.
type Fuzzer interface {
	// Fuzz modifies the HTTP request depending
	// on the fuzzer implementation. It may retrun
	// any encountered error during modification.
	//
	// If Fuzz returns an non-nil error the state
	// of the request is undefined.
	Fuzz(*http.Request) error
}

// LoopFuzzer is the generic interface for a
// statefull and adjusting fuzzer.
//
// A LoopFuzzer can be seen as a Fuzzer which
// modifies an HTTP request before sending it.
// When the HTTP response is received the
// LoopFuzzer looks at the response and may
// adjusts its fuzzing strategy.
type LoopFuzzer interface {
	Fuzzer

	// Adjust can extract information about the
	// HTTP response so that the fuzzer can adjust
	// its fuzzing strategy. It MUST NOT modify the
	// response.
	Adjust(*http.Response)
}

// MultiFuzzer combines a list of fuzzers into a
// single fuzzer.
type MultiFuzzer []Fuzzer

// Fuzz applies every fuzzer to the provided request.
// It returns the first error encountered.
func (mf MultiFuzzer) Fuzz(req *http.Request) error {
	for _, f := range mf {
		if err := f.Fuzz(req); err != nil {
			return err
		}
	}
	return nil
}

// Adjust passes the HTTP response to every LoopFuzzer
// which is part of the MultiFuzzer.
func (mf MultiFuzzer) Adjust(resp *http.Response) {
	for _, f := range mf {
		if lf, ok := f.(LoopFuzzer); ok {
			lf.Adjust(resp)
		}
	}
}

// RegisterFuzzer wraps a http.RounderTripper. The returned
// http.RounderTripper modifies all HTTP request using the
// provided fuzzers before the request is processed any further.
func RegisterFuzzer(rt http.RoundTripper, f ...Fuzzer) http.RoundTripper {
	return fuzzer{rt: rt, fz: MultiFuzzer(f)}
}

type fuzzer struct {
	rt http.RoundTripper
	fz LoopFuzzer
}

func (f fuzzer) RoundTrip(r *http.Request) (*http.Response, error) {
	if err := f.fz.Fuzz(r); err != nil {
		return nil, err
	}
	resp, err := f.rt.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	f.fz.Adjust(resp)
	return resp, nil
}
