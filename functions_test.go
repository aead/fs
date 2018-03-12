package sf

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

var now = time.Now()

func equal(h1, h2 http.Header) bool {
	if len(h1) != len(h2) {
		return false
	}

	for k1, v1 := range h1 {
		v2 := h2[k1]
		if len(v1) != len(v2) {
			return false
		}
		for i := range v1 {
			if v1[i] != v2[i] {
				return false
			}
		}
	}
	return true
}

func TestFilter(t *testing.T) {
	for i, test := range FilterTests {
		req := &http.Request{Header: test.headers}
		if err := test.f.Fuzz(req); err != nil {
			t.Errorf("Test %d: Fuzzer failed with: %v", i, err)
		}
		if !equal(req.Header, test.want) {
			t.Errorf("Test %d: filtered headers do not match wanted headers", i)
		}
	}
}

func TestMap(t *testing.T) {
	for i, test := range MapTests {
		req := &http.Request{Header: test.headers}
		if err := test.f.Fuzz(req); err != nil {
			t.Errorf("Test %d: Fuzzer failed with: %v", i, err)
		}
		if !equal(req.Header, test.want) {
			t.Errorf("Test %d: mapped headers do not match wanted headers", i)
		}
	}
}
func TestInsert(t *testing.T) {
	for i, test := range InsertTests {
		req := &http.Request{Header: test.headers}
		if err := test.f.Fuzz(req); err != nil {
			t.Errorf("Test %d: Fuzzer failed with: %v", i, err)
		}
		if !equal(req.Header, test.want) {
			t.Errorf("Test %d: mapped headers do not match wanted headers", i)
		}
	}
}

var InsertTests = []struct {
	headers, want http.Header
	f             Insert
}{
	{
		headers: http.Header{
			"Date": []string{now.Format(time.RFC1123)},
		},
		want: http.Header{
			"Date":       []string{now.Format(time.RFC1123)},
			"X-Amz-Date": []string{now.Format(time.RFC1123)},
		},
		f: func() (string, string) { return "X-Amz-Date", now.Format(time.RFC1123) },
	},
	{headers: http.Header{
		"Date":           []string{now.Format(time.RFC1123)},
		"Content-Length": []string{"42"},
	},
		want: http.Header{
			"Date":           []string{now.Format(time.RFC1123)},
			"Content-Length": []string{"0"},
		},
		f: func() (string, string) { return "Content-Length", "0" },
	},
}

var MapTests = []struct {
	headers, want http.Header
	f             Map
}{
	{
		headers: http.Header{
			"Date":       []string{now.Format(time.RFC1123)},
			"X-Amz-Date": []string{now.Format(time.RFC1123)},
		},
		want: http.Header{
			"Date":       []string{now.Format(time.RFC1123)},
			"X-Amz-Date": []string{now.Format(time.RFC1123)},
		},
		f: func(k, v string) (string, string) { return k, v },
	},
	{
		headers: http.Header{
			"Date":       []string{now.Format(time.RFC1123)},
			"X-Amz-Date": []string{now.Format(time.RFC1123)},
		},
		want: http.Header{
			"Date":       []string{"Mon, 12 Mar 2018 14:30:56 CET"},
			"X-Amz-Date": []string{"Mon, 12 Mar 2018 14:30:56 CET"},
		},
		f: func(k, v string) (string, string) { return k, "Mon, 12 Mar 2018 14:30:56 CET" },
	},
	{
		headers: http.Header{
			"X-Amz-Date":                   []string{now.Format(time.RFC1123)},
			"X-Amz-Server-Side-Encryption": []string{"AES256"},
			"Content-Length":               []string{"42"},
		},
		want: http.Header{
			"Date": []string{now.Format(time.RFC1123)},
			"Server-Side-Encryption": []string{"AES256"},
			"X-Amz-Content-Length":   []string{"42"},
		},
		f: func(k, v string) (string, string) {
			if strings.HasPrefix(k, "X-Amz-") {
				return strings.Replace(k, "X-Amz-", "", 1), v
			}
			return "X-Amz-" + k, v
		},
	},
}

var FilterTests = []struct {
	headers, want http.Header
	f             Filter
}{
	{
		headers: http.Header{
			"Date":       []string{now.Format(time.RFC1123)},
			"X-Amz-Date": []string{now.Format(time.RFC1123)},
		},
		want: http.Header{},
		f:    func(k string) bool { return strings.Contains(k, "Date") },
	},
	{
		headers: http.Header{
			"Date":                         []string{now.Format(time.RFC1123)},
			"X-Amz-Date":                   []string{now.Format(time.RFC1123)},
			"X-Amz-Server-Side-Encryption": []string{"AES256"},
		},
		want: http.Header{
			"Date": []string{now.Format(time.RFC1123)},
			"X-Amz-Server-Side-Encryption": []string{"AES256"},
		},
		f: Filter(func(k string) bool {
			return strings.Contains(k, "X-Amz-")
		}).And(func(k string) bool {
			return strings.Contains(k, "Date")
		}),
	},
	{
		headers: http.Header{
			"Date":                         []string{now.Format(time.RFC1123)},
			"X-Amz-Date":                   []string{now.Format(time.RFC1123)},
			"X-Amz-Server-Side-Encryption": []string{"AES256"},
		},
		want: http.Header{
			"Date":       []string{now.Format(time.RFC1123)},
			"X-Amz-Date": []string{now.Format(time.RFC1123)},
		},
		f: Filter(func(k string) bool {
			return strings.Contains(k, "Date")
		}).Not(),
	},
	{
		headers: http.Header{
			"Date":                         []string{now.Format(time.RFC1123)},
			"X-Amz-Date":                   []string{now.Format(time.RFC1123)},
			"X-Amz-Server-Side-Encryption": []string{"AES256"},
			"Content-Length":               []string{"42"},
		},
		want: http.Header{
			"Content-Length": []string{"42"},
		},
		f: Filter(func(k string) bool {
			return strings.Contains(k, "Date")
		}).Or(func(k string) bool {
			return strings.Contains(k, "X-Amz-")
		}),
	},
}
