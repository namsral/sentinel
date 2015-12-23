// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package datastore

import (
	"log"
	"os"
	"strings"

	"sentinel"

	"code.google.com/p/go-uuid/uuid"
)

var users []*sentinel.User
var services []*sentinel.Service

func init() {
	dbname := os.Getenv("PGDATABASE")
	if len(dbname) == 0 {
		dbname = "sentinel_test"
	}
	if !strings.HasSuffix(dbname, "_test") {
		dbname += "_test"
	}
	if err := os.Setenv("PGDATABASE", dbname); err != nil {
		log.Fatal(err)
	}

	// Reset the database
	Connect()
	Drop()
	Create()
	SetupTest()
}

func SetupTest() {
	d := NewDatastore(DB)

	users = []*sentinel.User{
		&sentinel.User{
			UID:              uuid.Parse("5c5c21e5-3286-4db4-bae9-d72cf3fcf1ec"),
			Name:             "Bob",
			PasswordHash:     "plain:ninja",
			DefaultAuthLevel: 1,
			IsArchived:       false,
			AuthEmailList: []*sentinel.AuthEmail{&sentinel.AuthEmail{
				Email:      "bob@example.com",
				UID:        uuid.Parse("4d300cb7-74f2-476c-b4c6-41c1189e4986"),
				IsVerified: false,
			}},
		},
		&sentinel.User{
			UID:              uuid.Parse("d806b977-7f8c-479f-84c7-5625b6b82863"),
			Name:             "Jane",
			PasswordHash:     "plain:princess",
			DefaultAuthLevel: 2,
			IsArchived:       false,
			AuthEmailList: []*sentinel.AuthEmail{&sentinel.AuthEmail{
				Email:      "jane@example.com",
				UID:        uuid.Parse("6fdab498-63c9-4b83-8d51-9b18c103b451"),
				IsVerified: false,
			}},
		},
	}

	for _, e := range users {
		_, err := d.Users.(*usersStore).Submit(e)
		if err != nil {
			panic(err)
		}
	}

	services = []*sentinel.Service{
		&sentinel.Service{
			UID:       uuid.Parse("309f7158-b4bf-4181-acfd-30cf6f7a9d19"),
			Name:      "Secure Mail",
			BaseURL:   "https://api.securemail.example.com/status",
			LogoURL:   "https://cdn.securemail.example.com/i/logo.png",
			AuthLevel: 0,
		},
		&sentinel.Service{
			UID:       uuid.Parse("c658145a-acba-4b15-a97d-2ccdd857e6de"),
			Name:      "Doc Cloud",
			BaseURL:   "https://api.doccloud.example.com/status",
			LogoURL:   "https://cdn.doccloud.example.com/i/logo.png",
			AuthLevel: 1,
		},
	}

	for _, e := range services {
		_, err := d.Services.(*servicesStore).submit(e)
		if err != nil {
			panic(err)
		}
	}
}
