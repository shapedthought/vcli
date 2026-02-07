# Raw JSON Refactor: Moving the Declarative Workflow from Typed Structs to `map[string]interface{}`

## Context

The owlctl project has a declarative workflow for managing VBR resources: **jobs**, **repositories**, **SOBRs**, **encryption passwords**, and **KMS servers**. Four commands form the core of this workflow:

- **adopt** -- import an existing VBR resource into state without modifying it
- **apply** -- create or update a VBR resource from a YAML spec
- **diff** -- detect configuration drift between applied state and current VBR
- **export** -- export a VBR resource to declarative YAML

The current implementation for **jobs** uses typed Go structs -- `models.VbrJobGet` for API responses and `models.VbrJobPost` for API requests. These structs are modeled specifically for **VSphere backup jobs** (`type: "VSphereBackup"`). The VBR API is polymorphic: the shape of a job's request and response body varies depending on the `type` field (e.g., `VSphereBackup`, `EntraIDTenantBackupCopy`, `NASBackup`). Each type includes different sections with type-specific fields.

When a non-VSphere job is deserialized into `VbrJobGet`, Go's JSON unmarshaler silently ignores any fields that do not have a corresponding struct field. This means type-specific data is dropped without any error or warning.

Notably, the repository, SOBR, encryption, and KMS commands already use `json.RawMessage` and `map[string]interface{}` for this exact reason -- those resource models were designed to avoid the typed-struct problem from the start.

---

## The Problem (Technical)

### 1. `findJobByName` (cmd/apply.go, line 415)

```go
func findJobByName(name string, profile models.Profile) (models.VbrJobGet, bool) {
    type JobsResponse struct {
        Data []models.VbrJobGet `json:"data"`
    }
    response := vhttp.GetData[JobsResponse]("jobs", profile)
    for _, job := range response.Data {
        if job.Name == name {
            return job, true
        }
    }
    return models.VbrJobGet{}, false
}
```

This function deserializes the entire `/jobs` list endpoint into `[]models.VbrJobGet`. For the purpose of name-to-ID lookup, this works because `name` and `id` are common fields present on every job type. However, the returned `VbrJobGet` object has already lost any fields that are not defined in the struct. This matters because `apply` uses the returned object as the base for merge operations (see below).

This function is used by both **apply** (to check if a job exists before creating/updating) and **adopt** (to resolve a name to an ID before fetching the full resource).

### 2. `apply` (cmd/apply.go)

The full apply pipeline for job creation:

```
YAML spec file
  -> resources.LoadResourceSpec()   -> ResourceSpec (spec is map[string]interface{})
  -> specToVBRJob()                 -> models.VbrJobPost (typed struct)
  -> json.Marshal()                 -> JSON bytes
  -> HTTP POST to /api/v1/jobs
```

The `specToVBRJob` function (line 349) marshals the spec map to JSON, then unmarshals it into `VbrJobPost`. Any fields in the spec that do not match `VbrJobPost` struct fields are silently dropped at this step.

For job updates, the pipeline is:

```
findJobByName()                     -> models.VbrJobGet (already lost non-VSphere fields)
mergeJobUpdates(existing, desired)  -> models.VbrJobPost (merge operates on typed structs)
Convert to VbrJobGet for PUT body   -> HTTP PUT to /api/v1/jobs/{id}
```

The `mergeJobUpdates` function (line 302) copies fields from the existing `VbrJobGet` into a new `VbrJobPost`, then selectively applies desired changes on top. Both structs only contain VSphere-specific fields (`VirtualMachines`, `Storage`, `GuestProcessing`, `Schedule`), so any type-specific fields from the existing job are lost before the merge even begins. The PUT request then sends back an incomplete job object, which would either fail at the API or silently remove fields.

Additionally, `mergeJobUpdates` includes VSphere-specific cleanup logic:
- `cleanVMExcludes` (line 397) operates on `models.VirtualMachines`, which is a VSphere concept
- Credential defaults (line 334) assume a `GuestProcessing.GuestCredentials` structure that does not exist on all job types

### 3. `adopt` (cmd/adopt.go) -- Already Safe

The adopt command's `fetchCurrentJob` function (line 171) calls `findJobByName` only to resolve the name to an ID. It then makes a second request to fetch the full job details as `json.RawMessage`:

