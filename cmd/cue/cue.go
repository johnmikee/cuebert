package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/johnmikee/cuebert/internal/db"
	"github.com/johnmikee/cuebert/internal/device"
	"github.com/johnmikee/cuebert/internal/env"
	"github.com/johnmikee/cuebert/internal/user"
	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/mdm/client"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/johnmikee/cuebert/pkg/version"
	"github.com/rs/zerolog/log"
)

// CueFlags holds the flags for the program
type CueFlags struct {
	envType     string
	logLevel    string
	logToFile   bool
	mdm         string
	serviceName string
	reqVers     string
	reqDate     string
}

// Config holds the credentials for the various apis and db
type Config struct {
	DBAddress  string `json:"db_address"`
	DBName     string `json:"db_name"`
	DBPass     string `json:"db_pass"`
	DBPort     string `json:"db_port"`
	Domain     string `json:"domain"`
	DBUser     string `json:"db_user"`
	MDMKey     string `json:"mdm_key"`
	MDMURL     string `json:"mdm_url"`
	SlackToken string `json:"slack_token"`
	SlackUrl   string `json:"slack_url"`
}

// CueConfig holds the configuation for the program, flags, and connection to the db/
type CueConfig struct {
	conf  *Config
	db    *db.Config
	mdm   mdm.Provider
	flags *CueFlags
	log   logger.Logger

	devices *device.Device
	users   *user.User
}

func main() {
	cfg, args := loadEnv()
	if len(args) < 2 {
		usage()
		os.Exit(1)
	}

	// check for version before we load the connection
	switch strings.ToLower(args[1]) {
	case "version", "-version":
		version.Print()
		return
	}

	conn, err := db.Connect(
		&db.Conf{
			Host:     cfg.conf.DBAddress,
			Name:     cfg.conf.DBName,
			Password: cfg.conf.DBPass,
			Port:     cfg.conf.DBPort,
			User:     cfg.conf.DBUser,
		})

	if err != nil {
		cfg.log.Info().AnErr("connecting to db", err).Send()
		os.Exit(3)
	}
	mdmclient := client.New(
		&client.MDM{
			MDM: mdm.MDM(cfg.flags.mdm),
			Config: mdm.Config{
				Domain:                 cfg.conf.Domain,
				MDM:                    mdm.MDM(cfg.flags.mdm),
				URL:                    cfg.conf.MDMURL,
				Token:                  cfg.conf.MDMKey,
				Client:                 nil,
				Log:                    logger.Logger{},
				ProviderSpecificConfig: nil,
			},
		},
	)
	cfg.mdm = mdmclient
	cfg.db = db.New(conn.DB, &cfg.log)
	cfg.devices = device.New(
		&device.Config{
			Client: mdmclient,
			DB:     conn.DB,
			Log:    &cfg.log,
		},
	)
	cfg.users = user.New(
		&user.UserConfig{
			Client:     mdmclient,
			SlackToken: cfg.conf.SlackToken,
			SlackUrl:   cfg.conf.SlackUrl,
			DB:         conn.DB,
			Log:        &cfg.log,
		},
	)

	var run func([]string) error
	switch strings.ToLower(args[1]) {
	case "br":
		run = cfg.BotRes
	case "db":
		run = cfg.DB
	case "exclusions":
		run = cfg.Exclusions
	case "devices":
		run = cfg.Devices
	case "users":
		run = cfg.Users
	default:
		usage()
		os.Exit(1)
	}

	if err := run(args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// top level usage
func usage() {
	const help = `Usage: cue [global options] command [command args] <subcommand> [subcommand args]

Main commands:
  br            Interact with the bot_results table.
  db            Interact with the db storing information for cue.
  devices       Perform actions on devices or the devices table.
  exclusions    Perform actions on exclusions table.
  users         Perform actions on the users or the users table.
  version       An alias for the "version" subcommand.

Global options:
  -env-type={dev|prod}  Set the environment for the program. Options are dev or prod.
                        If using the dev environment any of the api tokens will be stored
                        in the system keychain.
  -log-level=info   	Set the level of the logger (defaults to info).
                        Available options:
                          * debug
                          * info
                          * trace
  -log-to-file=true     Set the output of the logger to a file.
  -required-date        The date the current required OS must be installed by.
  -required-os          The version devices must upgrade to.
  -service-name         Change the service name from cue to something custom.
  -help                 Show this help output, or the help for a specified subcommand.

Use cue <command> -h for additional usage of each command.
Example: cue db -h
	`
	fmt.Print(help)

}

// usageFor is used by all sub-comands
func usageFor(fs *flag.FlagSet, short string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  %s\n", short)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")
		w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(w, "\t-%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
		})
		w.Flush()
		fmt.Fprintf(os.Stderr, "\n")
	}
}

