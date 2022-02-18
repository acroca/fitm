package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Bucket struct {
	ID string `json:"id"`
}

type User struct {
	ID      string   `json:"id"`
	Tokens  []string `json:"tokens,omitempty"`
	Buckets []string `json:"buckets,omitempty"`
}

type API struct {
	vaultClient *VaultClient
}

func newAPI(vaultClient *VaultClient) *API {
	return &API{
		vaultClient: vaultClient,
	}
}

func (api *API) ListBuckets(c *gin.Context) {
	ids, err := api.vaultClient.ListBuckets()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	buckets := []*Bucket{}
	for _, id := range ids {
		buckets = append(buckets, &Bucket{ID: id})
	}
	c.JSON(http.StatusOK, buckets)
}

func (api *API) GetBucket(c *gin.Context) {
	id := c.Param("id")
	exists, err := api.vaultClient.BucketExists(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !exists {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, Bucket{ID: id})
}

func (api *API) DeleteBucket(c *gin.Context) {
	id := c.Param("id")
	exists, err := api.vaultClient.BucketExists(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !exists {
		c.Status(http.StatusNotFound)
		return
	}
	err = api.vaultClient.DeleteBucket(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, Bucket{ID: id})
}

func (api *API) CreateBucket(c *gin.Context) {
	var bucket Bucket
	if err := c.ShouldBindJSON(&bucket); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(bucket.ID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is not provided"})
		return
	}
	err := api.vaultClient.CreateBucket(bucket.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, bucket)
}

func (api *API) ListUsers(c *gin.Context) {
	ids, err := api.vaultClient.ListUsers()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	users := []*User{}
	for _, id := range ids {
		users = append(users, &User{ID: id})
	}
	c.JSON(http.StatusOK, users)
}

func (api *API) GetUser(c *gin.Context) {
	id := c.Param("id")
	exists, err := api.vaultClient.UserExists(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !exists {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, User{ID: id})
}

func (api *API) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	exists, err := api.vaultClient.UserExists(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !exists {
		c.Status(http.StatusNotFound)
		return
	}
	err = api.vaultClient.DeleteUser(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, User{ID: id})
}

func (api *API) CreateUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(user.ID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is not provided"})
		return
	}
	err := api.vaultClient.CreateUser(user.ID, user.Tokens, user.Buckets)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}
