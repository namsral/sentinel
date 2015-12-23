// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sentinel

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"sentinel/router"
	"sentinel/validate"

	"code.google.com/p/go-uuid/uuid"
)

const (
	AuthLevelUnknown int = iota
	AuthLevelNotify
	AuthLevelFast
	AuthLevelSecure
)

var (
	ErrUserNotFound = errors.New("user not found")
)

// User is a reflection of the enduser's profile.
type User struct {
	ID               int          `json:"-"`
	UID              uuid.UUID    `db:"uid" json:"id"`
	Name             string       `json:"name"`
	PasswordHash     string       `db:"password_hash" json:"-"`
	LastLoginAt      time.Time    `db:"lastlogin_at" json:"lastLogin"`
	DefaultAuthLevel int          `json:"defaultAuthLevel"`
	CreatedAt        time.Time    `db:"created_at" json:"-"`
	UpdatedAt        time.Time    `db:"updated_at" json:"-"`
	IsArchived       bool         `db:"is_archived" json:"-"`
	AuthEmailList    []*AuthEmail `json:"authEmailList"`
	DeviceToken      string       `json:"deviceToken"`
}

type UserUpdateOptions struct {
	Name, Password, DeviceToken string
	DefaultAuthLevel            int
}

func (o *UserUpdateOptions) ParseForm(v url.Values) error {
	var parsed bool
	if s := v.Get("name"); s != "" {
		if len(s) != 0 && len(s) > 256 {
			return errors.New("invalid name parameter; exceeds maximum of 256 characters")
		}
		o.Name = s
		parsed = true
	}

	if s := v.Get("password"); s != "" {
		if err := validate.Password(s); err != nil {
			return errors.New("invalid password parameter; does not match regexp " + validate.RulePassword.String())
		}
		o.Password = s
		parsed = true
	}
	if s := v.Get("defaultAuthLevel"); s != "" {
		switch s {
		case "1":
			o.DefaultAuthLevel = AuthLevelNotify
			parsed = true
		case "2":
			o.DefaultAuthLevel = AuthLevelFast
			parsed = true
		case "3":
			o.DefaultAuthLevel = AuthLevelSecure
			parsed = true
		default:
			return errors.New("invalid defaultAuthLevel parameter; options are 1:Notify, 2:Fast or 3:Secure")
		}
	}
	if s := v.Get("deviceToken"); s != "" {
		o.DeviceToken = s
		parsed = true
	}
	if !parsed {
		return errors.New("found no paramters to parse")
	}
	return nil
}

// AuthEmail is the authentication of a single user.
type AuthEmail struct {
	ID         int       `json:"-"`
	UID        uuid.UUID `db:"uid" json:"id"`
	UserID     int       `db:"user_id" json:"-"`
	Email      string    `json:"email"`
	IsVerified bool      `db:"is_verified" json:"isVerified"`
	CreatedAt  time.Time `db:"created_at" json:"-"`
	UpdatedAt  time.Time `db:"updated_at" json:"-"`
}

// AuthEmailListOptions is a filter instance.
type AuthEmailListOptions struct {
	User *uuid.UUID
	*ListOptions
}

// UsersService interacts with the user-related endpoint in Sentinel's API.
type UsersService interface {
	Signup(email, password string) (*User, error)
	GetUserDetails(uid uuid.UUID) (*User, error)
	UpdateDetails(userID uuid.UUID, opt UserUpdateOptions) (*User, error)
	List(UserListOptions) ([]*User, error)
	ListEmail(*AuthEmailListOptions) ([]*AuthEmail, error)
	AddEmail(userID uuid.UUID, email string) (*AuthEmail, error)
	AckEmail(uid uuid.UUID) error
	GetEmail(uid uuid.UUID) (*AuthEmail, error)
	DelEmail(uid uuid.UUID) error
}

type usersService struct {
	client *Client
}

func (s *usersService) GetUserDetails(uid uuid.UUID) (*User, error) {
	url, err := s.client.url(router.GetUserDetails, nil, nil)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	if err := s.client.Authorize(req); err != nil {
		return nil, err
	}

	var user *User
	if _, err = s.client.Do(req, &user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *usersService) Signup(email, password string) (*User, error) {
	u, err := s.client.url(router.Signup, nil, nil)
	if err != nil {
		return nil, err
	}

	form := &url.Values{
		"email":    {email},
		"password": {password},
	}
	req, err := s.client.NewRequest("POST", u.String(), form)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Prefer", "return=representation")

	var user User
	resp, err := s.client.Do(req, &user)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New("API reponded with status " + http.StatusText(resp.StatusCode))
	}

	return &user, nil
}

func (s *usersService) List(opt UserListOptions) ([]*User, error) {
	//TODO:implement
	return nil, nil
}

