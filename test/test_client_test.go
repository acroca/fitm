package test

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestClient struct {
	t          *testing.T
	httpClient *http.Client
	addr       string
}

func newTestClient(t *testing.T, host, addr, bucket, token string) *TestClient {
	return &TestClient{
		t:    t,
		addr: addr,
		httpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(&url.URL{
					Scheme: "http",
					Host:   host,
					User:   url.UserPassword(bucket, token),
				}),
			},
		},
	}
}

func (c *TestClient) Get() string {
	res, err := c.httpClient.Get(c.addr)
	require.NoError(c.t, err)
	require.Equal(c.t, 200, res.StatusCode)
	for _, cookie := range res.Cookies() {
		require.NotEqual(c.t, "test-cookie", cookie.Name)
	}
	return c.readBody(res)
}

func (c *TestClient) readBody(res *http.Response) string {
	all, err := ioutil.ReadAll(res.Body)
	require.NoError(c.t, err)
	defer res.Body.Close()

	return string(all)
}
