package main

import (
	"github.com/dghubble/oauth1"
)

var FanfouEndpoint = oauth1.Endpoint{
	RequestTokenURL: "http://fanfou.com/oauth/request_token",
	AuthorizeURL:    "http://fanfou.com/oauth/authorize",
	AccessTokenURL:  "http://fanfou.com/oauth/access_token",
}
