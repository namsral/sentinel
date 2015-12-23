// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package datastore

import (
	"errors"
	"time"

	"sentinel"

	"code.google.com/p/go-uuid/uuid"
)

const serviceTable = "services"
const serviceTableCreateStmt = `
CREATE TABLE services (
    id SERIAL PRIMARY KEY, -- internal identifier
    uid uuid UNIQUE not null, -- uuid identifier
    name TEXT NOT NULL,
    baseurl TEXT NOT NULL,
    logourl TEXT NOT NULL,
    authlevel INTEGER NOT NULL DEFAULT 0, -- 0:notify 1:fast 2:secure
    lastentry_at TIMESTAMP(0),
    created_at TIMESTAMP(0),
    updated_at TIMESTAMP(0),
    is_archived BOOLEAN NOT NULL DEFAULT FALSE
);
`

const serviceInsertStmt = `
INSERT INTO services(uid, name, baseurl, logourl, authlevel, lastentry_at, 
    created_at, updated_at, is_archived)
VALUES (:uid, :name, :baseurl, :logourl, :authlevel, :lastentry_at,
    :created_at, :updated_at, :is_archived) RETURNING id
;`

type servicesStore struct {
	*Datastore
}

func (s *servicesStore) submit(service *sentinel.Service) (*sentinel.Service, error) {
	var err error

	if service.UID == nil {
		service.UID = uuid.NewRandom()
	}

	now := time.Now().UTC()
	if service.ID == 0 {
		service.CreatedAt = now
	}
	service.UpdatedAt = now

	stmt, err := s.db.PrepareNamed(serviceInsertStmt)
	if err != nil {
		return nil, err
	}
	var serviceID int
	err = stmt.QueryRowx(service).Scan(&serviceID)
	if err != nil {
		return nil, err
	}

	service.ID = serviceID
	return service, nil
}

func (s *servicesStore) Get(uid uuid.UUID) (*sentinel.Service, error) {
	var service sentinel.Service
	err := s.db.QueryRowx(`SELECT * FROM services WHERE is_archived=FALSE AND uid=$1;`, uid).StructScan(&service)
	if err != nil {
		return nil, err
	}

	return &service, nil
}

func (s *servicesStore) Auth(uid uuid.UUID, email, status string) error {
	service, err := s.Get(uid)
	if err != nil {
		return err
	}

	opt := &sentinel.UserListOptions{Email: []string{email}}
	users, err := s.Datastore.Users.List(*opt)
	if err != nil {
		return err
	}

	if len(users) == 0 {
		return errors.New("email address not found")
	}

	//TODO: finish implement
	_ = service

	return nil
}
