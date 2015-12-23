// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tokens

// Keypair was generated using the following commands:
//   $ openssl genrsa -out sentinel 2048
//   $ openssl rsa -in sentinel -pubout

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/zhevron/jwt"
)

var (
	privateKey, publicKey string
)

func init() {
	f, err := ioutil.ReadFile("./testdata/sentinel")
	if err != nil {
		panic(err)
	}
	privateKey = string(f)
	f, err = ioutil.ReadFile("./testdata/sentinel.pub")
	if err != nil {
		panic(err)
	}
	publicKey = string(f)
}

func TestSignAndVerify(t *testing.T) {
	expect := Claims{
		"email":   "bob@example.com",
		"user_id": "373707eb-db20-4b1c-bf8c-505f19d9ccf5",
	}
	opt := DefaultOptions

	tokenStr, err := Sign(expect, privateKey, &opt)
	if err != nil {
		t.Fatal(err)
	}

	result, err := Verify(tokenStr, publicKey, &opt)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(result, expect) {
		t.Errorf("Result should have been %v, but it was %v", expect, result)
	}
}

func TestFailVerify(t *testing.T) {
	expect := jwt.ErrNoneAlgorithmWithSecret

	claims := Claims{
		"email":   "bob@example.com",
		"user_id": "373707eb-db20-4b1c-bf8c-505f19d9ccf5",
	}
	opt1 := DefaultOptions
	opt2 := opt1
	opt2.Algorithm = jwt.None

	tokenStr, err := Sign(claims, privateKey, &opt1)
	if err != nil {
		t.Fatal(err)
	}

	_, result := Verify(tokenStr, publicKey, &opt2)

	if result != expect {
		t.Errorf("Result should have been %v, but it was %v", expect, result)
	}
}
