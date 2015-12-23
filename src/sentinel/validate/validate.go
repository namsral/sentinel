// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validate

import (
	"fmt"
	"net"
	"regexp"
)

var (
	// Validation rules
	RuleFloat        = regexp.MustCompile(`^[0-9]+\.[0-9]+$`)
	RuleInteger      = regexp.MustCompile(`^[0-9]+$`)
	RuleAlphanumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	RuleAlphabetic   = regexp.MustCompile(`^[a-zA-Z]+$`)
	RuleURL          = regexp.MustCompile(`^[a-zA-Z0-9]+:\/\/.+`)
	RuleUUID         = regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)
	RuleUUIDv4       = regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89aAbB][a-f0-9]{3}-[a-f0-9]{12}$`) //strict for v4 UUIDs
	RuleEmail        = regexp.MustCompile(`^[^@\s]+@[^@\s]+$`)
	RulePassword     = regexp.MustCompile(`.{8,}$`)
)

type Error struct {
	Err string
}

func (e *Error) Error() string {
	return fmt.Sprintf("validate: %s", e.Err)
}

// Return error if the provided input is empty
func NotEmpty(input string) error {
	if input == "" {
		return ErrEmpty
	}
	return nil
}

// Return error if the provided input is not empty
func Empty(input string) error {
	if input != "" {
		return ErrNotEmpty
	}
	return nil
}

// Float return an error when the given string is not a valid float.
func Float(input string) error {
	if RuleFloat.MatchString(input) == false {
		return ErrNotFloat
	}
	return nil
}

// Return error if the provided input is not a valid Integer
func Integer(input string) error {
	if RuleInteger.MatchString(input) == false {
		return ErrNotInteger
	}
	return nil
}

// Return error if the provided input is not a valid Alphanumeric
func Alphanumeric(input string) error {
	if RuleAlphanumeric.MatchString(input) == false {
		return ErrNotAlphanumeric
	}
	return nil
}

// Return error if the provided input is not a valid Alphabetic
func Alphabetic(input string) error {
	if RuleAlphabetic.MatchString(input) == false {
		return ErrNotAlphabetic
	}
	return nil
}

// Return error if the provided input is not a valid URL
func URL(input string) error {
	if RuleURL.MatchString(input) == false {
		return ErrNotURL
	}
	return nil
}

// Return error if the provided input is not a valid IPv4 or IPv6
func IP(input string) error {
	ip := net.ParseIP(input)
	if ip == nil {
		return ErrNotIP
	}
	return nil
}

// Return error if the provided input is not a valid UUID
func UUID(input string) error {
	if RuleUUID.MatchString(input) == false {
		return ErrNotUUID
	}
	return nil
}

// Return error if the provided input is not a valid UUIDv4
func UUIDv4(input string) error {
	if RuleUUIDv4.MatchString(input) == false {
		return ErrNotUUIDv4
	}
	return nil
}

// Return error if the provided input is not a valid email address.
func Email(input string) error {
	if RuleEmail.MatchString(input) == false {
		return ErrNotEmail
	}
	return nil
}

// Return error if the provided input is not a valid email address.
func Password(input string) error {
	if RulePassword.MatchString(input) == false {
		return ErrNotPassword
	}
	return nil
}

// Return error if the provided input is not validated against
// the chain of validation rules.
//
// Example:
//     err := Chain(input, validate.NotEmpty, validate.Alphabetic)
func Chain(input string, links ...func(string) error) error {
	var err error
	for _, link := range links {
		err = link(input)
		if err != nil {
			return err
		}
	}
	return nil
}