func loadEnv() (*CueConfig, []string) {
	globalFlags := &CueFlags{
		envType:     "dev",
		logLevel:    "info",
		logToFile:   false,
		mdm:         "kandji",
		serviceName: "cue",
		reqVers:     "",
		reqDate:     "",
	}
	flag.StringVar(
		&globalFlags.envType,
		"env-type",
		globalFlags.envType,
		"Set the env type. Options are [prod, dev].",
	)
	flag.StringVar(
		&globalFlags.logLevel,
		"log-level",
		globalFlags.logLevel,
		"Set the log level.",
	)
	flag.BoolVar(&globalFlags.logToFile,
		"log-to-file",
		globalFlags.logToFile,
		"Log results to file.")
	flag.StringVar(
		&globalFlags.mdm,
		"mdm",
		globalFlags.mdm,
		"Set the MDM to use. Options are [kandji, jamf].",
	)
	flag.StringVar(
		&globalFlags.reqVers,
		"required-os",
		globalFlags.reqVers,
		"The version devices must upgrade to.",
	)
	flag.StringVar(
		&globalFlags.reqDate,
		"required-date",
		globalFlags.reqDate,
		"The date the current required OS must be installed by.",
	)
	flag.StringVar(
		&globalFlags.serviceName,
		"service-name",
		globalFlags.serviceName,
		"if using the dev env the service name to store keys under.",
	)

	flag.Usage = func() {
		usage()
	}

	args := os.Args[0:]

	// yoink the global flags and set them
	for i := 0; i < len(args); i++ {
		switch {
		case strings.Contains(args[i], "env-type"):
			val, err := extractFlagValue("env-type", args)
			if err == nil {
				globalFlags.envType = val
				args = helpers.Remove(args, i)
				i--
			}
		case strings.Contains(args[i], "log-level"):
			val, err := extractFlagValue("log-level", args)
			if err == nil {
				globalFlags.logLevel = val
				args = helpers.Remove(args, i)
				i--
			}
		case strings.Contains(args[i], "log-to-file"):
			val, err := extractFlagValue("log-to-file", args)
			if err == nil {
				lb, err := strconv.ParseBool(val)
				if err == nil {
					globalFlags.logToFile = lb
					args = helpers.Remove(args, i)
					i--
				}
			}
		case strings.Contains(args[i], "service-name"):
			val, err := extractFlagValue("service-name", args)
			if err == nil {
				globalFlags.serviceName = val
				args = helpers.Remove(args, i)
				i--
			}
		case strings.Contains(args[i], "required-os"):
			val, err := extractFlagValue("required-os", args)
			if err == nil {
				globalFlags.reqVers = val
				args = helpers.Remove(args, i)
				i--
			}
		case strings.Contains(args[i], "required-date"):
			val, err := extractFlagValue("required-date", args)
			if err == nil {
				globalFlags.reqDate = val
				args = helpers.Remove(args, i)
				i--
			}
		}
	}

	err := flag.CommandLine.Parse(args)
	if err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	cfg := &CueConfig{
		flags: globalFlags,
		log: logger.NewLogger(&logger.Config{
			ToFile:  globalFlags.logToFile,
			Level:   globalFlags.logLevel,
			Service: globalFlags.serviceName,
			Env:     globalFlags.envType,
		},
		),
	}

	var config Config
	err = env.Get(
		env.EnvType(globalFlags.envType),
		&env.GetConfig{
			Name:         globalFlags.serviceName,
			Type:         env.JSON,
			EnvPrefix:    "CUE",
			ConfigStruct: &config,
		},
	)

	if err != nil {
		log.Info().AnErr("error", err).Msg("Error loading config from env")
		os.Exit(1)
	}

	cfg.conf = &config

	return cfg, args
}
