// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package datastore

import (
	"errors"
	"strings"
	"time"

	"sentinel"

	"code.google.com/p/go-uuid/uuid"
	sq "github.com/lann/squirrel"
	"golang.org/x/crypto/bcrypt"
)

const userTable = "users"
const userTableCreateStmt = `
CREATE TABLE users (
	id SERIAL PRIMARY KEY, -- internal identifier
	uid uuid UNIQUE not null, -- uuid identifier
	name TEXT NOT NULL,
	password_hash TEXT NOT NULL DEFAULT 'plain:secret', -- format: <hash type>:<password hash>
	devicetoken TEXT NOT NULL DEFAULT '', -- device token for push services like APN and GCM
	lastlogin_at TIMESTAMP(0),
	defaultauthlevel INTEGER NOT NULL DEFAULT 0, -- 0:unknown 1:notify 2:fast 3:secure
	created_at TIMESTAMP(0),
	updated_at TIMESTAMP(0),
	is_archived BOOLEAN NOT NULL DEFAULT FALSE
);`

const userInsertStmt = `
INSERT INTO users(uid, name, password_hash, devicetoken, lastlogin_at, defaultauthlevel,
	created_at, updated_at, is_archived)
VALUES (:uid, :name, :password_hash, :devicetoken, :lastlogin_at,
	:defaultauthlevel, :created_at, :updated_at, :is_archived)
RETURNING id
;`

const userUpdateStmt = `
UPDATE users SET
	(name, password_hash, devicetoken, lastlogin_at, defaultauthlevel, is_archived) =
	(:name, :password_hash, :devicetoken, :lastlogin_at, :defaultauthlevel, :is_archived)
WHERE uid=:uid
;`

const userListStmt = `SELECT * FROM users WHERE is_archived=FALSE;`
const userGetStmt = `SELECT * FROM users WHERE is_archived=FALSE AND id=$1;`

var (
	psq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
)

type usersStore struct {
	*Datastore
}

