package test

import (
	"net/http"

	"github.com/google/uuid"
)

func (s *Suite) TestCookieIsSavedInProxy() {
	s.runFakeServerGeneratingCookieIfNotPresentAndReturnsItsValue()

	bucket := uuid.NewString()
	user := uuid.NewString()

	s.createBucket(bucket)
	s.createUser(user)
	s.grantAccess(user, bucket)
	token := s.generateToken(user, bucket)

	client := s.testClient(bucket, token)
	first := client.Get()
	second := client.Get()

	s.logMitmLogs()
	s.Require().Equal(first, second)
}

func (s *Suite) TestCookieIsSavedInProxyForAllUsersInTheSameBucket() {
	s.runFakeServerGeneratingCookieIfNotPresentAndReturnsItsValue()

	bucket := uuid.NewString()
	user1 := uuid.NewString()
	user2 := uuid.NewString()

	s.createBucket(bucket)
	s.createUser(user1)
	s.createUser(user2)
	s.grantAccess(user1, bucket)
	s.grantAccess(user2, bucket)
	token1 := s.generateToken(user1, bucket)
	token2 := s.generateToken(user2, bucket)

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
	user := uuid.NewString()

	s.createBucket(bucket1)
	s.createBucket(bucket2)
	s.createUser(user)
	s.grantAccess(user, bucket1)
	s.grantAccess(user, bucket2)
	token := s.generateToken(user, bucket1, bucket2)

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
	user := uuid.NewString()

	s.createBucket(bucket)
	s.createUser(user)
	s.grantAccess(user, bucket)
	token := s.generateToken(user, bucket)

	client := s.testClient(uuid.NewString(), token)
	res, err := client.RawGet()
	s.Require().NoError(err)
	s.Require().Equal(http.StatusProxyAuthRequired, res.StatusCode)
}
