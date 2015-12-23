// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"testing"

	"sentinel"

	"code.google.com/p/go-uuid/uuid"
)

var testServices = []*sentinel.Service{
	&sentinel.Service{
		UID:       uuid.Parse("0a991da9-b01d-418d-8d56-9fb56fa78b22"),
		Name:      "Shoeland",
		BaseURL:   "https://api.shoeland.example.com/status",
		LogoURL:   "https://cdn.shoeland.example.com/i/logo.png",
		AuthLevel: 0,
	},
	&sentinel.Service{
		UID:       uuid.Parse("c590bab6-09cf-404d-b4fb-07880bbdd4fe"),
		Name:      "Rent-A-Dog",
		BaseURL:   "https://api.rent-a-dog.example.com/status",
		LogoURL:   "https://cdn.rent-a-dog.example.com/i/logo.png",
		AuthLevel: 1,
	},
}

func TestserveGetService(t *testing.T) {
	setup()

	service := testServices[0]

	calledGet := false
	store.Services.(*sentinel.MockServicesService).GetFn = func(uid uuid.UUID) (*sentinel.Service, error) {
		if !uuid.Equal(service.UID, uid) {
			t.Errorf("Result should have been %v, but it was %v", service.UID, uid)
		}
		calledGet = true
		return service, nil
	}

	s, err := apiClient.Services.Get(service.UID)
	if err != nil {
		t.Error(err)
	}

	if !calledGet {
		t.Error("!calledGet")
	}

	expect := service.UID
	result := s.UID
	if !uuid.Equal(expect, result) {
		t.Errorf("Result should have been %v, but it was %v", expect, result)
	}
}

func TestserveSetServiceStatus(t *testing.T) {
	setup()

	expectUID := uuid.NewRandom()
	expectEmail := "jack@example.com"
	expectStatus := "accepted"

	calledSetServiceStatus := false
	store.Services.(*sentinel.MockServicesService).AuthFn = func(uid uuid.UUID, email, status string) error {
		if !uuid.Equal(expectUID, uid) {
			t.Errorf("Result should have been %v, but it was %v", expectUID, uid)
		}
		if expectEmail != email {
			t.Errorf("Result should have been %v, but it was %v", expectEmail, email)
		}
		if expectStatus != status {
			t.Errorf("Result should have been %v, but it was %v", expectStatus, status)
		}
		calledSetServiceStatus = true
		return nil
	}

	err := apiClient.Services.Auth(expectUID, expectEmail, expectStatus)
	if err != nil {
		t.Error(err)
	}

	if !calledSetServiceStatus {
		t.Error("!calledSetServiceStatus")
	}
}
