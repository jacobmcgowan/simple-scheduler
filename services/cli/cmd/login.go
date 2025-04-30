package cmd

import (
	"fmt"
	"sync"
	"time"

	"github.com/jacobmcgowan/simple-scheduler/services/cli/cmd/options"
	"github.com/jacobmcgowan/simple-scheduler/services/cli/services"
	"github.com/spf13/cobra"
)

var loginOptions = options.LoginOptions{}

var loginCmd = &cobra.Command{
	Use:     "login",
	Aliases: []string{"l"},
	Short:   "Logins into the Simple Scheduler API",
	Long:    `Logs into the Simple Scheduler API using OIDC.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		wg := sync.WaitGroup{}
		authSvc := services.AuthService{
			ApiUrl: ApiUrl,
		}
		if err := authSvc.Start(cmd.Context(), loginOptions.ClientId, loginOptions.ClientSecret, loginOptions.ProviderType, &wg); err != nil {
			return fmt.Errorf("failed to start auth service: %s", err)
		}
		if err := authSvc.Login(); err != nil {
			return fmt.Errorf("failed to login: %s", err)
		}

		time.Sleep(30 * time.Second)

		if err := authSvc.Stop(); err != nil {
			return fmt.Errorf("failed to stop auth service: %s", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVarP(&loginOptions.ClientId, "client-id", "i", "", "The client id to use for login.")
	loginCmd.MarkFlagRequired("client-id")
	loginCmd.Flags().StringVarP(&loginOptions.ClientSecret, "client-secret", "s", "", "The client secret to use for login.")
	loginCmd.MarkFlagRequired("client-secret")
	loginCmd.Flags().VarP(&loginOptions.ProviderType, "provider", "p", "The provider to use for login. Supported values are: github")
	loginCmd.MarkFlagRequired("provider")
}
