package common

import (
	"github.com/johnmikee/cuebert/bot/ops"
	"github.com/shomali11/slacker/v2"
)

func (c *Common) LoadInput(ctx *slacker.InteractionContext, loadType string) {
	values := ctx.Callback().View.State.Values

	c.flags.AuthUsers = ops.OptionalStrs(values["auth_users"]["auth_users_opt"].SelectedUsers, c.flags.AuthUsers)
	c.flags.AuthUsersFromIDP = ops.OptionalBool(values["auth_idp"]["auth_idp"].SelectedOption, c.flags.AuthUsersFromIDP)
	c.flags.LogLevel = ops.OptionalStrSelect(values["log_level"]["log_level_opt"].SelectedOption, c.flags.LogLevel)
	c.flags.LogToFile = ops.OptionalBool(values["log_to_file"]["log_to_file"].SelectedOption, c.flags.LogToFile)
	c.flags.RequiredVers = ops.OptionalStr(values["required_version"]["required_version"].Value, c.flags.RequiredVers)
	c.flags.Testing = ops.OptionalBool(values["testing"]["testing"].SelectedOption, c.flags.Testing)
	c.flags.TestingStartTime = ops.OptionalStr(values["testing_start"]["testing_start"].SelectedTime, c.flags.TestingStartTime)
	c.flags.TestingEndTime = ops.OptionalStr(values["testing_end"]["testing_end"].SelectedTime, c.flags.TestingEndTime)
	c.flags.Deadline = ops.OptionalStr(values["date_picker"]["date_picker"].SelectedDate, c.flags.Deadline)
	c.flags.CutoffTime = ops.OptionalStr(values["cutoff_time"]["cutoff_time_picker"].SelectedTime, c.flags.CutoffTime)
	c.flags.DiffInterval = ops.OptionalInt(values["diff_interval"]["diff_interval"].Value, c.flags.DiffInterval)
	c.flags.CheckInterval = ops.OptionalInt(values["check_interval"]["check_interval"].Value, c.flags.CheckInterval)
	c.flags.PollInterval = ops.OptionalInt(values["poll_interval"]["poll_interval"].Value, c.flags.PollInterval)
	c.flags.HelpDocsURL = ops.OptionalStr(values["docs_url"]["docs_url"].Value, c.flags.HelpDocsURL)
	c.flags.HelpRepoURL = ops.OptionalStr(values["help_url"]["help_url"].Value, c.flags.HelpRepoURL)
	c.flags.HelpTicketURL = ops.OptionalStr(values["repo_url"]["repo_url"].Value, c.flags.HelpTicketURL)

	c.shared.LoadInput(ctx, loadType)
}
