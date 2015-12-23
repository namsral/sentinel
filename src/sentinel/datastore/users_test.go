// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package datastore

import (
	"database/sql"
	"reflect"
	"testing"

	"sentinel"
)

func TestUsersStoreGetUserDetails(t *testing.T) {
	expected := users[0]
	d := NewDatastore(DB)

	result, err := d.Users.GetUserDetails(expected.UID)
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(expected, result) {
		t.Fatalf("Result should have been %v, but it was %v", expected, result)
	}
}

func TestUsersStoreSignup(t *testing.T) {
	email := "anna@example.com"
	password := "unicorn"

	d := NewDatastore(DB)
	_, err := d.Users.Signup(email, password)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUsersStoreList(t *testing.T) {
	d := NewDatastore(DB)

	opt := sentinel.UserListOptions{
		Email: []string{
			"bob@example.com",
			"jane@example.com",
		},
	}

	users, err := d.Users.List(opt)
	if err != nil {
		t.Fatal(err)
	}

	expected := len(opt.Email)
	result := len(users)
	if expected != result {
		t.Fatalf("Result should have been %v, but it was %v", expected, result)
	}
}

func TestUsersUpdateDetails(t *testing.T) {
	d := NewDatastore(DB)
	expect := "Jesse Pinkman"

	user := users[0]
	opt := sentinel.UserUpdateOptions{
		Name: expect,
	}
	u, err := d.Users.UpdateDetails(user.UID, opt)
	if err != nil {
		t.Fatal(err)
	}
	result := u.Name
	if expect != result {
		t.Fatalf("Result should have been %v, but it was %v", expect, result)
	}
}

func TestAckEmail(t *testing.T) {
	d := NewDatastore(DB)

	emailID := users[0].AuthEmailList[0].UID
	err := d.Users.AckEmail(emailID)
	if err != nil {
		t.Fatal(err)
	}

	err = d.Users.AckEmail(emailID)
	if err != sql.ErrNoRows {
		t.Fatal(err)
	}
}

func TestAddEmail(t *testing.T) {
	d := NewDatastore(DB)

	user := users[0]
	expected := "bob.smith@example.com"

	authEmail, err := d.Users.AddEmail(user.UID, expected)
	if err != nil {
		t.Fatal(err)
	}
	result := authEmail.Email
	if expected != result {
		t.Fatalf("Result should have been %v, but it was %v", expected, result)
	}
}

func TestDelEmail(t *testing.T) {
	d := NewDatastore(DB)

	user := users[0]
	authEmail := user.AuthEmailList[0]
	err := d.Users.DelEmail(authEmail.UID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestListEmail(t *testing.T) {
	d := NewDatastore(DB)

	emails, err := d.Users.ListEmail(nil)
	if err != nil {
		t.Error(err)
	}
	result := len(emails) > 0
	expect := true
	if expect != result {
		t.Fatalf("Result should have been %v, but it was %v", expect, result)
	}
}
