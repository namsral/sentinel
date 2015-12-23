// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// Reads all .raml files in the current folder
// and encode them as strings literals in docs_generated.go
func main() {
	fs, _ := ioutil.ReadDir(".")
	out, _ := os.Create("docs_generated.go")
	out.Write([]byte("//Do not edit this file, it is generated.\npackage api\n\nconst (\n"))
	for _, f := range fs {
		if strings.HasSuffix(f.Name(), ".raml") {
			out.Write([]byte(strings.TrimSuffix(f.Name(), ".raml") + "Raml = `"))
			f, _ := os.Open(f.Name())
			io.Copy(out, f)
			out.Write([]byte("`\n"))
			f.Close()
		}
	}
	out.Write([]byte(")\n"))
	out.Close()
}
