/*
	Copyright (c) 2012 Brian Hetro <whee@smaertness.net>
	Use of this source code is governed by the ISC
		license that can be found in the LICENSE file.
*/
package appnet

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Application struct {
	Id             string
	Secret         string
	RedirectURI    string
	Scopes         Scopes
	PasswordSecret string
	Token          string // Access token, Added, DAW, 06-Nov-2013
	UserName       string // User name, Added, DAW, 06-Nov-2013
	UserId         string // User ID, Added, DAW, 06-Nov-2013
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

	ep := ApiEndpoints[name]
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

/*
	Do handles all API requests.

	The Request contains the authentication token and optional body.

	The name comes from ApiEndpoints, with template arguments
		provided in args.

	The response is unpacked into v.

	In the future, you would not call this function directly, but
		instead use this helper function for the specific action.
*/
func (c *Application) Do(r *Request, name string, args EpArgs, v interface{}) (err error) {
	body, err := c.request(r, name, args)

	if err != nil {
		//		fmt.Printf("(appnet.Do 1) err = '%s'\n", err)
		return
	}

	defer body.Close()

	resp, err := ioutil.ReadAll(body)

	if err != nil {
		//		fmt.Printf("(appnet.Do 2) err = '%s'\n", err)
		return
	}

	epOptions := ApiEndpoints[name].Options

	if epOptions == nil || epOptions.ResponseEnvelope {
		err = json.Unmarshal(resp, v)

		if err != nil {
			fmt.Printf("(appnet.Do 3) err = '%s'\n", err)
			return
		}

		if re.Meta.ErrorId != "" {
			return APIError(re.Meta)
		}
	} else {
		err = json.Unmarshal(resp, v)

		if err != nil {
			fmt.Printf("(appnet.Do 4) err = '%s'\n", err)
			return
		}
	}

	return
}

// Generate the authentication URL for the server-side flow.
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

/*
	During server-side flow, the user will be redirected back with a
		code.

	AccessToken() uses this code to request an access token for the
		user, which is returned as a string.
*/
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

/*
	PasswordToken is used to carry out the password flow. The function
		submits the username and password to get an access token. This
		token is returned as a string.

		** Works **
*/
func (c *Application) PasswordToken(userName, userPassword string) (aToken string, err error) {
	type Response struct {
		AccessToken string `json:"access_token"`
		Error       string
	}

	data := url.Values{}
	data.Set("client_id", c.Id)
	data.Set("password_grant_secret", c.PasswordSecret)
	data.Set("grant_type", "password")
	data.Set("username", userName)
	data.Set("password", userPassword)
	data.Set("scope", c.Scopes.Spaced())

	r := &Request{
		Body:     strings.NewReader(data.Encode()),
		BodyType: "application/x-www-form-urlencoded",
	}

	resp := &Response{}

	err = c.Do(r, "get access token", EpArgs{}, resp)

	aToken = ""

	if err != nil {
		return
	}

	if resp.Error != "" {
		err = errors.New(resp.Error)
	} else {
		aToken = resp.AccessToken
	}

	return
}
