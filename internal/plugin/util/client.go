// Package util implements utility and convenience methods for plugins. It is
// not intended for the main stash code to access.
package util

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/shurcooL/graphql"

	"stash-plugin-duplicate-finder/internal/plugin/common"
)

// NewClient creates a graphql Client connecting to the stash server using
// the provided server connection details.
func NewClient(provider common.StashServerConnection, addr string) *graphql.Client {
	u, _ := url.Parse(fmt.Sprintf("http://%s:%d/graphql", addr, provider.Port))
	u.Scheme = provider.Scheme

	cookieJar, _ := cookiejar.New(nil)

	cookie := provider.SessionCookie
	if cookie != nil {
		cookieJar.SetCookies(u, []*http.Cookie{
			cookie,
		})
	}

	httpClient := &http.Client{
		Jar: cookieJar,
	}

	return graphql.NewClient(u.String(), httpClient)
}
