// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"sentinel"
	"sentinel/api"
	"sentinel/datastore"
)

var (
	baseURLStr = flag.String("baseurl", "http://sentinel.sh", "baseurl of the Sentinel backend service")
	baseURL    *url.URL
	apiclient  = sentinel.NewClient(nil)
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, `sentinel is the API backend for sentinel.sh

Usage:

	sentinel [options] command [arguments]

The options are:
`)
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, `
The commands are:
`)
		for _, c := range subcmds {
			fmt.Fprintf(os.Stderr, "\t%-24s %s\n", c.name, c.description)
		}
		fmt.Fprintln(os.Stderr, `
	Use "sentinel command -h" for more information about the command
`)
		os.Exit(1)
	}
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
	}
	log.SetFlags(0)

	var err error
	baseURL, err = url.Parse(*baseURLStr)
	if err != nil {
		log.Fatal(err)
	}
	apiclient.BaseURL.ResolveReference(&url.URL{Path: "/api/v1"})

	subcmd := flag.Arg(0)
	for _, c := range subcmds {
		if c.name == subcmd {
			c.run(flag.Args()[1:])
			return
		}
	}

	fmt.Fprintf(os.Stderr, "unknown subcommand %q\n", subcmd)
	fmt.Fprintf(os.Stderr, "Run \"sentinel -h\" for usage.\n")
	for _, c := range subcmds {
		fmt.Fprintf(os.Stderr, "\t%s\t%s\n", c.name, c.description)
	}
	os.Exit(1)
}

type subcmd struct {
	name, description string
	run               func(args []string)
}

var subcmds = []subcmd{
	{"serve", "run the API backend service", serveCmd},
	{"createdb", "create the database schema", createDBCmd},
}

func serveCmd(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	httpAddr := fs.String("http", "localhost:6002", "HTTP service address")
	fs.Parse(args)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `Usage: %s serve [options]

Starts the backend service

Options:
`, fs.Args()[0])
		fs.PrintDefaults()
		os.Exit(1)
	}

	if fs.NArg() != 0 {
		fs.Usage()
	}

	datastore.Connect()

	m := http.NewServeMux()
	api.SetbaseURL(baseURL.ResolveReference(&url.URL{Path: "/api/v1/"}))
	m.Handle("/api/v1/", api.Handler())

	log.Print("Listening on ", *httpAddr)
	err := http.ListenAndServe(*httpAddr, m)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func createDBCmd(args []string) {
	fs := flag.NewFlagSet("createdb", flag.ExitOnError)
	drop := fs.Bool("drop", false, "drop DB before creating")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `Usage: %s createdb [options]

Creates the necessary DB tables and indexes.

Options:
`, fs.Args()[0])
		fs.PrintDefaults()
		os.Exit(1)
	}
	fs.Parse(args)

	if fs.NArg() != 0 {
		fs.Usage()
	}

	datastore.Connect()
	if *drop {
		datastore.Drop()
	}
	datastore.Create()
}
