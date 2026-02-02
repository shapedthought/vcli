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
	repoSnapshotAll  bool
	repoDiffAll      bool
	sobrSnapshotAll  bool
	sobrDiffAll      bool
	repoExportAll    bool
	repoExportOutput string
	sobrExportAll    bool
	sobrExportOutput string
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Repository management commands",
	Long: `Repository related commands for state management, drift detection, and export.

ONLY WORKS WITH VBR AT THE MOMENT.

Subcommands:

Snapshot repository configuration
  vcli repo snapshot "Default Backup Repository"
  vcli repo snapshot --all

Detect configuration drift
  vcli repo diff "Default Backup Repository"
  vcli repo diff --all

Export repository to declarative YAML
  vcli repo export "Default Backup Repository"
  vcli repo export --all -o /tmp/repos/

Scale-out backup repositories
  vcli repo sobr-snapshot "Scale-out Backup Repository 1"
  vcli repo sobr-snapshot --all
  vcli repo sobr-diff "Scale-out Backup Repository 1"
  vcli repo sobr-diff --all
  vcli repo sobr-export "Scale-out Backup Repository 1"
  vcli repo sobr-export --all -o /tmp/sobrs/
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
  3 - Drift detected (INFO or WARNING)
  4 - Critical security drift detected
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
	return saveResourceToState("VBRRepository", name, id, rawData)
}

func diffSingleRepo(repoName string) {
	loadSeverityOverrides()
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

	// Compare, classify, filter
	drifts := detectDrift(resource.Spec, currentMap, repoIgnoreFields)
	drifts = classifyDrifts(drifts, repoSeverityMap)
	minSev := parseSeverityFlag()
	drifts = filterDriftsBySeverity(drifts, minSev)

	if len(drifts) == 0 {
		fmt.Println(noDriftMessage("Repository matches snapshot state.", minSev))
		os.Exit(0)
	}

	// Display drift
	printSecuritySummary(drifts)
	fmt.Println("Drift detected:")
	for _, drift := range drifts {
		printDriftWithSeverity(drift)
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("  - %d drifts detected\n", len(drifts))
	fmt.Printf("  - Highest severity: %s\n", getMaxSeverity(drifts))
	fmt.Printf("  - Last snapshot: %s\n", resource.LastApplied.Format("2006-01-02 15:04:05"))
	fmt.Printf("  - Last snapshot by: %s\n", resource.LastAppliedBy)

	fmt.Println("\nThe repository has drifted from the snapshot configuration.")
	fmt.Printf("\nTo update the snapshot, run:\n")
	fmt.Printf("  vcli repo snapshot \"%s\"\n", repoName)

	os.Exit(exitCodeForDrifts(drifts))
}

func diffAllRepos() {
	loadSeverityOverrides()
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

	minSev := parseSeverityFlag()
	driftedCount := 0
	cleanCount := 0
	var allDrifts []Drift

	for _, resource := range resources {
		// Fetch current from VBR
		endpoint := fmt.Sprintf("backupInfrastructure/repositories/%s", resource.ID)
		currentRaw := vhttp.GetData[json.RawMessage](endpoint, profile)

		var currentMap map[string]interface{}
		if err := json.Unmarshal(currentRaw, &currentMap); err != nil {
			fmt.Printf("  %s: Failed to unmarshal repository data: %v\n", resource.Name, err)
			continue
		}

		// Detect, classify, filter
		drifts := detectDrift(resource.Spec, currentMap, repoIgnoreFields)
		drifts = classifyDrifts(drifts, repoSeverityMap)
		drifts = filterDriftsBySeverity(drifts, minSev)

		if len(drifts) > 0 {
			maxSev := getMaxSeverity(drifts)
			fmt.Printf("  %s %s: %d drifts detected\n", maxSev, resource.Name, len(drifts))
			allDrifts = append(allDrifts, drifts...)
			driftedCount++
		} else {
			fmt.Printf("  %s: No drift\n", resource.Name)
			cleanCount++
		}
	}

	if len(allDrifts) > 0 {
		printSecuritySummary(allDrifts)
	}
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  - %d repositories clean\n", cleanCount)
	fmt.Printf("  - %d repositories drifted\n", driftedCount)

	if driftedCount > 0 {
		os.Exit(exitCodeForDrifts(allDrifts))
	}
	os.Exit(0)
}

// --- Scale-Out Backup Repository commands ---

var sobrSnapshotCmd = &cobra.Command{
	Use:   "sobr-snapshot [sobr-name]",
	Short: "Snapshot scale-out backup repository configuration to state",
	Long: `Capture the current configuration of a scale-out backup repository and store it in state.

