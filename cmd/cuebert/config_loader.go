package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/johnmikee/cuebert/db"
	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

func ynToBool(s string) bool {
	return strings.ToLower(s) == "yes"
}

func valToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

func (b *Bot) optsToStrs(opts []slack.OptionBlockObject) string {
	var strs []string
	for _, o := range opts {
		strs = append(strs, o.Value)
	}

	if len(strs) > 1 {
		return strings.Join(strs, ",")
	}

	return strs[0]
}

func usersToStrs(users []string) string {
	if len(users) > 1 {
		return strings.Join(users, ",")
	}

	return users[0]
}

const (
	START  = "start_settings_modal"
	RELOAD = "reload_settings_modal"
)

// take the values submitted by loadProgram and set the config for the program
func (b *Bot) loadInput(s *slacker.Slacker, event *socketmode.Event, callback *slack.InteractionCallback, loadType string) {
	values := callback.View.State.Values
	tables := values["table_names"]["table_names_opt"].SelectedOptions
	level := values["log_level"]["log_level_opt"].SelectedOption.Text.Text

	b.cfg.authUsers = values["auth_users"]["auth_users_opt"].SelectedUsers
	b.cfg.flags.authUsers = usersToStrs(values["auth_users"]["auth_users_opt"].SelectedUsers)
	b.cfg.flags.authUsersFromIDP = ynToBool(values["auth_idp"]["auth_idp"].SelectedOption.Text.Text)
	b.cfg.flags.clearTables = ynToBool(values["clear_db"]["clear_db"].SelectedOption.Text.Text)
	b.cfg.flags.tableNames = b.optsToStrs(tables)
	b.cfg.flags.logLevel = level
	b.cfg.flags.logToFile = ynToBool(values["log_to_file"]["log_to_file"].SelectedOption.Text.Text)
	b.cfg.flags.requiredVers = values["required_version"]["required_version"].Value
	b.cfg.flags.testing = ynToBool(values["testing"]["testing"].SelectedOption.Text.Text)
	b.cfg.flags.testingStartTime = values["testing_start"]["testing_start"].SelectedTime
	b.cfg.flags.testingEndTime = values["testing_end"]["testing_end"].SelectedTime
	b.cfg.testUsers = values["testing_users"]["testing_users_opt"].SelectedUsers
	b.cfg.flags.testingUsers = usersToStrs(values["testing_users"]["testing_users_opt"].SelectedUsers)
	b.cfg.flags.deadline = values["date_picker"]["date_picker"].SelectedDate
	b.cfg.flags.cutoffTime = values["cutoff_time"]["cutoff_time_picker"].SelectedTime
	b.cfg.flags.deviceDiffInterval = valToInt(values["device_diff_interval"]["device_diff_interval"].Value)
	b.cfg.flags.checkInterval = valToInt(values["check_interval"]["check_interval"].Value)
	b.cfg.flags.pollInterval = valToInt(values["poll_interval"]["poll_interval"].Value)
	b.cfg.flags.helpDocsURL = values["docs_url"]["docs_url"].Value
	b.cfg.flags.helpRepoURL = values["help_url"]["help_url"].Value
	b.cfg.flags.helpTicketURL = values["repo_url"]["repo_url"].Value

	s.SocketMode().Ack(*event.Request)

	switch loadType {
	case START:
		b.start()
	case RELOAD:
		b.update()
	}
}

