package main

import (
	"errors"
	"fmt"

	vault "github.com/hashicorp/vault/api"
)

type VaultClient struct {
	client *vault.Client
}

func newVaultClient(client *vault.Client) *VaultClient {
	return &VaultClient{
		client: client,
	}
}

func (c *VaultClient) ListBuckets() ([]string, error) {
	res, err := c.client.Logical().List("secret/metadata/buckets")
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
	return []string{}, nil
}

func (c *VaultClient) BucketExists(id string) (bool, error) {
	res, err := c.client.Logical().Read("secret/metadata/buckets/" + id)
	if err != nil {
		return false, err
	}
	if res == nil {
		return false, errors.New("Path not found")
	}
	return true, nil
}

func (c *VaultClient) DeleteBucket(id string) error {
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

func (c *VaultClient) CreateBucket(id string) error {
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

func (c *VaultClient) ListUsers() ([]string, error) {
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
	return []string{}, nil
}

func (c *VaultClient) UserExists(id string) (bool, error) {
	res, err := c.client.Logical().Read("auth/token/roles/" + id)
	if err != nil {
		return false, err
	}
	if res == nil {
		return false, errors.New("Path not found")
	}
	return true, nil
}

func (c *VaultClient) DeleteUser(id string) error {
	_, err := c.client.Logical().Delete("auth/token/roles/" + id)
	if err != nil {
		return err
	}
	return nil
}

func (c *VaultClient) CreateUser(id string, tokens []string, buckets []string) error {
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