func (s *usersService) AckEmail(uid uuid.UUID) error {
	u, err := s.client.url(router.AckEmail, nil, nil)
	if err != nil {
		return err
	}

	form := &url.Values{
		"token": {s.client.token},
	}
	req, err := s.client.NewRequest("POST", u.String(), form)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New("API reponded with status " + http.StatusText(resp.StatusCode))
	}

	return nil
}

func (s *usersService) AddEmail(userID uuid.UUID, email string) (*AuthEmail, error) {
	u, err := s.client.url(router.AddEmail, nil, nil)
	if err != nil {
		return nil, err
	}

	form := &url.Values{
		"email": {email},
	}
	req, err := s.client.NewRequest("POST", u.String(), form)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Prefer", "return=representation")

	var e AuthEmail
	resp, err := s.client.Do(req, &e)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New("API reponded with status " + http.StatusText(resp.StatusCode))
	}

	return &e, nil
}

func (s *usersService) GetEmail(uid uuid.UUID) (*AuthEmail, error) {
	//TODO: implement
	return nil, nil
}

func (s *usersService) ListEmail(*AuthEmailListOptions) ([]*AuthEmail, error) {
	//TODO: implement
	return nil, nil
}

func (s *usersService) DelEmail(userID uuid.UUID) error {
	u, err := s.client.url(router.DelEmail, nil, nil)
	if err != nil {
		return err
	}

	req, err := s.client.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New("API reponded with status " + http.StatusText(resp.StatusCode))
	}

	return nil
}

func (s *usersService) UpdateDetails(userID uuid.UUID, opt UserUpdateOptions) (*User, error) {
	//TODO: implement
	return nil, nil
}

// UserListOptions is an instance to filter users from a list of users.
type UserListOptions struct {
	// IncludeArchived will include archived users (inactive accounts)
	IncludeArchived bool

	// Email belonging to users
	Email []string

	*ListOptions
}

// MockUsersService is a mock of the UsersService.
type MockUsersService struct {
	SignupFn         func(email, password string) (*User, error)
	GetUserDetailsFn func(uid uuid.UUID) (*User, error)
	ListFn           func(UserListOptions) ([]*User, error)
	AddEmailFn       func(userID uuid.UUID, email string) (*AuthEmail, error)
	AckEmailFn       func(uid uuid.UUID) error
	GetEmailFn       func(uid uuid.UUID) (*AuthEmail, error)
	DelEmailFn       func(id uuid.UUID) error
	ListEmailFn      func(opt *AuthEmailListOptions) ([]*AuthEmail, error)
	UpdateDetailsFn  func(userID uuid.UUID, opt UserUpdateOptions) (*User, error)
}

var _ UsersService = &MockUsersService{}

// GetUserDetails returns a User instance with the given uuid.
func (s *MockUsersService) GetUserDetails(uid uuid.UUID) (*User, error) {
	if s.GetUserDetailsFn == nil {
		return nil, nil
	}
	return s.GetUserDetailsFn(uid)
}

// Signup returns a new User instance with the given email and password.
func (s *MockUsersService) Signup(email, password string) (*User, error) {
	if s.SignupFn == nil {
		return nil, nil
	}
	return s.SignupFn(email, password)
}

// List returns a filter list of User instances.
func (s *MockUsersService) List(opt UserListOptions) ([]*User, error) {
	if s.ListFn == nil {
		return nil, nil
	}
	return s.ListFn(opt)
}

// AckEmail sets the email with the given UUID as acknologed.
func (s *MockUsersService) AckEmail(uid uuid.UUID) error {
	if s.AckEmailFn == nil {
		return nil
	}
	return s.AckEmailFn(uid)
}

// AddEmail creates a new email for the user with the given UUID.
func (s *MockUsersService) AddEmail(userID uuid.UUID, email string) (*AuthEmail, error) {
	if s.AddEmailFn == nil {
		return nil, nil
	}
	return s.AddEmailFn(userID, email)
}

func (s *MockUsersService) GetEmail(emailID uuid.UUID) (*AuthEmail, error) {
	if s.GetEmailFn == nil {
		return nil, nil
	}
	return s.GetEmailFn(emailID)
}

func (s *MockUsersService) DelEmail(id uuid.UUID) error {
	if s.DelEmailFn == nil {
		return nil
	}
	return s.DelEmailFn(id)
}

func (s *MockUsersService) ListEmail(opt *AuthEmailListOptions) ([]*AuthEmail, error) {
	if s.ListEmailFn == nil {
		return nil, nil
	}
	return s.ListEmailFn(opt)
}

func (s *MockUsersService) UpdateDetails(userID uuid.UUID, opt UserUpdateOptions) (*User, error) {
	if s.UpdateDetailsFn == nil {
		return nil, nil
	}
	return s.UpdateDetailsFn(userID, opt)
}
