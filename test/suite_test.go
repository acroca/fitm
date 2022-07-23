package test

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	fitm "github.com/acroca/fitm/pkg"
	"github.com/stretchr/testify/require"
)

type Suite struct {
	t    *testing.T
	isCI bool

	repoClient *fitm.Repository

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

func (s *Suite) testClient(bucket, token string) *TestClient {
	return newTestClient(s.t, fmt.Sprintf("localhost:%v", s.mitmPort), s.fakeServerAddr, bucket, token)
}

func (s *Suite) teardown() {
	s.cmd("docker rm -f " + s.vaultContainer + " " + s.mitmContainer)
	if !s.isCI {
		s.cmd("docker network rm fitm_test")
	}
}

func (s *Suite) network() {
	if !s.isCI {
		s.cmd("docker network create fitm_test")
	}
}

func (s *Suite) runVault() {
	if s.isCI {
		s.vaultContainer = s.cmd("docker run --name fitm_test_vault -d --cap-add=IPC_LOCK --network host -e VAULT_DEV_ROOT_TOKEN_ID=myroot vault:1.9.4")
		s.localVaultAddr = "http://localhost:8200"
	} else {
		s.vaultContainer = s.cmd("docker run --name fitm_test_vault -d --cap-add=IPC_LOCK --network fitm_test -p 8200 -e VAULT_DEV_ROOT_TOKEN_ID=myroot vault:1.9.4")

		vaultAddr := s.cmd("docker port " + s.vaultContainer + " 8200")
		firstLine := strings.Split(string(vaultAddr), "\n")[0]
		mappedVaultPort := strings.Split(firstLine, ":")[1]

		s.localVaultAddr = fmt.Sprintf("http://localhost:%v", mappedVaultPort)
	}
	s.repoClient = fitm.NewVaultRepository(s.localVaultAddr, "myroot")
}

func (s *Suite) runMitm() {
	mitmImage := s.cmd("docker build ../proxy -q")

	if s.isCI {
		s.mitmContainer = s.cmd("docker run -d --network host -e VAULT_ADDRESS=http://localhost:8200 " + mitmImage)
		s.mitmPort = "8080"
	} else {
		mitmImage := s.cmd("docker build ../proxy -q")
		s.mitmContainer = s.cmd("docker run -d --add-host=fake-server:host-gateway --network fitm_test -p 8080 -e VAULT_ADDRESS=http://fitm_test_vault:8200 " + mitmImage)

		mitmAddr := s.cmd("docker port " + s.mitmContainer + " 8080")
		firstLine := strings.Split(string(mitmAddr), "\n")[0]
		s.mitmPort = strings.Split(firstLine, ":")[1]
	}
	s.waitForMitm()
}

func (s *Suite) waitForMitm() {
	max := 40
	mustContain := "Proxy server listening at"

	var logs string
	for max > 0 {
		logs = s.cmd("docker logs " + s.mitmContainer)
		if strings.Contains(logs, mustContain) {
			return
		}
		time.Sleep(50 * time.Millisecond)
		max--
	}
	require.Contains(s.t, logs, mustContain)
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

func (s *Suite) runFakeServerGeneratingCookieIfNotPresentAndReturnsItsValue() {
	s.runFakeServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func (s *Suite) createBucket(id string) {
	require.NoError(s.t, s.repoClient.CreateBucket(id))
}

func (s *Suite) createUser(id string, tokens []string, buckets []string) {
	require.NoError(s.t, s.repoClient.CreateUser(id, tokens, buckets))
}

func (s *Suite) cmd(c string) string {
	parts := strings.Split(c, " ")
	res, err := exec.Command(parts[0], parts[1:]...).Output()
	require.NoError(s.t, err)
	return strings.TrimSpace(string(res))
}
