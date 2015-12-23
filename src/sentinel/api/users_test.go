// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"io/ioutil"
	"net/http"
	"testing"

	"sentinel"
	"sentinel/datastore"
	"sentinel/tokens"

	"code.google.com/p/go-uuid/uuid"
)

func init() {
	f, err := ioutil.ReadFile("../tokens/testdata/sentinel")
	if err != nil {
		panic(err)
	}
	privateKey = string(f)
	f, err = ioutil.ReadFile("../tokens/testdata/sentinel.pub")
	if err != nil {
		panic(err)
	}
	publicKey = string(f)
}

func TestUserGetUserDetails(t *testing.T) {
	setup()

	email := "anna@example.com"
	password := "secretninja"
	user := &sentinel.User{
		UID: uuid.NewRandom(),
		AuthEmailList: []*sentinel.AuthEmail{
			&sentinel.AuthEmail{Email: email},
		},
	}
	datastore.SetPassword(user, password)

	store.Users.(*sentinel.MockUsersService).ListFn = func(opt sentinel.UserListOptions) ([]*sentinel.User, error) {
		users := []*sentinel.User{user}
		return users, nil
	}

	calledGet := false
	store.Users.(*sentinel.MockUsersService).GetUserDetailsFn = func(uid uuid.UUID) (*sentinel.User, error) {
		if !uuid.Equal(user.UID, uid) {
			t.Errorf("Result should have been %v, but it was %v", user.UID, uid)
		}
		calledGet = true
		return user, nil
	}

	err := apiClient.Authenticate(email, password)
	if err != nil {
		t.Error(err)
	}

	u, err := apiClient.Users.GetUserDetails(user.UID)
	if err != nil {
		t.Error(err)
	}

	if !calledGet {
		t.Error("!calledGet")
	}

	expected := user.UID
	result := u.UID
	if !uuid.Equal(expected, result) {
		t.Errorf("Result should have been %v, but it was %v", expected, result)
	}
}

func TestSignup(t *testing.T) {
	setup()

	expectedEmail := "jess@example.com"
	expectedPassword := "princess"
	expectedUID := uuid.NewRandom()

	user := &sentinel.User{
		UID:          expectedUID,
		PasswordHash: "plain:" + expectedPassword,
		AuthEmailList: []*sentinel.AuthEmail{
			&sentinel.AuthEmail{Email: expectedEmail},
		},
	}

	calledSubmit := false
	store.Users.(*sentinel.MockUsersService).SignupFn = func(email, password string) (*sentinel.User, error) {
		if expectedEmail != email {
			t.Errorf("Expected request for user %+v, but received %+v", expectedEmail, email)
		}
		if expectedPassword != password {
			t.Errorf("Expected request for password %+v, but received %+v", expectedPassword, password)
		}
		calledSubmit = true
		return user, nil
	}

	user, err := apiClient.Users.Signup(expectedEmail, expectedPassword)
	if err != nil {
		t.Error(err)
	}
	if !calledSubmit {
		t.Error("!calledSubmit")
	}
	if !uuid.Equal(expectedUID, user.UID) {
		t.Errorf("Result should have been %v, but it was %v", expectedUID, user.UID)
	}
}

func TestAuthorized(t *testing.T) {
	expected := uuid.NewRandom()

	claims := tokens.Claims{
		"email":   "jess@example.com",
		"user_id": expected.String(),
	}
	user := &sentinel.User{UID: expected}
	store.Users.(*sentinel.MockUsersService).GetUserDetailsFn = func(uid uuid.UUID) (*sentinel.User, error) {
		if !uuid.Equal(uid, user.UID) {
			return nil, ErrInvalidClient
		}
		return user, nil
	}

	tokenStr, err := tokens.Sign(claims, privateKey, &tokens.AccessTokenOptions)
	if err != nil {
		t.Fatal(err)
	}
	c := sentinel.NewClient(nil)
	c.SetToken(tokenStr)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Authorize(req); err != nil {
		t.Fatal(err)
	}
	u, err := Authorized(req)
	if err != nil {
		t.Fatal(err)
	}

	result := u.UID
	if !uuid.Equal(expected, result) {
		t.Errorf("Result should have been %v, but it was %v", expected, result)
	}
}

func TestUnAuthorizedInvalidToken(t *testing.T) {
	claims := tokens.Claims{
		"email":   "jess@example.com",
		"user_id": "82f051f1-977d-430d-8119-134d3abb8171",
	}
	tokenStr, err := tokens.Sign(claims, privateKey, &tokens.DefaultOptions)
	if err != nil {
		t.Fatal(err)
	}

	c := sentinel.NewClient(nil)
	c.SetToken(tokenStr)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Authorize(req); err != nil {
		t.Fatal(err)
	}

	_, err = Authorized(req)

	expected := ErrInvalidAuthenticationToken
	result := err
	if expected != result {
		t.Errorf("Result should have been %v, but it was %v", expected, result)
	}
}

