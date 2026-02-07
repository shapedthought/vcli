package resources

import (
	"fmt"

	"github.com/shapedthought/owlctl/utils"
	"github.com/shapedthought/owlctl/vhttp"
)

// Resolver handles name-to-ID resolution for VBR resources
type Resolver struct {
	cache map[string]string // Cache for name->ID mappings
}

// NewResolver creates a new resolver instance
func NewResolver() *Resolver {
	return &Resolver{
		cache: make(map[string]string),
	}
}

// Repository represents a VBR backup repository
type Repository struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// RepositoryList represents the list response from VBR
type RepositoryList struct {
	Data []Repository `json:"data"`
}

// VMObject represents a VM from VBR hierarchy
type VMObject struct {
	ObjectID string `json:"objectId"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	HostName string `json:"hostName"`
}

// VMList represents VM list response
type VMList struct {
	Data []VMObject `json:"data"`
}

// ResolveRepositoryID resolves a repository name to its ID
func (r *Resolver) ResolveRepositoryID(name string) (string, error) {
	// Check cache first
	cacheKey := "repo:" + name
	if id, found := r.cache[cacheKey]; found {
		return id, nil
	}

	// Fetch repositories from VBR
	profile := utils.GetCurrentProfile()
	repos := vhttp.GetData[RepositoryList]("backupInfrastructure/repositories", profile)

	// Find repository by name
	for _, repo := range repos.Data {
		if repo.Name == name {
			r.cache[cacheKey] = repo.ID
			return repo.ID, nil
		}
	}

	return "", fmt.Errorf("repository %q not found", name)
}

// ResolveVMID resolves a VM name to its object ID
func (r *Resolver) ResolveVMID(name string, hostName string) (string, error) {
	// Check cache first
	cacheKey := "vm:" + hostName + ":" + name
	if id, found := r.cache[cacheKey]; found {
		return id, nil
	}

	// Fetch VMs from VBR hierarchy
	profile := utils.GetCurrentProfile()
	vms := vhttp.GetData[VMList]("vmware/hierarchyRoots", profile)

	// Find VM by name (and optionally hostname)
	for _, vm := range vms.Data {
		if vm.Name == name {
			if hostName == "" || vm.HostName == hostName {
				r.cache[cacheKey] = vm.ObjectID
				return vm.ObjectID, nil
			}
		}
	}

	if hostName != "" {
		return "", fmt.Errorf("VM %q not found on host %q", name, hostName)
	}
	return "", fmt.Errorf("VM %q not found", name)
}

// ResolveRepositoryName resolves a repository ID to its name
func (r *Resolver) ResolveRepositoryName(id string) (string, error) {
	// Fetch repositories from VBR
	profile := utils.GetCurrentProfile()
	repos := vhttp.GetData[RepositoryList]("backupInfrastructure/repositories", profile)

	// Find repository by ID
	for _, repo := range repos.Data {
		if repo.ID == id {
			return repo.Name, nil
		}
	}

	return "", fmt.Errorf("repository with ID %q not found", id)
}

// ResolveVMName resolves a VM object ID to its name
func (r *Resolver) ResolveVMName(objectID string) (string, string, error) {
	// Fetch VMs from VBR hierarchy
	profile := utils.GetCurrentProfile()
	vms := vhttp.GetData[VMList]("vmware/hierarchyRoots", profile)

	// Find VM by object ID
	for _, vm := range vms.Data {
		if vm.ObjectID == objectID {
			return vm.Name, vm.HostName, nil
		}
	}

	return "", "", fmt.Errorf("VM with object ID %q not found", objectID)
}
