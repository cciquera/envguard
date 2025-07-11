package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan your environment for misconfigurations",
	RunE: func(cmd *cobra.Command, args []string) error {
		tfDir := viper.GetString("terraform.dir")
		if tfDir == "" {
			tfDir, _ = cmd.Flags().GetString("tf-dir") // fallback to CLI flag
		}

		jsonOutput := false
		outputFormat := viper.GetString("output.format")
		if outputFormat == "json" {
			jsonOutput = true
		}

		fmt.Println("🔍 Scanning Terraform for drift...")

		results, err := runTerraformDriftCheck(tfDir)
		if err != nil {
			return err // exit code 1 on error
		}

		if jsonOutput {
			output, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to encode JSON output: %w", err)
			}
			fmt.Println(string(output))
		} else {
			if len(results) == 0 {
				fmt.Println("✅ No unmanaged drift detected.")
			} else {
				fmt.Println("⚠️ Drift summary:")
				for _, res := range results {
					fmt.Printf("- [%s] %s: %s\n", res.Severity, res.Resource, res.Message)
				}
			}
		}

		failOn := viper.GetString("output.failOnSeverity")
		highest := highestSeverity(results)

		if severityLevel(highest) >= severityLevel(failOn) {
			os.Exit(severityToExitCode(highest))
		}

		// Determine highest severity to set exit code
		exitCode := severityToExitCode(highestSeverity(results))
		os.Exit(exitCode)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().String("tf-dir", ".", "Path to Terraform code")
	scanCmd.Flags().Bool("json", false, "Output results as JSON")
}

func runTerraformDriftCheck(tfDir string) ([]ScanResult, error) {
	results := make([]ScanResult, 0)

	absPath, err := filepath.Abs(tfDir)
	if err != nil {
		return results, err
	}

	// 1. terraform init
	fmt.Println("→ Running terraform init...")
	cmd := exec.Command("terraform", "init", "-input=false")
	cmd.Dir = absPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return results, fmt.Errorf("terraform init failed: %w", err)
	}

	// 2. terraform plan
	fmt.Println("→ Running terraform plan...")
	planPath := filepath.Join(absPath, "plan.tfplan")
	cmd = exec.Command("terraform", "plan", "-out=plan.tfplan", "-input=false")
	cmd.Dir = absPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return results, fmt.Errorf("terraform plan failed: %w", err)
	}

	// 3. terraform show -json
	fmt.Println("→ Analyzing plan for drift...")
	cmd = exec.Command("terraform", "show", "-json", planPath)
	cmd.Dir = absPath
	output, err := cmd.Output()
	if err != nil {
		return results, fmt.Errorf("terraform show failed: %w", err)
	}

	var plan tfjson.Plan
	if err := json.Unmarshal(output, &plan); err != nil {
		return results, fmt.Errorf("failed to parse plan JSON: %w", err)
	}

	// 4. Collect drift results
	for _, rc := range plan.ResourceChanges {
		actions := rc.Change.Actions
		resName := fmt.Sprintf("%s.%s", rc.Type, rc.Name)

		if contains(actions, tfjson.ActionUpdate) {
			results = append(results, ScanResult{
				Source:     "terraform",
				Resource:   resName,
				ChangeType: "update",
				Severity:   "warning",
				Message:    "Terraform resource drift detected (update)",
			})
		}
		if contains(actions, tfjson.ActionDelete) {
			results = append(results, ScanResult{
				Source:     "terraform",
				Resource:   resName,
				ChangeType: "delete",
				Severity:   "critical",
				Message:    "Resource marked for deletion due to drift",
			})
		}
		if contains(actions, tfjson.ActionCreate) {
			results = append(results, ScanResult{
				Source:     "terraform",
				Resource:   resName,
				ChangeType: "create",
				Severity:   "warning",
				Message:    "Terraform plans to create missing resource",
			})
		}
	}

	return results, nil
}

func contains(actions tfjson.Actions, target tfjson.Action) bool {
	for _, a := range actions {
		if a == target {
			return true
		}
	}
	return false
}

func highestSeverity(results []ScanResult) string {
	levels := map[string]int{"info": 1, "warning": 2, "critical": 3}
	highest := "info"
	for _, r := range results {
		if levels[r.Severity] > levels[highest] {
			highest = r.Severity
		}
	}
	return highest
}

func severityToExitCode(severity string) int {
	switch severity {
	case "critical":
		return 3
	case "warning":
		return 2
	default:
		return 0
	}
}

func severityLevel(s string) int {
	levels := map[string]int{"info": 1, "warning": 2, "critical": 3}
	return levels[s]
}
