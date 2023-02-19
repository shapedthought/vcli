package models

type VbrJobGet struct {
	IsHighPriority  bool `json:"isHighPriority"`
	VirtualMachines struct {
		Includes []struct {
			InventoryObject struct {
				Type     string `json:"type"`
				HostName string `json:"hostName"`
				Name     string `json:"name"`
				ObjectID string `json:"objectId"`
			} `json:"inventoryObject"`
			Size string `json:"size"`
		} `json:"includes"`
		Excludes struct {
			Vms   []interface{} `json:"vms"`
			Disks []struct {
				DisksToProcess string `json:"disksToProcess"`
				VMObject       struct {
					Type     string `json:"type"`
					HostName string `json:"hostName"`
					Name     string `json:"name"`
					ObjectID string `json:"objectId"`
				} `json:"vmObject"`
				Disks                     []interface{} `json:"disks"`
				RemoveFromVMConfiguration bool          `json:"removeFromVMConfiguration"`
			} `json:"disks"`
			Templates struct {
				IsEnabled              bool `json:"isEnabled"`
				ExcludeFromIncremental bool `json:"excludeFromIncremental"`
			} `json:"templates"`
		} `json:"excludes"`
	} `json:"virtualMachines"`
	Storage struct {
		BackupRepositoryID string `json:"backupRepositoryId"`
		BackupProxies      struct {
			AutoSelection bool          `json:"autoSelection"`
			ProxyIds      []interface{} `json:"proxyIds"`
		} `json:"backupProxies"`
		RetentionPolicy struct {
			Type     string `json:"type"`
			Quantity int    `json:"quantity"`
		} `json:"retentionPolicy"`
		GfsPolicy struct {
			IsEnabled bool `json:"isEnabled"`
			Weekly    struct {
				DesiredTime          string `json:"desiredTime"`
				IsEnabled            bool   `json:"isEnabled"`
				KeepForNumberOfWeeks int    `json:"keepForNumberOfWeeks"`
			} `json:"weekly"`
			Monthly struct {
				DesiredTime           string `json:"desiredTime"`
				IsEnabled             bool   `json:"isEnabled"`
				KeepForNumberOfMonths int    `json:"keepForNumberOfMonths"`
			} `json:"monthly"`
			Yearly struct {
				DesiredTime          string `json:"desiredTime"`
				IsEnabled            bool   `json:"isEnabled"`
				KeepForNumberOfYears int    `json:"keepForNumberOfYears"`
			} `json:"yearly"`
		} `json:"gfsPolicy"`
		AdvancedSettings struct {
			BackupModeType  string `json:"backupModeType"`
			SynthenticFulls struct {
				IsEnabled bool     `json:"isEnabled"`
				Days      []string `json:"days"`
			} `json:"synthenticFulls"`
			ActiveFulls struct {
				IsEnabled bool `json:"isEnabled"`
				Weekly    struct {
					IsEnabled bool     `json:"isEnabled"`
					Days      []string `json:"days"`
				} `json:"weekly"`
				Monthly struct {
					DayOfWeek        string   `json:"dayOfWeek"`
					DayNumberInMonth string   `json:"dayNumberInMonth"`
					IsEnabled        bool     `json:"isEnabled"`
					DayOfMonths      int      `json:"dayOfMonths"`
					Months           []string `json:"months"`
				} `json:"monthly"`
			} `json:"activeFulls"`
			BackupHealth struct {
				IsEnabled bool `json:"isEnabled"`
				Weekly    struct {
					IsEnabled bool     `json:"isEnabled"`
					Days      []string `json:"days"`
				} `json:"weekly"`
				Monthly struct {
					DayOfWeek        string   `json:"dayOfWeek"`
					DayNumberInMonth string   `json:"dayNumberInMonth"`
					IsEnabled        bool     `json:"isEnabled"`
					DayOfMonths      int      `json:"dayOfMonths"`
					Months           []string `json:"months"`
				} `json:"monthly"`
			} `json:"backupHealth"`
			FullBackupMaintenance struct {
				RemoveData struct {
					IsEnabled bool `json:"isEnabled"`
					AfterDays int  `json:"afterDays"`
				} `json:"RemoveData"`
				DefragmentAndCompact struct {
					IsEnabled bool `json:"isEnabled"`
					Weekly    struct {
						IsEnabled bool     `json:"isEnabled"`
						Days      []string `json:"days"`
					} `json:"weekly"`
					Monthly struct {
						DayOfWeek        string   `json:"dayOfWeek"`
						DayNumberInMonth string   `json:"dayNumberInMonth"`
						IsEnabled        bool     `json:"isEnabled"`
						DayOfMonths      int      `json:"dayOfMonths"`
						Months           []string `json:"months"`
					} `json:"monthly"`
				} `json:"defragmentAndCompact"`
			} `json:"fullBackupMaintenance"`
			StorageData struct {
				CompressionLevel         string `json:"compressionLevel"`
				StorageOptimization      string `json:"storageOptimization"`
				EnableInlineDataDedup    bool   `json:"enableInlineDataDedup"`
				ExcludeSwapFileBlocks    bool   `json:"excludeSwapFileBlocks"`
				ExcludeDeletedFileBlocks bool   `json:"excludeDeletedFileBlocks"`
				Encryption               struct {
					IsEnabled                  bool        `json:"isEnabled"`
					EncryptionPasswordIDOrNull string      `json:"encryptionPasswordIdOrNull"`
					EncryptionPasswordTag      interface{} `json:"encryptionPasswordTag"`
				} `json:"encryption"`
			} `json:"storageData"`
			Notifications struct {
				SendSNMPNotifications bool `json:"sendSNMPNotifications"`
				EmailNotifications    struct {
					NotificationType           interface{}   `json:"notificationType"`
					IsEnabled                  bool          `json:"isEnabled"`
					Recipients                 []interface{} `json:"recipients"`
					CustomNotificationSettings interface{}   `json:"customNotificationSettings"`
				} `json:"emailNotifications"`
				VMAttribute struct {
					IsEnabled              bool   `json:"isEnabled"`
					Notes                  string `json:"notes"`
					AppendToExisitingValue bool   `json:"appendToExisitingValue"`
				} `json:"vmAttribute"`
			} `json:"notifications"`
			VSphere struct {
				EnableVMWareToolsQuiescence bool `json:"enableVMWareToolsQuiescence"`
				ChangedBlockTracking        struct {
					IsEnabled              bool `json:"isEnabled"`
					EnableCBTautomatically bool `json:"enableCBTautomatically"`
					ResetCBTonActiveFull   bool `json:"resetCBTonActiveFull"`
				} `json:"changedBlockTracking"`
			} `json:"vSphere"`
			StorageIntegration struct {
				IsEnabled                bool `json:"isEnabled"`
				LimitProcessedVM         bool `json:"limitProcessedVm"`
				LimitProcessedVMCount    int  `json:"limitProcessedVmCount"`
				FailoverToStandardBackup bool `json:"failoverToStandardBackup"`
			} `json:"storageIntegration"`
			Scripts struct {
				PeriodicityType string `json:"periodicityType"`
				PreCommand      struct {
					IsEnabled bool   `json:"isEnabled"`
					Command   string `json:"command"`
				} `json:"preCommand"`
				PostCommand struct {
					IsEnabled bool   `json:"isEnabled"`
					Command   string `json:"command"`
				} `json:"postCommand"`
				RunScriptEvery int      `json:"runScriptEvery"`
				DayOfWeek      []string `json:"dayOfWeek"`
			} `json:"scripts"`
		} `json:"advancedSettings"`
	} `json:"storage"`
	GuestProcessing struct {
		AppAwareProcessing struct {
			IsEnabled   bool `json:"isEnabled"`
			AppSettings []struct {
				Vss             string `json:"vss"`
				TransactionLogs string `json:"transactionLogs"`
				VMObject        struct {
					Type     string `json:"type"`
					HostName string `json:"hostName"`
					Name     string `json:"name"`
					ObjectID string `json:"objectId"`
				} `json:"vmObject"`
				UsePersistentGuestAgent bool `json:"usePersistentGuestAgent"`
				Sql                     struct {
					LogsProcessing     string      `json:"logsProcessing"`
					RetainLogBackups   interface{} `json:"retainLogBackups"`
					BackupMinsCount    interface{} `json:"backupMinsCount"`
					KeepDaysCount      interface{} `json:"keepDaysCount"`
					LogShippingServers interface{} `json:"logShippingServers"`
				} `json:"sql"`
				Oracle struct {
					ArchiveLogs         string      `json:"archiveLogs"`
					RetainLogBackups    string      `json:"retainLogBackups"`
					UseGuestCredentials bool        `json:"useGuestCredentials"`
					CredentialsID       interface{} `json:"credentialsId"`
					DeleteHoursCount    interface{} `json:"deleteHoursCount"`
					DeleteGBsCount      interface{} `json:"deleteGBsCount"`
					BackupLogs          bool        `json:"backupLogs"`
					BackupMinsCount     int         `json:"backupMinsCount"`
					KeepDaysCount       int         `json:"keepDaysCount"`
					LogShippingServers  struct {
						AutoSelection     bool          `json:"autoSelection"`
						ShippingServerIds []interface{} `json:"shippingServerIds"`
					} `json:"logShippingServers"`
				} `json:"oracle"`
				Exclusions struct {
					ExclusionPolicy string        `json:"exclusionPolicy"`
					ItemsList       []interface{} `json:"itemsList"`
				} `json:"exclusions"`
				Scripts struct {
					ScriptProcessingMode string      `json:"scriptProcessingMode"`
					WindowsScripts       interface{} `json:"windowsScripts"`
					LinuxScripts         interface{} `json:"linuxScripts"`
				} `json:"scripts"`
			} `json:"appSettings"`
		} `json:"appAwareProcessing"`
		GuestFSIndexing struct {
			IsEnabled        bool          `json:"isEnabled"`
			IndexingSettings []interface{} `json:"indexingSettings"`
		} `json:"guestFSIndexing"`
		GuestInteractionProxies struct {
			AutoSelection bool          `json:"autoSelection"`
			ProxyIds      []interface{} `json:"proxyIds"`
		} `json:"guestInteractionProxies"`
		GuestCredentials struct {
			CredsType             string        `json:"credsType"`
			CredsID               string        `json:"credsId"`
			CredentialsPerMachine []interface{} `json:"credentialsPerMachine"`
		} `json:"guestCredentials"`
	} `json:"guestProcessing"`
	Schedule struct {
		RunAutomatically bool `json:"runAutomatically"`
		Daily            struct {
			DailyKind string   `json:"dailyKind"`
			IsEnabled bool     `json:"isEnabled"`
			LocalTime string   `json:"localTime"`
			Days      []string `json:"days"`
		} `json:"daily"`
		Monthly struct {
			DayOfWeek        string   `json:"dayOfWeek"`
			DayNumberInMonth string   `json:"dayNumberInMonth"`
			IsEnabled        bool     `json:"isEnabled"`
			LocalTime        string   `json:"localTime"`
			DayOfMonth       int      `json:"dayOfMonth"`
			Months           []string `json:"months"`
		} `json:"monthly"`
		Periodically struct {
			PeriodicallyKind string `json:"periodicallyKind"`
			IsEnabled        bool   `json:"isEnabled"`
			Frequency        int    `json:"frequency"`
			BackupWindow     struct {
				Days []struct {
					Day   string `json:"day"`
					Hours string `json:"hours"`
				} `json:"days"`
			} `json:"backupWindow"`
			StartTimeWithinAnHour int `json:"startTimeWithinAnHour"`
		} `json:"periodically"`
		Continuously struct {
			IsEnabled    bool `json:"isEnabled"`
			BackupWindow struct {
				Days []struct {
					Day   string `json:"day"`
					Hours string `json:"hours"`
				} `json:"days"`
			} `json:"backupWindow"`
		} `json:"continuously"`
		AfterThisJob struct {
			IsEnabled bool        `json:"isEnabled"`
			JobName   interface{} `json:"jobName"`
		} `json:"afterThisJob"`
		Retry struct {
			IsEnabled    bool `json:"isEnabled"`
			RetryCount   int  `json:"retryCount"`
			AwaitMinutes int  `json:"awaitMinutes"`
		} `json:"retry"`
		BackupWindow struct {
			IsEnabled    bool `json:"isEnabled"`
			BackupWindow struct {
				Days []struct {
					Day   string `json:"day"`
					Hours string `json:"hours"`
				} `json:"days"`
			} `json:"backupWindow"`
		} `json:"backupWindow"`
	} `json:"schedule"`
	Type        string `json:"type"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDisabled  bool   `json:"isDisabled"`
}

type Includes struct {
	Type     string `json:"type"`
	HostName string `json:"hostName"`
	Name     string `json:"name"`
	ObjectID string `json:"objectId"`
}

type VirtualMachines struct {
	Includes []Includes `json:"includes"`
	Excludes struct {
		Vms   []interface{} `json:"vms"`
		Disks []struct {
			DisksToProcess string `json:"disksToProcess"`
			VMObject       struct {
				Type     string `json:"type"`
				HostName string `json:"hostName"`
				Name     string `json:"name"`
				ObjectID string `json:"objectId"`
			} `json:"vmObject"`
			Disks                     []interface{} `json:"disks"`
			RemoveFromVMConfiguration bool          `json:"removeFromVMConfiguration"`
		} `json:"disks"`
		Templates struct {
			IsEnabled              bool `json:"isEnabled"`
			ExcludeFromIncremental bool `json:"excludeFromIncremental"`
		} `json:"templates"`
	} `json:"excludes"`
}

type VbrJobPost struct {
	IsHighPriority  bool            `json:"isHighPriority"`
	VirtualMachines VirtualMachines `json:"virtualMachines"`
	Storage         struct {
		BackupRepositoryID string `json:"backupRepositoryId"`
		BackupProxies      struct {
			AutoSelection bool          `json:"autoSelection"`
			ProxyIds      []interface{} `json:"proxyIds"`
		} `json:"backupProxies"`
		RetentionPolicy struct {
			Type     string `json:"type"`
			Quantity int    `json:"quantity"`
		} `json:"retentionPolicy"`
		GfsPolicy struct {
			IsEnabled bool `json:"isEnabled"`
			Weekly    struct {
				DesiredTime          string `json:"desiredTime"`
				IsEnabled            bool   `json:"isEnabled"`
				KeepForNumberOfWeeks int    `json:"keepForNumberOfWeeks"`
			} `json:"weekly"`
			Monthly struct {
				DesiredTime           string `json:"desiredTime"`
				IsEnabled             bool   `json:"isEnabled"`
				KeepForNumberOfMonths int    `json:"keepForNumberOfMonths"`
			} `json:"monthly"`
			Yearly struct {
				DesiredTime          string `json:"desiredTime"`
				IsEnabled            bool   `json:"isEnabled"`
				KeepForNumberOfYears int    `json:"keepForNumberOfYears"`
			} `json:"yearly"`
		} `json:"gfsPolicy"`
		AdvancedSettings struct {
			BackupModeType  string `json:"backupModeType"`
			SynthenticFulls struct {
				IsEnabled bool     `json:"isEnabled"`
				Days      []string `json:"days"`
			} `json:"synthenticFulls"`
			ActiveFulls struct {
				IsEnabled bool `json:"isEnabled"`
				Weekly    struct {
					IsEnabled bool     `json:"isEnabled"`
					Days      []string `json:"days"`
				} `json:"weekly"`
				Monthly struct {
					DayOfWeek        string   `json:"dayOfWeek"`
					DayNumberInMonth string   `json:"dayNumberInMonth"`
					IsEnabled        bool     `json:"isEnabled"`
					DayOfMonths      int      `json:"dayOfMonths"`
					Months           []string `json:"months"`
				} `json:"monthly"`
			} `json:"activeFulls"`
			BackupHealth struct {
				IsEnabled bool `json:"isEnabled"`
				Weekly    struct {
					IsEnabled bool     `json:"isEnabled"`
					Days      []string `json:"days"`
				} `json:"weekly"`
				Monthly struct {
					DayOfWeek        string   `json:"dayOfWeek"`
					DayNumberInMonth string   `json:"dayNumberInMonth"`
					IsEnabled        bool     `json:"isEnabled"`
					DayOfMonths      int      `json:"dayOfMonths"`
					Months           []string `json:"months"`
				} `json:"monthly"`
			} `json:"backupHealth"`
			FullBackupMaintenance struct {
				RemoveData struct {
					IsEnabled bool `json:"isEnabled"`
					AfterDays int  `json:"afterDays"`
				} `json:"RemoveData"`
				DefragmentAndCompact struct {
					IsEnabled bool `json:"isEnabled"`
					Weekly    struct {
						IsEnabled bool     `json:"isEnabled"`
						Days      []string `json:"days"`
					} `json:"weekly"`
					Monthly struct {
						DayOfWeek        string   `json:"dayOfWeek"`
						DayNumberInMonth string   `json:"dayNumberInMonth"`
						IsEnabled        bool     `json:"isEnabled"`
						DayOfMonths      int      `json:"dayOfMonths"`
						Months           []string `json:"months"`
					} `json:"monthly"`
				} `json:"defragmentAndCompact"`
			} `json:"fullBackupMaintenance"`
			StorageData struct {
				CompressionLevel         string `json:"compressionLevel"`
				StorageOptimization      string `json:"storageOptimization"`
				EnableInlineDataDedup    bool   `json:"enableInlineDataDedup"`
				ExcludeSwapFileBlocks    bool   `json:"excludeSwapFileBlocks"`
				ExcludeDeletedFileBlocks bool   `json:"excludeDeletedFileBlocks"`
				Encryption               struct {
					IsEnabled                  bool        `json:"isEnabled"`
					EncryptionPasswordIDOrNull string      `json:"encryptionPasswordIdOrNull"`
					EncryptionPasswordTag      interface{} `json:"encryptionPasswordTag"`
				} `json:"encryption"`
			} `json:"storageData"`
			Notifications struct {
				SendSNMPNotifications bool `json:"sendSNMPNotifications"`
				EmailNotifications    struct {
					NotificationType           interface{}   `json:"notificationType"`
					IsEnabled                  bool          `json:"isEnabled"`
					Recipients                 []interface{} `json:"recipients"`
					CustomNotificationSettings interface{}   `json:"customNotificationSettings"`
				} `json:"emailNotifications"`
				VMAttribute struct {
					IsEnabled              bool   `json:"isEnabled"`
					Notes                  string `json:"notes"`
					AppendToExisitingValue bool   `json:"appendToExisitingValue"`
				} `json:"vmAttribute"`
			} `json:"notifications"`
			VSphere struct {
				EnableVMWareToolsQuiescence bool `json:"enableVMWareToolsQuiescence"`
				ChangedBlockTracking        struct {
					IsEnabled              bool `json:"isEnabled"`
					EnableCBTautomatically bool `json:"enableCBTautomatically"`
					ResetCBTonActiveFull   bool `json:"resetCBTonActiveFull"`
				} `json:"changedBlockTracking"`
			} `json:"vSphere"`
			StorageIntegration struct {
				IsEnabled                bool `json:"isEnabled"`
				LimitProcessedVM         bool `json:"limitProcessedVm"`
				LimitProcessedVMCount    int  `json:"limitProcessedVmCount"`
				FailoverToStandardBackup bool `json:"failoverToStandardBackup"`
			} `json:"storageIntegration"`
			Scripts struct {
				PeriodicityType string `json:"periodicityType"`
				PreCommand      struct {
					IsEnabled bool   `json:"isEnabled"`
					Command   string `json:"command"`
				} `json:"preCommand"`
				PostCommand struct {
					IsEnabled bool   `json:"isEnabled"`
					Command   string `json:"command"`
				} `json:"postCommand"`
				RunScriptEvery int      `json:"runScriptEvery"`
				DayOfWeek      []string `json:"dayOfWeek"`
			} `json:"scripts"`
		} `json:"advancedSettings"`
	} `json:"storage"`
	GuestProcessing struct {
		AppAwareProcessing struct {
			IsEnabled   bool `json:"isEnabled"`
			AppSettings []struct {
				Vss             string `json:"vss"`
				TransactionLogs string `json:"transactionLogs"`
				VMObject        struct {
					Type     string `json:"type"`
					HostName string `json:"hostName"`
					Name     string `json:"name"`
					ObjectID string `json:"objectId"`
				} `json:"vmObject"`
				UsePersistentGuestAgent bool `json:"usePersistentGuestAgent"`
				Sql                     struct {
					LogsProcessing     string      `json:"logsProcessing"`
					RetainLogBackups   interface{} `json:"retainLogBackups"`
					BackupMinsCount    interface{} `json:"backupMinsCount"`
					KeepDaysCount      interface{} `json:"keepDaysCount"`
					LogShippingServers interface{} `json:"logShippingServers"`
				} `json:"sql"`
				Oracle struct {
					ArchiveLogs         string      `json:"archiveLogs"`
					RetainLogBackups    string      `json:"retainLogBackups"`
					UseGuestCredentials bool        `json:"useGuestCredentials"`
					CredentialsID       interface{} `json:"credentialsId"`
					DeleteHoursCount    interface{} `json:"deleteHoursCount"`
					DeleteGBsCount      interface{} `json:"deleteGBsCount"`
					BackupLogs          bool        `json:"backupLogs"`
					BackupMinsCount     int         `json:"backupMinsCount"`
					KeepDaysCount       int         `json:"keepDaysCount"`
					LogShippingServers  struct {
						AutoSelection     bool          `json:"autoSelection"`
						ShippingServerIds []interface{} `json:"shippingServerIds"`
					} `json:"logShippingServers"`
				} `json:"oracle"`
				Exclusions struct {
					ExclusionPolicy string        `json:"exclusionPolicy"`
					ItemsList       []interface{} `json:"itemsList"`
				} `json:"exclusions"`
				Scripts struct {
					ScriptProcessingMode string      `json:"scriptProcessingMode"`
					WindowsScripts       interface{} `json:"windowsScripts"`
					LinuxScripts         interface{} `json:"linuxScripts"`
				} `json:"scripts"`
			} `json:"appSettings"`
		} `json:"appAwareProcessing"`
		GuestFSIndexing struct {
			IsEnabled        bool          `json:"isEnabled"`
			IndexingSettings []interface{} `json:"indexingSettings"`
		} `json:"guestFSIndexing"`
		GuestInteractionProxies struct {
			AutoSelection bool          `json:"autoSelection"`
			ProxyIds      []interface{} `json:"proxyIds"`
		} `json:"guestInteractionProxies"`
		GuestCredentials struct {
			CredsType             string        `json:"credsType"`
			CredsID               string        `json:"credsId"`
			CredentialsPerMachine []interface{} `json:"credentialsPerMachine"`
		} `json:"guestCredentials"`
	} `json:"guestProcessing"`
	Schedule struct {
		RunAutomatically bool `json:"runAutomatically"`
		Daily            struct {
			DailyKind string   `json:"dailyKind"`
			IsEnabled bool     `json:"isEnabled"`
			LocalTime string   `json:"localTime"`
			Days      []string `json:"days"`
		} `json:"daily"`
		Monthly struct {
			DayOfWeek        string   `json:"dayOfWeek"`
			DayNumberInMonth string   `json:"dayNumberInMonth"`
			IsEnabled        bool     `json:"isEnabled"`
			LocalTime        string   `json:"localTime"`
			DayOfMonth       int      `json:"dayOfMonth"`
			Months           []string `json:"months"`
		} `json:"monthly"`
		Periodically struct {
			PeriodicallyKind string `json:"periodicallyKind"`
			IsEnabled        bool   `json:"isEnabled"`
			Frequency        int    `json:"frequency"`
			BackupWindow     struct {
				Days []struct {
					Day   string `json:"day"`
					Hours string `json:"hours"`
				} `json:"days"`
			} `json:"backupWindow"`
			StartTimeWithinAnHour int `json:"startTimeWithinAnHour"`
		} `json:"periodically"`
		Continuously struct {
			IsEnabled    bool `json:"isEnabled"`
			BackupWindow struct {
				Days []struct {
					Day   string `json:"day"`
					Hours string `json:"hours"`
				} `json:"days"`
			} `json:"backupWindow"`
		} `json:"continuously"`
		AfterThisJob struct {
			IsEnabled bool        `json:"isEnabled"`
			JobName   interface{} `json:"jobName"`
		} `json:"afterThisJob"`
		Retry struct {
			IsEnabled    bool `json:"isEnabled"`
			RetryCount   int  `json:"retryCount"`
			AwaitMinutes int  `json:"awaitMinutes"`
		} `json:"retry"`
		BackupWindow struct {
			IsEnabled    bool `json:"isEnabled"`
			BackupWindow struct {
				Days []struct {
					Day   string `json:"day"`
					Hours string `json:"hours"`
				} `json:"days"`
			} `json:"backupWindow"`
		} `json:"backupWindow"`
	} `json:"schedule"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDisabled  bool   `json:"isDisabled"`
}