// Copyright (c) 2012 Brian Hetro <whee@smaertness.net>
// Use of this source code is governed by the ISC
// license that can be found in the LICENSE file.

package adn

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Application struct {
	Id          string
	Secret      string
	RedirectURI string
	Scopes      Scopes
}

var DefaultApplication = &Application{}
var apiHttpClient = &http.Client{}

type Request struct {
	Token    string    // Authentication token for the user or ""
	Body     io.Reader // Data for the body
	BodyType string    // Value for the Content-Type header
}

func (c *Application) request(r *Request, name string, args EpArgs) (body io.ReadCloser, err error) {
	var path bytes.Buffer
	err = epTemplates.ExecuteTemplate(&path, name, args)
	if err != nil {
		return
	}

	ep := apiEndpoints[name]
	url := path.String()
	req, err := http.NewRequest(string(ep.Method), url, r.Body)
	if err != nil {
		return
	}

	req.Header.Set("X-ADN-Migration-Overrides", "response_envelope=1")
	if r.Token != "" {
		req.Header.Set("Authorization", "Bearer "+r.Token)
	}
	if r.BodyType != "" {
		req.Header.Set("Content-Type", r.BodyType)
	}

	resp, err := apiHttpClient.Do(req)
	if err != nil {
		return
	}
	body = resp.Body

	return
}

func (c *Application) Do(r *Request, name string, args EpArgs, v interface{}) error {
	body, err := c.request(r, name, args)
	if err != nil {
		return err
	}
	defer body.Close()

	resp, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	epOptions := apiEndpoints[name].Options
	if epOptions == nil || epOptions.ResponseEnvelope {
		re := &responseEnvelope{Data: v}
		err = json.Unmarshal(resp, re)
		if err != nil {
			return err
		}

		if re.Meta.ErrorId != "" {
			return APIError(re.Meta)
		}
	} else {
		err = json.Unmarshal(resp, v)
		if err != nil {
			return err
		}
	}

	return err
}

func (c *Application) AuthenticationURL(state string) (string, error) {
	var url bytes.Buffer
	args := struct {
		*Application
		State string
	}{c, state}
	err := epTemplates.ExecuteTemplate(&url, "authentication url", args)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (c *Application) AccessToken(code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", c.Id)
	data.Set("client_secret", c.Secret)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", c.RedirectURI)
	data.Set("code", code)

	r := &Request{
		Body:     strings.NewReader(data.Encode()),
		BodyType: "application/x-www-form-urlencoded",
	}

	//{"error": "This code has already been used."}
	//{"access_token": "x", "username": "whee", "user_id": 19058}

	resp := &struct {
		AccessToken string `json:"access_token"`
		Error       string
	}{}
	err := c.Do(r, "get access token", EpArgs{}, resp)
	if err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", errors.New(resp.Error)
	}
	return resp.AccessToken, nil
}
