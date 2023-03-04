package models

type VbrJobGet struct {
	IsHighPriority  bool `json:"isHighPriority" yaml:"isHighPriority"`
	VirtualMachines struct {
		Includes []struct {
			InventoryObject struct {
				Type     string `json:"type" yaml:"type"`
				HostName string `json:"hostName" yaml:"hostName"`
				Name     string `json:"name" yaml:"name"`
				ObjectID string `json:"objectId" yaml:"objectId"`
			} `json:"inventoryObject" yaml:"inventoryObject"`
			Size string `json:"size" yaml:"size"`
		} `json:"includes" yaml:"includes"`
		Excludes struct {
			Vms   []interface{} `json:"vms" yaml:"vms"`
			Disks []struct {
				DisksToProcess string `json:"disksToProcess" yaml:"disksToProcess"`
				VMObject       struct {
					Type     string `json:"type" yaml:"type"`
					HostName string `json:"hostName" yaml:"hostName"`
					Name     string `json:"name" yaml:"name"`
					ObjectID string `json:"objectId" yaml:"objectId"`
				} `json:"vmObject" yaml:"vmObject"`
				Disks                     []interface{} `json:"disks" yaml:"disks"`
				RemoveFromVMConfiguration bool          `json:"removeFromVMConfiguration" yaml:"removeFromVMConfiguration"`
			} `json:"disks" yaml:"disks"`
			Templates struct {
				IsEnabled              bool `json:"isEnabled" yaml:"isEnabled"`
				ExcludeFromIncremental bool `json:"excludeFromIncremental" yaml:"excludeFromIncremental"`
			} `json:"templates" yaml:"templates"`
		} `json:"excludes" yaml:"excludes"`
	} `json:"virtualMachines" yaml:"virtualMachines"`
	Storage         Storage         `json:"storage" yaml:"storage"`
	GuestProcessing GuestProcessing `json:"guestProcessing" yaml:"guestProcessing"`
	Schedule        Schedule        `json:"schedule" yaml:"schedule"`
	Type            string          `json:"type" yaml:"type"`
	ID              string          `json:"id" yaml:"id"`
	Name            string          `json:"name" yaml:"name"`
	Description     string          `json:"description" yaml:"description"`
	IsDisabled      bool            `json:"isDisabled" yaml:"isDisabled"`
}

type Includes struct {
	Type     string `json:"type" yaml:"type"`
	HostName string `json:"hostName" yaml:"hostName"`
	Name     string `json:"name" yaml:"name"`
	ObjectID string `json:"objectId" yaml:"objectId"`
}

type VirtualMachines struct {
	Includes []Includes `json:"includes" yaml:"includes"`
	Excludes struct {
		Vms   []interface{} `json:"vms" yaml:"vms"`
		Disks []struct {
			DisksToProcess string `json:"disksToProcess" yaml:"disksToProcess"`
			VMObject       struct {
				Type     string `json:"type" yaml:"type"`
				HostName string `json:"hostName" yaml:"hostName"`
				Name     string `json:"name" yaml:"name"`
				ObjectID string `json:"objectId" yaml:"objectId"`
			} `json:"vmObject" yaml:"vmObject"`
			Disks                     []interface{} `json:"disks" yaml:"disks"`
			RemoveFromVMConfiguration bool          `json:"removeFromVMConfiguration" yaml:"removeFromVMConfiguration"`
		} `json:"disks" yaml:"disks"`
		Templates struct {
			IsEnabled              bool `json:"isEnabled" yaml:"isEnabled"`
			ExcludeFromIncremental bool `json:"excludeFromIncremental" yaml:"excludeFromIncremental"`
		} `json:"templates" yaml:"templates"`
	} `json:"excludes" yaml:"excludes"`
}

