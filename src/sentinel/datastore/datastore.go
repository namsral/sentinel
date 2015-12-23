// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package datastore

import (
	"sentinel"

	"github.com/jmoiron/sqlx"
)

type Datastore struct {
	Users    sentinel.UsersService
	Services sentinel.ServicesService
	db       *sqlx.DB
}

func NewDatastore(db *sqlx.DB) *Datastore {
	if db == nil {
		db = DB
	}

	d := &Datastore{db: db}
	d.Users = &usersStore{Datastore: d}
	d.Services = &servicesStore{Datastore: d}
	return d
}

func NewMockDatastore() *Datastore {
	return &Datastore{
		Users:    &sentinel.MockUsersService{},
		Services: &sentinel.MockServicesService{},
	}
}