Examples:
  # Snapshot a single SOBR
  vcli repo sobr-snapshot "Scale-out Backup Repository 1"

  # Snapshot all SOBRs
  vcli repo sobr-snapshot --all
`,
	Run: func(cmd *cobra.Command, args []string) {
		if sobrSnapshotAll {
			snapshotAllSobrs()
		} else if len(args) > 0 {
			snapshotSingleSobr(args[0])
		} else {
			log.Fatal("Provide SOBR name or use --all")
		}
	},
}

var sobrDiffCmd = &cobra.Command{
	Use:   "sobr-diff [sobr-name]",
	Short: "Detect configuration drift for scale-out backup repositories",
	Long: `Compare current VBR scale-out backup repository configuration against the
last snapshot state to detect manual changes or drift.

Examples:
  # Check single SOBR for drift
  vcli repo sobr-diff "Scale-out Backup Repository 1"

  # Check all SOBRs
  vcli repo sobr-diff --all

Exit Codes:
  0 - No drift detected
  3 - Drift detected (INFO or WARNING)
  4 - Critical security drift detected
  1 - Error occurred`,
	Run: func(cmd *cobra.Command, args []string) {
		if sobrDiffAll {
			diffAllSobrs()
		} else if len(args) > 0 {
			diffSingleSobr(args[0])
		} else {
			log.Fatal("Provide SOBR name or use --all")
		}
	},
}

func snapshotSingleSobr(sobrName string) {
	profile := utils.GetProfile()

	if profile.Name != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	sobrList := vhttp.GetData[models.VbrSobrList]("backupInfrastructure/scaleOutRepositories", profile)

	var found *models.VbrSobrGet
	for i := range sobrList.Data {
		if sobrList.Data[i].Name == sobrName {
			found = &sobrList.Data[i]
			break
		}
	}

	if found == nil {
		log.Fatalf("Scale-out repository '%s' not found in VBR.", sobrName)
	}

	endpoint := fmt.Sprintf("backupInfrastructure/scaleOutRepositories/%s", found.ID)
	sobrData := vhttp.GetData[json.RawMessage](endpoint, profile)

	if err := saveResourceToState("VBRScaleOutRepository", sobrName, found.ID, sobrData); err != nil {
		log.Fatalf("Failed to save SOBR state: %v", err)
	}

	fmt.Printf("Snapshot saved for scale-out repository: %s\n", sobrName)
}

func snapshotAllSobrs() {
	profile := utils.GetProfile()

	if profile.Name != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	sobrList := vhttp.GetData[models.VbrSobrList]("backupInfrastructure/scaleOutRepositories", profile)

	if len(sobrList.Data) == 0 {
		fmt.Println("No scale-out repositories found.")
		return
	}

	fmt.Printf("Snapshotting %d scale-out repositories...\n", len(sobrList.Data))

	for _, sobr := range sobrList.Data {
		endpoint := fmt.Sprintf("backupInfrastructure/scaleOutRepositories/%s", sobr.ID)
		sobrData := vhttp.GetData[json.RawMessage](endpoint, profile)

		if err := saveResourceToState("VBRScaleOutRepository", sobr.Name, sobr.ID, sobrData); err != nil {
			fmt.Printf("Warning: Failed to save state for '%s': %v\n", sobr.Name, err)
			continue
		}

		fmt.Printf("  Snapshot saved: %s\n", sobr.Name)
	}

	stateMgr := state.NewManager()
	fmt.Printf("\nState updated: %s\n", stateMgr.GetStatePath())
}

func diffSingleSobr(sobrName string) {
	loadSeverityOverrides()
	profile := utils.GetProfile()

	if profile.Name != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	stateMgr := state.NewManager()
	resource, err := stateMgr.GetResource(sobrName)
	if err != nil {
		log.Fatalf("Scale-out repository '%s' not found in state. Has it been snapshotted?\n", sobrName)
	}

	if resource.Type != "VBRScaleOutRepository" {
		log.Fatalf("Resource '%s' is not a scale-out repository (type: %s).\n", sobrName, resource.Type)
	}

	fmt.Printf("Checking drift for scale-out repository: %s\n\n", sobrName)

	endpoint := fmt.Sprintf("backupInfrastructure/scaleOutRepositories/%s", resource.ID)
	currentRaw := vhttp.GetData[json.RawMessage](endpoint, profile)

	var currentMap map[string]interface{}
	if err := json.Unmarshal(currentRaw, &currentMap); err != nil {
		log.Fatalf("Failed to unmarshal current SOBR data: %v", err)
	}

	// Compare, classify, filter
	drifts := detectDrift(resource.Spec, currentMap, sobrIgnoreFields)
	drifts = classifyDrifts(drifts, sobrSeverityMap)
	minSev := parseSeverityFlag()
	drifts = filterDriftsBySeverity(drifts, minSev)

	if len(drifts) == 0 {
		fmt.Println(noDriftMessage("Scale-out repository matches snapshot state.", minSev))
		os.Exit(0)
	}

	printSecuritySummary(drifts)
	fmt.Println("Drift detected:")
	for _, drift := range drifts {
		printDriftWithSeverity(drift)
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("  - %d drifts detected\n", len(drifts))
	fmt.Printf("  - Highest severity: %s\n", getMaxSeverity(drifts))
	fmt.Printf("  - Last snapshot: %s\n", resource.LastApplied.Format("2006-01-02 15:04:05"))
	fmt.Printf("  - Last snapshot by: %s\n", resource.LastAppliedBy)

	fmt.Println("\nThe scale-out repository has drifted from the snapshot configuration.")
	fmt.Printf("\nTo update the snapshot, run:\n")
	fmt.Printf("  vcli repo sobr-snapshot \"%s\"\n", sobrName)

	os.Exit(exitCodeForDrifts(drifts))
}

func diffAllSobrs() {
	loadSeverityOverrides()
	profile := utils.GetProfile()

	if profile.Name != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	stateMgr := state.NewManager()
	resources, err := stateMgr.ListResources("VBRScaleOutRepository")
	if err != nil {
		log.Fatalf("Failed to load state: %v\n", err)
	}

	if len(resources) == 0 {
		fmt.Println("No scale-out repositories in state.")
		return
	}

	fmt.Printf("Checking %d scale-out repositories for drift...\n\n", len(resources))

	minSev := parseSeverityFlag()
	driftedCount := 0
	cleanCount := 0
	var allDrifts []Drift

	for _, resource := range resources {
		endpoint := fmt.Sprintf("backupInfrastructure/scaleOutRepositories/%s", resource.ID)
		currentRaw := vhttp.GetData[json.RawMessage](endpoint, profile)

		var currentMap map[string]interface{}
		if err := json.Unmarshal(currentRaw, &currentMap); err != nil {
			fmt.Printf("  %s: Failed to unmarshal SOBR data: %v\n", resource.Name, err)
			continue
		}

		// Detect, classify, filter
		drifts := detectDrift(resource.Spec, currentMap, sobrIgnoreFields)
		drifts = classifyDrifts(drifts, sobrSeverityMap)
		drifts = filterDriftsBySeverity(drifts, minSev)

		if len(drifts) > 0 {
			maxSev := getMaxSeverity(drifts)
			fmt.Printf("  %s %s: %d drifts detected\n", maxSev, resource.Name, len(drifts))
			allDrifts = append(allDrifts, drifts...)
			driftedCount++
		} else {
			fmt.Printf("  %s: No drift\n", resource.Name)
			cleanCount++
		}
	}

	if len(allDrifts) > 0 {
		printSecuritySummary(allDrifts)
	}
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  - %d scale-out repositories clean\n", cleanCount)
	fmt.Printf("  - %d scale-out repositories drifted\n", driftedCount)

	if driftedCount > 0 {
		os.Exit(exitCodeForDrifts(allDrifts))
	}
	os.Exit(0)
}

// --- Repository Export commands ---

var repoExportCmd = &cobra.Command{
	Use:   "export [repo-name]",
	Short: "Export repository configuration to declarative YAML",
	Long: `Export a backup repository configuration to declarative YAML format.