type Storage struct {
	BackupRepositoryID string `json:"backupRepositoryId" yaml:"backupRepositoryId"`
	BackupProxies      struct {
		AutoSelection bool          `json:"autoSelection" yaml:"autoSelection"`
		ProxyIds      []interface{} `json:"proxyIds" yaml:"proxyIds"`
	} `json:"backupProxies" yaml:"backupProxies"`
	RetentionPolicy struct {
		Type     string `json:"type" yaml:"type"`
		Quantity int    `json:"quantity" yaml:"quantity"`
	} `json:"retentionPolicy" yaml:"retentionPolicy"`
	GfsPolicy struct {
		IsEnabled bool `json:"isEnabled" yaml:"isEnabled"`
		Weekly    struct {
			DesiredTime          string `json:"desiredTime" yaml:"desiredTime"`
			IsEnabled            bool   `json:"isEnabled" yaml:"isEnabled"`
			KeepForNumberOfWeeks int    `json:"keepForNumberOfWeeks" yaml:"keepForNumberOfWeeks"`
		} `json:"weekly" yaml:"weekly"`
		Monthly struct {
			DesiredTime           string `json:"desiredTime" yaml:"desiredTime"`
			IsEnabled             bool   `json:"isEnabled" yaml:"isEnabled"`
			KeepForNumberOfMonths int    `json:"keepForNumberOfMonths" yaml:"keepForNumberOfMonths"`
		} `json:"monthly" yaml:"monthly"`
		Yearly struct {
			DesiredTime          string `json:"desiredTime" yaml:"desiredTime"`
			IsEnabled            bool   `json:"isEnabled" yaml:"isEnabled"`
			KeepForNumberOfYears int    `json:"keepForNumberOfYears" yaml:"keepForNumberOfYears"`
		} `json:"yearly" yaml:"yearly"`
	} `json:"gfsPolicy" yaml:"gfsPolicy"`
	AdvancedSettings struct {
		BackupModeType  string `json:"backupModeType" yaml:"backupModeType"`
		SynthenticFulls struct {
			IsEnabled bool     `json:"isEnabled" yaml:"isEnabled"`
			Days      []string `json:"days" yaml:"days"`
		} `json:"synthenticFulls" yaml:"synthenticFulls"`
		ActiveFulls struct {
			IsEnabled bool `json:"isEnabled" yaml:"isEnabled"`
			Weekly    struct {
				IsEnabled bool     `json:"isEnabled" yaml:"isEnabled"`
				Days      []string `json:"days" yaml:"days"`
			} `json:"weekly" yaml:"weekly"`
			Monthly struct {
				DayOfWeek        string   `json:"dayOfWeek" yaml:"dayOfWeek"`
				DayNumberInMonth string   `json:"dayNumberInMonth" yaml:"dayNumberInMonth"`
				IsEnabled        bool     `json:"isEnabled" yaml:"isEnabled"`
				DayOfMonths      int      `json:"dayOfMonths" yaml:"dayOfMonths"`
				Months           []string `json:"months" yaml:"months"`
			} `json:"monthly" yaml:"monthly"`
		} `json:"activeFulls" yaml:"activeFulls"`
		BackupHealth struct {
			IsEnabled bool `json:"isEnabled" yaml:"isEnabled"`
			Weekly    struct {
				IsEnabled bool     `json:"isEnabled" yaml:"isEnabled"`
				Days      []string `json:"days" yaml:"days"`
			} `json:"weekly" yaml:"weekly"`
			Monthly struct {
				DayOfWeek        string   `json:"dayOfWeek" yaml:"dayOfWeek"`
				DayNumberInMonth string   `json:"dayNumberInMonth" yaml:"dayNumberInMonth"`
				IsEnabled        bool     `json:"isEnabled" yaml:"isEnabled"`
				DayOfMonths      int      `json:"dayOfMonths" yaml:"dayOfMonths"`
				Months           []string `json:"months" yaml:"months"`
			} `json:"monthly" yaml:"monthly"`
		} `json:"backupHealth" yaml:"backupHealth"`
		FullBackupMaintenance struct {
			RemoveData struct {
				IsEnabled bool `json:"isEnabled" yaml:"isEnabled"`
				AfterDays int  `json:"afterDays" yaml:"afterDays"`
			} `json:"RemoveData" yaml:"RemoveData"`
			DefragmentAndCompact struct {
				IsEnabled bool `json:"isEnabled" yaml:"isEnabled"`
				Weekly    struct {
					IsEnabled bool     `json:"isEnabled" yaml:"isEnabled"`
					Days      []string `json:"days" yaml:"days"`
				} `json:"weekly" yaml:"weekly"`
				Monthly struct {
					DayOfWeek        string   `json:"dayOfWeek" yaml:"dayOfWeek"`
					DayNumberInMonth string   `json:"dayNumberInMonth" yaml:"dayNumberInMonth"`
					IsEnabled        bool     `json:"isEnabled" yaml:"isEnabled"`
					DayOfMonths      int      `json:"dayOfMonths" yaml:"dayOfMonths"`
					Months           []string `json:"months" yaml:"months"`
				} `json:"monthly" yaml:"monthly"`
			} `json:"defragmentAndCompact" yaml:"defragmentAndCompact"`
		} `json:"fullBackupMaintenance" yaml:"fullBackupMaintenance"`
		StorageData struct {
			CompressionLevel         string `json:"compressionLevel" yaml:"compressionLevel"`
			StorageOptimization      string `json:"storageOptimization" yaml:"storageOptimization"`
			EnableInlineDataDedup    bool   `json:"enableInlineDataDedup" yaml:"enableInlineDataDedup"`
			ExcludeSwapFileBlocks    bool   `json:"excludeSwapFileBlocks" yaml:"excludeSwapFileBlocks"`
			ExcludeDeletedFileBlocks bool   `json:"excludeDeletedFileBlocks" yaml:"excludeDeletedFileBlocks"`
			Encryption               struct {
				IsEnabled                  bool        `json:"isEnabled" yaml:"isEnabled"`
				EncryptionPasswordIDOrNull string      `json:"encryptionPasswordIdOrNull" yaml:"encryptionPasswordIdOrNull"`
				EncryptionPasswordTag      interface{} `json:"encryptionPasswordTag" yaml:"encryptionPasswordTag"`
			} `json:"encryption" yaml:"encryption"`
		} `json:"storageData" yaml:"storageData"`
		Notifications struct {
			SendSNMPNotifications bool `json:"sendSNMPNotifications" yaml:"sendSNMPNotifications"`
			EmailNotifications    struct {
				NotificationType           interface{}   `json:"notificationType" yaml:"notificationType"`
				IsEnabled                  bool          `json:"isEnabled" yaml:"isEnabled"`
				Recipients                 []interface{} `json:"recipients" yaml:"recipients"`
				CustomNotificationSettings interface{}   `json:"customNotificationSettings" yaml:"customNotificationSettings"`
			} `json:"emailNotifications" yaml:"emailNotifications"`
			VMAttribute struct {
				IsEnabled              bool   `json:"isEnabled" yaml:"isEnabled"`
				Notes                  string `json:"notes" yaml:"notes"`
				AppendToExisitingValue bool   `json:"appendToExisitingValue" yaml:"appendToExisitingValue"`
			} `json:"vmAttribute" yaml:"vmAttribute"`
		} `json:"notifications" yaml:"notifications"`
		VSphere struct {
			EnableVMWareToolsQuiescence bool `json:"enableVMWareToolsQuiescence" yaml:"enableVMWareToolsQuiescence"`
			ChangedBlockTracking        struct {
				IsEnabled              bool `json:"isEnabled" yaml:"isEnabled"`
				EnableCBTautomatically bool `json:"enableCBTautomatically" yaml:"enableCBTautomatically"`
				ResetCBTonActiveFull   bool `json:"resetCBTonActiveFull" yaml:"resetCBTonActiveFull"`
			} `json:"changedBlockTracking" yaml:"changedBlockTracking"`
		} `json:"vSphere" yaml:"vSphere"`
		StorageIntegration struct {
			IsEnabled                bool `json:"isEnabled" yaml:"isEnabled"`
			LimitProcessedVM         bool `json:"limitProcessedVm" yaml:"limitProcessedVm"`
			LimitProcessedVMCount    int  `json:"limitProcessedVmCount" yaml:"limitProcessedVmCount"`
			FailoverToStandardBackup bool `json:"failoverToStandardBackup" yaml:"failoverToStandardBackup"`
		} `json:"storageIntegration" yaml:"storageIntegration"`
		Scripts struct {
			PeriodicityType string `json:"periodicityType" yaml:"periodicityType"`
			PreCommand      struct {
				IsEnabled bool   `json:"isEnabled" yaml:"isEnabled"`
				Command   string `json:"command" yaml:"command"`
			} `json:"preCommand" yaml:"preCommand"`
			PostCommand struct {
				IsEnabled bool   `json:"isEnabled" yaml:"isEnabled"`
				Command   string `json:"command" yaml:"command"`
			} `json:"postCommand" yaml:"postCommand"`
			RunScriptEvery int      `json:"runScriptEvery" yaml:"runScriptEvery"`
			DayOfWeek      []string `json:"dayOfWeek" yaml:"dayOfWeek"`
		} `json:"scripts" yaml:"scripts"`
	} `json:"advancedSettings" yaml:"advancedSettings"`
}