```go
func fetchCurrentJob(name string, profile models.Profile) (json.RawMessage, string, error) {
    job, found := findJobByName(name, profile)  // Only used for name -> ID lookup
    if !found {
        return nil, "", fmt.Errorf("job '%s' not found in VBR", name)
    }
    endpoint := fmt.Sprintf("jobs/%s", job.ID)
    rawData := vhttp.GetData[json.RawMessage](endpoint, profile)  // Full raw JSON
    return rawData, job.ID, nil
}
```

The comparison in `adoptResource` (line 21) uses `map[string]interface{}` on both sides -- the spec map from the YAML file and the VBR data unmarshaled from raw JSON. No data is lost. The adopt workflow is already type-agnostic and works correctly for any job type.

### 4. `diff` (cmd/jobs.go, lines 344 and 408)

For job diff, the current implementation fetches via typed struct then round-trips through JSON:

```go
// diffSingleJob (line 363)
current := vhttp.GetData[models.VbrJobGet](endpoint, profile)
currentBytes, err := json.Marshal(current)
var currentMap map[string]interface{}
json.Unmarshal(currentBytes, &currentMap)
```

This `VbrJobGet -> JSON -> map` round-trip drops any fields not in the struct definition. When the state contains a full spec (because it was stored via adopt, which uses raw JSON), the drift detection will find "removed" drifts for every field that `VbrJobGet` does not model. These are false positives -- the fields exist in VBR but were silently dropped during deserialization.

By contrast, repository diff (cmd/repo.go, line 205) fetches as `json.RawMessage` directly:

```go
currentRaw := vhttp.GetData[json.RawMessage](endpoint, profile)
var currentMap map[string]interface{}
json.Unmarshal(currentRaw, &currentMap)
```

This preserves all fields and produces accurate drift results. The same pattern is used by SOBR diff, encryption diff, and KMS diff.

### 5. `export` (cmd/export.go, lines 71 and 93)

The export command fetches jobs via typed struct:

```go
func exportSingleJob(jobID string, profile models.Profile) {
    endpoint := fmt.Sprintf("jobs/%s", jobID)
    vbrJob := vhttp.GetData[models.VbrJobGet](endpoint, profile)
    yamlContent, err := convertJobToYAML(&vbrJob)
    ...
}
```

The `convertJobToYAMLFull` function (line 163) marshals the `VbrJobGet` to YAML then unmarshals to a map. Any fields not in the struct are already gone by this point, so the "full" export is actually an incomplete export for non-VSphere jobs.

The `exportAllJobs` function (line 93) compounds this by deserializing the list endpoint into `[]models.VbrJobGet` and then fetching each job individually through `VbrJobGet` again.

Again, by contrast, the repository, SOBR, encryption, and KMS export commands all use `json.RawMessage` and the shared `convertResourceToYAML` helper (cmd/export_helpers.go), which operates on raw JSON and produces complete exports.

---

## The Problem (User Perspective)

Currently, owlctl's declarative workflow (apply, diff, export, adopt) **only works correctly for VSphere backup jobs**. For any other job type supported by the VBR API, users will encounter silent data loss:

- **Apply**: If a user writes a YAML spec for an EntraID tenant backup job and runs `owlctl job apply`, the apply command will strip all EntraID-specific fields (such as tenant configuration, EntraID-specific storage settings, or identity provider details) from the request body before sending it to VBR. This could create a broken job, create a job with default values instead of the specified ones, or cause an API error. There is no warning that fields were dropped.

- **Export**: If a user runs `owlctl export <job-id>` on a non-VSphere job, the exported YAML will be missing all type-specific fields. The user might believe they have a complete backup of their job configuration, but reapplying the export would produce a different (and likely broken) job.

- **Diff**: If a user runs `owlctl job diff` on a non-VSphere job, the drift detection will report false "removed" drifts because the typed struct dropped fields when reading from VBR. The user would see phantom drift that does not actually exist, making the diff output untrustworthy.

- **Adopt**: This command already works correctly for all job types because it uses `json.RawMessage` internally. No changes needed here.

The core issue is an asymmetry: repositories, SOBRs, encryption passwords, and KMS servers all work correctly with any schema variation because they use raw JSON throughout. Jobs are the only resource type that funnels data through typed structs in the declarative workflow, and this limits jobs to a single type.

---

## The Proposed Solution

### Core Idea

