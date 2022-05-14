package vbrmodels

type VbrJobStates struct {
	Data []struct {
		ID             string `json:"id" yaml:"id"`
		Name           string `json:"name" yaml:"name"`
		Type           string `json:"type" yaml:"type"`
		Description    string `json:"description" yaml:"description"`
		Status         string `json:"status" yaml:"status"`
		LastRun        string `json:"lastRun" yaml:"lastRun"`
		LastResult     string `json:"lastResult" yaml:"lastResult"`
		NextRun        string `json:"nextRun" yaml:"nextRun"`
		Workload       string `json:"workload" yaml:"workload"`
		RepositoryID   string `json:"repositoryId" yaml:"repositoryId"`
		RepositoryName string `json:"repositoryName" yaml:"repositoryName"`
		ObjectsCount   int    `json:"objectsCount" yaml:"objectsCount"`
		SessionID      string `json:"sessionId" yaml:"sessionId"`
	} `json:"data" yaml:"data"`
	Pagination struct {
		Total int `json:"total" yaml:"total"`
		Count int `json:"count" yaml:"count"`
		Skip  int `json:"skip" yaml:"skip"`
		Limit int `json:"limit" yaml:"limit"`
	} `json:"pagination" yaml:"pagination"`
}

type VbrJobs struct {
	Data []struct {
		IsHighPriority  bool   `json:"isHighPriority" yaml:"isHighPriority"`
		Type            string `json:"type" yaml:"type"`
		Name            string `json:"name" yaml:"name"`
		ID              string `json:"id" yaml:"id"`
		Description     string `json:"description" yaml:"description"`
		IsDisabled      bool   `json:"isDisabled" yaml:"isDisabled"`
		VirtualMachines struct {
			Includes []struct {
				InventoryObject struct {
					HostName string `json:"hostName" yaml:"hostName"`
					Name     string `json:"name" yaml:"name"`
					Type     string `json:"type"  yaml:"type"`
					ObjectID string `json:"objectId" yaml:"objectId"`
				} `json:"inventoryObject" yaml:"inventoryObject"`
				Size string `json:"size" yaml:"size"`
			} `json:"includes" yaml:"includes"`
			Excludes struct {
				Vms []struct {
					InventoryObject struct {
						HostName string `json:"hostName" yaml:"hostName"`
						Name     string `json:"name" yaml:"name"`
						Type     string `json:"type" yaml:"type"`
						ObjectID string `json:"objectId" yaml:"objectId"`
					} `json:"inventoryObject" yaml:"inventoryObject"`
					Size string `json:"size" yaml:"size"`
				} `json:"vms" yaml:"vms"`
				Disks []struct {
					VMObject struct {
						HostName string `json:"hostName" yaml:"hostName"`
						Name     string `json:"name" yaml:"name"`
						Type     string `json:"type" yaml:"type"`
						ObjectID string `json:"objectId" yaml:"objectId"`
					} `json:"vmObject" yaml:"vmObject"`
					DisksToProcess            string   `json:"disksToProcess" yaml:"disksToProcess"`
					Disks                     []string `json:"disks" yaml:"disks"`
					RemoveFromVMConfiguration bool     `json:"removeFromVMConfiguration" yaml:"removeVMConfiguration"`
				} `json:"disks" yaml:"disks"`
				Templates struct {
					IsEnabled              bool `json:"isEnabled" yaml:"isEnabled"`
					ExcludeFromIncremental bool `json:"excludeFromIncremental" yaml:"excludeFromIncremental"`
				} `json:"templates" yaml:"templates"`
			} `json:"excludes" yaml:"excludes"`
		} `json:"virtualMachines" yaml:"virtualMachines"`
		Storage struct {
			BackupRepositoryID string `json:"backupRepositoryId" yaml:"backupRepositoryId"`
			BackupProxies      struct {
				AutoSelection bool     `json:"autoSelection" yaml:"autoSelection"`
				ProxyIds      []string `json:"proxyIds" yaml:"proxyIds"`
			} `json:"backupProxies" yaml:"backupProxies"`
			RetentionPolicy struct {
				Type     string `json:"type" yaml:"type"`
				Quantity int    `json:"quantity" yaml:"quantity"`
			} `json:"retentionPolicy" yaml:"retentionPolicy"`
			GfsPolicy struct {
				IsEnabled bool `json:"isEnabled" yaml:"isEnabled"`
				Weekly    struct {
					IsEnabled            bool   `json:"isEnabled" yaml:"isEnabled"`
					KeepForNumberOfWeeks int    `json:"keepForNumberOfWeeks" yaml:"keyForNumberOfWeeks"`
					DesiredTime          string `json:"desiredTime" yaml:"desiredTime"`
				} `json:"weekly" yaml:"weekly"`
				Monthly struct {
					IsEnabled             bool   `json:"isEnabled" yaml:"isEnabled"`
					KeepForNumberOfMonths int    `json:"keepForNumberOfMonths" yaml:"keepForNumberOfMonths"`
					DesiredTime           string `json:"desiredTime" yaml:"desiredTime"`
				} `json:"monthly" yaml:"monthly"`
				Yearly struct {
					IsEnabled            bool   `json:"isEnabled" yaml:"isEnabled"`
					KeepForNumberOfYears int    `json:"keepForNumberOfYears" yaml:"keepforNumberOfYears"`
					DesiredTime          string `json:"desiredTime" yaml:"desiredTime"`
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
						IsEnabled        bool     `json:"isEnabled" yaml:"isEnabled"`
						DayOfWeek        string   `json:"dayOfWeek" yaml:"dayOfWeek"`
						DayNumberInMonth string   `json:"dayNumberInMonth" yaml:"dayNumberInMonth"`
						DayOfMonths      int      `json:"dayOfMonths" yaml:"dayOfMonths"`
						Months           []string `json:"months" yaml:"months"`
					} `json:"monthly" yaml:"months"`
				} `json:"activeFulls" yaml:"activeFulls"`
				BackupHealth struct {
					IsEnabled bool `json:"isEnabled" yaml:"isEnabled"`
					Weekly    struct {
						IsEnabled bool     `json:"isEnabled" yaml:"isEnabled"`
						Days      []string `json:"days" yaml:"days"`
					} `json:"weekly" yaml:"weekly"`
					Monthly struct {
						IsEnabled        bool     `json:"isEnabled" yaml:"isEnabled"`
						DayOfWeek        string   `json:"dayOfWeek" yaml:"dayOfWeek"`
						DayNumberInMonth string   `json:"dayNumberInMonth" yaml:"dayNumberInMonth"`
						DayOfMonths      int      `json:"dayOfMonths" yaml:"dayOfMonth"`
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
							IsEnabled        bool     `json:"isEnabled" yaml:"isEnabled"`
							DayOfWeek        string   `json:"dayOfWeek" yaml:"dayOfWeek"`
							DayNumberInMonth string   `json:"dayNumberInMonth" yaml:"dayNumberInMonth"`
							DayOfMonths      int      `json:"dayOfMonths" yaml:"dayOfMonths"`
							Months           []string `json:"months" yaml:"months"`
						} `json:"monthly" yaml:"monthly"`
					} `json:"defragmentAndCompact" yaml:"defragmentAndCompact"`
				} `json:"fullBackupMaintenance" yaml:"fullBackupMaintenance"`
				StorageData struct {
					EnableInlineDataDedup    bool   `json:"enableInlineDataDedup" yaml:"enabledInlineDataDedup"`
					ExcludeSwapFileBlocks    bool   `json:"excludeSwapFileBlocks" yaml:"excludeSwapFileBlocks"`
					ExcludeDeletedFileBlocks bool   `json:"excludeDeletedFileBlocks" yaml:"excludeDeletedFileBlocks"`
					CompressionLevel         string `json:"compressionLevel" yaml:"compressionLevel"`
					StorageOptimization      string `json:"storageOptimization" yaml:"storageOptimization"`
					Encryption               struct {
						IsEnabled                  bool   `json:"isEnabled" yaml:"isEnabled"`
						EncryptionPasswordIDOrNull string `json:"encryptionPasswordIdOrNull" yaml:"encryptionPasswordIdOrNull"`
						EncryptionPasswordTag      string `json:"encryptionPasswordTag" yaml:"encryptionPasswordTag"`
					} `json:"encryption" yaml:"encryption"`
				} `json:"storageData" yaml:"storageData"`
				Notifications struct {
					SendSNMPNotifications bool `json:"sendSNMPNotifications"`
					EmailNotifications    struct {
						IsEnabled                  bool     `json:"isEnabled"`
						Recipients                 []string `json:"recipients"`
						NotificationType           string   `json:"notificationType"`
						CustomNotificationSettings struct {
							Subject                            string `json:"subject"`
							NotifyOnSuccess                    bool   `json:"notifyOnSuccess"`
							NotifyOnWarning                    bool   `json:"notifyOnWarning"`
							NotifyOnError                      bool   `json:"notifyOnError"`
							SuppressNotificationUntilLastRetry bool   `json:"SuppressNotificationUntilLastRetry"`
						} `json:"customNotificationSettings" yaml:"customNotificationSettings"`
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
					PreCommand struct {
						IsEnabled bool   `json:"isEnabled" yaml:"isEnabled"`
						Command   string `json:"command" yaml:"command"`
					} `json:"preCommand" yaml:"preCommand"`
					PostCommand struct {
						IsEnabled bool   `json:"isEnabled" yaml:"isEnabled"`
						Command   string `json:"command" yaml:"command"`
					} `json:"postCommand" yaml:"postCommand"`
					PeriodicityType string   `json:"periodicityType" yaml:"periodicityType"`
					RunScriptEvery  int      `json:"runScriptEvery" yaml:"runScriptEvery"`
					DayOfWeek       []string `json:"dayOfWeek" yaml:"dayOfWeek"`
				} `json:"scripts" yaml:"scripts"`
			} `json:"advancedSettings" yaml:"advancedSettings"`
		} `json:"storage" yaml:"storage"`
		GuestProcessing struct {
			AppAwareProcessing struct {
				IsEnabled   bool `json:"isEnabled" yaml:"isEnabled"`
				AppSettings []struct {
					VMObject struct {
						HostName string `json:"hostName" yaml:"hostName"`
						Name     string `json:"name" yaml:"name"`
						Type     string `json:"type" yaml:"type"`
						ObjectID string `json:"objectId" yaml:"objectId"`
					} `json:"vmObject" yaml:"vmObject"`
					Vss                     string `json:"vss" yaml:"vss"`
					UsePersistentGuestAgent bool   `json:"usePersistentGuestAgent" yaml:"usePersistentGuestAgent"`
					TransactionLogs         string `json:"transactionLogs" yaml:"transactionLogs"`
					SQL                     struct {
						LogsProcessing     string `json:"logsProcessing" yaml:"logsProcessing"`
						BackupMinsCount    int    `json:"backupMinsCount" yaml:"backupMinsCount"`
						RetainLogBackups   string `json:"retainLogBackups" yaml:"retainLogBackups"`
						KeepDaysCount      int    `json:"keepDaysCount" yaml:"keepDaysCount"`
						LogShippingServers struct {
							AutoSelection     bool     `json:"autoSelection" yaml:"autoSelection"`
							ShippingServerIds []string `json:"shippingServerIds" yaml:"shippingServerIds"`
						} `json:"logShippingServers" yaml:"logShippingServers"`
					} `json:"sql" yaml:"sql"`
					Oracle struct {
						UseGuestCredentials bool   `json:"useGuestCredentials" yaml:"useGuestCredentials"`
						CredentialsID       string `json:"credentialsId" yaml:"credentialsId"`
						ArchiveLogs         string `json:"archiveLogs" yaml:"archiveLogs"`
						DeleteHoursCount    int    `json:"deleteHoursCount" yaml:"deleteHoursCount"`
						DeleteGBsCount      int    `json:"deleteGBsCount" yaml:"deleteGBsCount"`
						BackupLogs          bool   `json:"backupLogs" yaml:"backupLogs"`
						BackupMinsCount     int    `json:"backupMinsCount" yaml:"backupMinsCount"`
						RetainLogBackups    string `json:"retainLogBackups" yaml:"retainLogBackups"`
						KeepDaysCount       int    `json:"keepDaysCount" yaml:"keepDaysCount"`
						LogShippingServers  struct {
							AutoSelection     bool     `json:"autoSelection" yaml:"autoSelection"`
							ShippingServerIds []string `json:"shippingServerIds" yaml:"shippingServerIds"`
						} `json:"logShippingServers" yaml:"logShippingServers"`
					} `json:"oracle" yaml:"oracle"`
					Exclusions struct {
						ExclusionPolicy string   `json:"exclusionPolicy" yaml:"exclusionPolicy"`
						ItemsList       []string `json:"itemsList" yaml:"itemsList"`
					} `json:"exclusions" yaml:"exclusions"`
					Scripts struct {
						ScriptProcessingMode string `json:"scriptProcessingMode" yaml:"scriptProcessingMode"`
						WindowsScripts       struct {
							PreFreezeScript string `json:"preFreezeScript" yaml:"preFreezeScript"`
							PostThawScript  string `json:"postThawScript" yaml:"postThawScript"`
						} `json:"windowsScripts" yaml:"windowsScripts"`
						LinuxScripts struct {
							PreFreezeScript string `json:"preFreezeScript" yaml:"preFreezeScript"`
							PostThawScript  string `json:"postThawScript" yaml:"postThawScript"`
						} `json:"linuxScripts" yaml:"linuxScripts"`
					} `json:"scripts" yaml:"scripts"`
				} `json:"appSettings" yaml:"appSettings"`
			} `json:"appAwareProcessing" yaml:"appAwareProcessing"`
			GuestFSIndexing struct {
				IsEnabled        bool `json:"isEnabled" yaml:"isEnabled"`
				IndexingSettings []struct {
					VMObject struct {
						HostName string `json:"hostName" yaml:"hostName"`
						Name     string `json:"name" yaml:"name"`
						Type     string `json:"type" yaml:"type"`
						ObjectID string `json:"objectId" yaml:"objectId"`
					} `json:"vmObject" yaml:"vmObject"`
					WindowsIndexing struct {
						GuestFSIndexingMode string   `json:"guestFSIndexingMode" yaml:"guestFSIndexingMode"`
						IndexingList        []string `json:"indexingList" yaml:"indexingList"`
					} `json:"WindowsIndexing" yaml:"WindowsIndexing"`
					LinuxIndexing struct {
						GuestFSIndexingMode string   `json:"guestFSIndexingMode" yaml:"guestFSIndexingMode"`
						IndexingList        []string `json:"indexingList" yaml:"indexingList"`
					} `json:"LinuxIndexing" yaml:"LinuxIndexing"`
				} `json:"indexingSettings" yaml:"indexingSettings"`
			} `json:"guestFSIndexing" yaml:"guestFSIndexing"`
			GuestInteractionProxies struct {
				AutoSelection bool     `json:"autoSelection" yaml:"autoSelection"`
				ProxyIds      []string `json:"proxyIds" yaml:"proxyIds"`
			} `json:"guestInteractionProxies" yaml:"guestInteractionProxies"`
			GuestCredentials struct {
				CredsID               string `json:"credsId" yaml:"credsId"`
				CredsType             string `json:"credsType" yaml:"credsType"`
				CredentialsPerMachine []struct {
					WindowsCredsID string `json:"windowsCredsId" yaml:"windowsCredsId"`
					LinuxCredsID   string `json:"linuxCredsId" yaml:"linuxCredsId"`
					VMObject       struct {
						HostName string `json:"hostName" yaml:"hostName"`
						Name     string `json:"name" yaml:"name"`
						Type     string `json:"type" yaml:"type"`
						ObjectID string `json:"objectId" yaml:"objectId"`
					} `json:"vmObject" yaml:"vmObject"`
				} `json:"credentialsPerMachine" yaml:"credentialsPerMachine"`
			} `json:"guestCredentials" yaml:"guestCredentials"`
		} `json:"guestProcessing" yaml:"guestProcessing"`
		Schedule struct {
			RunAutomatically bool `json:"runAutomatically" yaml:"runAutomatically"`
			Daily            struct {
				IsEnabled bool     `json:"isEnabled" yaml:"isEnabled"`
				LocalTime string   `json:"localTime" yaml:"localTIme"`
				DailyKind string   `json:"dailyKind" yaml:"dailyKind"`
				Days      []string `json:"days" yaml:"days"`
			} `json:"daily" yaml:"daily"`
			Monthly struct {
				IsEnabled        bool     `json:"isEnabled" yaml:"isEnabled"`
				LocalTime        string   `json:"localTime" yaml:"localTime"`
				DayOfWeek        string   `json:"dayOfWeek" yaml:"dayOfWeek"`
				DayNumberInMonth string   `json:"dayNumberInMonth" yaml:"dayNumnberInMonth"`
				DayOfMonth       int      `json:"dayOfMonth" yaml:"dayOfMonth"`
				Months           []string `json:"months" yaml:"months"`
			} `json:"monthly" yaml:"monthly"`
			Periodically struct {
				IsEnabled        bool   `json:"isEnabled" yaml:"isEnabled"`
				PeriodicallyKind string `json:"periodicallyKind" yaml:"periodicallyKind"`
				Frequency        int    `json:"frequency" yaml:"frequency"`
				BackupWindow     struct {
					Days []struct {
						Day   string `json:"day" yaml:"day"`
						Hours string `json:"hours" yaml:"hours"`
					} `json:"days" yaml:"days"`
				} `json:"backupWindow" yaml:"backupWindow"`
			} `json:"periodically" yaml:"periodically"`
			Continuously struct {
				IsEnabled     bool `json:"isEnabled" yaml:"isEnabled"`
				WindowSetting struct {
					Days []struct {
						Day   string `json:"day" yaml:"day"`
						Hours string `json:"hours" yaml:"hours"`
					} `json:"days" yaml:"days"`
				} `json:"WindowSetting" yaml:"WindowSetting"`
			} `json:"continuously" yaml:"continuously"`
			AfterThisJob struct {
				IsEnabled bool   `json:"isEnabled" yaml:"isEnabled"`
				JobName   string `json:"jobName" yaml:"jobName"`
			} `json:"afterThisJob" yaml:"afterThisJob"`
			Retry struct {
				IsEnabled    bool `json:"isEnabled" yaml:"isEnabled"`
				RetryCount   int  `json:"retryCount" yaml:"retryCount"`
				AwaitMinutes int  `json:"awaitMinutes" yaml:"awaitMinutes"`
			} `json:"retry" yaml:"retry"`
			BackupWindow struct {
				IsEnabled     bool `json:"isEnabled" yaml:"isEnabled"`
				WindowSetting struct {
					Days []struct {
						Day   string `json:"day" yaml:"day"`
						Hours string `json:"hours" yaml:"hours"`
					} `json:"days" yaml:"days"`
				} `json:"WindowSetting" yaml:"WindowSettings"`
			} `json:"backupWindow" yaml:"backupWindow"`
		} `json:"schedule" yaml:"schedule"`
	} `json:"data" yaml:"data"`
}
