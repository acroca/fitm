package fitm

import (
	"errors"
	"fmt"
	"log"

	vault "github.com/hashicorp/vault/api"
)

type Repository struct {
	client *vault.Client
}

func NewVaultRepository(address, token string) *Repository {
	config := vault.DefaultConfig()
	config.Address = address
	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
	}
	client.SetToken(token)

	return &Repository{
		client: client,
	}
}

func (c *Repository) ListBuckets() ([]string, error) {
	res, err := c.client.Logical().List("secret/metadata/buckets")
	if err != nil {
		return nil, err
	}
	secrets := []string{}
	if res == nil {
		return secrets, nil
	}
	for _, k := range res.Data["keys"].([]interface{}) {
		secrets = append(secrets, k.(string))
	}
	return secrets, nil
}

func (c *Repository) BucketExists(id string) (bool, error) {
	res, err := c.client.Logical().Read("secret/metadata/buckets/" + id)
	if err != nil {
		return false, err
	}
	if res == nil {
		return false, errors.New("Path not found")
	}
	return true, nil
}

func (c *Repository) DeleteBucket(id string) error {
	_, err := c.client.Logical().Delete("secret/metadata/buckets/" + id)
	if err != nil {
		return err
	}
	_, err = c.client.Logical().Delete("sys/policy/b." + id)
	if err != nil {
		return err
	}
	return nil
}

func (c *Repository) CreateBucket(id string) error {
	_, err := c.client.Logical().Write("secret/data/buckets/"+id, map[string]interface{}{
		"options": map[string]int{
			"cas": 0,
		},
		"data": map[string]string{
			"cookies": "[]",
		},
	})
	if err != nil {
		return err
	}

	_, err = c.client.Logical().Write("sys/policy/b."+id, map[string]interface{}{
		"policy": fmt.Sprintf(`
			path "secret/data/buckets/%[1]s" {
				capabilities = ["read", "update"]
			}

			path "secret/metadata/buckets/*" {
				capabilities = ["list"]
			}
		`, id),
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Repository) ListUsers() ([]string, error) {
	res, err := c.client.Logical().List("auth/token/roles")
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, errors.New("Path not found")
	}
	secrets := []string{}
	for _, k := range res.Data["keys"].([]interface{}) {
		secrets = append(secrets, k.(string))
	}
	return secrets, nil
}

func (c *Repository) UserExists(id string) (bool, error) {
	res, err := c.client.Logical().Read("auth/token/roles/" + id)
	if err != nil {
		return false, err
	}
	if res == nil {
		return false, errors.New("Path not found")
	}
	return true, nil
}

func (c *Repository) DeleteUser(id string) error {
	_, err := c.client.Logical().Delete("auth/token/roles/" + id)
	if err != nil {
		return err
	}
	return nil
}

func (c *Repository) CreateUser(id string, tokens []string, buckets []string) error {
	policies := []string{"default"}
	for _, bucket := range buckets {
		policies = append(policies, "b."+bucket)
	}
	_, err := c.client.Logical().Write("auth/token/roles/"+id, map[string]interface{}{
		"options": map[string]int{
			"cas": 0,
		},
		"allowed_policies": policies,
	})
	if err != nil {
		return err
	}

	for _, token := range tokens {
		_, err = c.client.Logical().Write("auth/token/create/"+id, map[string]interface{}{
			"id":       token,
			"policies": policies,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
