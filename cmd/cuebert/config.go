package main

import (
	"github.com/johnmikee/cuebert/pkg/logger"
)

// CuebertConfig is a struct to hold config for the program
type CuebertConfig struct {
	AdminGroupID      string `json:"admin_group_id"`
	DBAddress         string `json:"db_address"`
	DBName            string `json:"db_name"`
	DBPass            string `json:"db_pass"`
	DBPort            string `json:"db_port"`
	DBUser            string `json:"db_user"`
	IDPDomain         string `json:"idp_domain"`
	IDPToken          string `json:"idp_token"`
	IDPURL            string `json:"idp_url"`
	MDMKey            string `json:"mdm_key"`
	MDMURL            string `json:"mdm_url"`
	SlackAppToken     string `json:"slack_app_token"`
	SlackAlertChannel string `json:"slack_alert_channel"`
	SlackBotToken     string `json:"slack_bot_token"`
	SlackBotID        string `json:"slack_bot_id"`

	authUsers []string
	testUsers []string
	flags     *Flags
	log       logger.Logger
}
