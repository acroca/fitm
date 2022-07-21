package test

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	fitm "github.com/acroca/fitm/pkg"
	"github.com/stretchr/testify/require"
)

type Suite struct {
	t *testing.T

	vaultContainer string
	localVaultAddr string

	mitmContainer string
	mitmPort      string

	fakeServerPort int
}

func newSuite(t *testing.T) *Suite {
	suite := &Suite{
		t: t,
	}
	suite.network()
	suite.runVault()
	suite.runMitm()
	suite.runFakeServer()
	return suite
}

func (s *Suite) repo() *fitm.Repository {
	return fitm.NewVaultRepository(s.localVaultAddr, "myroot")
}

func (s *Suite) httpClient() *http.Client {
	localMitmURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("localhost:%v", s.mitmPort),
		User:   url.UserPassword("testBucket", "testToken"),
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(localMitmURL),
		},
	}
}

func (s *Suite) teardown() {
	s.cmd(s.t, "docker rm -f "+s.vaultContainer+" "+s.mitmContainer)
	s.cmd(s.t, "docker network rm fitm_test")
}

func (s *Suite) network() {
	s.cmd(s.t, "docker network create fitm_test")
}

func (s *Suite) runVault() {
	s.vaultContainer = s.cmd(s.t, "docker run --name fitm_test_vault --network fitm_test -d --cap-add=IPC_LOCK -p 8200 -e VAULT_DEV_ROOT_TOKEN_ID=myroot vault:1.9.4")

	vaultAddr := s.cmd(s.t, "docker port "+s.vaultContainer+" 8200")
	mappedVaultPort := strings.Split(string(vaultAddr), ":")[1]

	s.localVaultAddr = fmt.Sprintf("http://localhost:%v", mappedVaultPort)
}

func (s *Suite) runMitm() {
	mitmImage := s.cmd(s.t, "docker build ../proxy -q")
	s.mitmContainer = s.cmd(s.t, "docker run -d -p 8080 --add-host=host.docker.internal:host-gateway --network fitm_test -e VAULT_ADDRESS=http://fitm_test_vault:8200 "+mitmImage)

	defer func() {
	}()

	mitmAddr := s.cmd(s.t, "docker port "+s.mitmContainer+" 8080")
	s.mitmPort = strings.Split(string(mitmAddr), ":")[1]
}

func (s *Suite) runFakeServer() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(s.t, err)
	s.fakeServerPort = listener.Addr().(*net.TCPAddr).Port

	go http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.t.Log(r.Header)

		number := 0

		currentCookie, err := r.Cookie("test-cookie")
		if err == http.ErrNoCookie {
			number = rand.Intn(10000000)
		} else {
			number, err = strconv.Atoi(currentCookie.Value)
			require.NoError(s.t, err)
		}
		w.Header().Add("Set-Cookie", fmt.Sprintf("test-cookie=%v,", number))

		w.WriteHeader(200)
		w.Write([]byte(strconv.Itoa(number)))
	}))
}

func (s *Suite) cmd(t *testing.T, c string) string {
	parts := strings.Split(c, " ")
	res, err := exec.Command(parts[0], parts[1:]...).Output()
	require.NoError(t, err)
	return strings.TrimSpace(string(res))
}
