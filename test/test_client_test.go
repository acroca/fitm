package test

import (
	"io/ioutil"
	"net/http"
	"net/url"
)

type TestClient struct {
	s          *Suite
	httpClient *http.Client
	addr       string
}

func newTestClient(s *Suite, host, addr, bucket, token string) *TestClient {
	return &TestClient{
		s:    s,
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
	c.s.Require().NoError(err)
	if res.StatusCode != 200 {
		c.s.logMitmLogs()
		c.s.Require().Equal(200, res.StatusCode)
	}
	for _, cookie := range res.Cookies() {
		c.s.Require().NotEqual("test-cookie", cookie.Name)
	}
	return c.readBody(res)
}

func (c *TestClient) readBody(res *http.Response) string {
	all, err := ioutil.ReadAll(res.Body)
	c.s.Require().NoError(err)
	defer res.Body.Close()

	return string(all)
}
