package rtspclient

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

const (
	AuthenticatorTypeNone = iota
	AuthenticatorTypeDigest
	AuthenticatorTypeBasic
)

type Authenticator struct {
	realm            string
	nonce            string
	username         string
	password         string
	authenicatorType int
}

func NewAuthenticator() *Authenticator {
	return &Authenticator{
		authenicatorType: AuthenticatorTypeNone,
	}
}

func MD5StringToString(src string) string {
	context := md5.New()
	context.Write([]byte(src))
	return hex.EncodeToString(context.Sum(nil))
}

func (authenticator *Authenticator) createAuthenticatorString(cmd, url string) string {
	var authenticatorString string
	if authenticator.authenicatorType == AuthenticatorTypeBasic {
		authenticatorString = "Authorization: Basic %s\r\n"
		strResponse := base64.StdEncoding.EncodeToString([]byte(authenticator.username + ":" + authenticator.password))
		authenticatorString = fmt.Sprintf(authenticatorString, strResponse)
	} else if authenticator.authenicatorType == AuthenticatorTypeDigest {
		// The "response" field is computed as:
		//    md5(md5(<username>:<realm>:<password>):<nonce>:md5(<cmd>:<url>))
		// or, if "PasswordIsMD5" is True:
		//    md5(<password>:<nonce>:md5(<cmd>:<url>))
		strResponse := MD5StringToString(MD5StringToString(authenticator.username+":"+authenticator.realm+":"+authenticator.password) + ":" + authenticator.nonce + ":" + MD5StringToString(cmd+":"+url))
		authenticatorString = "Authorization: Digest username=\"%s\", realm=\"%s\", nonce=\"%s\", uri=\"%s\", response=\"%s\"\r\n"
		authenticatorString = fmt.Sprintf(authenticatorString, authenticator.username, authenticator.realm, authenticator.nonce, url, strResponse)
	}
	return authenticatorString
}