Examples:
  # Export single repository to stdout
  vcli repo export "Default Backup Repository"

  # Export single repository to file
  vcli repo export "Default Backup Repository" -o repo.yaml

  # Export all repositories to current directory
  vcli repo export --all

  # Export all repositories to specific directory
  vcli repo export --all -o /tmp/repos/
`,
	Run: func(cmd *cobra.Command, args []string) {
		profile := utils.GetProfile()
		if profile.Name != "vbr" {
			log.Fatal("This command only works with VBR at the moment.")
		}

		if repoExportAll {
			exportAllRepos(profile)
		} else if len(args) > 0 {
			exportSingleRepo(args[0], profile)
		} else {
			log.Fatal("Provide repository name or use --all")
		}
	},
}

func exportSingleRepo(repoName string, profile models.Profile) {
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

	endpoint := fmt.Sprintf("backupInfrastructure/repositories/%s", found.ID)
	repoData := vhttp.GetData[json.RawMessage](endpoint, profile)

	cfg := ResourceExportConfig{
		Kind:         "VBRRepository",
		IgnoreFields: repoIgnoreFields,
		HeaderLines: []string{
			"# VBR Repository Configuration (Full Export)",
			"# Exported from VBR",
		},
	}

	yamlContent, err := convertResourceToYAML(repoName, repoData, cfg)
	if err != nil {
		log.Fatalf("Failed to convert repository to YAML: %v", err)
	}

	writeExportOutput(yamlContent, repoExportOutput, repoName)
}

func exportAllRepos(profile models.Profile) {
	repoList := vhttp.GetData[models.VbrRepoList]("backupInfrastructure/repositories", profile)

	if len(repoList.Data) == 0 {
		fmt.Println("No repositories found.")
		return
	}

	outputDir := repoExportOutput
	if outputDir == "" {
		outputDir = "."
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	fmt.Printf("Exporting %d repositories...\n", len(repoList.Data))

	successCount := 0
	failedCount := 0

	cfg := ResourceExportConfig{
		Kind:         "VBRRepository",
		IgnoreFields: repoIgnoreFields,
		HeaderLines: []string{
			"# VBR Repository Configuration (Full Export)",
			"# Exported from VBR",
		},
	}

	for i, repo := range repoList.Data {
		endpoint := fmt.Sprintf("backupInfrastructure/repositories/%s", repo.ID)
		repoData := vhttp.GetData[json.RawMessage](endpoint, profile)

		yamlContent, err := convertResourceToYAML(repo.Name, repoData, cfg)
		if err != nil {
			fmt.Printf("Warning: Failed to convert repository %s: %v\n", repo.Name, err)
			failedCount++
			continue
		}

		if err := writeExportAllOutput(yamlContent, outputDir, repo.Name, i, len(repoList.Data)); err != nil {
			fmt.Printf("Warning: %v\n", err)
			failedCount++
			continue
		}

		successCount++
	}

	fmt.Printf("\nExport complete: %d successful, %d failed\n", successCount, failedCount)
}

// --- SOBR Export commands ---

var sobrExportCmd = &cobra.Command{
	Use:   "sobr-export [sobr-name]",
	Short: "Export scale-out backup repository to declarative YAML",
	Long: `Export a scale-out backup repository configuration to declarative YAML format.

