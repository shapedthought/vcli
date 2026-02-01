package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/shapedthought/vcli/models"
	"github.com/shapedthought/vcli/state"
	"github.com/shapedthought/vcli/utils"
	"github.com/shapedthought/vcli/vhttp"
	"github.com/spf13/cobra"
)

var (
	encSnapshotAll bool
	encDiffAll     bool
	kmsSnapshotAll bool
	kmsDiffAll     bool
)

var encryptionCmd = &cobra.Command{
	Use:   "encryption",
	Short: "Encryption password and KMS server management commands",
	Long: `Encryption related commands for state management and drift detection.

ONLY WORKS WITH VBR AT THE MOMENT.

Note: Only password metadata (IDs, hints) is tracked — never actual password values.

Subcommands:

Snapshot encryption password inventory
  vcli encryption snapshot "My backup password"
  vcli encryption snapshot --all

Detect encryption password drift
  vcli encryption diff "My backup password"
  vcli encryption diff --all

KMS server management
  vcli encryption kms-snapshot "My KMS Server"
  vcli encryption kms-snapshot --all
  vcli encryption kms-diff "My KMS Server"
  vcli encryption kms-diff --all
`,
}

// --- Encryption Password commands ---

var encSnapshotCmd = &cobra.Command{
	Use:   "snapshot [password-hint]",
	Short: "Snapshot encryption password metadata to state",
	Long: `Capture the current encryption password metadata and store it in state.
Only metadata (ID, hint, import status) is captured — never actual password values.

Examples:
  # Snapshot a single encryption password by hint
  vcli encryption snapshot "My backup password"

  # Snapshot all encryption passwords
  vcli encryption snapshot --all
`,
	Run: func(cmd *cobra.Command, args []string) {
		if encSnapshotAll {
			snapshotAllEncryptionPasswords()
		} else if len(args) > 0 {
			snapshotSingleEncryptionPassword(args[0])
		} else {
			log.Fatal("Provide password hint or use --all")
		}
	},
}

var encDiffCmd = &cobra.Command{
	Use:   "diff [password-hint]",
	Short: "Detect encryption password inventory drift",
	Long: `Compare current VBR encryption password inventory against the last snapshot
to detect removed, added, or modified password records.

When using --all, performs inventory-level checks:
  - Detects passwords removed from VBR (CRITICAL)
  - Detects passwords added since last snapshot
  - Detects metadata changes on existing passwords

Examples:
  # Check single password for drift
  vcli encryption diff "My backup password"

  # Check entire encryption inventory
  vcli encryption diff --all

Exit Codes:
  0 - No drift detected
  3 - Drift detected (INFO or WARNING)
  4 - Critical security drift detected
  1 - Error occurred`,
	Run: func(cmd *cobra.Command, args []string) {
		if encDiffAll {
			diffAllEncryptionPasswords()
		} else if len(args) > 0 {
			diffSingleEncryptionPassword(args[0])
		} else {
			log.Fatal("Provide password hint or use --all")
		}
	},
}

func snapshotSingleEncryptionPassword(hint string) {
	profile := utils.GetProfile()

	if profile.Name != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	// Fetch all passwords and find by hint
	passwordList := vhttp.GetData[models.VbrEncryptionPasswordList]("encryptionPasswords", profile)

	var found *models.VbrEncryptionPasswordGet
	for i := range passwordList.Data {
		if passwordList.Data[i].Hint == hint {
			found = &passwordList.Data[i]
			break
		}
	}

	if found == nil {
		log.Fatalf("Encryption password with hint '%s' not found in VBR.", hint)
	}

	if err := saveEncryptionPasswordToState(found, nil); err != nil {
		log.Fatalf("Failed to save encryption password state: %v", err)
	}

	fmt.Printf("Snapshot saved for encryption password: %s\n", hint)
}

func snapshotAllEncryptionPasswords() {
	profile := utils.GetProfile()

	if profile.Name != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	passwordList := vhttp.GetData[models.VbrEncryptionPasswordList]("encryptionPasswords", profile)

	if len(passwordList.Data) == 0 {
		fmt.Println("No encryption passwords found.")
		return
	}

	fmt.Printf("Snapshotting %d encryption passwords...\n", len(passwordList.Data))

	// Build hint counts to detect duplicates
	hintCounts := make(map[string]int, len(passwordList.Data))
	for _, p := range passwordList.Data {
		hintCounts[p.Hint]++
	}

	for i := range passwordList.Data {
		p := &passwordList.Data[i]
		if err := saveEncryptionPasswordToState(p, hintCounts); err != nil {
			fmt.Printf("Warning: Failed to save state for '%s': %v\n", p.Hint, err)
			continue
		}

		displayName := p.Hint
		if hintCounts[p.Hint] > 1 {
			displayName = fmt.Sprintf("%s-%s", p.Hint, p.ID)
		}
		fmt.Printf("  Snapshot saved: %s\n", displayName)
	}

	stateMgr := state.NewManager()
	fmt.Printf("\nState updated: %s\n", stateMgr.GetStatePath())
}

