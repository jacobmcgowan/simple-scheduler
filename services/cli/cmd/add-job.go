package cmd

import (
	"fmt"
	"time"

	"github.com/jacobmcgowan/simple-scheduler/services/cli/cmd/options"
	"github.com/jacobmcgowan/simple-scheduler/services/cli/services"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"github.com/spf13/cobra"
)

var addJobOptions = options.JobOptions{}

var addJobCmd = &cobra.Command{
	Use:     "job",
	Aliases: []string{"j"},
	Short:   "Adds a job",
	Long:    `Schedules a job.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		nextRunAtTime, err := time.Parse(time.RFC3339, addJobOptions.NextRunAt)
		if err != nil {
			return fmt.Errorf("nextRunAt, %s, is not a valid RFC3339 datetime", addJobOptions.NextRunAt)
		}

		job := dtos.Job{
			Name:                addJobOptions.Name,
			Enabled:             addJobOptions.Enabled,
			NextRunAt:           nextRunAtTime,
			Interval:            addJobOptions.Interval,
			RunExecutionTimeout: addJobOptions.RunExecutionTimeout,
			RunStartTimeout:     addJobOptions.RunStartTimeout,
			MaxQueueCount:       addJobOptions.MaxQueueCount,
			AllowConcurrentRuns: addJobOptions.AllowConcurrentRuns,
			HeartbeatTimeout:    addJobOptions.HeartbeatTimeout,
		}
		svc := services.JobService{
			ApiUrl: ApiUrl,
		}

		if _, err := svc.Add(job); err != nil {
			return fmt.Errorf("failed to add job: %s", err)
		}

		return nil
	},
}

func init() {
	addCmd.AddCommand(addJobCmd)
	addJobCmd.Flags().StringVarP(&addJobOptions.Name, "name", "n", "", "The name of the job.")
	addJobCmd.MarkFlagRequired("name")
	addJobCmd.Flags().BoolVarP(&addJobOptions.Enabled, "enabled", "e", true, "Whether the job is enabled.")
	addJobCmd.Flags().StringVarP(&addJobOptions.NextRunAt, "next-run-at", "r", "", "The next time the job should run.")
	addJobCmd.MarkFlagRequired("next-run-at")
	addJobCmd.Flags().IntVarP(&addJobOptions.Interval, "interval", "i", 0, "The interval to run the job in milliseconds.")
	addJobCmd.Flags().IntVarP(&addJobOptions.RunExecutionTimeout, "run-execution-timeout", "x", 0, "The time in milliseconds to wait for each run to complete.")
	addJobCmd.Flags().IntVarP(&addJobOptions.RunStartTimeout, "run-start-timeout", "s", 0, "The time in milliseconds to wait for each run to start to start.")
	addJobCmd.Flags().IntVarP(&addJobOptions.MaxQueueCount, "max-queue-count", "q", 0, "The maximum number of runs that can be queued.")
	addJobCmd.Flags().BoolVarP(&addJobOptions.AllowConcurrentRuns, "allow-concurrent-runs", "c", false, "Whether to allow concurrent runs of the job.")
	addJobCmd.Flags().IntVarP(&addJobOptions.HeartbeatTimeout, "heartbeat-timeout", "t", 0, "The time in milliseconds to wait for each heartbeat of a run.")
}
