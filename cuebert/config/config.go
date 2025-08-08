package config

// Config holds the sensitive values for the program.
type Config struct {
	AdminGroupID           string `json:"admin_group_id"`
	DBAddress              string `json:"db_address"`
	DBName                 string `json:"db_name"`
	DBPass                 string `json:"db_pass"`
	DBPort                 string `json:"db_port"`
	DBUser                 string `json:"db_user"`
	GoogleAdminEmail       string `json:"google_admin_email"`
	GoogleServiceCreds     string `json:"google_service_creds"` // base64 encoded
	GoogleSheetID          string `json:"google_sheet_id"`
	IDPToken               string `json:"idp_token"`
	IDPURL                 string `json:"idp_url"`
	IDPBlockGroup          string `json:"idp_block_group"`
	MDMKey                 string `json:"mdm_key"`
	MDMURL                 string `json:"mdm_url"`
	SlackAppToken          string `json:"slack_app_token"`
	SlackAlertChannel      string `json:"slack_alert_channel"`
	SlackBotToken          string `json:"slack_bot_token"`
	SlackBotID             string `json:"slack_bot_id"`
	Subdomain              string `json:"subdomain"`
	ComplianceServiceToken string `json:"compliance_service_token"`
	DrataToken             string `json:"drata_token"`
	LessonlyToken          string `json:"lessonly_token"`
}