func saveEncryptionPasswordToState(p *models.VbrEncryptionPasswordGet, hintCounts map[string]int) error {
	// Marshal to JSON then to map for state storage
	pBytes, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal password data: %w", err)
	}

	name := p.Hint
	if name == "" {
		name = p.ID
	} else if hintCounts != nil && hintCounts[p.Hint] > 1 {
		// Hints are not guaranteed unique; append ID to avoid overwriting
		name = fmt.Sprintf("%s-%s", name, p.ID)
	}

	return saveResourceToState("VBREncryptionPassword", name, p.ID, json.RawMessage(pBytes))
}

func diffSingleEncryptionPassword(hint string) {
	loadSeverityOverrides()
	profile := utils.GetProfile()

	if profile.Name != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	// Load from state
	stateMgr := state.NewManager()
	resource, err := stateMgr.GetResource(hint)
	if err != nil {
		log.Fatalf("Encryption password '%s' not found in state. Has it been snapshotted?\n", hint)
	}

	if resource.Type != "VBREncryptionPassword" {
		log.Fatalf("Resource '%s' is not an encryption password (type: %s).\n", hint, resource.Type)
	}

	fmt.Printf("Checking drift for encryption password: %s\n\n", hint)

	// Fetch current from VBR by ID
	passwordList := vhttp.GetData[models.VbrEncryptionPasswordList]("encryptionPasswords", profile)

	var currentMap map[string]interface{}
	found := false
	for _, p := range passwordList.Data {
		if p.ID == resource.ID {
			pBytes, err := json.Marshal(p)
			if err != nil {
				log.Fatalf("Failed to marshal current password data: %v", err)
			}
			if err := json.Unmarshal(pBytes, &currentMap); err != nil {
				log.Fatalf("Failed to unmarshal current password data: %v", err)
			}
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("CRITICAL - Encryption password '%s' (ID: %s) has been removed from VBR!\n", hint, resource.ID)
		fmt.Println("\nThis password may be referenced by active backup jobs.")
		fmt.Println("Encrypted backups using this password may become unrecoverable.")
		os.Exit(4) // CRITICAL drift
	}

	// Compare, classify, filter
	drifts := detectDrift(resource.Spec, currentMap, encryptionIgnoreFields)
	drifts = classifyDrifts(drifts, encryptionSeverityMap)
	minSev := parseSeverityFlag()
	drifts = filterDriftsBySeverity(drifts, minSev)

	if len(drifts) == 0 {
		fmt.Println(noDriftMessage("Encryption password matches snapshot state.", minSev))
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

	fmt.Println("\nThe encryption password has drifted from the snapshot.")
	fmt.Printf("\nTo update the snapshot, run:\n")
	fmt.Printf("  vcli encryption snapshot \"%s\"\n", hint)

	os.Exit(exitCodeForDrifts(drifts))
}

func diffAllEncryptionPasswords() {
	loadSeverityOverrides()
	profile := utils.GetProfile()

	if profile.Name != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	stateMgr := state.NewManager()
	stateResources, err := stateMgr.ListResources("VBREncryptionPassword")
	if err != nil {
		log.Fatalf("Failed to load state: %v\n", err)
	}

	// Fetch current inventory from VBR
	passwordList := vhttp.GetData[models.VbrEncryptionPasswordList]("encryptionPasswords", profile)

	// Build lookup maps by ID
	stateByID := make(map[string]*state.Resource)
	for _, r := range stateResources {
		stateByID[r.ID] = r
	}

	currentByID := make(map[string]map[string]interface{})
	currentHintByID := make(map[string]string)
	for _, p := range passwordList.Data {
		pBytes, err := json.Marshal(p)
		if err != nil {
			fmt.Printf("Warning: Failed to marshal password '%s': %v\n", p.Hint, err)
			continue
		}
		var pMap map[string]interface{}
		if err := json.Unmarshal(pBytes, &pMap); err != nil {
			fmt.Printf("Warning: Failed to unmarshal password '%s': %v\n", p.Hint, err)
			continue
		}
		currentByID[p.ID] = pMap
		currentHintByID[p.ID] = p.Hint
	}

	if len(stateResources) == 0 {
		fmt.Println("No encryption passwords in state.")
		return
	}

	fmt.Printf("Checking encryption password inventory...\n")
	fmt.Printf("  State: %d passwords, VBR: %d passwords\n\n", len(stateResources), len(passwordList.Data))

	minSev := parseSeverityFlag()
	driftedCount := 0
	cleanCount := 0
	var allDrifts []Drift

	// Check for removed passwords (in state but not in VBR) — always CRITICAL
	for id, stateRes := range stateByID {
		if _, exists := currentByID[id]; !exists {
			removedDrift := Drift{Path: "inventory", Action: "removed", State: stateRes.Name, Severity: SeverityCritical}
			if severityRank(SeverityCritical) >= severityRank(minSev) {
				fmt.Printf("  CRITICAL - %s (ID: %s): Removed from VBR\n", stateRes.Name, id)
				allDrifts = append(allDrifts, removedDrift)
				driftedCount++
			}
		}
	}

	// Check for added passwords (in VBR but not in state) — INFO
	for id, hint := range currentHintByID {
		if _, exists := stateByID[id]; !exists {
			addedDrift := Drift{Path: "inventory", Action: "added", VBR: hint, Severity: SeverityInfo}
			if severityRank(SeverityInfo) >= severityRank(minSev) {
				fmt.Printf("  INFO + %s (ID: %s): Added since last snapshot\n", hint, id)
				allDrifts = append(allDrifts, addedDrift)
				driftedCount++
			}
		}
	}

	// Check field-level drift on matching passwords
	for id, stateRes := range stateByID {
		if currentMap, exists := currentByID[id]; exists {
			drifts := detectDrift(stateRes.Spec, currentMap, encryptionIgnoreFields)
			drifts = classifyDrifts(drifts, encryptionSeverityMap)
			drifts = filterDriftsBySeverity(drifts, minSev)

			if len(drifts) > 0 {
				fmt.Printf("  %s: %d drifts detected\n", stateRes.Name, len(drifts))
				for _, d := range drifts {
					printDriftWithSeverity(d)
				}
				allDrifts = append(allDrifts, drifts...)
				driftedCount++
			} else {
				fmt.Printf("  %s: No drift\n", stateRes.Name)
				cleanCount++
			}
		}
	}

	if len(allDrifts) > 0 {
		printSecuritySummary(allDrifts)
	}
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  - %d passwords clean\n", cleanCount)
	fmt.Printf("  - %d passwords drifted/changed\n", driftedCount)

	if driftedCount > 0 {
		os.Exit(exitCodeForDrifts(allDrifts))
	}
	os.Exit(0)
}

// --- KMS Server commands ---

var kmsSnapshotCmd = &cobra.Command{
	Use:   "kms-snapshot [kms-name]",
	Short: "Snapshot KMS server configuration to state",
	Long: `Capture the current KMS server configuration and store it in state.

Examples:
  # Snapshot a single KMS server
  vcli encryption kms-snapshot "My KMS Server"

  # Snapshot all KMS servers
  vcli encryption kms-snapshot --all
`,
	Run: func(cmd *cobra.Command, args []string) {
		if kmsSnapshotAll {
			snapshotAllKmsServers()
		} else if len(args) > 0 {
			snapshotSingleKmsServer(args[0])
		} else {
			log.Fatal("Provide KMS server name or use --all")
		}
	},
}

var kmsDiffCmd = &cobra.Command{
	Use:   "kms-diff [kms-name]",
	Short: "Detect KMS server configuration drift",
	Long: `Compare current VBR KMS server configuration against the last snapshot
to detect changes or drift.

Examples:
  # Check single KMS server for drift
  vcli encryption kms-diff "My KMS Server"

  # Check all KMS servers
  vcli encryption kms-diff --all

Exit Codes:
  0 - No drift detected
  3 - Drift detected (INFO or WARNING)
  4 - Critical security drift detected
  1 - Error occurred`,
	Run: func(cmd *cobra.Command, args []string) {
		if kmsDiffAll {
			diffAllKmsServers()
		} else if len(args) > 0 {
			diffSingleKmsServer(args[0])
		} else {
			log.Fatal("Provide KMS server name or use --all")
		}
	},
}

func snapshotSingleKmsServer(name string) {
	profile := utils.GetProfile()

	if profile.Name != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	kmsList := vhttp.GetData[models.VbrKmsServerList]("kmsServers", profile)

	var found *models.VbrKmsServerGet
	for i := range kmsList.Data {
		if kmsList.Data[i].Name == name {
			found = &kmsList.Data[i]
			break
		}
	}

	if found == nil {
		log.Fatalf("KMS server '%s' not found in VBR.", name)
	}

	kBytes, err := json.Marshal(found)
	if err != nil {
		log.Fatalf("Failed to marshal KMS server data: %v", err)
	}

	if err := saveResourceToState("VBRKmsServer", found.Name, found.ID, json.RawMessage(kBytes)); err != nil {
		log.Fatalf("Failed to save KMS server state: %v", err)
	}

	fmt.Printf("Snapshot saved for KMS server: %s\n", name)
}

func snapshotAllKmsServers() {
	profile := utils.GetProfile()

	if profile.Name != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	kmsList := vhttp.GetData[models.VbrKmsServerList]("kmsServers", profile)

	if len(kmsList.Data) == 0 {
		fmt.Println("No KMS servers found.")
		return
	}

	fmt.Printf("Snapshotting %d KMS servers...\n", len(kmsList.Data))

	// Build name counts to detect duplicates
	nameCounts := make(map[string]int, len(kmsList.Data))
	for _, k := range kmsList.Data {
		nameCounts[k.Name]++
	}

	for _, k := range kmsList.Data {
		kBytes, err := json.Marshal(k)
		if err != nil {
			fmt.Printf("Warning: Failed to marshal KMS server '%s': %v\n", k.Name, err)
			continue
		}

		resourceName := k.Name
		if nameCounts[k.Name] > 1 {
			resourceName = fmt.Sprintf("%s-%s", k.Name, k.ID)
		}

		if err := saveResourceToState("VBRKmsServer", resourceName, k.ID, json.RawMessage(kBytes)); err != nil {
			fmt.Printf("Warning: Failed to save state for '%s': %v\n", resourceName, err)
			continue
		}

		fmt.Printf("  Snapshot saved: %s\n", resourceName)
	}

	stateMgr := state.NewManager()
	fmt.Printf("\nState updated: %s\n", stateMgr.GetStatePath())
}

func diffSingleKmsServer(name string) {
	loadSeverityOverrides()
	profile := utils.GetProfile()

	if profile.Name != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	stateMgr := state.NewManager()
	resource, err := stateMgr.GetResource(name)
	if err != nil {
		log.Fatalf("KMS server '%s' not found in state. Has it been snapshotted?\n", name)
	}

	if resource.Type != "VBRKmsServer" {
		log.Fatalf("Resource '%s' is not a KMS server (type: %s).\n", name, resource.Type)
	}

	fmt.Printf("Checking drift for KMS server: %s\n\n", name)

	// Fetch current from VBR
	kmsList := vhttp.GetData[models.VbrKmsServerList]("kmsServers", profile)

	var currentMap map[string]interface{}
	found := false
	for _, k := range kmsList.Data {
		if k.ID == resource.ID {
			kBytes, err := json.Marshal(k)
			if err != nil {
				log.Fatalf("Failed to marshal current KMS server data: %v", err)
			}
			if err := json.Unmarshal(kBytes, &currentMap); err != nil {
				log.Fatalf("Failed to unmarshal current KMS server data: %v", err)
			}
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("CRITICAL - KMS server '%s' (ID: %s) has been removed from VBR!\n", name, resource.ID)
		os.Exit(4) // CRITICAL drift
	}

	// Compare, classify, filter
	drifts := detectDrift(resource.Spec, currentMap, kmsIgnoreFields)
	drifts = classifyDrifts(drifts, kmsSeverityMap)
	minSev := parseSeverityFlag()
	drifts = filterDriftsBySeverity(drifts, minSev)

	if len(drifts) == 0 {
		fmt.Println(noDriftMessage("KMS server matches snapshot state.", minSev))
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

	fmt.Println("\nThe KMS server has drifted from the snapshot configuration.")
	fmt.Printf("\nTo update the snapshot, run:\n")
	fmt.Printf("  vcli encryption kms-snapshot \"%s\"\n", name)

	os.Exit(exitCodeForDrifts(drifts))
}

func diffAllKmsServers() {
	loadSeverityOverrides()
	profile := utils.GetProfile()

	if profile.Name != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	stateMgr := state.NewManager()
	stateResources, err := stateMgr.ListResources("VBRKmsServer")
	if err != nil {
		log.Fatalf("Failed to load state: %v\n", err)
	}

	// Fetch current from VBR
	kmsList := vhttp.GetData[models.VbrKmsServerList]("kmsServers", profile)

	// Build lookup maps by ID
	stateByID := make(map[string]*state.Resource)
	for _, r := range stateResources {
		stateByID[r.ID] = r
	}

	currentByID := make(map[string]map[string]interface{})
	currentNameByID := make(map[string]string)
	for _, k := range kmsList.Data {
		kBytes, err := json.Marshal(k)
		if err != nil {
			fmt.Printf("Warning: Failed to marshal KMS server '%s': %v\n", k.Name, err)
			continue
		}
		var kMap map[string]interface{}
		if err := json.Unmarshal(kBytes, &kMap); err != nil {
			fmt.Printf("Warning: Failed to unmarshal KMS server '%s': %v\n", k.Name, err)
			continue
		}
		currentByID[k.ID] = kMap
		currentNameByID[k.ID] = k.Name
	}

	if len(stateResources) == 0 {
		fmt.Println("No KMS servers in state.")
		return
	}

	fmt.Printf("Checking KMS server inventory...\n")
	fmt.Printf("  State: %d servers, VBR: %d servers\n\n", len(stateResources), len(kmsList.Data))

	minSev := parseSeverityFlag()
	driftedCount := 0
	cleanCount := 0
	var allDrifts []Drift

	// Check for removed KMS servers (in state but not in VBR) — CRITICAL
	for id, stateRes := range stateByID {
		if _, exists := currentByID[id]; !exists {
			removedDrift := Drift{Path: "inventory", Action: "removed", State: stateRes.Name, Severity: SeverityCritical}
			if severityRank(SeverityCritical) >= severityRank(minSev) {
				fmt.Printf("  CRITICAL - %s (ID: %s): Removed from VBR\n", stateRes.Name, id)
				allDrifts = append(allDrifts, removedDrift)
				driftedCount++
			}
		}
	}

	// Check for added KMS servers (in VBR but not in state) — INFO
	for id, name := range currentNameByID {
		if _, exists := stateByID[id]; !exists {
			addedDrift := Drift{Path: "inventory", Action: "added", VBR: name, Severity: SeverityInfo}
			if severityRank(SeverityInfo) >= severityRank(minSev) {
				fmt.Printf("  INFO + %s (ID: %s): Added since last snapshot\n", name, id)
				allDrifts = append(allDrifts, addedDrift)
				driftedCount++
			}
		}
	}

	// Check field-level drift on matching servers
	for id, stateRes := range stateByID {
		if currentMap, exists := currentByID[id]; exists {
			drifts := detectDrift(stateRes.Spec, currentMap, kmsIgnoreFields)
			drifts = classifyDrifts(drifts, kmsSeverityMap)
			drifts = filterDriftsBySeverity(drifts, minSev)

			if len(drifts) > 0 {
				fmt.Printf("  %s: %d drifts detected\n", stateRes.Name, len(drifts))
				for _, d := range drifts {
					printDriftWithSeverity(d)
				}
				allDrifts = append(allDrifts, drifts...)
				driftedCount++
			} else {
				fmt.Printf("  %s: No drift\n", stateRes.Name)
				cleanCount++
			}
		}
	}

	if len(allDrifts) > 0 {
		printSecuritySummary(allDrifts)
	}
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  - %d KMS servers clean\n", cleanCount)
	fmt.Printf("  - %d KMS servers drifted/changed\n", driftedCount)

	if driftedCount > 0 {
		os.Exit(exitCodeForDrifts(allDrifts))
	}
	os.Exit(0)
}

func init() {
	encSnapshotCmd.Flags().BoolVar(&encSnapshotAll, "all", false, "Snapshot all encryption passwords")
	encDiffCmd.Flags().BoolVar(&encDiffAll, "all", false, "Check drift for all encryption passwords in state")
	addSeverityFlags(encDiffCmd)
	kmsSnapshotCmd.Flags().BoolVar(&kmsSnapshotAll, "all", false, "Snapshot all KMS servers")
	kmsDiffCmd.Flags().BoolVar(&kmsDiffAll, "all", false, "Check drift for all KMS servers in state")
	addSeverityFlags(kmsDiffCmd)

	encryptionCmd.AddCommand(encSnapshotCmd)
	encryptionCmd.AddCommand(encDiffCmd)
	encryptionCmd.AddCommand(kmsSnapshotCmd)
	encryptionCmd.AddCommand(kmsDiffCmd)
	rootCmd.AddCommand(encryptionCmd)
}
