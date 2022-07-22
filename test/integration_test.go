package test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHappyPath(t *testing.T) {
	suite := newSuite(t)
	defer suite.teardown()

	suite.runFakeServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		number := 0

		currentCookie, err := r.Cookie("test-cookie")
		if err == http.ErrNoCookie {
			number = rand.Intn(10000000)
		} else {
			number, err = strconv.Atoi(currentCookie.Value)
			require.NoError(t, err)
		}
		w.Header().Add("Set-Cookie", fmt.Sprintf("test-cookie=%v,", number))

		w.WriteHeader(200)
		w.Write([]byte(strconv.Itoa(number)))
	}))

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
