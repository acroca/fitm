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

func (c *Repository) CreateUser(id string) error {
	_, err := c.client.Logical().Write("auth/token/roles/"+id, map[string]interface{}{
		"options": map[string]int{
			"cas": 0,
		},
		"allowed_policies": []string{"default"},
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Repository) GrantAccess(userID, bucketID string) error {
	policyName := "b." + bucketID

	res, err := c.client.Logical().Read("auth/token/roles/" + userID)
	if err != nil {
		return err
	}

	policies := res.Data["allowed_policies"].([]interface{})
	for _, policy := range policies {
		if policy.(string) == policyName {
			return nil
		}
	}
	policies = append(policies, policyName)

	_, err = c.client.Logical().Write("auth/token/roles/"+userID, map[string]interface{}{
		"allowed_policies": policies,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Repository) RevokeAccess(userID, bucketID string) error {
	policyName := "b." + bucketID

	res, err := c.client.Logical().Read("auth/token/roles/" + userID)
	if err != nil {
		return err
	}

	initialPolicies := res.Data["allowed_policies"].([]interface{})
	newPolicies := make([]string, 0)
	for _, policy := range initialPolicies {
		if policy.(string) != policyName {
			newPolicies = append(newPolicies, policy.(string))
		}
	}
	if len(initialPolicies) == len(newPolicies) {
		return nil
	}

	_, err = c.client.Logical().Write("auth/token/roles/"+userID, map[string]interface{}{
		"allowed_policies": newPolicies,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Repository) GenerateAuthToken(userID string, bucketIDs []string) (string, error) {
	policies := make([]string, 0)
	for _, bucketID := range bucketIDs {
		policies = append(policies, "b."+bucketID)
	}

	res, err := c.client.Logical().Write("auth/token/create/"+userID, map[string]interface{}{
		"policies": policies,
	})
	if err != nil {
		return "", err
	}

	return res.Auth.ClientToken, nil
}
