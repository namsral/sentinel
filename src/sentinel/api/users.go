// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"database/sql"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"sentinel"
	"sentinel/datastore"
	"sentinel/router"
	"sentinel/tokens"
	"sentinel/validate"

	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/mux"
	"github.com/keighl/mandrill"
	"github.com/lib/pq"
)

const (
	AuthenticationScheme = "Bearer"
	AuthenticationRealm  = "https://sentinel.sh"
)

var (
	privateKey string
	publicKey  string

	mc *mandrill.Client
)

func init() {
	privateKey = os.Getenv("PRIVATE_KEY")
	publicKey = os.Getenv("PUBLIC_KEY")

	key := os.Getenv("MANDRILL_KEY")
	if key == "" {
		key = "SANDBOX_ERROR"
	}
	mc = mandrill.ClientWithKey(key)
}

func Authorized(r *http.Request) (*sentinel.User, error) {
	prefix := AuthenticationScheme + " "

	auth := r.Header.Get("Authorization")
	if auth == "" {
		return nil, ErrNoAuthentionMethodIncluded
	}
	if !strings.HasPrefix(auth, prefix) {
		return nil, ErrUnsupportedAuthenticationMethod
	}

	tokenStr := strings.TrimPrefix(auth, prefix)
	claims, err := tokens.Verify(tokenStr, publicKey, &tokens.AccessTokenOptions)
	if err != nil {
		return nil, ErrInvalidAuthenticationToken
	}
	userIDStr := claims["user_id"].(string)
	if err := validate.UUIDv4(userIDStr); err != nil {
		return nil, ErrInvalidAuthenticationToken.Append("value of claim 'user_id' was invalid")
	}
	userID := uuid.Parse(userIDStr)

	user, err := store.Users.GetUserDetails(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidClient
		}
		return nil, err
	}

	return user, nil
}

func serveGetUserDetails(w http.ResponseWriter, r *http.Request) error {
	user, err := Authorized(r)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, user)
}

func serveSignup(w http.ResponseWriter, r *http.Request) error {
	var email, password string

	expectMediatype := "application/x-www-form-urlencoded"
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != expectMediatype {
		return ErrUnsupportedMediatype.Append("expected " + expectMediatype)
	}

	if err := r.ParseForm(); err != nil {
		return err
	}

	// Get form input: email, password
	email = strings.TrimSpace(r.PostForm.Get("email"))
	password = r.PostForm.Get("password")

	// Validate form input
	if err := validate.Email(email); err != nil {
		return ErrInvalidEmail
	}
	if err := validate.Password(password); err != nil {
		return ErrInvalidPassword
	}
	user, err := store.Users.Signup(email, password)
	if err, ok := err.(*pq.Error); ok {
		// duplicate key value, pg code 23505
		if err.Code == "23505" {
			return ErrEmailRegistered
		}
	}
	if err != nil {
		return err
	}

	// Response
	u, err := apiRouter.Get(router.GetUserDetails).URL()
	if err != nil {
		return err
	}
	w.Header().Set("Location", u.String())
	if v := r.Header.Get("Prefer"); v == "return=representation" {
		if err := writeJSON(w, http.StatusCreated, user); err != nil {
			return err
		}
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	// Create token
	claims := tokens.Claims{
		"email_id": user.AuthEmailList[0].UID.String(),
		"user_id":  user.UID.String(),
	}
	tokenStr, err := tokens.Sign(claims, privateKey, &tokens.VerifyEmailOptions)
	if err != nil {
		log.Println("signing verify-email token failed due error:", err)
		return nil
	}

	// Send email verification
	msg := NewVerifyEmailMessage(tokenStr)
	msg.AddRecipient(email, "", "to")
	a, err := mc.MessagesSend(msg)
	if err != nil {
		log.Println("calling Mandrill failed with error:", err)
		return nil
	}
	for _, e := range a {
		log.Println("sent email-verification message with Mandrill id:", e.Id)
	}

	return nil
}

func serveCreateToken(w http.ResponseWriter, r *http.Request) error {
	prefix := "Basic "

	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ErrNoAuthentionMethodIncluded
	}
	if !strings.HasPrefix(auth, prefix) {
		return ErrUnsupportedAuthenticationMethod
	}

	email, password, ok := r.BasicAuth()
	if !ok {
		return ErrInvalidClient
	}

	email = strings.TrimSpace(email)

	// Validate email and password
	if err := validate.Email(email); err != nil {
		return ErrInvalidAuthenticationCredentials
	}
	if err := validate.Password(password); err != nil {
		return ErrInvalidAuthenticationCredentials
	}

	users, err := store.Users.List(sentinel.UserListOptions{Email: []string{email}})
	if err != nil {
		return err
	}
	if len(users) != 1 {
		return ErrUnknownClient
	}
	user := users[0]
	if err := datastore.ComparePassword(user, password); err != nil {
		return ErrInvalidAuthenticationCredentials
	}

	// Generate Token
	// Set token options
	opt := tokens.AccessTokenOptions
	expectMediatype := "application/x-www-form-urlencoded"
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err == nil || mt == expectMediatype {
		if err := r.ParseForm(); err != nil {
			return err
		}
		if s := r.PostForm.Get("client_id"); len(s) > 255 {
			return ErrInvalidRequest.Append("client_id exceeds max of 255 characters")
		} else {
			opt.Audience = s
		}
	}
	// Set token claims
	claims := tokens.Claims{
		"user_id": user.UID.String(),
	}
	// Sign token
	tokenStr, err := tokens.Sign(claims, privateKey, &opt)
	if err != nil {
		return err
	}

	// Response
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Pragma", "no-cache")
	data := map[string]interface{}{
		"token_type": "Bearer",
		"expires_in": (opt.TTL / time.Second).Nanoseconds(),
		"id_token":   tokenStr,
	}
	if err := writeJSON(w, http.StatusOK, data); err != nil {
		return err
	}
	return nil
}

