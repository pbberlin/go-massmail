package main

import (
	"errors"
	"net/smtp"
)

type auth2 struct {
	username, password string
}

// Auth2 is used for smtp login auth
// stackoverflow.com/questions/42305763/connecting-to-exchange-with-golang
func Auth2(username, password string) smtp.Auth {
	return &auth2{username, password}
}

func (a *auth2) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(a.username), nil
}

func (a *auth2) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("unknown from smtp server")
		}
	}
	return nil, nil
}
