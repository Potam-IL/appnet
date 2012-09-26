// Copyright (c) 2012 Brian Hetro <whee@smaertness.net>
// Use of this source code is governed by the ISC
// license that can be found in the LICENSE file.

package adn

import (
	"time"
)

type User struct {
	Id          string      `json:"id"`
	Username    string      `json:"username"`
	Name        string      `json:"name"` // The user-supplied Name may be a pseudonym.
	Description Description `json:"description"`
	Timezone    string      `json:"timezone"` // The Timezone is in tzinfo format.
	Locale      string      `json:"locale"`
	AvatarImage Image       `json:"avatar_image"` // The URL and original size of the user's avatar.
	CoverImage  Image       `json:"cover_image"`  // The URL and original size of the user's cover image.
	Type        string      `json:"type"`         // An account can be human, bot, corporate, or feed.
	CreatedAt   time.Time   `json:"created_at"`   // The time at which the User was created in ISO 8601 format.
	Counts      Counts      `json:"counts"`
	AppData     interface{} `json:"app_data"`    // TODO: Opaque information an application may store.
	FollowsYou  bool        `json:"follows_you"` // Does this user follow the user making the request? May be omitted if this is not an authenticated request.
	YouFollow   bool        `json:"you_follow"`  // Does this user making the request follow this user? May be omitted if this is not an authenticated request.
	YouMuted    bool        `json:"you_muted"`   // Has the user making the request blocked this user? May be omitted if this is not an authenticated request.
}

type Description struct {
	Text     string   `json:"text"` // Biographical information
	HTML     string   `json:"html"` // Server-generated annotated HTML version of Text.
	Entities Entities `json:"entities"`
}

type Image struct {
	Height int    `json:"height"`
	Width  int    `json:"width"`
	Url    string `json:"url"`
}

type Counts struct {
	Following int `json:"following"`
	Followers int `json:"followers"`
	Posts     int `json:"posts"`
	Stars     int `json:"stars"`
}

// Get will not actually work until authentication is implemented.
func (c *Application) GetUser(id string) (u *User, err error) {
	u = &User{}
	err = c.Get(nil, "retrieve user", EpArgs{User: id}, u)
	return
}

func GetUser(id string) (*User, error) {
	return DefaultApplication.GetUser(id)
}
