// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validate

import (
	"testing"
)

func TestValidateNotEmpty(t *testing.T) {
	var err error

	err = NotEmpty(" ")
	if err != nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}

	err = NotEmpty("")
	if err == nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}
}

func TestValidateEmpty(t *testing.T) {
	var err error

	err = Empty("")
	if err != nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}

	err = Empty(" ")
	if err == nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}
}

func TestValidateFloat(t *testing.T) {
	var err error

	err = Float("0.1")
	if err != nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}

	err = Float(".1")
	if err == nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}
}

func TestValidateInteger(t *testing.T) {
	var err error

	err = Integer("42")
	if err != nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}

	err = Integer("0.1")
	if err == nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}
}

func TestValidateAlphanumeric(t *testing.T) {
	var err error

	err = Alphanumeric("A4")
	if err != nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}

	err = Alphanumeric("$")
	if err == nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}
}

func TestValidateAlphabetic(t *testing.T) {
	var err error

	err = Alphabetic("aB")
	if err != nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}

	err = Alphabetic("A4")
	if err == nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}
}

func TestValidateURL(t *testing.T) {
	var err error

	err = URL("http://example.com/path/?query=value#anchor")
	if err != nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}

	err = URL("http:/hello")
	if err == nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}
}

func TestValidateIP(t *testing.T) {
	var err error

	err = IP("39.21.31.4")
	if err != nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}

	err = IP("555.2.2.4")
	if err == nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}
}

func TestValidateUUIDv4(t *testing.T) {
	var err error

	err = UUIDv4("16fd2706-8baf-433b-82eb-8c7fada847da")
	if err != nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}

	err = UUIDv4("a8098c1a-f86e-11da-bd1a-00112444be1e")
	if err == nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}

	err = UUIDv4("886313e1-3b8a-5372-9b90-0c9aee199e5d")
	if err == nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}
}

func TestValidateEmail(t *testing.T) {
	var err error

	err = Email("a@a")
	if err != nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}

	err = Email(" @a")
	if err == nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}

	err = Email("a@@a")
	if err == nil {
		t.Fatalf("Failed validation: %s", err.Error())
	}
}