// loadProgram is the modal that admins can use to modify the configuration of the program
// from within slack. the results of this modal will either update the config or start the program.
//
// this takes the flags from the config and creates a modal with the current values.
// note: this modal is only available to admins.
func (b *Bot) loadProgram(triggerID, loadType string) {
	headerText := slack.NewTextBlockObject(slack.MarkdownType, "Modify the program", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	// yes/no options
	//
	yes := slack.NewTextBlockObject(slack.PlainTextType, "Yes", false, false)
	no := slack.NewTextBlockObject(slack.PlainTextType, "No", false, false)
	// auth from IDP
	authIDPSection := slack.NewTextBlockObject(slack.PlainTextType, "Pull auth users from iDP?", false, false)
	authIDPYesOption := slack.NewOptionBlockObject("auth_idp_yes", yes, nil)
	authIDPNoOption := slack.NewOptionBlockObject("auth_idp_no", no, nil)
	authIDPOpts := slack.NewRadioButtonsBlockElement("auth_idp", authIDPYesOption, authIDPNoOption)
	authIDPBlock := slack.NewInputBlock("auth_idp", authIDPSection, nil, authIDPOpts)
	// clearing the database
	clearDBSection := slack.NewTextBlockObject(slack.PlainTextType, "Clear the database?", false, false)
	clearDBYesOption := slack.NewOptionBlockObject("clear_db_yes", yes, nil)
	clearDBNoOption := slack.NewOptionBlockObject("clear_db_no", no, nil)
	cleardDBOpts := slack.NewRadioButtonsBlockElement("clear_db", clearDBYesOption, clearDBNoOption)
	clearDBBlock := slack.NewInputBlock("clear_db", clearDBSection, nil, cleardDBOpts)
	// logtofile?
	logToFileSection := slack.NewTextBlockObject(slack.PlainTextType, "Log to file?", false, false)
	logToFileYesOption := slack.NewOptionBlockObject("log_to_file_yes", yes, nil)
	logToFileNoOption := slack.NewOptionBlockObject("log_to_file_no", no, nil)
	logToFileOpts := slack.NewRadioButtonsBlockElement("log_to_file", logToFileYesOption, logToFileNoOption)
	logToFileBlock := slack.NewInputBlock("log_to_file", logToFileSection, nil, logToFileOpts)
	// are we testing?
	testingSection := slack.NewTextBlockObject(slack.PlainTextType, "Testing?", false, false)
	testingYesOption := slack.NewOptionBlockObject("testing_yes", yes, nil)
	testingNoOption := slack.NewOptionBlockObject("testing_no", no, nil)
	testingOpts := slack.NewRadioButtonsBlockElement("testing", testingYesOption, testingNoOption)
	testingBlock := slack.NewInputBlock("testing", testingSection, nil, testingOpts)

	// multi-select options
	//
	// auth users
	authUsers := slack.NewTextBlockObject(slack.PlainTextType, "Which users are authorized to administrate Cuebert?", false, false)
	authUsersOptBlock := slack.NewOptionsMultiSelectBlockElement(slack.MultiOptTypeUser, nil, "auth_users_opt")
	authUsersBlock := slack.NewInputBlock("auth_users", authUsers, nil, authUsersOptBlock)
	// logLevel
	logLevel := slack.NewTextBlockObject(slack.PlainTextType, "Log Level", false, false)
	logLevelOptions := createOptionBlockObjects([]string{"info", "debug", "trace", "warn", "error"})
	logLevelOptBlock := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, nil, "log_level_opt", logLevelOptions...)
	logLevelBlock := slack.NewInputBlock("log_level", logLevel, nil, logLevelOptBlock)
	// table names
	tableNames := slack.NewTextBlockObject(slack.PlainTextType, "Table Names", false, false)
	tableNamesOptions := createOptionBlockObjects(db.CueTables)
	tableNamesOptBlock := slack.NewOptionsMultiSelectBlockElement(slack.MultiOptTypeStatic, nil, "table_names_opt", tableNamesOptions...)
	tableNamesBlock := slack.NewInputBlock("table_names", tableNames, nil, tableNamesOptBlock)
	// testing users.
	testingUsers := slack.NewTextBlockObject(slack.PlainTextType, "Testing Users", false, false)
	testingUsersOptBlock := slack.NewOptionsMultiSelectBlockElement(slack.MultiOptTypeUser, nil, "testing_users_opt")
	testingUsersBlock := slack.NewInputBlock("testing_users", testingUsers, nil, testingUsersOptBlock)

	// input options
	//
	// set the cutoff time
	cutoffTime := slack.NewTextBlockObject(slack.PlainTextType, "Cutoff Time", false, false)
	cutoffTimePlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "ex: 5:00 PM", false, false)
	cutoffTimeBlock := slack.NewTimePickerBlockElement("cutoff_time_picker")
	cutoffTimeInput := slack.NewInputBlock("cutoff_time", cutoffTime, cutoffTimePlaceHolder, cutoffTimeBlock)
	// set deadline date dd:mm:yyyy
	today := time.Now().Format("2006-01-02")

	deadline := slack.NewTextBlockObject(slack.PlainTextType, "Date", false, false)
	dealineOpts := slack.NewTextBlockObject(slack.PlainTextType, fmt.Sprintf("ex: %s", today), false, false)
	deadlineBox := slack.NewDatePickerBlockElement("date_picker")
	deadlineBlock := slack.NewInputBlock("date_picker", deadline, dealineOpts, deadlineBox)
	// set the required version
	requiredVersion := slack.NewTextBlockObject(slack.PlainTextType, "Required Version", false, false)
	requiredVersionPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "ex: 13.4", false, false)
	requiredVersionBlock := slack.NewPlainTextInputBlockElement(requiredVersionPlaceHolder, "required_version")
	requiredVersionInput := slack.NewInputBlock("required_version", requiredVersion, nil, requiredVersionBlock)
	// testing start time
	testingStart := slack.NewTextBlockObject(slack.PlainTextType, "Testing Start Time", false, false)
	testingStartPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "ex: 9:00 AM", false, false)
	testingStartBlock := slack.NewTimePickerBlockElement("testing_start")
	testingStartInput := slack.NewInputBlock("testing_start", testingStart, testingStartPlaceHolder, testingStartBlock)
	// testing end time
	testingEnd := slack.NewTextBlockObject(slack.PlainTextType, "Testing End Time", false, false)
	testingEndPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "ex: 5:00 PM", false, false)
	testingEndBlock := slack.NewTimePickerBlockElement("testing_end")
	testingEndInput := slack.NewInputBlock("testing_end", testingEnd, testingEndPlaceHolder, testingEndBlock)
	// check interval
	checkInterval := slack.NewTextBlockObject(slack.PlainTextType, "Check Interval", false, false)
	checkIntervalPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "ex: 15m", false, false)
	checkIntervalBlock := slack.NewPlainTextInputBlockElement(checkIntervalPlaceHolder, "check_interval")
	checkIntervalInput := slack.NewInputBlock("check_interval", checkInterval, nil, checkIntervalBlock)
	// interval for deviceDiff
	deviceDiffInterval := slack.NewTextBlockObject(slack.PlainTextType, "Device Diff Interval", false, false)
	deviceDiffIntervalPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "ex: 15m", false, false)
	deviceDiffIntervalBlock := slack.NewPlainTextInputBlockElement(deviceDiffIntervalPlaceHolder, "device_diff_interval")
	deviceDiffIntervalInput := slack.NewInputBlock("device_diff_interval", deviceDiffInterval, nil, deviceDiffIntervalBlock)
	// interval for poll
	pollInterval := slack.NewTextBlockObject(slack.PlainTextType, "Poll Interval", false, false)
	pollIntervalPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "ex: 15m", false, false)
	pollIntervalBlock := slack.NewPlainTextInputBlockElement(pollIntervalPlaceHolder, "poll_interval")
	pollIntervalInput := slack.NewInputBlock("poll_interval", pollInterval, nil, pollIntervalBlock)

	// help modal inputs
	//
	// docs url
	docsURL := slack.NewTextBlockObject(slack.PlainTextType, "Docs URL", false, false)
	docsURLPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "ex: https://docs.example.com", false, false)
	docsURLBlock := slack.NewPlainTextInputBlockElement(docsURLPlaceHolder, "docs_url")
	docsURLInput := slack.NewInputBlock("docs_url", docsURL, nil, docsURLBlock)
	// help url
	helpURL := slack.NewTextBlockObject(slack.PlainTextType, "Help URL", false, false)
	helpURLPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "ex: https://help.example.com", false, false)
	helpURLBlock := slack.NewPlainTextInputBlockElement(helpURLPlaceHolder, "help_url")
	helpURLInput := slack.NewInputBlock("help_url", helpURL, nil, helpURLBlock)
	// code repo url
	repoURL := slack.NewTextBlockObject(slack.PlainTextType, "Repo URL", false, false)
	repoURLPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "ex: https://github.com/example/example", false, false)
	repoURLBlock := slack.NewPlainTextInputBlockElement(repoURLPlaceHolder, "repo_url")
	repoURLInput := slack.NewInputBlock("repo_url", repoURL, nil, repoURLBlock)

	// build the modal
	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			headerSection,
			slack.NewDividerBlock(),
			authIDPBlock,
			clearDBBlock,
			logToFileBlock,
			testingBlock,
			slack.NewDividerBlock(),
			authUsersBlock,
			tableNamesBlock,
			logLevelBlock,
			testingUsersBlock,
			slack.NewDividerBlock(),
			cutoffTimeInput,
			deadlineBlock,
			requiredVersionInput,
			testingStartInput,
			testingEndInput,
			checkIntervalInput,
			deviceDiffIntervalInput,
			pollIntervalInput,
			slack.NewDividerBlock(),
			docsURLInput,
			helpURLInput,
			repoURLInput,
		},
	}

	modalRequest := slack.ModalViewRequest{
		Type:       slack.ViewType(slack.VTModal),
		Title:      titleText,
		Close:      closeText,
		Submit:     submitText,
		Blocks:     blocks,
		CallbackID: loadType,
	}

	// open the modal
	vr, err := b.bot.Client().OpenView(triggerID, modalRequest)
	if err != nil {
		b.log.Error().Err(err).Msg("error opening modal")
	}

	b.log.Trace().Interface("view_response", vr).Msg("modal opened")
}

// loadPrompt leads us to the modal to update cuebert
func (b *Bot) loadPrompt(user, callback string) {
	b.modalGateway(
		&modalGateway{
			text:       "Do you want to update the configuration?",
			callbackID: callback,
			yesName:    "yes_update_config",
			yesText:    "Yes",
			yesValue:   "yes_update_config",
			yesStyle:   "primary",
			noName:     "no_update_config",
			noText:     "No",
			noValue:    "no_update_config",
			noStyle:    "danger",
			channel:    user,
			msg:        "updating config prompt",
		},
	)
}
