// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sentinel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"sentinel/router"

	"github.com/google/go-querystring/query"
	"github.com/gorilla/mux"
)

const (
	version   = "0.0.1"
	userAgent = "sentinel-client/" + version
)

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	baseURL, _ := url.Parse("http://sentinel.sh/")
	c := &Client{
		BaseURL:    baseURL,
		apiRouter:  router.API(baseURL),
		UserAgent:  userAgent,
		httpClient: httpClient,
	}
	c.Users = &usersService{c}
	c.Services = &servicesService{c}
	return c
}

type Client struct {
	Users    UsersService
	Services ServicesService

	// BaseURL to Sentinel API
	BaseURL *url.URL

	// apiRouter is used to generate URLs for the Sentinel API
	apiRouter *mux.Router

	// User Agent string used for HTTP request to Sentinel's API
	UserAgent string

	httpClient *http.Client

	// Token used to authenticate HTTP requests to Sentinel's API
	token string
}

func (c *Client) url(apiRouterName string, routeVars map[string]string, opt interface{}) (*url.URL, error) {
	router := c.apiRouter.Get(apiRouterName)
	if router == nil {
		return nil, fmt.Errorf("no API router named %s", apiRouterName)
	}

	routeVarsList := make([]string, 2*len(routeVars))
	i := 0
	for k, v := range routeVars {
		routeVarsList[i*2] = k
		routeVarsList[i*2+1] = v
		i++
	}
	url, err := router.URL(routeVarsList...)
	if err != nil {
		return nil, err
	}

	url.Path = strings.TrimPrefix(url.Path, "/")

	if opt != nil {
		err = addOptions(url, opt)
		if err != nil {
			return nil, err
		}
	}
	return url, nil
}

func addOptions(u *url.URL, opt interface{}) error {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}

	qs, err := query.Values(opt)
	if err != nil {
		return err
	}
	u.RawQuery = qs.Encode()

	return nil
}

func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	buf := bytes.NewBuffer(nil)

	var contentType string
	switch body.(type) {
	case nil:
	case *url.Values:
		s := body.(*url.Values).Encode()
		if _, err := buf.WriteString(s); err != nil {
			return nil, err
		}
		contentType = "application/x-www-form-urlencoded; charset=utf-8"
	default:
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
		contentType = "application/json; charset=utf-8"
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", c.UserAgent)
	return req, nil
}

func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = CheckResponse(resp)
	if err != nil {
		return resp, err
	}

	if v != nil {
		if bp, ok := v.(*[]byte); ok {
			*bp, err = ioutil.ReadAll(resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("error reading response from %s %s: %s", req.Method, req.URL.RequestURI(), err)
	}

	return resp, nil
}

// Authorize sets the Authorization header for the given Request.
func (c *Client) Authorize(r *http.Request) error {
	if c.token == "" {
		return errors.New("no valid token to sign request")
	}
	r.Header.Set("Authorization", "Bearer "+c.token)
	return nil
}

func (c *Client) Authenticate(email, password string) error {
	token, err := c.createToken(email, password)
	if err != nil {
		return err
	}
	c.SetToken(token)

	return nil
}

func (c *Client) createToken(email, password string) (string, error) {
	u, err := c.url(router.CreateToken, nil, nil)
	if err != nil {
		return "", err
	}

	req, err := c.NewRequest("POST", u.String(), nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(email, password)

	data := make(map[string]interface{})
	resp, err := c.Do(req, &data)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("API reponded with status " + http.StatusText(resp.StatusCode))
	}

	token := data["id_token"].(string)

	return token, nil
}

func (c *Client) SetToken(token string) {
	c.token = token
}

// DefaultPerPage is the default number of results to return in a result set.
const DefaultLimit = 20

// ListOptions specifies general range options for fetching a list of
// results.
type ListOptions struct {
	First uint64
	Last  uint64
}

func (o ListOptions) Limit() uint64 {
	if o.First >= o.Last {
		return o.First + DefaultLimit
	}
	return (o.Last - o.First) + 1
}

func (o ListOptions) Offset() uint64 {
	return o.First
}
