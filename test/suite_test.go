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
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite

	IsCI bool

	RepoClient *fitm.Repository

	VaultContainer string
	LocalVaultAddr string

	MitmContainer string
	MitmPort      string

	FakeServer     *http.Server
	FakeServerAddr string
}

func (s *Suite) SetupSuite() {
	rand.Seed(time.Now().UnixNano())
	s.IsCI = os.Getenv("CI") == "true"
	s.network()
	s.runVault()
	s.runMitm()
}

func (s *Suite) TearDownSuite() {
	s.cmd("docker rm -f " + s.VaultContainer + " " + s.MitmContainer)
	if !s.IsCI {
		s.cmd("docker network rm fitm_test")
	}
}

func (s *Suite) TearDownTest() {
	if s.FakeServer != nil {
		s.FakeServer.Close()
	}
	s.FakeServer = nil
}

func (s *Suite) testClient(bucket, token string) *TestClient {
	return newTestClient(s, fmt.Sprintf("localhost:%v", s.MitmPort), s.FakeServerAddr, bucket, token)
}

func (s *Suite) network() {
	if !s.IsCI {
		s.cmd("docker network create fitm_test")
	}
}

func (s *Suite) runVault() {
	if s.IsCI {
		s.VaultContainer = s.cmd("docker run --name fitm_test_vault -d --cap-add=IPC_LOCK --network host -e VAULT_DEV_ROOT_TOKEN_ID=myroot vault:1.9.4")
		s.LocalVaultAddr = "http://localhost:8200"
	} else {
		s.VaultContainer = s.cmd("docker run --name fitm_test_vault -d --cap-add=IPC_LOCK --network fitm_test -p 8200 -e VAULT_DEV_ROOT_TOKEN_ID=myroot vault:1.9.4")

		vaultAddr := s.cmd("docker port " + s.VaultContainer + " 8200")
		firstLine := strings.Split(string(vaultAddr), "\n")[0]
		mappedVaultPort := strings.Split(firstLine, ":")[1]

		s.LocalVaultAddr = fmt.Sprintf("http://localhost:%v", mappedVaultPort)
	}
	s.RepoClient = fitm.NewVaultRepository(s.LocalVaultAddr, "myroot")
}

func (s *Suite) runMitm() {
	mitmImage := s.cmd("docker build ../proxy -q")

	if s.IsCI {
		s.MitmContainer = s.cmd("docker run -d --network host -e VAULT_ADDRESS=http://localhost:8200 " + mitmImage)
		s.MitmPort = "8080"
	} else {
		mitmImage := s.cmd("docker build ../proxy -q")
		s.MitmContainer = s.cmd("docker run -d --add-host=fake-server:host-gateway --network fitm_test -p 8080 -e VAULT_ADDRESS=http://fitm_test_vault:8200 " + mitmImage)

		mitmAddr := s.cmd("docker port " + s.MitmContainer + " 8080")
		firstLine := strings.Split(string(mitmAddr), "\n")[0]
		s.MitmPort = strings.Split(firstLine, ":")[1]
	}
	dir, err := os.Getwd()
	s.Require().NoError(err)
	s.cmd("docker cp " + dir + "/../proxy/fitm.py " + s.MitmContainer + ":/root/fitm.py")
	s.waitForMitm()
}

func (s *Suite) waitForMitm() {
	max := 40
	mustContain := "Proxy server listening at"

	var logs string
	for max > 0 {
		logs = s.cmd("docker logs " + s.MitmContainer)
		if strings.Contains(logs, mustContain) {
			return
		}
		time.Sleep(50 * time.Millisecond)
		max--
	}
	s.Require().Contains(logs, mustContain)
}

func (s *Suite) runFakeServer(fn http.HandlerFunc) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)
	port := listener.Addr().(*net.TCPAddr).Port

	if s.IsCI {
		s.FakeServerAddr = fmt.Sprintf("http://localhost:%v", port)
	} else {
		s.FakeServerAddr = fmt.Sprintf("http://fake-server:%v", port)
	}

	s.FakeServer = &http.Server{Handler: fn}
	go s.FakeServer.Serve(listener)
}

func (s *Suite) runFakeServerGeneratingCookieIfNotPresentAndReturnsItsValue() {
	s.runFakeServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		number := 0

		currentCookie, err := r.Cookie("test-cookie")
		if err == http.ErrNoCookie {
			number = rand.Intn(10000000)
		} else {
			number, err = strconv.Atoi(currentCookie.Value)
			s.Require().NoError(err)
		}
		w.Header().Add("Set-Cookie", fmt.Sprintf("test-cookie=%v,", number))

		w.WriteHeader(200)
		w.Write([]byte(strconv.Itoa(number)))
	}))
}

func (s *Suite) createBucket(id string) {
	s.Require().NoError(s.RepoClient.CreateBucket(id))
}

func (s *Suite) createUser(userID string) {
	s.Require().NoError(s.RepoClient.CreateUser(userID))
}

func (s *Suite) grantAccess(userID, bucketID string) {
	s.Require().NoError(s.RepoClient.GrantAccess(userID, bucketID))
}

func (s *Suite) revokeAccess(userID, bucketID string) {
	s.Require().NoError(s.RepoClient.RevokeAccess(userID, bucketID))
}

func (s *Suite) generateToken(userID string, bucketIDs ...string) string {
	token, err := s.RepoClient.GenerateAuthToken(userID, bucketIDs)
	s.Require().NoError(err)
	return token
}

func (s *Suite) cmd(c string) string {
	parts := strings.Split(c, " ")
	res, err := exec.Command(parts[0], parts[1:]...).Output()

	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			s.T().Log("STDERR: ", string(ee.Stderr))
		}
		s.Require().NoError(err)
	}
	return strings.TrimSpace(string(res))
}

func (s *Suite) logMitmLogs() {
	logs := s.cmd("docker logs " + s.MitmContainer)
	s.T().Log("MITM Logs: ", logs)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(Suite))
}