Instead of deserializing VBR API responses into typed Go structs (`VbrJobGet`, `VbrJobPost`) and then re-serializing them, the declarative workflow should work with `json.RawMessage` and `map[string]interface{}` throughout. The YAML spec already contains the correct structure for any job type as a `map[string]interface{}` (see `ResourceSpec.Spec` in `resources/spec.go`). Go does not need to "understand" the API schema -- it just needs to pass the data through faithfully.

This is the same approach already used by the repo, SOBR, encryption, and KMS commands.

### What Changes

#### 1. New shared helper: `findResourceInList`

```go
func findResourceInList(
    listEndpoint string,
    matchField string,
    matchValue string,
    profile models.Profile,
) (json.RawMessage, string, error)
```

This function would:
- Fetch the list endpoint and deserialize the `data` array as `[]json.RawMessage`
- Walk each entry, unmarshal to `map[string]interface{}`, and check the match field
- Return the matching raw JSON entry and its `id`
- Work for any resource type: jobs match on `"name"`, encryption passwords match on `"hint"`, etc.

This replaces `findJobByName` and the inline find-by-name loops that currently exist in `repo.go` (lines 122-128, 588-597), `encryption.go` (lines 131-136, 436-445, 583-592, 734-743), and other files. Each of those loops uses a typed struct just to extract the name and ID -- a pattern that `findResourceInList` would handle generically.

Implementation note: The list response structure `{"data": [...], "pagination": {...}}` is consistent across VBR API endpoints. The function would deserialize into a structure like:

```go
type GenericListResponse struct {
    Data []json.RawMessage `json:"data"`
}
```

#### 2. Adopt fetch functions simplified

The existing fetch functions (`fetchCurrentJob`, `fetchCurrentRepo`, `fetchCurrentSobr`, `fetchCurrentEncryptionPassword`, `fetchCurrentKmsServer`) would become thin wrappers around `findResourceInList`:

**Before** (`fetchCurrentJob`):
```go
func fetchCurrentJob(name string, profile models.Profile) (json.RawMessage, string, error) {
    job, found := findJobByName(name, profile)  // Typed struct lookup
    if !found { return nil, "", ... }
    rawData := vhttp.GetData[json.RawMessage](fmt.Sprintf("jobs/%s", job.ID), profile)
    return rawData, job.ID, nil
}
```

**After**:
```go
func fetchCurrentJob(name string, profile models.Profile) (json.RawMessage, string, error) {
    _, id, err := findResourceInList("jobs", "name", name, profile)
    if err != nil { return nil, "", err }
    rawData := vhttp.GetData[json.RawMessage](fmt.Sprintf("jobs/%s", id), profile)
    return rawData, id, nil
}
```

Some resources (repos, SOBRs, jobs) need a second GET for full details because the list endpoint returns summary data. Others (encryption passwords, KMS servers) have all data in the list response, so `findResourceInList` would return the complete raw JSON directly.

#### 3. Job diff refactored

**Before** (cmd/jobs.go, lines 363-373):
```go
current := vhttp.GetData[models.VbrJobGet](endpoint, profile)
currentBytes, _ := json.Marshal(current)
var currentMap map[string]interface{}
json.Unmarshal(currentBytes, &currentMap)
```

**After**:
```go
currentRaw := vhttp.GetData[json.RawMessage](endpoint, profile)
var currentMap map[string]interface{}
json.Unmarshal(currentRaw, &currentMap)
```

This is a two-line change per call site (one in `diffSingleJob`, one in `diffAllJobs`). It brings job diff in line with how repo, SOBR, encryption, and KMS diff already work.

#### 4. Job export refactored (full mode)

**Before** (cmd/export.go, line 71-91):
```go
func exportSingleJob(jobID string, profile models.Profile) {
    endpoint := fmt.Sprintf("jobs/%s", jobID)
    vbrJob := vhttp.GetData[models.VbrJobGet](endpoint, profile)
    yamlContent, err := convertJobToYAML(&vbrJob)
    ...
}
```

**After**:
```go
func exportSingleJob(jobID string, profile models.Profile) {
    endpoint := fmt.Sprintf("jobs/%s", jobID)
    rawData := vhttp.GetData[json.RawMessage](endpoint, profile)

    // Extract name from raw JSON for metadata
    var meta struct{ Name string `json:"name"` }
    json.Unmarshal(rawData, &meta)

    cfg := ResourceExportConfig{
        Kind:         "VBRJob",
        IgnoreFields: jobIgnoreFields,
        HeaderLines:  []string{"# VBR Job Configuration (Full Export)", "# Exported from VBR"},
    }
    yamlContent, err := convertResourceToYAML(meta.Name, rawData, cfg)
    ...
}
```

