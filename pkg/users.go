package fitm

import (
	"fmt"

	cli "github.com/urfave/cli/v2"
)

func ListUsersAction(c *cli.Context) error {
	repo := newVaultRepository(c.String("vault_address"), c.String("vault_token"))

	users, err := repo.ListUsers()
	if err != nil {
		return err
	}
	if len(users) == 0 {
		fmt.Println("No users found.")
	} else {
		fmt.Println("Users:")

		for _, bucket := range users {
			fmt.Printf("* %v", bucket)
		}
	}

	return nil
}

func CreateUsersAction(c *cli.Context) error {
	repo := newVaultRepository(c.String("vault_address"), c.String("vault_token"))

	err := repo.CreateUser(c.String("id"), c.StringSlice("token"), c.StringSlice("bucket"))
	if err != nil {
		return err
	}
	fmt.Println("User created.")
	return nil
}

func DeleteUsersAction(c *cli.Context) error {
	repo := newVaultRepository(c.String("vault_address"), c.String("vault_token"))

	err := repo.DeleteUser(c.String("id"))
	if err != nil {
		return err
	}
	fmt.Println("User deleted.")
	return nil
}
