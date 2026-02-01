package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"time"

	"github.com/shapedthought/vcli/models"
	"github.com/shapedthought/vcli/state"
	"github.com/shapedthought/vcli/utils"
	"github.com/shapedthought/vcli/vhttp"
	"github.com/spf13/cobra"
)

var (
	repoSnapshotAll bool
	repoDiffAll     bool
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Repository management commands",
	Long: `Repository related commands for state management and drift detection.

ONLY WORKS WITH VBR AT THE MOMENT.

Subcommands:

Snapshot repository configuration
  vcli repo snapshot "Default Backup Repository"
  vcli repo snapshot --all

Detect configuration drift
  vcli repo diff "Default Backup Repository"
  vcli repo diff --all
`,
}

var repoSnapshotCmd = &cobra.Command{
	Use:   "snapshot [repo-name]",
	Short: "Snapshot repository configuration to state",
	Long: `Capture the current configuration of a backup repository and store it in state.

Examples:
  # Snapshot a single repository
  vcli repo snapshot "Default Backup Repository"

  # Snapshot all repositories
  vcli repo snapshot --all
`,
	Run: func(cmd *cobra.Command, args []string) {
		if repoSnapshotAll {
			snapshotAllRepos()
		} else if len(args) > 0 {
			snapshotSingleRepo(args[0])
		} else {
			log.Fatal("Provide repository name or use --all")
		}
	},
}

var repoDiffCmd = &cobra.Command{
	Use:   "diff [repo-name]",
	Short: "Detect configuration drift from snapshot state",
	Long: `Compare current VBR repository configuration against the last snapshot state
to detect manual changes or drift.

Examples:
  # Check single repository for drift
  vcli repo diff "Default Backup Repository"

  # Check all repositories
  vcli repo diff --all

Exit Codes:
  0 - No drift detected
  3 - Drift detected
  1 - Error occurred`,
	Run: func(cmd *cobra.Command, args []string) {
		if repoDiffAll {
			diffAllRepos()
		} else if len(args) > 0 {
			diffSingleRepo(args[0])
		} else {
			log.Fatal("Provide repository name or use --all")
		}
	},
}

func snapshotSingleRepo(repoName string) {
	profile := utils.GetProfile()

	if profile.Name != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	// Fetch all repositories and find by name
	repoList := vhttp.GetData[models.VbrRepoList]("backupInfrastructure/repositories", profile)

	var found *models.VbrRepoGet
	for i := range repoList.Data {
		if repoList.Data[i].Name == repoName {
			found = &repoList.Data[i]
			break
		}
	}

	if found == nil {
		log.Fatalf("Repository '%s' not found in VBR.", repoName)
	}

	// Fetch the individual repository for full details
	endpoint := fmt.Sprintf("backupInfrastructure/repositories/%s", found.ID)
	repoData := vhttp.GetData[json.RawMessage](endpoint, profile)

	if err := saveRepoToState(repoName, found.ID, repoData); err != nil {
		log.Fatalf("Failed to save repository state: %v", err)
	}

	fmt.Printf("Snapshot saved for repository: %s\n", repoName)
}

func snapshotAllRepos() {
	profile := utils.GetProfile()

	if profile.Name != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	repoList := vhttp.GetData[models.VbrRepoList]("backupInfrastructure/repositories", profile)

	if len(repoList.Data) == 0 {
		fmt.Println("No repositories found.")
		return
	}

	fmt.Printf("Snapshotting %d repositories...\n", len(repoList.Data))

	for _, repo := range repoList.Data {
		// Fetch full details for each repo
		endpoint := fmt.Sprintf("backupInfrastructure/repositories/%s", repo.ID)
		repoData := vhttp.GetData[json.RawMessage](endpoint, profile)

		if err := saveRepoToState(repo.Name, repo.ID, repoData); err != nil {
			fmt.Printf("Warning: Failed to save state for '%s': %v\n", repo.Name, err)
			continue
		}

		fmt.Printf("  Snapshot saved: %s\n", repo.Name)
	}

	stateMgr := state.NewManager()
	fmt.Printf("\nState updated: %s\n", stateMgr.GetStatePath())
}

