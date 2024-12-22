package cmd

import (
	"fmt"
	"time"

	"github.com/jacobmcgowan/simple-scheduler/services/cli/services"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"github.com/spf13/cobra"
)

var name string
var enabled bool
var nextRunAt string
var interval int
var runExecutionTimeout int
var runStartTimeout int
var maxQueueCount int
var allowConcurrentRuns bool
var heartbeatTimeout int

var jobCmd = &cobra.Command{
	Use:     "job",
	Aliases: []string{"j"},
	Short:   "Adds a job",
	Long:    `Schedules a job.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		nextRunAtTime, err := time.Parse(time.RFC3339, nextRunAt)
		if err != nil {
			return fmt.Errorf("nextRunAt, %s, is not a valid RFC3339 datetime", nextRunAt)
		}

		job := dtos.Job{
			Name:                name,
			Enabled:             enabled,
			NextRunAt:           nextRunAtTime,
			Interval:            interval,
			RunExecutionTimeout: runExecutionTimeout,
			RunStartTimeout:     runStartTimeout,
			MaxQueueCount:       maxQueueCount,
			AllowConcurrentRuns: allowConcurrentRuns,
			HeartbeatTimeout:    heartbeatTimeout,
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
	addCmd.AddCommand(jobCmd)
	jobCmd.Flags().StringVarP(&name, "name", "n", "", "The name of the job.")
	jobCmd.MarkFlagRequired("name")
	jobCmd.Flags().BoolVarP(&enabled, "enabled", "e", true, "Whether the job is enabled.")
	jobCmd.Flags().StringVarP(&nextRunAt, "next-run-at", "r", "", "The next time the job should run.")
	jobCmd.MarkFlagRequired("next-run-at")
	jobCmd.Flags().IntVarP(&interval, "interval", "i", 0, "The interval to run the job in milliseconds.")
	jobCmd.Flags().IntVarP(&runExecutionTimeout, "run-execution-timeout", "x", 0, "The time in milliseconds to wait for each run to complete.")
	jobCmd.Flags().IntVarP(&runStartTimeout, "run-start-timeout", "s", 0, "The time in milliseconds to wait for each run to start to start.")
	jobCmd.Flags().IntVarP(&maxQueueCount, "max-queue-count", "q", 0, "The maximum number of runs that can be queued.")
	jobCmd.Flags().BoolVarP(&allowConcurrentRuns, "allow-concurrent-runs", "c", false, "Whether to allow concurrent runs of the job.")
	jobCmd.Flags().IntVarP(&heartbeatTimeout, "heartbeat-timeout", "t", 0, "The time in milliseconds to wait for each heartbeat of a run.")
}
