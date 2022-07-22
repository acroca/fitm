package test

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"

	fitm "github.com/acroca/fitm/pkg"
	"github.com/stretchr/testify/require"
)

type Suite struct {
	t    *testing.T
	isCI bool

	vaultContainer string
	localVaultAddr string

	mitmContainer string
	mitmPort      string

	fakeServerAddr string
}

func newSuite(t *testing.T) *Suite {
	suite := &Suite{
		t:    t,
		isCI: os.Getenv("CI") == "true",
	}
	suite.network()
	suite.runVault()
	suite.runMitm()
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
	if !s.isCI {
		s.cmd(s.t, "docker network rm fitm_test")
	}
}

func (s *Suite) network() {
	if !s.isCI {
		s.cmd(s.t, "docker network create fitm_test")
	}
}

func (s *Suite) runVault() {
	if s.isCI {
		s.vaultContainer = s.cmd(s.t, "docker run --name fitm_test_vault -d --cap-add=IPC_LOCK --network host -e VAULT_DEV_ROOT_TOKEN_ID=myroot vault:1.9.4")
		s.localVaultAddr = "http://localhost:8200"
	} else {
		s.vaultContainer = s.cmd(s.t, "docker run --name fitm_test_vault -d --cap-add=IPC_LOCK --network fitm_test -p 8200 -e VAULT_DEV_ROOT_TOKEN_ID=myroot vault:1.9.4")

		vaultAddr := s.cmd(s.t, "docker port "+s.vaultContainer+" 8200")
		firstLine := strings.Split(string(vaultAddr), "\n")[0]
		mappedVaultPort := strings.Split(firstLine, ":")[1]

		s.localVaultAddr = fmt.Sprintf("http://localhost:%v", mappedVaultPort)
	}
}

func (s *Suite) runMitm() {
	mitmImage := s.cmd(s.t, "docker build ../proxy -q")

	if s.isCI {
		s.mitmContainer = s.cmd(s.t, "docker run -d --network host -e VAULT_ADDRESS=http://localhost:8200 "+mitmImage)
		s.mitmPort = "8080"
	} else {
		mitmImage := s.cmd(s.t, "docker build ../proxy -q")
		s.mitmContainer = s.cmd(s.t, "docker run -d --add-host=fake-server:host-gateway --network fitm_test -p 8080 -e VAULT_ADDRESS=http://fitm_test_vault:8200 "+mitmImage)

		mitmAddr := s.cmd(s.t, "docker port "+s.mitmContainer+" 8080")
		firstLine := strings.Split(string(mitmAddr), "\n")[0]
		s.mitmPort = strings.Split(firstLine, ":")[1]
	}
}

func (s *Suite) runFakeServer(fn http.HandlerFunc) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(s.t, err)
	port := listener.Addr().(*net.TCPAddr).Port

	if s.isCI {
		s.fakeServerAddr = fmt.Sprintf("http://localhost:%v", port)
	} else {
		s.fakeServerAddr = fmt.Sprintf("http://fake-server:%v", port)
	}

	go http.Serve(listener, fn)
}

func (s *Suite) cmd(t *testing.T, c string) string {
	parts := strings.Split(c, " ")
	res, err := exec.Command(parts[0], parts[1:]...).Output()
	require.NoError(t, err)
	return strings.TrimSpace(string(res))
}
