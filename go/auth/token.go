// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

//
// Code used to support authentication tokens for arbitrary purposes.
//
package auth

import (
	"bytes"
	"encoding/json"
	"math"
	"time"

	libkb "github.com/keybase/client/go/libkb"
	keybase1 "github.com/keybase/client/go/protocol"
)

const (
	TokenType           = "auth"
	CurrentTokenVersion = 1
)

type TokenAuth struct {
	Server string `json:"server"`
}

type TokenKey struct {
	UID      keybase1.UID             `json:"uid"`
	Username libkb.NormalizedUsername `json:"username"`
	KID      keybase1.KID             `json:"kid"`
}

type TokenBody struct {
	Auth    TokenAuth `json:"auth"`
	Key     TokenKey  `json:"key"`
	Type    string    `json:"type"`
	Version int       `json:"version"`
}

type TokenClient struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Token struct {
	Body         TokenBody   `json:"body"`
	Client       TokenClient `json:"client"`
	CreationTime int64       `json:"ctime"`
	ExpireIn     int         `json:"expire_in"`
	Tag          string      `json:"tag"`
}

func NewToken(uid keybase1.UID, username libkb.NormalizedUsername, kid keybase1.KID,
	server string, expireIn int, clientName string, clientVersion string) *Token {
	return &Token{
		Body: TokenBody{
			Auth: TokenAuth{
				Server: server,
			},
			Key: TokenKey{
				UID:      uid,
				Username: username,
				KID:      kid,
			},
			Type:    TokenType,
			Version: CurrentTokenVersion,
		},
		Client: TokenClient{
			Name:    clientName,
			Version: clientVersion,
		},
		CreationTime: time.Now().Unix(),
		ExpireIn:     expireIn,
		Tag:          "signature",
	}
}

func (t Token) Bytes() []byte {
	bytes, err := json.Marshal(&t)
	if err != nil {
		return []byte{}
	}
	return bytes
}

func (t Token) String() string {
	return string(t.Bytes())
}

func VerifyToken(signature, server string, maxExpireIn int) (*Token, error) {
	var t *Token
	key, token, _, err := libkb.NaclVerifyAndExtract(signature)
	if err != nil {
		return nil, err
	}
	if t, err = parseToken(token); err != nil {
		return nil, err
	}
	if key.GetKID() != t.KID() {
		return nil, InvalidTokenKeyError{
			expected: key.GetKID().String(),
			received: t.KID().String(),
		}
	}
	if TokenType != t.Type() {
		return nil, InvalidTokenTypeError{
			expected: TokenType,
			received: t.Type(),
		}
	}
	if server != t.Server() {
		return nil, InvalidTokenServerError{
			expected: server,
			received: t.Server(),
		}
	}
	remaining := t.TimeRemaining()
	if remaining > maxExpireIn {
		return nil, MaxTokenExpiresError{
			creationTime: t.CreationTime,
			expireIn:     t.ExpireIn,
			now:          time.Now().Unix(),
			maxExpireIn:  maxExpireIn,
			remaining:    remaining,
		}
	}
	if remaining <= 0 {
		return nil, TokenExpiredError{
			creationTime: t.CreationTime,
			expireIn:     t.ExpireIn,
			now:          time.Now().Unix(),
		}
	}
	return t, nil
}

func (t Token) TimeRemaining() int {
	ctime := time.Unix(t.CreationTime, 0)
	expires := ctime.Add(time.Duration(t.ExpireIn) * time.Second)
	return int(math.Ceil(expires.Sub(time.Now()).Seconds()))
}

func (t Token) Server() string {
	return t.Body.Auth.Server
}

func (t Token) UID() keybase1.UID {
	return t.Body.Key.UID
}

func (t Token) KID() keybase1.KID {
	return t.Body.Key.KID
}

func (t Token) Username() libkb.NormalizedUsername {
	return t.Body.Key.Username
}

func (t Token) Type() string {
	return t.Body.Type
}

func (t Token) Version() int {
	return t.Body.Version
}

func (t Token) ClientName() string {
	return t.Client.Name
}

func (t Token) ClientVersion() string {
	return t.Client.Version
}

func parseToken(token []byte) (*Token, error) {
	decoder := json.NewDecoder(bytes.NewReader(token))
	decoder.UseNumber()
	var t Token
	if err := decoder.Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}
