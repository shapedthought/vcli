package vbrmodels

type Repos struct {
	Data []struct {
		HostID      string `json:"hostId" yaml:"hostId"`
		Type        string `json:"type" yaml:"type"`
		ID          string `json:"id" yaml:"id"`
		Name        string `json:"name" yaml:"name"`
		Description string `json:"description" yaml:"description"`
		Repository  struct {
			Path             string `json:"path"`
			MaxTaskCount     int    `json:"maxTaskCount"`
			ReadWriteRate    int    `json:"readWriteRate"`
			AdvancedSettings struct {
				AlignDataBlocks         bool `json:"alignDataBlocks"`
				DecompressBeforeStoring bool `json:"decompressBeforeStoring"`
				RotatedDrives           bool `json:"rotatedDrives"`
				PerVMBackup             bool `json:"perVmBackup"`
			} `json:"advancedSettings"`
		} `json:"repository" yaml:"repository"`
		MountServer struct {
			MountServerID         string `json:"mountServerId"`
			WriteCacheFolder      string `json:"writeCacheFolder"`
			VPowerNFSEnabled      bool   `json:"vPowerNFSEnabled"`
			VPowerNFSPortSettings struct {
				MountPort     int `json:"mountPort"`
				VPowerNFSPort int `json:"vPowerNFSPort"`
			} `json:"vPowerNFSPortSettings"`
		} `json:"mountServer" yaml:"mountServer"`
	} `json:"data" yaml:"data"`
	Pagination struct {
		Total int `json:"total" yaml:"total"`
		Count int `json:"count" yaml:"count"`
		Skip  int `json:"skip" yaml:"skip"`
		Limit int `json:"limit" yaml:"limit"`
	} `json:"pagination" yaml:"pagination"`
}

func (p *Repos) GetName(id string) string {
	repoName := "None"
	for _, r := range p.Data {
		if r.ID == id {
			repoName = r.Name
		}
	}
	return repoName
}

type RepoStates struct {
	Data []struct {
		ID          string  `json:"id" yaml:"id"`
		Name        string  `json:"name" yaml:"name"`
		Type        string  `json:"type" yaml:"type"`
		Description string  `json:"description" yaml:"description"`
		HostID      string  `json:"hostId" yaml:"hostId"`
		HostName    string  `json:"hostName" yaml:"hostName"`
		Path        string  `json:"path" yaml:"path"`
		CapacityGB  float32 `json:"capacityGB" yaml:"capacityGB"`
		FreeGB      float32 `json:"freeGB" yaml:"freeGB"`
		UsedSpaceGB float32 `json:"usedSpaceGB" yaml:"usedSpaceGB"`
	} `json:"data" yaml:"data"`
	Pagination struct {
		Total int `json:"total" yaml:"total"`
		Count int `json:"count" yaml:"count"`
		Skip  int `json:"skip" yaml:"skip"`
		Limit int `json:"limit" yaml:"limit"`
	} `json:"pagination" yaml:"pagination"`
}

type Repo struct {
	Name        string `json:"name" yaml:"name"`
	ID          string `json:"id" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	HostID      string `json:"hostId" yaml:"hostId"`
	Repository  struct {
		Path             string `json:"path" yaml:"path"`
		MaxTaskCount     int    `json:"maxTaskCount" yaml:"maxTaskCount"`
		ReadWriteRate    int    `json:"readWriteRate" yaml:"readWriteRate"`
		AdvancedSettings struct {
			AlignDataBlocks         bool `json:"alignDataBlocks" yaml:"alignDataBlocks"`
			DecompressBeforeStoring bool `json:"decompressBeforeStoring" yaml:"decompressBeforeStoring"`
			RotatedDrives           bool `json:"rotatedDrives" yaml:"rotateDrives"`
			PerVMBackup             bool `json:"perVmBackup" yaml:"perVmBackup"`
		} `json:"advancedSettings" yaml:"advancedSettings"`
	} `json:"repository" yaml:"repository"`
	MountServer struct {
		MountServerID         string `json:"mountServerId" yaml:"mountServerId"`
		WriteCacheFolder      string `json:"writeCacheFolder" yaml:"writeCacheFolder"`
		VPowerNFSEnabled      bool   `json:"vPowerNFSEnabled" yaml:"vPowerNFSEnabled"`
		VPowerNFSPortSettings struct {
			MountPort     int `json:"mountPort" yaml:"mountPort"`
			VPowerNFSPort int `json:"vPowerNFSPort" yaml:"vPowerNFSPort"`
		} `json:"vPowerNFSPortSettings" yaml:"vPowerNFSPortSettings"`
	} `json:"mountServer" yaml:"mountServer"`
	Type string `json:"type" yaml:"type"`
}

type Sobr struct {
	Data []struct {
		ID              string `json:"id"`
		Name            string `json:"name"`
		Description     string `json:"description"`
		Tag             string `json:"tag"`
		PerformanceTier struct {
			PerformanceExtents []struct {
				ID     string `json:"id"`
				Name   string `json:"name"`
				Status string `json:"status"`
			} `json:"performanceExtents"`
			AdvancedSettings struct {
				PerVMBackup           bool `json:"perVmBackup"`
				FullWhenExtentOffline bool `json:"fullWhenExtentOffline"`
			} `json:"advancedSettings"`
		} `json:"performanceTier"`
		PlacementPolicy struct {
			Type     string `json:"type"`
			Settings []struct {
				ExtentName     string `json:"extentName"`
				AllowedBackups string `json:"allowedBackups"`
			} `json:"settings"`
		} `json:"placementPolicy"`
		CapacityTier struct {
			Enabled       bool   `json:"enabled"`
			ExtentID      string `json:"extentId"`
			OffloadWindow struct {
				Days []struct {
					Day   string `json:"day"`
					Hours string `json:"hours"`
				} `json:"days"`
			} `json:"offloadWindow"`
			CopyPolicyEnabled            bool `json:"copyPolicyEnabled"`
			MovePolicyEnabled            bool `json:"movePolicyEnabled"`
			OperationalRestorePeriodDays int  `json:"operationalRestorePeriodDays"`
			OverridePolicy               struct {
				IsEnabled                      bool `json:"isEnabled"`
				OverrideSpaceThresholdPercents int  `json:"overrideSpaceThresholdPercents"`
			} `json:"overridePolicy"`
			Encryption struct {
				IsEnabled                  bool   `json:"isEnabled"`
				EncryptionPasswordIDOrNull string `json:"encryptionPasswordIdOrNull"`
				EncryptionPasswordTag      string `json:"encryptionPasswordTag"`
			} `json:"encryption"`
		} `json:"capacityTier"`
		ArchiveTier struct {
			IsEnabled         bool   `json:"isEnabled"`
			ExtentID          string `json:"extentId"`
			ArchivePeriodDays int    `json:"archivePeriodDays"`
			AdvancedSettings  struct {
				CostOptimizedArchiveEnabled bool `json:"costOptimizedArchiveEnabled"`
				ArchiveDeduplicationEnabled bool `json:"archiveDeduplicationEnabled"`
			} `json:"advancedSettings"`
		} `json:"archiveTier"`
	} `json:"data"`
	Pagination struct {
		Total int `json:"total"`
		Count int `json:"count"`
		Skip  int `json:"skip"`
		Limit int `json:"limit"`
	} `json:"pagination"`
}
