package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jacobmcgowan/simple-scheduler/services/cli/services"
	"github.com/spf13/cobra"
)

var jobsCmd = &cobra.Command{
	Use:     "jobs",
	Aliases: []string{"j"},
	Short:   "Lists the jobs",
	Long:    `Provides details on the current jobs that are scheduled.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		authSvc := services.AuthService{}
		token, err := authSvc.GetAccessToken()
		if err != nil {
			return fmt.Errorf("failed to get access token: %s", err)
		}

		svc := services.JobService{
			ApiUrl:      ApiUrl,
			AccessToken: token,
		}

		if jobs, err := svc.Browse(); err == nil {
			writer := tabwriter.NewWriter(os.Stdout, 1, 1, 4, ' ', 0)
			fmt.Fprintln(
				writer,
				"NAME\tENABLED\tNEXT RUN AT\tINTERVAL\tRUN EXECUTION TIMEOUT\tRUN START TIMEOUT\tMAX QUEUE COUNT\tALLOW CONCURRENT RUNS\tHEARTBEAT TIMEOUT")

			for _, job := range jobs {
				fmt.Fprintf(
					writer,
					"%s\t%t\t%s\t%d\t%d\t%d\t%d\t%t\t%d\n",
					job.Name,
					job.Enabled,
					job.NextRunAt,
					job.Interval,
					job.RunExecutionTimeout,
					job.RunStartTimeout,
					job.MaxQueueCount,
					job.AllowConcurrentRuns,
					job.HeartbeatTimeout)
			}

			writer.Flush()
		} else {
			return fmt.Errorf("failed to get runs: %s", err)
		}

		return nil
	},
}

func init() {
	listCmd.AddCommand(jobsCmd)
}
