package test

import (
	"github.com/google/uuid"
)

func (s *Suite) TestCookieIsSavedInProxy() {
	s.runFakeServerGeneratingCookieIfNotPresentAndReturnsItsValue()

	bucket := uuid.New().String()
	token := uuid.New().String()

	s.createBucket(bucket)
	s.createUser("testUser", []string{token}, []string{bucket})

	client := s.testClient(bucket, token)
	first := client.Get()
	second := client.Get()

	s.Require().Equal(first, second)
}

func (s *Suite) TestCookieIsSavedInProxyForAllUsersInTheSameBucket() {
	s.runFakeServerGeneratingCookieIfNotPresentAndReturnsItsValue()

	bucket := uuid.New().String()
	token1 := uuid.New().String()
	token2 := uuid.New().String()

	s.createBucket(bucket)
	s.createUser("testUser", []string{token1}, []string{bucket})
	s.createUser("testUser", []string{token2}, []string{bucket})

	client := s.testClient(bucket, token1)
	first := client.Get()
	client = s.testClient(bucket, token2)
	second := client.Get()

	s.Require().Equal(first, second)
}
