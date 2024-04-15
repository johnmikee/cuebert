package shared

import (
	"fmt"
	"strings"
	"time"

	"github.com/johnmikee/cuebert/bot"
	"github.com/johnmikee/cuebert/bot/ops"
	"github.com/johnmikee/cuebert/cuebert/config"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

var (
	titleText  = slack.NewTextBlockObject(slack.PlainTextType, "Cuebert", false, false)
	closeText  = slack.NewTextBlockObject(slack.PlainTextType, "Close", false, false)
	submitText = slack.NewTextBlockObject(slack.PlainTextType, "Submit", false, false)
	yes        = slack.NewTextBlockObject(slack.PlainTextType, "Yes", false, false)
	no         = slack.NewTextBlockObject(slack.PlainTextType, "No", false, false)
)

func (c *Shared) LoadInput(ctx *slacker.InteractionContext, loadType string) {
	values := ctx.Callback().View.State.Values

	var (
		taskBefore     = c.flags.Task
		taskImplBefore = c.flags.TaskImplementation
	)

	config.LogFlags(c.log, c.flags)
	c.flags.AuthUsers = ops.OptionalStrs(values["auth_users"]["auth_users_opt"].SelectedUsers, c.flags.AuthUsers)
	c.flags.AuthUsersFromIDP = ops.OptionalBool(values["auth_idp"]["auth_idp"].SelectedOption, c.flags.AuthUsersFromIDP)
	c.flags.LogLevel = ops.OptionalStrSelect(values["log_level"]["log_level_opt"].SelectedOption, c.flags.LogLevel)
	c.flags.LogToFile = ops.OptionalBool(values["log_to_file"]["log_to_file"].SelectedOption, c.flags.LogToFile)
	c.flags.Testing = ops.OptionalBool(values["testing"]["testing"].SelectedOption, c.flags.Testing)
	c.flags.TestingStartTime = ops.OptionalStr(values["testing_start"]["testing_start"].SelectedTime, c.flags.TestingStartTime)
	c.flags.TestingEndTime = ops.OptionalStr(values["testing_end"]["testing_end"].SelectedTime, c.flags.TestingEndTime)
	c.flags.Deadline = ops.OptionalStr(values["date_picker"]["date_picker"].SelectedDate, c.flags.Deadline)
	c.flags.Task = strings.ToLower(ops.OptionalStrSelect(values["task"]["task_opt"].SelectedOption, c.flags.Task))
	c.flags.TaskImplementation = strings.ToLower(ops.OptionalStrSelect(values["task_implementation"]["task_implementation_opt"].SelectedOption, c.flags.TaskImplementation))
	c.flags.TaskID = ops.OptionalStr(values["task_id"]["task_id"].Value, c.flags.TaskID)
	c.flags.CutoffTime = ops.OptionalStr(values["cutoff_time"]["cutoff_time_picker"].SelectedTime, c.flags.CutoffTime)
	c.flags.DiffInterval = ops.OptionalInt(values["diff_interval"]["diff_interval"].Value, c.flags.DiffInterval)
	c.flags.CheckInterval = ops.OptionalInt(values["check_interval"]["check_interval"].Value, c.flags.CheckInterval)
	c.flags.PollInterval = ops.OptionalInt(values["poll_interval"]["poll_interval"].Value, c.flags.PollInterval)
	c.flags.HelpDocsURL = ops.OptionalStr(values["docs_url"]["docs_url"].Value, c.flags.HelpDocsURL)
	c.flags.HelpRepoURL = ops.OptionalStr(values["help_url"]["help_url"].Value, c.flags.HelpRepoURL)
	c.flags.HelpTicketURL = ops.OptionalStr(values["repo_url"]["repo_url"].Value, c.flags.HelpTicketURL)

	// if they change use those to reload or start the program
	if taskBefore != c.flags.Task || taskImplBefore != c.flags.TaskImplementation {
		loadType = c.constFormat(Reload)
	}

	switch loadType {
	case c.constFormat(Start):
		c.lifecycle.Start()
	case c.constFormat(Reload):
		c.lifecycle.Update()
	}
}

type Loader struct {
	TriggerID string
	LoadType  string
	Tables    []string
}

func (c *Shared) LoadProgram(l *Loader, b ...*slack.InputBlock) {
	headerText := slack.NewTextBlockObject(slack.MarkdownType, "Modify the program", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)
	// yes/no options
	//
	// auth from IDP
	authIDPSection := slack.NewTextBlockObject(slack.PlainTextType, "Pull auth users from iDP?", false, false)
	authIDPYesOption := slack.NewOptionBlockObject("auth_idp_yes", yes, nil)
	authIDPNoOption := slack.NewOptionBlockObject("auth_idp_no", no, nil)
	authIDPOpts := slack.NewRadioButtonsBlockElement("auth_idp", authIDPYesOption, authIDPNoOption)
	authIDPBlock := slack.NewInputBlock("auth_idp", authIDPSection, nil, authIDPOpts)
	authIDPBlock.Optional = true
	// clearing the database
	clearDBSection := slack.NewTextBlockObject(slack.PlainTextType, "Clear the database?", false, false)
	clearDBYesOption := slack.NewOptionBlockObject("clear_db_yes", yes, nil)
	clearDBNoOption := slack.NewOptionBlockObject("clear_db_no", no, nil)
	cleardDBOpts := slack.NewRadioButtonsBlockElement("clear_db", clearDBYesOption, clearDBNoOption)
	clearDBBlock := slack.NewInputBlock("clear_db", clearDBSection, nil, cleardDBOpts)
	clearDBBlock.Optional = true
	// logtofile?
	logToFileSection := slack.NewTextBlockObject(slack.PlainTextType, "Log to file?", false, false)
	logToFileYesOption := slack.NewOptionBlockObject("log_to_file_yes", yes, nil)
	logToFileNoOption := slack.NewOptionBlockObject("log_to_file_no", no, nil)
	logToFileOpts := slack.NewRadioButtonsBlockElement("log_to_file", logToFileYesOption, logToFileNoOption)
	logToFileBlock := slack.NewInputBlock("log_to_file", logToFileSection, nil, logToFileOpts)
	logToFileBlock.Optional = true
	// are we testing?
	testingSection := slack.NewTextBlockObject(slack.PlainTextType, "Testing?", false, false)
	testingYesOption := slack.NewOptionBlockObject("testing_yes", yes, nil)
	testingNoOption := slack.NewOptionBlockObject("testing_no", no, nil)
	testingOpts := slack.NewRadioButtonsBlockElement("testing", testingYesOption, testingNoOption)
	testingBlock := slack.NewInputBlock("testing", testingSection, nil, testingOpts)
	testingBlock.Optional = true
	// multi-select options
	//
	// auth users
	authUsers := slack.NewTextBlockObject(slack.PlainTextType, "Which users are authorized to administrate Cuebert?", false, false)
	authUsersOptBlock := slack.NewOptionsMultiSelectBlockElement(slack.MultiOptTypeUser, nil, "auth_users_opt")
	authUsersBlock := slack.NewInputBlock("auth_users", authUsers, nil, authUsersOptBlock)
	authUsersBlock.Optional = true
	// logLevel
	logLevel := slack.NewTextBlockObject(slack.PlainTextType, "Log Level", false, false)
	logLevelOptions := bot.CreateOptionBlockObjects([]string{"info", "debug", "trace", "warn", "error"})
	logLevelOptBlock := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, nil, "log_level_opt", logLevelOptions...)
	logLevelBlock := slack.NewInputBlock("log_level", logLevel, nil, logLevelOptBlock)
	logLevelBlock.Optional = true
	// testing users.
	testingUsers := slack.NewTextBlockObject(slack.PlainTextType, "Testing Users", false, false)
	testingUsersOptBlock := slack.NewOptionsMultiSelectBlockElement(slack.MultiOptTypeUser, nil, "testing_users_opt")
	testingUsersBlock := slack.NewInputBlock("testing_users", testingUsers, nil, testingUsersOptBlock)
	testingUsersBlock.Optional = true
	// input options
	//
	// set the cutoff time
	cutoffTime := slack.NewTextBlockObject(slack.PlainTextType, "Cutoff Time", false, false)
	cutoffTimePlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "ex: 5:00 PM", false, false)
	cutoffTimeBlock := slack.NewTimePickerBlockElement("cutoff_time_picker")
	cutoffTimeInput := slack.NewInputBlock("cutoff_time", cutoffTime, cutoffTimePlaceHolder, cutoffTimeBlock)
	cutoffTimeInput.Optional = true
	// set deadline date dd:mm:yyyy
	today := time.Now().Format("2006-01-02")

	deadline := slack.NewTextBlockObject(slack.PlainTextType, "Date", false, false)
	dealineOpts := slack.NewTextBlockObject(slack.PlainTextType, fmt.Sprintf("ex: %s", today), false, false)
	deadlineBox := slack.NewDatePickerBlockElement("date_picker")
	deadlineBlock := slack.NewInputBlock("date_picker", deadline, dealineOpts, deadlineBox)
	deadlineBlock.Optional = true
	// set the required version
	requiredVersion := slack.NewTextBlockObject(slack.PlainTextType, "Required Version", false, false)
	requiredVersionPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "ex: 13.4", false, false)
	requiredVersionBlock := slack.NewPlainTextInputBlockElement(requiredVersionPlaceHolder, "required_version")
	requiredVersionInput := slack.NewInputBlock("required_version", requiredVersion, nil, requiredVersionBlock)
	requiredVersionInput.Optional = true
	// testing start time
	testingStart := slack.NewTextBlockObject(slack.PlainTextType, "Testing Start Time", false, false)
	testingStartPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "ex: 9:00 AM", false, false)
	testingStartBlock := slack.NewTimePickerBlockElement("testing_start")
	testingStartInput := slack.NewInputBlock("testing_start", testingStart, testingStartPlaceHolder, testingStartBlock)
	testingStartInput.Optional = true
	// testing end time
	testingEnd := slack.NewTextBlockObject(slack.PlainTextType, "Testing End Time", false, false)
	testingEndPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "ex: 5:00 PM", false, false)
	testingEndBlock := slack.NewTimePickerBlockElement("testing_end")
	testingEndInput := slack.NewInputBlock("testing_end", testingEnd, testingEndPlaceHolder, testingEndBlock)
	testingEndInput.Optional = true
	// check interval
	checkInterval := slack.NewTextBlockObject(slack.PlainTextType, "Check Interval", false, false)
	checkIntervalPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "ex: 15m", false, false)
	checkIntervalBlock := slack.NewPlainTextInputBlockElement(checkIntervalPlaceHolder, "check_interval")
	checkIntervalInput := slack.NewInputBlock("check_interval", checkInterval, nil, checkIntervalBlock)
	checkIntervalInput.Optional = true
	// interval for Diff
	DiffInterval := slack.NewTextBlockObject(slack.PlainTextType, "Diff Interval", false, false)
	DiffIntervalPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "ex: 15m", false, false)
	DiffIntervalBlock := slack.NewPlainTextInputBlockElement(DiffIntervalPlaceHolder, "diff_interval")
	DiffIntervalInput := slack.NewInputBlock("diff_interval", DiffInterval, nil, DiffIntervalBlock)
	DiffIntervalInput.Optional = true
	// interval for poll
	pollInterval := slack.NewTextBlockObject(slack.PlainTextType, "Poll Interval", false, false)
	pollIntervalPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "ex: 15m", false, false)
	pollIntervalBlock := slack.NewPlainTextInputBlockElement(pollIntervalPlaceHolder, "poll_interval")
	pollIntervalInput := slack.NewInputBlock("poll_interval", pollInterval, nil, pollIntervalBlock)
	pollIntervalInput.Optional = true
	// task choices
	task := slack.NewTextBlockObject(slack.PlainTextType, "Task", false, false)
	taskOptions := bot.CreateOptionBlockObjects([]string{"OS", "Compliance"})
	taskOptBlock := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, nil, "task_opt", taskOptions...)
	taskBlock := slack.NewInputBlock("task", task, nil, taskOptBlock)
	taskBlock.Optional = true
	// task implementation choices
	taskImplementation := slack.NewTextBlockObject(slack.PlainTextType, "Task Implementation", false, false)
	taskImplementationOptions := bot.CreateOptionBlockObjects([]string{"Drata", "Lessonly", "Employee Handbook"})
	taskImplementationOptBlock := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, nil, "task_implementation_opt", taskImplementationOptions...)
	taskImplementationBlock := slack.NewInputBlock("task_implementation", taskImplementation, nil, taskImplementationOptBlock)
	taskImplementationBlock.Optional = true
	// task ID input
	taskID := slack.NewTextBlockObject(slack.PlainTextType, "Task ID", false, false)
	taskIDPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "ex: 12345", false, false)
	taskIDBlock := slack.NewPlainTextInputBlockElement(taskIDPlaceHolder, "task_id")
	taskIDInput := slack.NewInputBlock("task_id", taskID, nil, taskIDBlock)
	taskIDInput.Optional = true
	// help modal inputs
	//
	// docs url
	docsURL := slack.NewTextBlockObject(slack.PlainTextType, "Docs URL", false, false)
	docsURLPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "ex: https://docs.example.com", false, false)
	docsURLBlock := slack.NewPlainTextInputBlockElement(docsURLPlaceHolder, "docs_url")
	docsURLInput := slack.NewInputBlock("docs_url", docsURL, nil, docsURLBlock)
	docsURLInput.Optional = true
	// help url
	helpURL := slack.NewTextBlockObject(slack.PlainTextType, "Help URL", false, false)
	helpURLPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "ex: https://help.example.com", false, false)
	helpURLBlock := slack.NewPlainTextInputBlockElement(helpURLPlaceHolder, "help_url")
	helpURLInput := slack.NewInputBlock("help_url", helpURL, nil, helpURLBlock)
	helpURLInput.Optional = true
	// code repo url
	repoURL := slack.NewTextBlockObject(slack.PlainTextType, "Repo URL", false, false)
	repoURLPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "ex: https://github.com/example/example", false, false)
	repoURLBlock := slack.NewPlainTextInputBlockElement(repoURLPlaceHolder, "repo_url")
	repoURLInput := slack.NewInputBlock("repo_url", repoURL, nil, repoURLBlock)
	repoURLInput.Optional = true
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
			logLevelBlock,
			testingUsersBlock,
			slack.NewDividerBlock(),
			taskBlock,
			taskImplementationBlock,
			slack.NewDividerBlock(),
			cutoffTimeInput,
			deadlineBlock,
			requiredVersionInput,
			testingStartInput,
			testingEndInput,
			checkIntervalInput,
			DiffIntervalInput,
			pollIntervalInput,
			slack.NewDividerBlock(),
		},
	}

	// add any additional blocks
	for _, block := range b {
		blocks.BlockSet = append(blocks.BlockSet, block)
	}

	blocks.BlockSet = append(blocks.BlockSet,
		slack.NewDividerBlock(),
		docsURLInput,
		helpURLInput,
		repoURLInput,
	)

	modalRequest := slack.ModalViewRequest{
		Type:       slack.ViewType(slack.VTModal),
		Title:      titleText,
		Close:      closeText,
		Submit:     submitText,
		Blocks:     blocks,
		CallbackID: l.LoadType,
	}

	_, err := c.bot.Bot().SlackClient().OpenView(l.TriggerID, modalRequest)
	if err != nil {
		c.log.Error().Err(err).Msg("error opening modal")
	}
}