func TestAuthenticate(t *testing.T) {
	setup()

	expectedEmail := "jess@example.com"
	expectedPassword := "princess"
	expectedUID := uuid.NewRandom()

	user := &sentinel.User{
		UID:          expectedUID,
		PasswordHash: "plain:" + expectedPassword,
		AuthEmailList: []*sentinel.AuthEmail{
			&sentinel.AuthEmail{Email: expectedEmail},
		},
	}

	store.Users.(*sentinel.MockUsersService).ListFn = func(opt sentinel.UserListOptions) ([]*sentinel.User, error) {
		email := opt.Email[0]
		if expectedEmail != email {
			t.Errorf("Expected request for user %+v, but received %+v", expectedEmail, email)
		}
		return []*sentinel.User{user}, nil
	}

	err := apiClient.Authenticate(expectedEmail, expectedPassword)
	if err != nil {
		t.Error(err)
	}
}

func TestserveAckEmail(t *testing.T) {
	setup()

	emailID := uuid.NewRandom()

	calledSubmit := false
	store.Users.(*sentinel.MockUsersService).AckEmailFn = func(uid uuid.UUID) error {
		if !uuid.Equal(uid, emailID) {
			t.Errorf("Expected request for user %+v, but received %+v", emailID, uid)
		}
		calledSubmit = true
		return nil
	}

	err := apiClient.Users.AckEmail(emailID)
	if err != nil {
		t.Fatal(err)
	}
	if !calledSubmit {
		t.Error("!calledSubmit")
	}
}

func TestserveAddEmail(t *testing.T) {
	setup()

	expectedEmail := "jess@example.com"

	e := &sentinel.AuthEmail{
		Email: expectedEmail,
	}
	calledSubmit := false
	store.Users.(*sentinel.MockUsersService).AddEmailFn = func(userID uuid.UUID, email string) (*sentinel.AuthEmail, error) {
		if expectedEmail != email {
			t.Errorf("Expected request for user %+v, but received %+v", expectedEmail, email)
		}
		calledSubmit = true
		return e, nil
	}

	e, err := apiClient.Users.AddEmail(uuid.NewRandom(), expectedEmail)
	if err != nil {
		t.Error(err)
	}
	if !calledSubmit {
		t.Error("!calledSubmit")
	}
	result := e.Email
	if expectedEmail != result {
		t.Errorf("Result should have been %v, but it was %v", expectedEmail, result)
	}
}

func TestserveDelEmail(t *testing.T) {
	setup()

	expectedEmailID := uuid.NewRandom()

	calledSubmit := false
	store.Users.(*sentinel.MockUsersService).DelEmailFn = func(id uuid.UUID) error {
		if !uuid.Equal(expectedEmailID, id) {
			t.Errorf("Expected request for user %+v, but received %+v", expectedEmailID, id)
		}
		calledSubmit = true
		return nil
	}

	if err := apiClient.Users.DelEmail(expectedEmailID); err != nil {
		t.Error(err)
	}
}

func TestserveUpdateUserDetails(t *testing.T) {
	setup()

	optExpect := sentinel.UserUpdateOptions{
		Name:             "Jane Brody",
		Password:         "bluebird",
		DefaultAuthLevel: sentinel.AuthLevelSecure,
	}

	calledUpdateDetails := false
	store.Users.(*sentinel.MockUsersService).UpdateDetailsFn = func(userID uuid.UUID, opt sentinel.UserUpdateOptions) (*sentinel.User, error) {
		if opt.Name != optExpect.Name {
			t.Errorf("Expected request for opt.Name %+v, but received %+v", optExpect.Name, opt.Name)
		}
		if opt.Password != optExpect.Password {
			t.Errorf("Expected request for opt.Password %+v, but received %+v", optExpect.Password, opt.Password)
		}
		if opt.DefaultAuthLevel != optExpect.DefaultAuthLevel {
			t.Errorf("Expected request for opt.DefaultAuthLevel %+v, but received %+v", optExpect.DefaultAuthLevel, opt.DefaultAuthLevel)
		}

		u := &sentinel.User{
			UID:              userID,
			Name:             opt.Name,
			DefaultAuthLevel: opt.DefaultAuthLevel,
		}
		datastore.SetPassword(u, opt.Password)

		calledUpdateDetails = true
		return u, nil
	}

	user, err := apiClient.Users.UpdateDetails(uuid.NewRandom(), optExpect)
	if err != nil {
		t.Error(err)
	}
	if !calledUpdateDetails {
		t.Error("!calledUpdateDetails")
	}

	if user.Name != optExpect.Name {
		t.Errorf("Result should have been %v, but it was %v", user.Name, optExpect.Name)
	}
	if user.DefaultAuthLevel != optExpect.DefaultAuthLevel {
		t.Errorf("Result should have been %v, but it was %v", user.DefaultAuthLevel, optExpect.DefaultAuthLevel)
	}
	if err := datastore.ComparePassword(user, optExpect.Password); err != nil {
		t.Error(err)
	}
}

func TestserveOneTimeLogin(t *testing.T) {
	//TODO: implement
}