type GuestProcessing struct {
	AppAwareProcessing struct {
		IsEnabled   bool `json:"isEnabled" yaml:"isEnabled"`
		AppSettings []struct {
			Vss             string `json:"vss" yaml:"vss"`
			TransactionLogs string `json:"transactionLogs" yaml:"transactionLogs"`
			VMObject        struct {
				Type     string `json:"type" yaml:"type"`
				HostName string `json:"hostName" yaml:"hostName"`
				Name     string `json:"name" yaml:"name"`
				ObjectID string `json:"objectId" yaml:"objectId"`
			} `json:"vmObject" yaml:"vmObject"`
			UsePersistentGuestAgent bool `json:"usePersistentGuestAgent" yaml:"usePersistentGuestAgent"`
			Sql                     struct {
				LogsProcessing     string      `json:"logsProcessing" yaml:"logsProcessing"`
				RetainLogBackups   interface{} `json:"retainLogBackups" yaml:"retainLogBackups"`
				BackupMinsCount    interface{} `json:"backupMinsCount" yaml:"backupMinsCount"`
				KeepDaysCount      interface{} `json:"keepDaysCount" yaml:"keepDaysCount"`
				LogShippingServers interface{} `json:"logShippingServers" yaml:"logShippingServers"`
			} `json:"sql" yaml:"sql"`
			Oracle struct {
				ArchiveLogs         string      `json:"archiveLogs" yaml:"archiveLogs"`
				RetainLogBackups    string      `json:"retainLogBackups" yaml:"retainLogBackups"`
				UseGuestCredentials bool        `json:"useGuestCredentials" yaml:"useGuestCredentials"` // true
				CredentialsID       interface{} `json:"credentialsId" yaml:"credentialsId"`
				DeleteHoursCount    interface{} `json:"deleteHoursCount" yaml:"deleteHoursCount"`
				DeleteGBsCount      interface{} `json:"deleteGBsCount" yaml:"deleteGBsCount"`
				BackupLogs          bool        `json:"backupLogs" yaml:"backupLogs"`
				BackupMinsCount     int         `json:"backupMinsCount" yaml:"backupMinsCount"`
				KeepDaysCount       int         `json:"keepDaysCount" yaml:"keepDaysCount"`
				LogShippingServers  struct {
					AutoSelection     bool          `json:"autoSelection" yaml:"autoSelection"`
					ShippingServerIds []interface{} `json:"shippingServerIds" yaml:"shippingServerIds"`
				} `json:"logShippingServers" yaml:"logShippingServers"`
			} `json:"oracle" yaml:"oracle"`
			Exclusions struct {
				ExclusionPolicy string        `json:"exclusionPolicy" yaml:"exclusionPolicy"`
				ItemsList       []interface{} `json:"itemsList" yaml:"itemsList"`
			} `json:"exclusions" yaml:"exclusions"`
			Scripts struct {
				ScriptProcessingMode string      `json:"scriptProcessingMode" yaml:"scriptProcessingMode"`
				WindowsScripts       interface{} `json:"windowsScripts" yaml:"windowsScripts"`
				LinuxScripts         interface{} `json:"linuxScripts" yaml:"linuxScripts"`
			} `json:"scripts" yaml:"scripts"`
		} `json:"appSettings" yaml:"appSettings"`
	} `json:"appAwareProcessing" yaml:"appAwareProcessing"`
	GuestFSIndexing struct {
		IsEnabled        bool          `json:"isEnabled" yaml:"isEnabled"`
		IndexingSettings []interface{} `json:"indexingSettings" yaml:"indexingSettings"`
	} `json:"guestFSIndexing" yaml:"guestFSIndexing"`
	GuestInteractionProxies struct {
		AutoSelection bool          `json:"autoSelection" yaml:"autoSelection"`
		ProxyIds      []interface{} `json:"proxyIds" yaml:"proxyIds"`
	} `json:"guestInteractionProxies" yaml:"guestInteractionProxies"`
	GuestCredentials struct {
		CredsType             string        `json:"credsType" yaml:"credsType"`
		CredsID               string        `json:"credsId" yaml:"credsId"`
		CredentialsPerMachine []interface{} `json:"credentialsPerMachine" yaml:"credentialsPerMachine"`
	} `json:"guestCredentials" yaml:"guestCredentials"`
}

