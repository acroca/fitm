package test

import (
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHappyPath(t *testing.T) {
	suite := newSuite(t)
	defer suite.teardown()

	repo := suite.repo()

	require.NoError(t, repo.CreateBucket("testBucket"))
	require.NoError(t, repo.CreateUser("testUser", []string{"testToken"}, []string{"testBucket"}))

	time.Sleep(2 * time.Second)
	client := suite.httpClient()

	res, err := client.Get(suite.fakeServerAddr)
	require.NoError(t, err)
	require.Equal(t, 200, res.StatusCode)
	first := body(t, res)
	for _, cookie := range res.Cookies() {
		require.NotEqual(t, "test-cookie", cookie.Name)
	}

	res, err = client.Get(suite.fakeServerAddr)
	require.NoError(t, err)
	require.Equal(t, 200, res.StatusCode)
	second := body(t, res)
	for _, cookie := range res.Cookies() {
		require.NotEqual(t, "test-cookie", cookie.Name)
	}

	require.Equal(t, first, second)
}

func body(t *testing.T, res *http.Response) string {
	all, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	res.Body.Close()
	return string(all)
}
