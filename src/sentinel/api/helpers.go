// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const DefaultContentRangeLast uint64 = 20

func writeJSON(w http.ResponseWriter, statusCode int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// setContentRange sets the Content-Range header using the given first, last and
// length integers.
// Example:
//     Content-Range: 0-19/20
func setContentRange(w http.ResponseWriter, first, last, length uint64) error {
	var value string
	if first > last || (last > length && length != 0) {
		return errors.New("invalid range")
	}
	switch {
	case length == 0:
		value = fmt.Sprintf("%d-%d/*", first, last)
	default:
		value = fmt.Sprintf("%d-%d/%d", first, last, length)
	}
	w.Header().Set("Content-Range", value)
	return nil
}

// parseRange parses the give Range header value and return the first and last
// as integers. Invalid ranges will return as 0, 0
// func parseRange(value string) (first, last uint64) {
// 	value = strings.TrimPrefix(value, "items=")
// 	a := strings.Split(value, "-")
// 	switch len(a) {
// 	case 1:
// 		if first, err = strconv.ParseUint(a[0], 10, 0); err != nil {
// 			break
// 		}
// 	case 2:
// 		if first, err = strconv.ParseUint(a[0], 10, 0); err != nil {
// 			break
// 		}
// 		if last, err = strconv.ParseUint(a[1], 10, 0); err != nil {
// 			break
// 		}
// 	}
// 	return first, last
// }

func NewContentRange(unit string, last uint64) *ContentRange {
	return &ContentRange{
		Unit: unit,
		Last: last,
	}
}

type ContentRange struct {
	Unit   string
	First  uint64
	Last   uint64
	Length uint64
}

func (r *ContentRange) SetContentRange(w http.ResponseWriter) error {
	var value string
	if err := r.Valid(); err != nil {
		return err
	}
	switch {
	case r.Length == 0:
		value = fmt.Sprintf("%d-%d/*", r.First, r.Last)
	default:
		value = fmt.Sprintf("%d-%d/%d", r.First, r.Last, r.Length)
	}
	w.Header().Set("Range-Units", r.Unit)
	w.Header().Set("Content-Range", value)
	return nil
}

// UpdateRange updates the range with the number of given units.
func (r *ContentRange) UpdateRange(n int) {
	if n > 0 {
		r.Last = (r.First + uint64(n-1))
	} else {
		r.Last = r.First
	}
}

func (r *ContentRange) Valid() error {
	if r.First > r.Last || (r.Last > r.Length && r.Length != 0) {
		return errors.New("invalid range")
	}
	return nil
}

// parseRange parses the Range header's value that matches the given unit and
// returns the first and last range when the range is valid.
func parseRange(req *http.Request, unit string) (first, last uint64, err error) {
	value := req.Header.Get("Range")
	if value == "" || req.Header.Get("Range-Unit") != unit {
		return first, last, errors.New("no valid Range headers found")
	}
	a := strings.Split(strings.TrimPrefix(value, unit+"="), "-")
	switch len(a) {
	case 1:
		if i, err := strconv.ParseUint(a[0], 10, 0); err == nil {
			first = i
		}
	case 2:
		if i, err := strconv.ParseUint(a[0], 10, 0); err == nil {
			first = i
		}
		if i, err := strconv.ParseUint(a[1], 10, 0); err == nil {
			last = i
		}
	}
	if last != 0 && first > last {
		return first, last, errors.New("invalid range")
	}
	return first, last, nil
}