type Schedule struct {
	RunAutomatically bool `json:"runAutomatically" yaml:"runAutomatically"`
	Daily            struct {
		DailyKind string   `json:"dailyKind" yaml:"dailyKind"`
		IsEnabled bool     `json:"isEnabled" yaml:"isEnabled"`
		LocalTime string   `json:"localTime" yaml:"localTime"`
		Days      []string `json:"days" yaml:"days"`
	} `json:"daily" yaml:"daily"`
	Monthly struct {
		DayOfWeek        string   `json:"dayOfWeek" yaml:"dayOfWeek"`
		DayNumberInMonth string   `json:"dayNumberInMonth" yaml:"dayNumberInMonth"`
		IsEnabled        bool     `json:"isEnabled" yaml:"isEnabled"`
		LocalTime        string   `json:"localTime" yaml:"localTime"`
		DayOfMonth       int      `json:"dayOfMonth" yaml:"dayOfMonth"`
		Months           []string `json:"months" yaml:"months"`
	} `json:"monthly" yaml:"monthly"`
	Periodically struct {
		PeriodicallyKind string `json:"periodicallyKind" yaml:"periodicallyKind"`
		IsEnabled        bool   `json:"isEnabled" yaml:"isEnabled"`
		Frequency        int    `json:"frequency" yaml:"frequency"`
		BackupWindow     struct {
			Days []struct {
				Day   string `json:"day" yaml:"day"`
				Hours string `json:"hours" yaml:"hours"`
			} `json:"days" yaml:"days"`
		} `json:"backupWindow" yaml:"backupWindow"`
		StartTimeWithinAnHour int `json:"startTimeWithinAnHour" yaml:"startTimeWithinAnHour"`
	} `json:"periodically" yaml:"periodically"`
	Continuously struct {
		IsEnabled    bool `json:"isEnabled" yaml:"isEnabled"`
		BackupWindow struct {
			Days []struct {
				Day   string `json:"day" yaml:"day"`
				Hours string `json:"hours" yaml:"hours"`
			} `json:"days" yaml:"days"`
		} `json:"backupWindow" yaml:"backupWindow"`
	} `json:"continuously" yaml:"continuously"`
	AfterThisJob struct {
		IsEnabled bool        `json:"isEnabled" yaml:"isEnabled"`
		JobName   interface{} `json:"jobName" yaml:"jobName"`
	} `json:"afterThisJob" yaml:"afterThisJob"`
	Retry struct {
		IsEnabled    bool `json:"isEnabled" yaml:"isEnabled"`
		RetryCount   int  `json:"retryCount" yaml:"retryCount"`
		AwaitMinutes int  `json:"awaitMinutes" yaml:"awaitMinutes"`
	} `json:"retry" yaml:"retry"`
	BackupWindow struct {
		IsEnabled    bool `json:"isEnabled" yaml:"isEnabled"`
		BackupWindow struct {
			Days []struct {
				Day   string `json:"day" yaml:"day"`
				Hours string `json:"hours" yaml:"hours"`
			} `json:"days" yaml:"days"`
		} `json:"backupWindow" yaml:"backupWindow"`
	} `json:"backupWindow" yaml:"backupWindow"`
}

type VbrJobPost struct {
	IsHighPriority  bool            `json:"isHighPriority" yaml:"isHighPriority"`
	VirtualMachines VirtualMachines `json:"virtualMachines" yaml:"virtualMachines"`
	Storage         Storage         `json:"storage" yaml:"storage"`
	GuestProcessing GuestProcessing `json:"guestProcessing" yaml:"guestProcessing"`
	Schedule        Schedule        `json:"schedule" yaml:"schedule"`
	Type            string          `json:"type" yaml:"type"`
	Name            string          `json:"name" yaml:"name"`
	Description     string          `json:"description" yaml:"description"`
	IsDisabled      bool            `json:"isDisabled" yaml:"isDisabled"`
}

type VbrJob struct {
	Type            string          `json:"type" yaml:"type"`
	Name            string          `json:"name" yaml:"name"`
	Description     string          `json:"description" yaml:"description"`
	IsDisabled      bool            `json:"isDisabled" yaml:"isDisabled"`
	VirtualMachines VirtualMachines `json:"virtualMachines" yaml:"virtualMachines"`
}