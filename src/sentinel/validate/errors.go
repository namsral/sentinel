// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validate

var (
	// Validation errors
	ErrNotEmpty        = &Error{`expecting an empty value`}
	ErrEmpty           = &Error{`expecting a non empty value`}
	ErrNotFloat        = &Error{`expecting a floating point number`}
	ErrNotInteger      = &Error{`expecting an integer`}
	ErrNotAlphanumeric = &Error{`expecting an alphanumeric string`}
	ErrNotAlphabetic   = &Error{`expecting an alphabetic string`}
	ErrNotURL          = &Error{`expecting an URL`}
	ErrNotIP           = &Error{`expecting an IPv4 or IPv6`}
	ErrNotUUID         = &Error{`expecting an UUID`}
	ErrNotUUIDv4       = &Error{`expecting an UUIDv4`}
	ErrNotEmail        = &Error{`expecting an email addres`}
	ErrNotPassword     = &Error{`expecting a password`}
)