This reuses the `convertResourceToYAML` helper from `export_helpers.go` that repo, SOBR, encryption, and KMS exports already use. The full export mode becomes type-agnostic.

The simplified and overlay export modes (`convertJobToYAMLSimplified`, `convertJobToYAMLOverlay`) would still need some understanding of fields for their specific purposes, but the full export -- which is the most commonly used and the one that should preserve all data -- becomes lossless.

#### 5. Apply refactored

This is the most significant change. The current apply flow converts the spec map into a typed struct and back. The refactored flow would work with maps throughout.

**For CREATE** (new job):
```go
// Current: spec map -> VbrJobPost struct -> JSON -> POST
// Proposed: spec map -> strip read-only fields -> JSON -> POST

body := make(map[string]interface{})
for k, v := range spec.Spec {
    if k == "id" { continue }  // Strip read-only fields
    body[k] = v
}
body["name"] = spec.Metadata.Name  // Metadata.name is authoritative
vhttp.PostDataRaw("jobs", body, profile)
```

**For UPDATE** (existing job):
```go
// Current: fetch as VbrJobGet -> mergeJobUpdates(existing, VbrJobPost) -> PUT
// Proposed: fetch as json.RawMessage -> unmarshal to map -> deep merge spec on top -> PUT

_, id, _ := findResourceInList("jobs", "name", spec.Metadata.Name, profile)
existingRaw := vhttp.GetData[json.RawMessage](fmt.Sprintf("jobs/%s", id), profile)
var existingMap map[string]interface{}
json.Unmarshal(existingRaw, &existingMap)

merged := deepMerge(existingMap, spec.Spec)
merged["id"] = id           // VBR requires ID in PUT body
merged["name"] = spec.Metadata.Name

vhttp.PutDataRaw(fmt.Sprintf("jobs/%s", id), merged, profile)
```

This eliminates the `specToVBRJob` and `mergeJobUpdates` functions entirely for the declarative path. The `VbrJobPost` and `VbrJobGet` structs remain for the legacy `job template` / `job create` commands (see next section).

### What Stays the Same

- **Legacy `owlctl job template` and `owlctl job create` commands** continue to use typed structs. These commands predate the declarative workflow, are VSphere-specific by design, and work fine for their intended use case. They are defined in `cmd/jobs.go` in the `getTemplates` and `createJob` functions.

- **The YAML spec format** does not change at all. `ResourceSpec.Spec` is already `map[string]interface{}`, so specs written for any job type already work -- the problem was never in the spec loading, but in the subsequent conversion to typed structs.

- **State file format** does not change. `state.Resource.Spec` is already `map[string]interface{}`.

- **Drift detection logic** (`detectDrift`, `classifyDrifts`, `filterDriftsBySeverity`) does not change. These functions already operate on `map[string]interface{}` inputs.

- **All repository, SOBR, encryption password, and KMS server commands** are already implemented using the raw JSON pattern and do not need changes.

---

## Impact Assessment

### What This Enables

- **Apply, diff, export, and adopt work for any VBR job type**, not just VSphere. Users can manage EntraID, NAS, physical, cloud, and any future job types through the same declarative workflow.

- **No need to add new Go structs when Veeam adds new job types to the API.** The current approach would require defining new typed structs for each new job type. The raw JSON approach handles any schema automatically.

- **Less code overall.** The `specToVBRJob`, `mergeJobUpdates`, and `cleanVMExcludes` functions are removed from the declarative path. The job-specific `convertJobToYAMLFull` function is replaced by the already-existing generic `convertResourceToYAML` helper.

- **Consistency across resource types.** All five resource types (jobs, repos, SOBRs, encryption passwords, KMS servers) would follow the same pattern in the declarative workflow.

### What This Does Not Change

- **User-facing CLI commands and flags** remain identical. No command syntax, flag names, or output formats change.

- **YAML spec format** is unchanged. Existing specs continue to work without modification.

- **State file format** is unchanged. Existing state files remain compatible.

- **Legacy template/create workflow** is unaffected. The `owlctl job template` and `owlctl job create` commands continue to work exactly as they do today.

### Risks

