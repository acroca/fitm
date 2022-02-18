package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	vault "github.com/hashicorp/vault/api"
)

func main() {

	config := vault.DefaultConfig()
	config.Address = os.Getenv("VAULT_ADDRESS")
	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
	}
	client.SetToken(os.Getenv("VAULT_TOKEN"))

	vaultClient := newVaultClient(client)
	api := newAPI(vaultClient)

	router := gin.Default()
	router.GET("/buckets/:id", api.GetBucket)
	router.DELETE("/buckets/:id", api.DeleteBucket)
	router.GET("/buckets", api.ListBuckets)
	router.POST("/buckets", api.CreateBucket)

	router.GET("/users/:id", api.GetUser)
	router.DELETE("/users/:id", api.DeleteUser)
	router.GET("/users", api.ListUsers)
	router.POST("/users", api.CreateUser)

	router.Run("0.0.0.0:4000")
}
