package test

import (
	"net/http"

	"github.com/google/uuid"
)

func (s *Suite) TestCookieIsSavedInProxy() {
	s.runFakeServerGeneratingCookieIfNotPresentAndReturnsItsValue()

	bucket := uuid.NewString()
	token := uuid.NewString()

	s.createBucket(bucket)
	s.createUser([]string{token}, []string{bucket})

	client := s.testClient(bucket, token)
	first := client.Get()
	second := client.Get()

	s.logMitmLogs()
	s.Require().Equal(first, second)
}

func (s *Suite) TestCookieIsSavedInProxyForAllUsersInTheSameBucket() {
	s.runFakeServerGeneratingCookieIfNotPresentAndReturnsItsValue()

	bucket := uuid.NewString()
	token1 := uuid.NewString()
	token2 := uuid.NewString()

	s.createBucket(bucket)
	s.createUser([]string{token1}, []string{bucket})
	s.createUser([]string{token2}, []string{bucket})

	client := s.testClient(bucket, token1)
	first := client.Get()
	client = s.testClient(bucket, token2)
	second := client.Get()

	s.Require().Equal(first, second)
}

func (s *Suite) TestCookieIsNotSavedAcrossBuckets() {
	s.runFakeServerGeneratingCookieIfNotPresentAndReturnsItsValue()

	bucket1 := uuid.NewString()
	bucket2 := uuid.NewString()
	token := uuid.NewString()

	s.createBucket(bucket1)
	s.createBucket(bucket2)
	s.createUser([]string{token}, []string{bucket1, bucket2})

	client := s.testClient(bucket1, token)
	first := client.Get()
	client = s.testClient(bucket2, token)
	second := client.Get()

	s.Require().NotEqual(first, second)
}

func (s *Suite) TestUnauthorisedUser() {
	s.runFakeServerGeneratingCookieIfNotPresentAndReturnsItsValue()

	bucket := uuid.NewString()
	token := uuid.NewString()

	s.createBucket(bucket)

	client := s.testClient(bucket, token)
	res, err := client.RawGet()
	s.Require().NoError(err)
	s.Require().Equal(http.StatusProxyAuthRequired, res.StatusCode)
}

func (s *Suite) TestRequiresValidBucket() {
	s.runFakeServerGeneratingCookieIfNotPresentAndReturnsItsValue()

	bucket := uuid.NewString()
	token := uuid.NewString()

	s.createBucket(bucket)
	s.createUser([]string{token}, []string{bucket})

	client := s.testClient(uuid.NewString(), token)
	res, err := client.RawGet()
	s.Require().NoError(err)
	s.Require().Equal(http.StatusProxyAuthRequired, res.StatusCode)
}
