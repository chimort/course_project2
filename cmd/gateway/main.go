package main

import (
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
)

func main() {

	e := echo.New()

	proxyTo := func(target string, prefixToStrip string) echo.HandlerFunc {
		url, _ := url.Parse(target)
		proxy := httputil.NewSingleHostReverseProxy(url)
		return func(c echo.Context) error {
			c.Request().URL.Path = strings.TrimPrefix(c.Request().URL.Path, prefixToStrip)
			c.Request().Host = url.Host
			proxy.ServeHTTP(c.Response(), c.Request())
			return nil
		}
	}

	e.Any("/user/*", proxyTo("http://user-service:8080", "/user"))
	e.Any("/matching/*", proxyTo("http://matching-service:8081", "/matching"))
	e.Any("/chat/*", proxyTo("http://chat-service:8082", "/chat"))

	e.Logger.Fatal(e.Start(":8000"))
}
