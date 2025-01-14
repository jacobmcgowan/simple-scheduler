package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jacobmcgowan/simple-scheduler/services/cli/cmd/options"
	"github.com/jacobmcgowan/simple-scheduler/services/cli/services"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
	"github.com/jacobmcgowan/simple-scheduler/shared/runStatuses"
	"github.com/jacobmcgowan/simple-scheduler/shared/validators"
	"github.com/spf13/cobra"
)

var listRunsOptions = options.RunFilterOptions{}

var statusChoices = fmt.Sprintf("%s|%s|%s|%s|%s|%s",
	runStatuses.Pending,
	runStatuses.Running,
	runStatuses.Cancelling,
	runStatuses.Cancelled,
	runStatuses.Failed,
	runStatuses.Completed)

var runsCmd = &cobra.Command{
	Use:     "runs",
	Aliases: []string{"r"},
	Short:   "Lists runs",
	Long: `Provides details on the runs for the current jobs that are
scheduled.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !validators.ValidateRunStatus(listRunsOptions.Status, true) {
			fmt.Printf("Invalid run status. Acceptable statuses are %s.", statusChoices)
			return
		}

		filter := dtos.RunFilter{}
		if listRunsOptions.JobName != "" {
			filter.JobName = &listRunsOptions.JobName
		}
		if listRunsOptions.Status != "" {
			runStatus := runStatuses.RunStatus(listRunsOptions.Status)
			filter.Status = &runStatus
		}

		svc := services.RunService{
			ApiUrl: ApiUrl,
		}

		if runs, err := svc.Browse(filter); err == nil {
			writer := tabwriter.NewWriter(os.Stdout, 1, 1, 4, ' ', 0)
			fmt.Fprintln(writer, "ID\tJOB\tSTATUS\tSTART TIME\tEND TIME")

			for _, run := range runs {
				fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n", run.Id, run.JobName, run.Status, run.StartTime, run.EndTime)
			}

			writer.Flush()
		} else {
			fmt.Printf("Failed to get runs: %s\n", err)
		}
	},
}

func init() {
	listCmd.AddCommand(runsCmd)
	runsCmd.Flags().StringVarP(&listRunsOptions.JobName, "job", "j", "", "The job to list the runs for.")
	runsCmd.Flags().StringVarP(&listRunsOptions.Status, "status", "s", "", fmt.Sprintf("The status of the runs to list (%s).", statusChoices))
}
