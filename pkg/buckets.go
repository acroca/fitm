package fitm

import (
	"fmt"

	cli "github.com/urfave/cli/v2"
)

func ListBucketsAction(c *cli.Context) error {
	repo := NewVaultRepository(c.String("vault_address"), c.String("vault_token"))

	buckets, err := repo.ListBuckets()
	if err != nil {
		return err
	}
	if len(buckets) == 0 {
		fmt.Println("No buckets found.")
	} else {
		fmt.Println("Buckets:")

		for _, bucket := range buckets {
			fmt.Printf("* %v", bucket)
		}
	}

	return nil
}

func CreateBucketsAction(c *cli.Context) error {
	repo := NewVaultRepository(c.String("vault_address"), c.String("vault_token"))

	err := repo.CreateBucket(c.String("id"))
	if err != nil {
		return err
	}
	fmt.Println("Bucket created.")
	return nil
}

func DeleteBucketsAction(c *cli.Context) error {
	repo := NewVaultRepository(c.String("vault_address"), c.String("vault_token"))

	err := repo.DeleteBucket(c.String("id"))
	if err != nil {
		return err
	}
	fmt.Println("Bucket deleted.")
	return nil
}
