package main

import (
	"encoding/base64"
	"strings"

	"github.com/spf13/viper"

	auth "github.com/korylprince/go-ad-auth"
	"github.com/valyala/fasthttp"
)

// basicAuth returns the username and password provided in the request's
// Authorization header, if the request uses HTTP Basic Authentication.
// See RFC 2617, Section 2.
func basicAuth(ctx *fasthttp.RequestCtx) (username, password string, ok bool) {
	auth := ctx.Request.Header.Peek("Authorization")
	if auth == nil {
		return
	}
	return parseBasicAuth(string(auth))
}

// parseBasicAuth parses an HTTP Basic Authentication string.
// "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==" returns ("Aladdin", "open sesame", true).
func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}

// AuthRequired is the auth handler
func AuthRequired(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {

		if viper.GetBool("authentication.enabled") && viper.GetString("authentication.kind") != "none" {
			user, password, hasAuth := basicAuth(ctx)
			if !hasAuth {
				ctx.Error(fasthttp.StatusMessage(fasthttp.StatusUnauthorized), fasthttp.StatusUnauthorized)
				ctx.Response.Header.Set("WWW-Authenticate", "Basic realm=Restricted")
				return
			}

			switch viper.GetString("authentication.kind") {
			case "ad":

				// Get the Basic Authentication credentials

				config := &auth.Config{
					Server:   viper.GetString("authentication.ad.server"),
					Port:     viper.GetInt("authentication.ad.port"),
					BaseDN:   viper.GetString("authentication.ad.baseDN"),
					Security: auth.SecurityNone,
				}
				if conn, err := config.Connect(); err == nil && conn != nil {
					upn, err := config.UPN(user)
					if err == nil {
						conn, err := config.Connect()
						if err == nil {
							defer conn.Conn.Close()

							//status, err := auth.Authenticate(config, user, password)
							status, err := conn.Bind(upn, password)
							if err == nil && status {
								entry, err := conn.GetAttributes("userPrincipalName", upn, []string{})
								if err == nil {

									for _, a := range entry.GetAttributeValues("memberOf") {
										for _, allowedGroup := range viper.GetStringSlice("authentication.ad.allowedGroups") {
											if strings.Split(a, ",")[0][3:] == allowedGroup {
												h(ctx)
												return
											}
										}
									}
									ctx.Error(fasthttp.StatusMessage(fasthttp.StatusUnauthorized), fasthttp.StatusUnauthorized)
									ctx.Response.Header.Set("WWW-Authenticate", "Basic realm=Restricted")
									return
								}
							}
						}
					}
				}
				break
			case "basic":
				for _, creds := range viper.GetStringSlice("authentication.basic.credentials") {
					cred := strings.SplitN(creds, ":", 2)
					if user == cred[0] && password == cred[1] {
						// Delegate request to the given handle
						h(ctx)
						return

					}
				}

				break
			}
			// Request Basic Authentication otherwise
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusUnauthorized), fasthttp.StatusUnauthorized)
			ctx.Response.Header.Set("WWW-Authenticate", "Basic realm=Restricted")
			return
		}
		h(ctx)
		return
	})
}