func (s *usersStore) Signup(email, password string) (*sentinel.User, error) {
	var user *sentinel.User

	// Create user
	now := time.Now().UTC()
	user = &sentinel.User{
		UID:       uuid.NewRandom(),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := SetPassword(user, password); err != nil {
		return nil, err
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}

	var id int
	stmt, err := tx.PrepareNamed(userInsertStmt)
	if err != nil {
		return nil, err
	}
	err = stmt.QueryRowx(user).Scan(&id)
	if err != nil {
		return nil, err
	}

	authEmail := &sentinel.AuthEmail{
		UID:       uuid.NewRandom(),
		UserID:    id,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := tx.NamedExec(authemailInsertStmt, authEmail); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	user.AuthEmailList = []*sentinel.AuthEmail{authEmail}

	return user, nil
}

func (s *usersStore) GetUserDetails(uid uuid.UUID) (*sentinel.User, error) {
	var user sentinel.User

	err := s.db.QueryRowx(`SELECT * FROM users WHERE is_archived=FALSE AND uid=$1;`, uid).StructScan(&user)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Queryx(`SELECT * FROM authemails WHERE user_id=$1`, user.ID)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	var a []*sentinel.AuthEmail
	for rows.Next() {
		var email sentinel.AuthEmail
		if err := rows.StructScan(&email); err != nil {
			return nil, err
		}
		a = append(a, &email)
	}
	if len(a) > 0 {
		user.AuthEmailList = a
	}

	return &user, nil
}

func (s *usersStore) get(id int) (*sentinel.User, error) {
	var user sentinel.User

	err := s.db.QueryRowx(userGetStmt, id).StructScan(&user)
	if err != nil {
		return nil, err
	}

	var a []*sentinel.AuthEmail
	rows, err := s.db.Queryx(`SELECT * FROM authemails WHERE user_id=$1`, user.ID)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var e sentinel.AuthEmail
		if err := rows.StructScan(&e); err != nil {
			return nil, err
		}
		a = append(a, &e)
	}
	if len(a) > 0 {
		user.AuthEmailList = a
	}

	return &user, nil
}

func (s *usersStore) Submit(user *sentinel.User) (uuid.UUID, error) {
	var err error

	if user.UID == nil {
		user.UID = uuid.NewRandom()
	}

	now := time.Now().UTC()
	if user.ID == 0 {
		user.CreatedAt = now
	}
	user.UpdatedAt = now

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}

	var userID int
	stmt, err := tx.PrepareNamed(userInsertStmt)
	if err != nil {
		return nil, err
	}
	err = stmt.QueryRowx(user).Scan(&userID)
	if err != nil {
		return nil, err
	}

	for _, v := range user.AuthEmailList {
		v.UID = uuid.NewRandom()
		v.UserID = userID
		v.CreatedAt = now
		v.UpdatedAt = now
		_, err := tx.NamedExec(authemailInsertStmt, v)
		if err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return user.UID, nil
}

func (s *usersStore) UpdateDetails(uid uuid.UUID, opt sentinel.UserUpdateOptions) (*sentinel.User, error) {
	user, err := s.GetUserDetails(uid)
	if err != nil {
		return nil, err
	}

	if opt.Name != "" {
		user.Name = opt.Name
	}

	if opt.Password != "" {
		SetPassword(user, opt.Password)
	}

	if opt.DeviceToken != "" {
		user.DeviceToken = opt.DeviceToken
	}

	if opt.DefaultAuthLevel != 0 {
		user.DefaultAuthLevel = int(opt.DefaultAuthLevel)
	}

	if err := s.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *usersStore) Update(user *sentinel.User) error {
	user.UpdatedAt = time.Now().UTC()

	result, err := s.db.NamedExec(userUpdateStmt, user)
	if err != nil {
		return err
	}
	i, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if i < 1 {
		return errors.New("user update failed")
	}
	return nil
}

func (s *usersStore) List(opt sentinel.UserListOptions) ([]*sentinel.User, error) {
	var users []*sentinel.User

	sb := psq.Select("users.*").From("users").Join("authemails ON(users.id = authemails.user_id)")

	if len(opt.Email) > 0 {
		sb = sb.Where(sq.Eq{"authemails.email": opt.Email})
	}

	if opt.IncludeArchived {
		sb = sb.Where(sq.Eq{"users.is_archived": true})
	}

	sql, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := s.db.Queryx(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var user sentinel.User
		if err := rows.StructScan(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	for _, user := range users {
		var a []*sentinel.AuthEmail
		rows, err := s.db.Queryx(`SELECT * FROM authemails WHERE user_id=$1`, user.ID)
		defer rows.Close()
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var e sentinel.AuthEmail
			if err := rows.StructScan(&e); err != nil {
				return nil, err
			}
			a = append(a, &e)
		}
		if len(a) > 0 {
			user.AuthEmailList = a
		}
	}

	return users, nil
}

func (s *usersStore) AckEmail(uid uuid.UUID) error {
	var isVerified bool
	if err := s.db.QueryRowx(`
		UPDATE authemails SET is_verified=TRUE
		WHERE uid=$1
		AND is_verified=FALSE
		RETURNING is_verified`, uid).Scan(&isVerified); err != nil {
		return err
	}
	if !isVerified {
		return errors.New("verfied remains false")
	}
	return nil
}

func (s *usersStore) GetEmail(uid uuid.UUID) (*sentinel.AuthEmail, error) {
	var email sentinel.AuthEmail

	err := s.db.QueryRowx(`SELECT * FROM authemails WHERE uid=$1;`, uid).StructScan(&email)
	if err != nil {
		return nil, err
	}

	return &email, nil
}

func (s *usersStore) ListEmail(opt *sentinel.AuthEmailListOptions) ([]*sentinel.AuthEmail, error) {
	sb := psq.Select("authemails.*").From("authemails").Join("users ON(users.id = authemails.user_id)")
	if opt != nil {
		if opt.User != nil {
			sb = sb.Where(sq.Eq{"users.uid": opt.User})
		}
		if opt.ListOptions != nil {
			sb = sb.Limit(opt.ListOptions.Limit()).Offset(opt.ListOptions.Offset())
		}
	}

	sql, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}

	var emails []*sentinel.AuthEmail
	rows, err := s.db.Queryx(sql, args...)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var e sentinel.AuthEmail
		if err := rows.StructScan(&e); err != nil {
			return nil, err
		}
		emails = append(emails, &e)
	}
	return emails, err
}

func (s *usersStore) AddEmail(userID uuid.UUID, email string) (*sentinel.AuthEmail, error) {
	now := time.Now().UTC()
	e := &sentinel.AuthEmail{
		UID:       uuid.NewRandom(),
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	var id int
	err := s.db.QueryRowx(authemailCreateStmt,
		e.UID,
		e.Email,
		e.IsVerified,
		e.CreatedAt,
		e.UpdatedAt,
		userID).Scan(&id)
	if err != nil {
		return nil, err
	}
	e.ID = id

	return e, nil
}

func (s *usersStore) DelEmail(id uuid.UUID) error {
	result, err := s.db.Exec(`DELETE FROM authemails WHERE uid=$1`, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("email not found")
	}

	return nil
}

func SetPassword(u *sentinel.User, password string) error {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = "bcrypt:" + string(b)
	return nil
}

func ComparePassword(u *sentinel.User, password string) error {
	passwordHash := u.PasswordHash
	if strings.HasPrefix(passwordHash, "plain:") {
		p := strings.TrimPrefix(passwordHash, "plain:")
		if p == password {
			return nil
		}
	}
	if strings.HasPrefix(passwordHash, "bcrypt:") {
		p := strings.TrimPrefix(passwordHash, "bcrypt:")
		err := bcrypt.CompareHashAndPassword([]byte(p), []byte(password))
		if err == nil {
			return nil
		}
	}
	return errors.New("invalid password")
}
