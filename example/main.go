package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/fastly/compute-sdk-go/fsthttp"
	"github.com/joeshaw/fastly-roundtripper/transport"
)

func main() {
	fsthttp.ServeFunc(func(ctx context.Context, w fsthttp.ResponseWriter, r *fsthttp.Request) {
		t := transport.New("ipv4")
		t.AddBackend("ipv6", "ipv6.joeshaw.org")
		t.Request = func(fstreq *fsthttp.Request) error {
			fstreq.CacheOptions.Pass = true
			fstreq.Header.Set("Fastly-Debug", "1")
			return nil
		}
		client := &http.Client{Transport: t}

		url := "https://ipv4.joeshaw.org/ip"
		if r.URL.Path == "/ipv6" {
			url = "https://ipv6.joeshaw.org/ip"
		}

		resp, err := client.Get(url)
		if err != nil {
			w.WriteHeader(fsthttp.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %s\n", err)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(resp.StatusCode)

		// Copy the response body to the response writer.
		if _, err := io.Copy(w, resp.Body); err != nil {
			w.WriteHeader(fsthttp.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %s\n", err)
			return
		}

		fmt.Fprintf(w, "\n---\n")

		d, _ := httputil.DumpResponse(resp, true)
		fmt.Fprintf(w, "%s\n", d)

		freq := transport.FastlyRequestFromContext(resp.Request.Context())
		fmt.Fprintf(w, "%v\n", freq.Host)

		fresp := transport.FastlyResponseFromContext(resp.Request.Context())
		fmt.Fprintf(w, "%v\n", fresp.Backend)
	})
}