func serveAckEmail(w http.ResponseWriter, r *http.Request) error {
	expectMediatype := "application/x-www-form-urlencoded"
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != expectMediatype {
		return ErrUnsupportedMediatype.Append("expected " + expectMediatype)
	}

	if err := r.ParseForm(); err != nil {
		return err
	}

	tokenStr := r.PostForm.Get("token")

	claims, err := tokens.Verify(tokenStr, publicKey, &tokens.VerifyEmailOptions)
	if err != nil {
		return ErrInvalidToken
	}

	emailIDStr := claims["email_id"].(string)
	userIDStr := claims["user_id"].(string)

	// Validate token data
	if err := validate.UUIDv4(emailIDStr); err != nil {
		return ErrInvalidToken.Append("value of claim 'email_id' was invalid")
	}
	if err := validate.UUIDv4(userIDStr); err != nil {
		return ErrInvalidToken.Append("value of claim 'user_id' was invalid")
	}

	// TODO: verify if the email address is associated with the user
	emailID := uuid.Parse(emailIDStr)

	// Set authemail to verified
	if err := store.Users.AckEmail(emailID); err != nil {
		if err == sql.ErrNoRows {
			return ErrInvalidRequest.Append("email already verified")
		}
		return err
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func serveGetEmail(w http.ResponseWriter, r *http.Request) error {
	user, err := Authorized(r)
	if err != nil {
		return err
	}

	s := mux.Vars(r)["uid"]
	if err := validate.UUIDv4(s); err != nil {
		return ErrNotFound
	}
	emailID := uuid.Parse(s)
	email, err := store.Users.GetEmail(emailID)
	if err != nil {
		return err
	}

	if email.UserID != user.ID {
		return ErrUnauthorizedClient
	}

	writeJSON(w, http.StatusOK, email)
	return nil
}

func serveAddEmail(w http.ResponseWriter, r *http.Request) error {
	user, err := Authorized(r)
	if err != nil {
		return err
	}

	expectMediatype := "application/x-www-form-urlencoded"
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != expectMediatype {
		return ErrUnsupportedMediatype.Append("expected " + expectMediatype)
	}

	if err := r.ParseForm(); err != nil {
		return err
	}

	// Get form input: email
	email := strings.TrimSpace(r.PostForm.Get("email"))

	// Validate form input
	if err := validate.Email(email); err != nil {
		return ErrInvalidEmail
	}

	authEmail, err := store.Users.AddEmail(user.UID, email)
	if err != nil {
		return err
	}

	// Response
	u, err := apiRouter.Get(router.GetEmail).URL("uid", authEmail.UID.String())
	if err != nil {
		return err
	}
	w.Header().Set("Location", u.String())
	if v := r.Header.Get("Prefer"); v == "return=representation" {
		if err := writeJSON(w, http.StatusCreated, authEmail); err != nil {
			return err
		}
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	// Create token
	claims := tokens.Claims{
		"email_id": authEmail.UID.String(),
		"user_id":  user.UID.String(),
	}
	tokenStr, err := tokens.Sign(claims, privateKey, &tokens.VerifyEmailOptions)
	if err != nil {
		log.Println("failed to sign verify-email token due to error:", err)
		return nil
	}

	// Send email verification
	msg := NewVerifyEmailMessage(tokenStr)
	msg.AddRecipient(email, "", "to")
	a, err := mc.MessagesSend(msg)
	if err != nil {
		log.Println("calling Mandrill failed with error:", err)
		return nil
	}
	for _, e := range a {
		log.Println("sent email-verification message with Mandrill id:", e.Id)
	}

	return nil
}

func serveDelEmail(w http.ResponseWriter, r *http.Request) error {
	user, err := Authorized(r)
	if err != nil {
		return err
	}

	s := mux.Vars(r)["uid"]
	if err := validate.UUIDv4(s); err != nil {
		return ErrNotFound
	}

	opt := sentinel.AuthEmailListOptions{
		User: &user.UID,
	}
	emails, err := store.Users.ListEmail(&opt)
	emailID := uuid.Parse(s)
	var email *sentinel.AuthEmail
	for _, e := range emails {
		if uuid.Equal(e.UID, emailID) {
			email = e
			break
		}
	}
	if email == nil {
		return ErrNotFound
	}
	if email != nil && len(emails) == 1 {
		return ErrConfilt.Append("cannot delete the only email address associated with the user.")
	}
	if err := store.Users.DelEmail(emailID); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func serveListEmail(w http.ResponseWriter, r *http.Request) error {
	user, err := Authorized(r)
	if err != nil {
		return err
	}

	cr := NewContentRange("items", DefaultContentRangeLast)
	if first, last, err := parseRange(r, "items"); err == nil {
		cr.First = first
		cr.Last = last
	}
	opt := sentinel.AuthEmailListOptions{
		User: &user.UID,
		ListOptions: &sentinel.ListOptions{
			First: cr.First,
			Last:  cr.Last,
		},
	}

	emails, err := store.Users.ListEmail(&opt)
	if err != nil {
		return err
	}
	cr.UpdateRange(len(emails))

	cr.SetContentRange(w)
	writeJSON(w, http.StatusPartialContent, emails)
	return nil
}

func servePublicKey(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(publicKey))
	return nil
}

func serveUpdateUserDetails(w http.ResponseWriter, r *http.Request) error {
	user, err := Authorized(r)
	if err != nil {
		return err
	}

	expectMediatype := "application/x-www-form-urlencoded"
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != expectMediatype {
		return ErrUnsupportedMediatype.Append("expected " + expectMediatype)
	}

	if err := r.ParseForm(); err != nil {
		return err
	}

	opt := sentinel.UserUpdateOptions{}
	if err := opt.ParseForm(r.PostForm); err != nil {
		return ErrInvalidRequest.Append(`; ` + err.Error())
	}

	user, err = store.Users.UpdateDetails(user.UID, opt)
	if err != nil {
		return err
	}

	if err := writeJSON(w, http.StatusOK, user); err != nil {
		return err
	}

	return nil
}

func serveOneTimeLogin(w http.ResponseWriter, r *http.Request) error {
	expectMediatype := "application/x-www-form-urlencoded"
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != expectMediatype {
		return ErrUnsupportedMediatype.Append("expected " + expectMediatype)
	}

	if err := r.ParseForm(); err != nil {
		return err
	}

	// Get form input: email
	email := r.PostForm.Get("email")

	// Validate form input
	if err := validate.Email(email); err != nil {
		return ErrInvalidEmail
	}

	// Check if email is registered
	opt := sentinel.UserListOptions{Email: []string{email}}
	users, err := store.Users.List(opt)
	if err != nil {
		return err
	}
	if len(users) != 1 {
		return ErrUnknownClient
	}
	// TODO: decide if email needs to be verified to continue

	// Generate AccessToken
	claims := tokens.Claims{
		"user_id": users[0].UID.String(),
	}
	tokenStr, err := tokens.Sign(claims, privateKey, &tokens.AccessTokenOptions)
	if err != nil {
		return err
	}

	// Send login link
	msg := NewEmailLoginLinkMessage(tokenStr)
	msg.AddRecipient(email, "", "to")
	a, err := mc.MessagesSend(msg)
	if err != nil {
		log.Println("calling Mandrill failed with error:", err)
	}
	for _, e := range a {
		log.Println("sent email-verification message with Mandrill id:", e.Id)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
