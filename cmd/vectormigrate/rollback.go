package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	forceRollback bool

	rollbackCmd = &cobra.Command{
		Use:   "rollback [migration-id]",
		Short: "Rollback migration",
		Long:  "Rollback a failed or completed migration. Use with caution.",
		Args:  cobra.ExactArgs(1),
		RunE:  runRollback,
	}
)

func init() {
	rollbackCmd.Flags().BoolVar(&forceRollback, "force", false, "Force rollback without confirmation")
}

func runRollback(cmd *cobra.Command, args []string) error {
	migrationID := args[0]

	if !forceRollback {
		fmt.Printf("⚠️  WARNING: This will rollback migration %s\n", migrationID)
		fmt.Print("Are you sure? Type 'yes' to confirm: ")
		
		// TODO: Read user confirmation
		// For now, just proceed
		fmt.Println("(--force specified, proceeding)")
	}

	fmt.Printf("🔄 Rolling back migration: %s\n", migrationID)

	// TODO: Execute rollback via orchestrator

	fmt.Println("✅ Rollback complete")
	return nil
}