Examples:
  # Export single SOBR to stdout
  vcli repo sobr-export "Scale-out Backup Repository 1"

  # Export single SOBR to file
  vcli repo sobr-export "Scale-out Backup Repository 1" -o sobr.yaml

  # Export all SOBRs to current directory
  vcli repo sobr-export --all

  # Export all SOBRs to specific directory
  vcli repo sobr-export --all -o /tmp/sobrs/
`,
	Run: func(cmd *cobra.Command, args []string) {
		profile := utils.GetProfile()
		if profile.Name != "vbr" {
			log.Fatal("This command only works with VBR at the moment.")
		}

		if sobrExportAll {
			exportAllSobrs(profile)
		} else if len(args) > 0 {
			exportSingleSobr(args[0], profile)
		} else {
			log.Fatal("Provide SOBR name or use --all")
		}
	},
}

func exportSingleSobr(sobrName string, profile models.Profile) {
	sobrList := vhttp.GetData[models.VbrSobrList]("backupInfrastructure/scaleOutRepositories", profile)

	var found *models.VbrSobrGet
	for i := range sobrList.Data {
		if sobrList.Data[i].Name == sobrName {
			found = &sobrList.Data[i]
			break
		}
	}

	if found == nil {
		log.Fatalf("Scale-out repository '%s' not found in VBR.", sobrName)
	}

	endpoint := fmt.Sprintf("backupInfrastructure/scaleOutRepositories/%s", found.ID)
	sobrData := vhttp.GetData[json.RawMessage](endpoint, profile)

	cfg := ResourceExportConfig{
		Kind:         "VBRScaleOutRepository",
		IgnoreFields: sobrIgnoreFields,
		HeaderLines: []string{
			"# VBR Scale-Out Repository Configuration (Full Export)",
			"# Exported from VBR",
		},
	}

	yamlContent, err := convertResourceToYAML(sobrName, sobrData, cfg)
	if err != nil {
		log.Fatalf("Failed to convert scale-out repository to YAML: %v", err)
	}

	writeExportOutput(yamlContent, sobrExportOutput, sobrName)
}

func exportAllSobrs(profile models.Profile) {
	sobrList := vhttp.GetData[models.VbrSobrList]("backupInfrastructure/scaleOutRepositories", profile)

	if len(sobrList.Data) == 0 {
		fmt.Println("No scale-out repositories found.")
		return
	}

	outputDir := sobrExportOutput
	if outputDir == "" {
		outputDir = "."
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	fmt.Printf("Exporting %d scale-out repositories...\n", len(sobrList.Data))

	successCount := 0
	failedCount := 0

	cfg := ResourceExportConfig{
		Kind:         "VBRScaleOutRepository",
		IgnoreFields: sobrIgnoreFields,
		HeaderLines: []string{
			"# VBR Scale-Out Repository Configuration (Full Export)",
			"# Exported from VBR",
		},
	}

	for i, sobr := range sobrList.Data {
		endpoint := fmt.Sprintf("backupInfrastructure/scaleOutRepositories/%s", sobr.ID)
		sobrData := vhttp.GetData[json.RawMessage](endpoint, profile)

		yamlContent, err := convertResourceToYAML(sobr.Name, sobrData, cfg)
		if err != nil {
			fmt.Printf("Warning: Failed to convert scale-out repository %s: %v\n", sobr.Name, err)
			failedCount++
			continue
		}

		if err := writeExportAllOutput(yamlContent, outputDir, sobr.Name, i, len(sobrList.Data)); err != nil {
			fmt.Printf("Warning: %v\n", err)
			failedCount++
			continue
		}

		successCount++
	}

	fmt.Printf("\nExport complete: %d successful, %d failed\n", successCount, failedCount)
}

// saveResourceToState is a shared helper for saving any resource type to state
func saveResourceToState(resourceType, name, id string, rawData json.RawMessage) error {
	var spec map[string]interface{}
	if err := json.Unmarshal(rawData, &spec); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	stateMgr := state.NewManager()

	currentUser := "unknown"
	if usr, err := user.Current(); err == nil {
		currentUser = usr.Username
	}

	resource := &state.Resource{
		Type:          resourceType,
		ID:            id,
		Name:          name,
		LastApplied:   time.Now(),
		LastAppliedBy: currentUser,
		Origin:        "observed",
		Spec:          spec,
	}

	if err := stateMgr.UpdateResource(resource); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	return nil
}

func init() {
	repoSnapshotCmd.Flags().BoolVar(&repoSnapshotAll, "all", false, "Snapshot all repositories")
	repoDiffCmd.Flags().BoolVar(&repoDiffAll, "all", false, "Check drift for all repositories in state")
	addSeverityFlags(repoDiffCmd)
	sobrSnapshotCmd.Flags().BoolVar(&sobrSnapshotAll, "all", false, "Snapshot all scale-out repositories")
	sobrDiffCmd.Flags().BoolVar(&sobrDiffAll, "all", false, "Check drift for all scale-out repositories in state")
	addSeverityFlags(sobrDiffCmd)

	repoExportCmd.Flags().BoolVar(&repoExportAll, "all", false, "Export all repositories")
	repoExportCmd.Flags().StringVarP(&repoExportOutput, "output", "o", "", "Output file or directory (default: stdout)")
	sobrExportCmd.Flags().BoolVar(&sobrExportAll, "all", false, "Export all scale-out repositories")
	sobrExportCmd.Flags().StringVarP(&sobrExportOutput, "output", "o", "", "Output file or directory (default: stdout)")

	repoCmd.AddCommand(repoSnapshotCmd)
	repoCmd.AddCommand(repoDiffCmd)
	repoCmd.AddCommand(repoExportCmd)
	repoCmd.AddCommand(sobrSnapshotCmd)
	repoCmd.AddCommand(sobrDiffCmd)
	repoCmd.AddCommand(sobrExportCmd)
	rootCmd.AddCommand(repoCmd)
}
