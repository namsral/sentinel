// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tokens

import (
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/zhevron/jwt"
)

var (
	DefaultOptions     = Options{Algorithm: jwt.RS256, Issuer: "https://sentinel.sh", TTL: time.Hour * 72}
	VerifyEmailOptions = Options{Algorithm: jwt.RS256, Issuer: "https://sentinel.sh/verify-email", TTL: time.Hour * 72}
	AccessTokenOptions = Options{Algorithm: jwt.RS256, Issuer: "https://sentinel.sh/access-token", TTL: time.Minute * 60}
)

type Options struct {
	Algorithm jwt.Algorithm
	Issuer    string
	Audience  string
	TTL       time.Duration
}

type Claims map[string]interface{}

func Sign(c Claims, privateKey string, opt *Options) (string, error) {
	if opt == nil {
		opt = &DefaultOptions
	}
	token := jwt.NewToken()
	token.Algorithm = opt.Algorithm
	token.Expires = token.IssuedAt.Add(opt.TTL)
	token.Issuer = opt.Issuer
	token.Claims = c
	if opt.Audience != "" {
		token.Audience = opt.Audience
		token.Subject = uuid.NewRandom().String()
	}
	return token.Sign(privateKey)
}

func Verify(token, publicKey string, opt *Options) (Claims, error) {
	if opt == nil {
		opt = &DefaultOptions
	}
	t, err := jwt.DecodeToken(token, opt.Algorithm, publicKey)
	if err != nil {
		return nil, err
	}
	if t.Algorithm != opt.Algorithm {
		return nil, jwt.ErrUnsupportedAlgorithm
	}
	if err := t.Verify(opt.Issuer, "", ""); err != nil {
		return nil, err
	}
	return t.Claims, nil
}
