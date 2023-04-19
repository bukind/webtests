package main

import (
	"fmt"
	"net/http"
	"strings"
)

func main() {
	fs := http.FileServer(http.Dir("assets"))
	err := http.ListenAndServe(":9090", http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		// Remove when dev is done.
		resp.Header().Add("Cache-Control", "no-cache")
		if strings.HasSuffix(req.URL.Path, ".wasm") {
			resp.Header().Set("content-type", "application/wasm")
		}
		fs.ServeHTTP(resp, req)
	}))
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