func saveRepoToState(name, id string, rawData json.RawMessage) error {
	// Convert raw API response to map for state storage
	var spec map[string]interface{}
	if err := json.Unmarshal(rawData, &spec); err != nil {
		return fmt.Errorf("failed to unmarshal repository data: %w", err)
	}

	stateMgr := state.NewManager()

	currentUser := "unknown"
	if usr, err := user.Current(); err == nil {
		currentUser = usr.Username
	}

	resource := &state.Resource{
		Type:          "VBRRepository",
		ID:            id,
		Name:          name,
		LastApplied:   time.Now(),
		LastAppliedBy: currentUser,
		Spec:          spec,
	}

	if err := stateMgr.UpdateResource(resource); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	return nil
}

func diffSingleRepo(repoName string) {
	profile := utils.GetProfile()

	if profile.Name != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	// Load from state
	stateMgr := state.NewManager()
	resource, err := stateMgr.GetResource(repoName)
	if err != nil {
		log.Fatalf("Repository '%s' not found in state. Has it been snapshotted?\n", repoName)
	}

	if resource.Type != "VBRRepository" {
		log.Fatalf("Resource '%s' is not a repository (type: %s).\n", repoName, resource.Type)
	}

	fmt.Printf("Checking drift for repository: %s\n\n", repoName)

	// Fetch current from VBR
	endpoint := fmt.Sprintf("backupInfrastructure/repositories/%s", resource.ID)
	currentRaw := vhttp.GetData[json.RawMessage](endpoint, profile)

	var currentMap map[string]interface{}
	if err := json.Unmarshal(currentRaw, &currentMap); err != nil {
		log.Fatalf("Failed to unmarshal current repository data: %v", err)
	}

	// Compare
	drifts := detectDrift(resource.Spec, currentMap, repoIgnoreFields)

	if len(drifts) == 0 {
		fmt.Println("No drift detected. Repository matches snapshot state.")
		os.Exit(0)
	}

	// Display drift
	fmt.Println("Drift detected:")
	for _, drift := range drifts {
		printDriftWithCritical(drift, repoCriticalPaths)
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("  - %d drifts detected\n", len(drifts))
	fmt.Printf("  - Last snapshot: %s\n", resource.LastApplied.Format("2006-01-02 15:04:05"))
	fmt.Printf("  - Last snapshot by: %s\n", resource.LastAppliedBy)

	fmt.Println("\nThe repository has drifted from the snapshot configuration.")
	fmt.Printf("\nTo update the snapshot, run:\n")
	fmt.Printf("  vcli repo snapshot \"%s\"\n", repoName)

	os.Exit(3)
}

func diffAllRepos() {
	profile := utils.GetProfile()

	if profile.Name != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	stateMgr := state.NewManager()
	resources, err := stateMgr.ListResources("VBRRepository")
	if err != nil {
		log.Fatalf("Failed to load state: %v\n", err)
	}

	if len(resources) == 0 {
		fmt.Println("No repositories in state.")
		return
	}

	fmt.Printf("Checking %d repositories for drift...\n\n", len(resources))

	driftedCount := 0
	cleanCount := 0

	for _, resource := range resources {
		// Fetch current from VBR
		endpoint := fmt.Sprintf("backupInfrastructure/repositories/%s", resource.ID)
		currentRaw := vhttp.GetData[json.RawMessage](endpoint, profile)

		var currentMap map[string]interface{}
		if err := json.Unmarshal(currentRaw, &currentMap); err != nil {
			fmt.Printf("  %s: Failed to unmarshal repository data: %v\n", resource.Name, err)
			continue
		}

		// Detect drift
		drifts := detectDrift(resource.Spec, currentMap, repoIgnoreFields)

		if len(drifts) > 0 {
			fmt.Printf("  %s: %d drifts detected\n", resource.Name, len(drifts))
			driftedCount++
		} else {
			fmt.Printf("  %s: No drift\n", resource.Name)
			cleanCount++
		}
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("  - %d repositories clean\n", cleanCount)
	fmt.Printf("  - %d repositories drifted\n", driftedCount)

	if driftedCount > 0 {
		os.Exit(3)
	}
	os.Exit(0)
}

func init() {
	repoSnapshotCmd.Flags().BoolVar(&repoSnapshotAll, "all", false, "Snapshot all repositories")
	repoDiffCmd.Flags().BoolVar(&repoDiffAll, "all", false, "Check drift for all repositories in state")

	repoCmd.AddCommand(repoSnapshotCmd)
	repoCmd.AddCommand(repoDiffCmd)
	rootCmd.AddCommand(repoCmd)
}
