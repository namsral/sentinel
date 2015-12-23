// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sentinel

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"sentinel/router"

	"code.google.com/p/go-uuid/uuid"
)

type Service struct {
	ID          int       `json:"-"`
	UID         uuid.UUID `db:"uid" json:"id"`
	Name        string    `json:"name"`
	BaseURL     string    `db:"baseurl" json:"serviceUrl"`
	LogoURL     string    `db:"logourl" json:"serviceLogoUrl"`
	AuthLevel   int       `db:"authlevel" json:"authLevel"`
	LastEntryAt time.Time `db:"lastentry_at" json:"lastEntryDate"`

	CreatedAt  time.Time `db:"created_at" json:"-"`
	UpdatedAt  time.Time `db:"updated_at" json:"-"`
	IsArchived bool      `db:"is_archived" json:"-"`
}

// ServicesService interacts with the service-related endpoint in Sentinel's API.
type ServicesService interface {
	Get(uid uuid.UUID) (*Service, error)
	Auth(uid uuid.UUID, email, status string) error
}

type servicesService struct {
	client *Client
}

func (s *servicesService) Get(uid uuid.UUID) (*Service, error) {
	url, err := s.client.url(router.Service, nil, nil)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	if err := s.client.Authorize(req); err != nil {
		return nil, err
	}

	var service *Service
	if _, err = s.client.Do(req, &service); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *servicesService) Auth(uid uuid.UUID, email, status string) error {
	u, err := s.client.url(router.AuthService, nil, nil)
	if err != nil {
		return err
	}

	form := &url.Values{
		"service_id": {uid.String()},
		"email":      {email},
		"status":     {status},
	}
	req, err := s.client.NewRequest("POST", u.String(), form)
	if err != nil {
		return err
	}

	var user User
	resp, err := s.client.Do(req, &user)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New("API reponded with status " + http.StatusText(resp.StatusCode))
	}

	return nil
}

type ServiceListOptions struct {
	// IncludeArchived will include archived/inactive services
	IncludeArchived bool

	ListOptions
}

type MockServicesService struct {
	GetFn  func(uid uuid.UUID) (*Service, error)
	AuthFn func(uid uuid.UUID, email, status string) error
}

var _ ServicesService = &MockServicesService{}

func (s *MockServicesService) Get(uid uuid.UUID) (*Service, error) {
	if s.GetFn == nil {
		return nil, nil
	}
	return s.GetFn(uid)
}

func (s *MockServicesService) Auth(uid uuid.UUID, email, status string) error {
	if s.AuthFn == nil {
		return nil
	}
	return s.AuthFn(uid, email, status)
}
