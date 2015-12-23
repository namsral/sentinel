// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate go run scripts/docs.go

package api

import (
	"net/http"
)

func serveAPIDocs(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "text/plain; chartset=utf-8")
	w.Write([]byte(docsRaml))
	return nil
}
