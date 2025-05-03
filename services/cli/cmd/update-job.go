package cmd

import (
	"fmt"
	"time"

	"github.com/jacobmcgowan/simple-scheduler/services/cli/cmd/options"
	"github.com/jacobmcgowan/simple-scheduler/services/cli/services"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"github.com/spf13/cobra"
)

var updateJobOptions = options.JobOptions{}

var updateJobCmd = &cobra.Command{
	Use:     "job",
	Aliases: []string{"j"},
	Short:   "Updates a job",
	Long: `Makes changes to a scheduled job. Only specified fields will be
changed; for example, "update job -n myjob --enabled false" will only change the
enabled status of the job named "myjob".`,
	RunE: func(cmd *cobra.Command, args []string) error {
		jobUpdate := dtos.JobUpdate{}
		if cmd.Flags().Changed("enabled") {
			jobUpdate.Enabled = &updateJobOptions.Enabled
		}
		if cmd.Flags().Changed("interval") {
			jobUpdate.Interval = &updateJobOptions.Interval
		}
		if cmd.Flags().Changed("run-execution-timeout") {
			jobUpdate.RunExecutionTimeout = &updateJobOptions.RunExecutionTimeout
		}
		if cmd.Flags().Changed("run-start-timeout") {
			jobUpdate.RunStartTimeout = &updateJobOptions.RunStartTimeout
		}
		if cmd.Flags().Changed("max-queue-count") {
			jobUpdate.MaxQueueCount = &updateJobOptions.MaxQueueCount
		}
		if cmd.Flags().Changed("allow-concurrent-runs") {
			jobUpdate.AllowConcurrentRuns = &updateJobOptions.AllowConcurrentRuns
		}
		if cmd.Flags().Changed("heartbeat-timeout") {
			jobUpdate.HeartbeatTimeout = &updateJobOptions.HeartbeatTimeout
		}

		if cmd.Flags().Changed("next-run-at") {
			nextRunAtTime, err := time.Parse(time.RFC3339, updateJobOptions.NextRunAt)
			if err != nil {
				return fmt.Errorf("nextRunAt, %s, is not a valid RFC3339 datetime", updateJobOptions.NextRunAt)
			}

			jobUpdate.NextRunAt = &nextRunAtTime
		}

		authSvc := services.AuthService{}
		token, err := authSvc.GetAccessToken()
		if err != nil {
			return fmt.Errorf("failed to get access token: %s", err.Error())
		}

		svc := services.JobService{
			ApiUrl:      ApiUrl,
			AccessToken: token,
		}

		if err := svc.Edit(updateJobOptions.Name, jobUpdate); err != nil {
			return fmt.Errorf("failed to update job: %s", err.Error())
		}

		return nil
	},
}

func init() {
	updateCmd.AddCommand(updateJobCmd)
	updateJobCmd.Flags().StringVarP(&updateJobOptions.Name, "name", "n", "", "The name of the job.")
	updateJobCmd.MarkFlagRequired("name")
	updateJobCmd.Flags().BoolVarP(&updateJobOptions.Enabled, "enabled", "e", true, "Whether the job is enabled.")
	updateJobCmd.Flags().StringVarP(&updateJobOptions.NextRunAt, "next-run-at", "r", "", "The next time the job should run.")
	updateJobCmd.Flags().IntVarP(&updateJobOptions.Interval, "interval", "i", 0, "The interval to run the job in milliseconds.")
	updateJobCmd.Flags().IntVarP(&updateJobOptions.RunExecutionTimeout, "run-execution-timeout", "x", 0, "The time in milliseconds to wait for each run to complete.")
	updateJobCmd.Flags().IntVarP(&updateJobOptions.RunStartTimeout, "run-start-timeout", "s", 0, "The time in milliseconds to wait for each run to start to start.")
	updateJobCmd.Flags().IntVarP(&updateJobOptions.MaxQueueCount, "max-queue-count", "q", 0, "The maximum number of runs that can be queued.")
	updateJobCmd.Flags().BoolVarP(&updateJobOptions.AllowConcurrentRuns, "allow-concurrent-runs", "c", false, "Whether to allow concurrent runs of the job.")
	updateJobCmd.Flags().IntVarP(&updateJobOptions.HeartbeatTimeout, "heartbeat-timeout", "t", 0, "The time in milliseconds to wait for each heartbeat of a run.")
}