1. **Deep merge for updates needs careful implementation.** The `deepMerge` function for the apply-update path must handle:
   - Nested objects (recursive merge)
   - Arrays (replacement vs. append semantics -- VBR typically expects full array replacement)
   - Null values (explicit null vs. absent field)
   - Type mismatches between spec and existing (e.g., spec has a string where existing has an object)

   The `resources/merge.go` package already contains YAML merge logic that could serve as a reference, but the apply-update merge operates on `map[string]interface{}` at the JSON level, not YAML nodes.

2. **Field cleanup in apply.** The current apply includes some VSphere-specific cleanup logic:
   - `cleanVMExcludes` (line 397) removes invalid or empty VM disk exclusion entries
   - Credential defaults (line 377) set `useAgentManagementCredentials: true` when no credentials are specified
   - Merged job updates set credential flags based on old vs. new format detection (line 334)

   These would need to be handled in one of three ways:
   - **(a)** Moved to spec validation -- run before the apply, operating on the spec map
   - **(b)** Done on the map directly -- the same logic but using map key access instead of struct field access
   - **(c)** Dropped entirely -- if these are VSphere-specific defaults, they arguably should not apply to non-VSphere job types

   Option (c) is likely correct for most of this logic: if a user is creating a VSphere job, their spec should include the correct credential configuration. The CLI should not be injecting defaults that only make sense for one job type.

3. **Less compile-time type safety in the declarative path.** With typed structs, the Go compiler catches field name typos and type mismatches at build time. With `map[string]interface{}`, these become runtime issues. This is an inherent trade-off: type safety vs. schema flexibility. Given that the VBR API schema is not controlled by owlctl and changes across API versions, the flexibility is more valuable here. The compile-time safety was providing a false sense of security anyway, since it only covered VSphere fields.

4. **The `vhttp.PostData` and `vhttp.PutData` functions** may need variants that accept `interface{}` or `map[string]interface{}` instead of a generic type parameter, depending on how the existing `sendData.go` and `modifyData.go` are structured. Alternatively, the existing generic functions should work with `map[string]interface{}` since `json.Marshal` handles it correctly.

---

## Suggested Implementation Order

Each step below is independently useful and can be merged as a separate pull request.

### Step 1: Add `findResourceInList` helper and refactor adopt fetch functions

**Risk: Low.** This is purely additive. The new helper function is created, and the existing fetch functions are updated to use it. The adopt command's behavior does not change -- it already uses raw JSON. This step just consolidates the name-to-ID lookup pattern.

**Files changed:**
- New helper function (could go in `cmd/utils.go` or a new `cmd/resource_helpers.go`)
- `cmd/adopt.go` -- update `fetchCurrentJob`
- `cmd/repo.go` -- update `fetchCurrentRepo`, `fetchCurrentSobr`
- `cmd/encryption.go` -- update `fetchCurrentEncryptionPassword`, `fetchCurrentKmsServer`

### Step 2: Refactor job diff to use `json.RawMessage`

**Risk: Low.** A minimal two-line change per call site. This brings job diff in line with how repo/SOBR/encryption/KMS diff already works. Existing diff behavior is preserved but now produces accurate results for non-VSphere jobs.

**Files changed:**
- `cmd/jobs.go` -- `diffSingleJob` and `diffAllJobs` functions

### Step 3: Refactor job export (full mode) to use raw JSON

**Risk: Medium.** The full export mode switches to the shared `convertResourceToYAML` helper. The simplified and overlay export modes remain unchanged (they are VSphere-specific features that may be deprecated or generalized later).

**Files changed:**
- `cmd/export.go` -- `exportSingleJob`, `exportAllJobs`, `convertJobToYAMLFull`

### Step 4: Refactor apply to use raw JSON for create and update

**Risk: Highest.** This is the most code-intensive change and affects the write path. Requires:
- Implementing a `deepMerge` function for map-based merging
- Removing or repurposing `specToVBRJob`, `mergeJobUpdates`, `cleanVMExcludes`
- Ensuring the apply command sends valid requests for both create and update
- Testing against the VBR API with different job types

**Files changed:**
- `cmd/apply.go` -- `applyVBRJob`, `specToVBRJob`, `mergeJobUpdates`, `findJobByName`
- Possibly `vhttp/sendData.go` or `vhttp/modifyData.go` if raw-body variants are needed

After this step, the `models.VbrJobPost` and `models.VbrJobGet` structs would only be used by the legacy `job template` / `job create` commands in `cmd/jobs.go`. They could be annotated with a comment indicating they are legacy-only, or eventually moved to a `legacy` sub-package if desired.
