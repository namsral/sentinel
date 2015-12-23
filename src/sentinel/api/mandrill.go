// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"fmt"

	"github.com/keighl/mandrill"
)

const (
	FromEmailSupport = "no-reply@sentinel"
	FromNameSupport  = "Sentinel Support"
)

//Subject: Confirm your email
var verifyEmailTmpl = `Hey, welcome to Sentinel! Before you get started, please verify your email address by visiting the following link:

    https://sentinel.sh/verify?token=%s

Sentinel Bot`

// Subject: Login link
var emailLoginLinkTmpl = `Hey, you requested that we send you a link to login to our application without a password. Here you go::

    https://sentinel.sh/login?token=%s

Using the one time login link can be convenient when you can't remember your password or when you don't want to enter your password on a public WiFi.

Sentinel Bot`

func NewVerifyEmailMessage(token string) *mandrill.Message {
	m := &mandrill.Message{
		FromEmail: FromEmailSupport,
		FromName:  FromNameSupport,
		Subject:   "Confirm your email",
	}
	m.Text = fmt.Sprintf(verifyEmailTmpl, token)
	return m
}

func NewEmailLoginLinkMessage(token string) *mandrill.Message {
	m := &mandrill.Message{
		FromEmail: FromEmailSupport,
		FromName:  FromNameSupport,
		Subject:   "Login link",
	}
	m.Text = fmt.Sprintf(verifyEmailTmpl, token)
	return m
}
