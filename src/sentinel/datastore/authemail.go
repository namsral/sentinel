// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package datastore

const authemailTable = "authemails"
const authemailTableCreateStmt = `
CREATE TABLE authemails (
    id SERIAL PRIMARY KEY, -- internal identifier
    uid uuid UNIQUE not null, -- uuid identifier
    user_id integer NOT NULL references users ON UPDATE CASCADE,
    email TEXT NOT NULL UNIQUE,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP(0),
    updated_at TIMESTAMP(0)
);
`
const authemailInsertStmt = `
INSERT INTO authemails (uid, user_id, email, is_verified, created_at, updated_at)
VALUES (:uid, :user_id, :email, :is_verified, :created_at, :updated_at)
;`

const authemailListStmt = `SELECT * FROM authemails`
const authemailGetStmt = `SELECT * FROM authemails WHERE id=$1;`

const authemailCreateStmt = `
INSERT INTO authemails (uid, user_id, email, is_verified, created_at, updated_at) (
    SELECT $1, id, $2, $3, $4, $5 FROM users WHERE uid=$6
) RETURNING id;`
