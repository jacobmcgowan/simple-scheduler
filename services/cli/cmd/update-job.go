package cmd

import (
	"fmt"
	"time"

	"github.com/jacobmcgowan/simple-scheduler/services/cli/cmd/options"
	"github.com/jacobmcgowan/simple-scheduler/services/cli/services"
	"github.com/jacobmcgowan/simple-scheduler/shared/common"
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
		jobUpdate := dtos.JobUpdate{
			Enabled: common.Undefinable[bool]{
				Value:   updateJobOptions.Enabled,
				Defined: cmd.Flags().Changed("enabled"),
			},
			Interval: common.Undefinable[int]{
				Value:   updateJobOptions.Interval,
				Defined: cmd.Flags().Changed("interval"),
			},
			RunExecutionTimeout: common.Undefinable[int]{
				Value:   updateJobOptions.RunExecutionTimeout,
				Defined: cmd.Flags().Changed("run-execution-timeout"),
			},
			RunStartTimeout: common.Undefinable[int]{
				Value:   updateJobOptions.RunStartTimeout,
				Defined: cmd.Flags().Changed("run-start-timeout"),
			},
			MaxQueueCount: common.Undefinable[int]{
				Value:   updateJobOptions.MaxQueueCount,
				Defined: cmd.Flags().Changed("max-queue-count"),
			},
			AllowConcurrentRuns: common.Undefinable[bool]{
				Value:   updateJobOptions.AllowConcurrentRuns,
				Defined: cmd.Flags().Changed("allow-concurrent-runs"),
			},
			HeartbeatTimeout: common.Undefinable[int]{
				Value:   updateJobOptions.HeartbeatTimeout,
				Defined: cmd.Flags().Changed("heartbeat-timeout"),
			},
		}

		if cmd.Flags().Changed("next-run-at") {
			nextRunAtTime, err := time.Parse(time.RFC3339, updateJobOptions.NextRunAt)
			if err != nil {
				return fmt.Errorf("nextRunAt, %s, is not a valid RFC3339 datetime", updateJobOptions.NextRunAt)
			}

			jobUpdate.NextRunAt = common.Undefinable[time.Time]{
				Value:   nextRunAtTime,
				Defined: true,
			}
		}

		svc := services.JobService{
			ApiUrl: ApiUrl,
		}

		if err := svc.Edit(updateJobOptions.Name, jobUpdate); err != nil {
			return fmt.Errorf("failed to update job: %s", err)
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
