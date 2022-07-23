package test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCookieIsSavedInProxy(t *testing.T) {
	suite := newSuite(t)
	defer suite.teardown()

	suite.runFakeServerGeneratingCookieIfNotPresentAndReturnsItsValue()

	bucket := uuid.New().String()
	token := uuid.New().String()

	suite.createBucket(bucket)
	suite.createUser("testUser", []string{token}, []string{bucket})

	client := suite.testClient(bucket, token)
	first := client.Get()
	second := client.Get()

	require.Equal(t, first, second)
}

func TestCookieIsSavedInProxyForAllUsersInTheSameBucket(t *testing.T) {
	suite := newSuite(t)
	defer suite.teardown()

	suite.runFakeServerGeneratingCookieIfNotPresentAndReturnsItsValue()

	bucket := uuid.New().String()
	token1 := uuid.New().String()
	token2 := uuid.New().String()

	suite.createBucket(bucket)
	suite.createUser("testUser", []string{token1}, []string{bucket})
	suite.createUser("testUser", []string{token2}, []string{bucket})

	client := suite.testClient(bucket, token1)
	first := client.Get()
	client = suite.testClient(bucket, token2)
	second := client.Get()

	require.Equal(t, first, second)
}
