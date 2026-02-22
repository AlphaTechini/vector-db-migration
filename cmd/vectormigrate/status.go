package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	statusCmd = &cobra.Command{
		Use:   "status [migration-id]",
		Short: "Get migration status",
		Long:  "Retrieve the current status and progress of a migration.",
		Args:  cobra.ExactArgs(1),
		RunE:  runStatus,
	}
)

func runStatus(cmd *cobra.Command, args []string) error {
	migrationID := args[0]

	// TODO: Get status from orchestrator
	// For now, just show placeholder
	fmt.Printf("Migration: %s\n", migrationID)
	fmt.Printf("Status: not_started\n")
	fmt.Printf("Progress: 0/0 records (0%%)\n")
	fmt.Printf("Batches: 0 processed\n")

	return nil
}
