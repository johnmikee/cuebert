package kandji

import "time"

// DeviceDetails returns details for a device.
//   - https://api.kandji.io/#efa2170d-e5f7-4b97-8f4c-da6f84ba58b5
type DeviceDetails struct {
	General          General          `json:"general"`
	Mdm              MDM              `json:"mdm"`
	ActivationLock   ActivationLock   `json:"activation_lock"`
	Filevault        Filevault        `json:"filevault"`
	HardwareOverview HardwareOverview `json:"hardware_overview"`
	Users            Users            `json:"users"`
}

// DeviceResult is returned when listing devices
//   - https://api.kandji.io/#78209960-31a7-4e3b-a2c0-95c7e65bb5f9
type DeviceResult struct {
	DeviceID        string      `json:"device_id"`
	DeviceName      string      `json:"device_name"`
	Model           string      `json:"model"`
	SerialNumber    string      `json:"serial_number"`
	Platform        string      `json:"platform"`
	OSVersion       string      `json:"os_version"`
	LastCheckIn     *time.Time  `json:"last_check_in"`
	User            *UserUnion  `json:"user"`
	AssetTag        interface{} `json:"asset_tag"`
	BlueprintID     string      `json:"blueprint_id"`
	MdmEnabled      bool        `json:"mdm_enabled"`
	AgentInstalled  bool        `json:"agent_installed"`
	IsMissing       bool        `json:"is_missing"`
	IsRemoved       bool        `json:"is_removed"`
	AgentVersion    string      `json:"agent_version"`
	FirstEnrollment string      `json:"first_enrollment"`
	LastEnrollment  string      `json:"last_enrollment"`
	BlueprintName   string      `json:"blueprint_name"`
}

// ActivationLock stores information on activation lock data on the device
type ActivationLock struct {
	BypassCodeFailed                     bool `json:"bypass_code_failed"`
	UserActivationLockEnabled            bool `json:"user_activation_lock_enabled"`
	DeviceActivationLockEnabled          bool `json:"device_activation_lock_enabled"`
	ActivationLockAllowedWhileSupervised bool `json:"activation_lock_allowed_while_supervised"`
	ActivationLockSupported              bool `json:"activation_lock_supported"`
}

// BlueprintName is the name of the blueprint assigned to the device
type BlueprintName string

// Filvault holds information on the filevault status of the machine
type Filevault struct {
	FilevaultEnabled         bool   `json:"filevault_enabled"`
	FilevaultRecoverykeyType string `json:"filevault_recoverykey_type"`
	FilevaultPrkEscrowed     bool   `json:"filevault_prk_escrowed"`
	FilevaultNextRotation    string `json:"filevault_next_rotation"`
	FilevaultRegenRequired   bool   `json:"filevault_regen_required"`
}

// General holds the general purpose information of the device
type General struct {
	DeviceID        string `json:"device_id"`
	DeviceName      string `json:"device_name"`
	LastEnrollment  string `json:"last_enrollment"`
	FirstEnrollment string `json:"first_enrollment"`
	Model           string `json:"model"`
	Platform        string `json:"platform"`
	OSVersion       string `json:"os_version"`
	SystemVersion   string `json:"system_version"`
	BootVolume      string `json:"boot_volume"`
	TimeSinceBoot   string `json:"time_since_boot"`
	LastUser        string `json:"last_user"`
	AssetTag        string `json:"asset_tag"`
	AssignedUser    string `json:"assigned_user"`
	BlueprintName   string `json:"blueprint_name"`
	BlueprintUUID   string `json:"blueprint_uuid"`
}

// HardwareOverview holds information related to the device hardware
type HardwareOverview struct {
	ModelName          string `json:"model_name"`
	ModelIdentifier    string `json:"model_identifier"`
	ProcessorName      string `json:"processor_name"`
	ProcessorSpeed     string `json:"processor_speed"`
	NumberOfProcessors string `json:"number_of_processors"`
	TotalNumberOfCores string `json:"total_number_of_cores"`
	Memory             string `json:"memory"`
	Udid               string `json:"udid"`
	SerialNumber       string `json:"serial_number"`
}

type Library struct {
	DeviceID     string         `json:"device_id"`
	LibraryItems []LibraryItems `json:"library_items"`
}

// LibraryItems holds information related to the library items on the device
type LibraryItems struct {
	ID                int         `json:"id"`
	Status            string      `json:"status"`
	ReportedAt        time.Time   `json:"reported_at"`
	Log               string      `json:"log"`
	LastAuditRun      time.Time   `json:"last_audit_run"`
	LastAuditLog      string      `json:"last_audit_log"`
	ControlLog        interface{} `json:"control_log"`
	ControlReportedAt interface{} `json:"control_reported_at"`
	ItemID            string      `json:"item_id"`
	Name              string      `json:"name"`
	Type              string      `json:"type"`
	Computer          Computer    `json:"computer"`
	Blueprint         Blueprint   `json:"blueprint"`
	RulesPresent      bool        `json:"rules_present"`
}

type Computer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type Blueprint struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// MDM holds information related to MDM data of the device
type MDM struct {
	MdmEnabled     string     `json:"mdm_enabled"`
	Supervised     string     `json:"supervised"`
	InstallDate    string     `json:"install_date"`
	LastCheckIn    *time.Time `json:"last_check_in"`
	MdmEnabledUser []string   `json:"mdm_enabled_user"`
}

// Users holds information about the console and system users on the machine
type Users struct {
	RegularUsers []User `json:"regular_users"`
	SystemUsers  []User `json:"system_users"`
}

// User represents a user object for the machine user
type User struct {
	Username string  `json:"username"`
	Uid      string  `json:"uid"`
	Path     string  `json:"path"`
	Name     *string `json:"name,omitempty"`
}
