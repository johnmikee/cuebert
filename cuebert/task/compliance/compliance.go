package compliance

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/johnmikee/cuebert/cuebert/config"
	"github.com/johnmikee/cuebert/cuebert/task"
	"github.com/johnmikee/cuebert/cuebert/task/compliance/common"
	"github.com/johnmikee/cuebert/cuebert/task/compliance/db/tables"
	"github.com/johnmikee/cuebert/cuebert/task/implementation"
	"github.com/johnmikee/cuebert/cuebert/task/info"
	"github.com/johnmikee/cuebert/cuebert/task/shared"
	"github.com/johnmikee/cuebert/db/shared/admin"
	"github.com/johnmikee/cuebert/db/shared/users"
	"github.com/johnmikee/cuebert/idp"
	"github.com/slack-go/slack"

	"github.com/johnmikee/cuebert/handlers"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/shomali11/slacker/v2"
)

// Client implements task.Rabbit.
type Client struct {
	common *common.Common
	caller Callback
	config *task.Config
	impl   Compliance
	idp    idp.Provider
	log    logger.Logger
	shared *shared.Shared
	tables *tables.Config
}

// CheckNotSent implements task.Rabbit.
func (c *Client) CheckNotSent(serviceID string, slackID string) (bool, error) {
	return c.common.CheckNotSent(serviceID, slackID)
}

// FinalMessageSent implements task.Rabbit.
func (c *Client) FinalMessageSent(serviceID string, slackID string) error {
	return c.common.FinalMessageSent(serviceID, slackID)
}

// IsExcluded implements task.Rabbit.
func (c *Client) IsExcluded(email string) (bool, bool) {
	return c.common.IsExcluded(email)
}

// Compliance implements task.Rabbit specific to the compliance task.
type Compliance interface {
	task.Messaging
	CheckComplianceStatus(user string) (bool, error)
	CheckComplianceAll() ([]info.PersistUsers, error) // this should return a slice of info.Users who should be present in the database
	GetTasks() ([]info.Tasks, error)
	GetUsers() ([]info.Users, error)
	Setup(*task.Config, *common.Common, *tables.Config)
}

// Callback implements task.Rabbit.
type Callback interface {
	shared.Caller
}

func (c *Client) TaskImplementation() error {
	comps := []Compliance{}

	for _, m := range comps {
		m.Setup(c.config, c.common, c.tables)
		iu, err := m.CheckComplianceAll()
		if err != nil {
			return err
		}

		c.log.Info().Interface("iu", iu).Msg("iu")
	}

	return nil
}

// Setup implements task.Rabbit.
func (c *Client) Setup(t *task.Config) {
	c.config = t
	c.log = t.Log

	switch implementation.Item(t.Flags.TaskImplementation) {
	/*
		here you would check the task implementation and setting the compliance task
	*/
	}

	share := shared.New(
		&shared.Config{
			Config:    c.config.Config,
			Bot:       c.config.Bot,
			LifeCycle: c.config.LifeCycle,
			DB:        c.config.DB,
			Flags:     c.config.Flags,
			Log:       c.config.Log,
			Status:    c.config.StatusHandler,
		},
	)
	c.shared = share

	tables := tables.New(
		&tables.Tables{
			DB:     c.config.DB,
			Config: c.config.DBConfig,
			Log:    c.log,
		},
	)
	c.tables = tables

	common := common.New(
		&common.Config{
			Bot:    c.config.Bot,
			Config: c.config.Config,
			Flags:  c.config.Flags,
			Log:    c.config.Log,
			Tables: tables,
			Status: c.config.StatusHandler,
			Shared: share,
		},
	)
	c.common = common

	c.impl.Setup(c.config, common, tables)
	c.shared.Update(
		shared.WithCaller(common),
		shared.WithImplChecker(c.impl),
	)
}

// FirstMessageDelivered implements task.Rabbit.
func (c *Client) FirstMessageDelivered(serviceID string, slackID string) error {
	return c.common.FirstMessageDelivered(serviceID, slackID)
}

// FirstMessageSent implements task.Rabbit.
func (c *Client) FirstMessageSent(serviceID string, slackID string) error {
	return c.common.FirstMessageSent(serviceID, slackID)
}

// ReminderMessage implements task.Rabbit.
func (c *Client) ReminderMessage(r *shared.ReminderPayload) error {
	return c.common.ReminderMessage(r)
}

// UserCompliant implements task.Rabbit.
func (c *Client) UserCompliant(userID string) error {
	ui, err := c.tables.User().By().Opts(users.UserSlackID, userID).Query()
	if err != nil {
		return err
	}

	if ui.Empty() {
		return fmt.Errorf("user not found")
	}
	return c.idp.RemoveFromBlocked(ui[0].UserEmail)
}

// Messagingreturns the messaging interface.
func (c *Client) Messaging() task.Messaging {
	return c.impl
}

// Caller returns the caller interface.
func (c *Client) Caller() shared.Caller {
	return c.caller
}

// UpdateReminderTime implements task.Rabbit updating the reminder time.
func (c *Client) UpdateReminderTime(dateVal string, timeVal string, slackID string) error {
	return c.common.UpdateReminderTime(dateVal, timeVal, slackID)
}

// GetUserTZ implements task.Rabbit getting the user's timezone.
func (c *Client) GetUserTZ(slackID string) (int64, error) {
	return c.common.GetUserTZ(slackID)
}

// ExclusionAddBlocks builds the exclusion add blocks.
func (c *Client) ExclusionAddBlocks() []*slack.InputBlock {
	return c.common.ExclusionAddBlocks()
}

// ExclusionRequestDecision executes the exclusion request decision helper
func (c *Client) ExclusionRequestDecision(ctx *slacker.InteractionContext) {
	c.common.ExclusionRequestDecision(ctx)
}

// ExclusionSubmit submits the exclusion request
func (c *Client) ExclusionSubmit(ctx *slacker.InteractionContext) {
	c.common.ExclusionSubmit(ctx)
}

// ExclusionYes runs when an exclusion is approved
func (c *Client) ExclusionYes(ctx *slacker.InteractionContext) {
	c.common.ExclusionYes(ctx)
}

// MessageAcknowledged runs when the message has been acknowledged
func (c *Client) MessageAcknowledged(userID string, when time.Time, which string) error {
	return c.common.MessageAcknowledged(userID, when, which)
}

// UserExclusionSubmit submits the user exclusion request
func (c *Client) UserExclusionSubmit(ctx *slacker.InteractionContext) {
	c.common.UserExclusionSubmit(ctx)
}

// UserUpdate runs the user update modal
func (c *Client) UserUpdate(ctx *slacker.InteractionContext) {
	c.common.UserUpdate(ctx)
}

// UserUpdateSubmit submits the user update request
func (c *Client) UserUpdateSubmit(ctx *slacker.InteractionContext) {
	c.common.UserUpdateSubmit(ctx)
}

func (c *Client) FinalMessage() string {
	return c.impl.FinalMessage()
}

func (c *Client) FirstMessage() string {
	return c.impl.FirstMessage()
}

func (c *Client) FirstMessageHeader() string {
	return c.impl.FirstMessageHeader()
}

func (c *Client) Get() []task.Job {
	return []task.Job{
		{
			Job:        c.Diff,
			Name:       "Diff",
			When:       time.Duration(c.config.Flags.DiffInterval) * time.Minute,
			Continuous: false,
		},
		{
			Job:        c.Check,
			Name:       "Check",
			When:       time.Duration(c.config.Flags.CheckInterval) * time.Minute,
			Continuous: false,
		},
		{
			Job:        c.Poll,
			Name:       "Poll",
			When:       time.Duration(c.config.Flags.PollInterval) * time.Minute,
			Continuous: false,
		},
		{
			Job:        c.CheckCompliance,
			Name:       "CheckCompliance",
			When:       2 * time.Minute,
			Continuous: true,
		},
	}
}

func (c *Client) Check(t time.Time) {
	c.common.Check(t,
		&shared.ReminderPayload{
			FirstMessage: c.impl.FirstMessage(),
			FinalMessage: c.impl.FinalMessage(),
			Text:         c.impl.FirstMessage(),
			Title:        c.impl.FirstMessageHeader(),
		},
	)
}

func (c *Client) Diff(t time.Time) {
	tasks, err := c.impl.GetTasks()
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("could not get tasks")
		return
	}
	err = c.common.UpdateAllTasks(tasks)
	if err != nil {
		c.log.Err(err).Msg("could not add tasks")
		return
	}

	c.common.Diff(t)
}

func (c *Client) Poll(t time.Time) {
	c.common.Poll(t)
}

// CheckCompliance implements task.Rabbit.
func (c *Client) CheckCompliance(t time.Time) {
	c.shared.CheckCompliance(t)
}

// Populate implements task.Rabbit.
func (c *Client) Populate(time.Time) error {
	users, err := c.impl.GetUsers()
	if err != nil {
		return err
	}

	c.log.Info().Interface("users", users).Msg("users")
	tasks, err := c.impl.GetTasks()
	if err != nil {
		return err
	}
	c.log.Info().Interface("tasks", tasks).Msg("tasks")
	_, err = c.common.TableBuilder(
		&common.TableBuilder{
			Users: users,
			Tasks: tasks,
		},
	)

	return err
}

func (c *Client) InitTables() ([]string, error) {
	users, err := c.impl.GetUsers()
	if err != nil {
		return nil, err
	}
	tasks, err := c.impl.GetTasks()
	if err != nil {
		return nil, err
	}
	return c.common.TableBuilder(
		&common.TableBuilder{
			Users: users,
			Tasks: tasks,
		},
	)
}

func (c *Client) InteractionHandler(ctx *slacker.InteractionContext) {
	c.shared.InteractionHandler(ctx)
}

func (c *Client) InteractiveHelper(ctx *slacker.InteractionContext) {
	c.shared.InteractiveHelper(ctx)
}

func (c *Client) LoadInput(ctx *slacker.InteractionContext, loadType string) {
	c.shared.LoadInput(ctx, loadType)
}

func (c *Client) LoadProgram(triggerID string, loadType string) {
	c.shared.LoadProgram(
		&shared.Loader{
			TriggerID: triggerID,
			LoadType:  loadType,
			Tables:    shared.Tables(c.config.Flags.TaskImplementation, shared.TaskType(strings.ToLower(c.config.Flags.Task))),
		},
	)
}

func (c *Client) LogFlags() {
	config.LogFlags(c.log, c.config.Flags)
}

func (c *Client) ModalSubmit(ctx *slacker.InteractionContext) {
	c.shared.ModalSubmit(ctx)
}

func (c *Client) RegisterAdminCommands() {
	c.common.GetUsersInfo()
	c.shared.AddExclusion()
	c.shared.InitSettings()
	c.shared.Stopper()
	c.shared.UpdateConfig()
	c.shared.GetStatus(c.config.StatusHandler)
}

func (c *Client) RegisterCommands() {
	c.shared.ExclusionHelp()
	c.shared.ReminderHelp()
	c.common.GetSelfInfo()
	c.shared.CheckComplianceStatus()
}

func (c *Client) RegisterHandlers() {
	c.shared.RegisterHandlers()
}

func (c *Client) RegisterHelp() {
	c.common.RegisterHelp()
}

func (c *Client) RegisterMiddlewares() {
	c.shared.RegisterMiddlewares()
}

func (c *Client) MakePie() []handlers.PieInfo {
	complete := c.common.CompleteCount()
	err := complete.AddUpdates(c.common)
	if err != nil {
		c.log.Info().Err(err).Msg("failed to add updates")
	}

	return c.makePies(complete)
}

func (c *Client) SetupReportHandler() *handlers.Handler {
	return &handlers.Handler{
		Status: c.config.StatusHandler,
		Grahps: c,
		Log:    c.log,
	}
}

func (c *Client) makePies(cc common.CSCount) []handlers.PieInfo {
	pies := []handlers.PieInfo{}

	complete := 0
	incomplete := 0
	for _, m := range cc {
		if m.Status {
			complete++
		} else {
			incomplete++
		}
	}

	pies = append(pies, handlers.PieInfo{
		Title:    "Tasks",
		Subtitle: "Completion Status",
		Items: []handlers.Items{
			{
				Name:  "Complete",
				Value: strconv.Itoa(complete),
			},
			{
				Name:  "Incomplete",
				Value: strconv.Itoa(incomplete),
			},
		},
	})

	return pies
}

func (c *Client) MakeLines() *handlers.LineInfo {
	return &handlers.LineInfo{
		Title:    "Compliance Report",
		Subtitle: "by day of the week",
		Series:   c.makeSeries(),
		XAxis:    []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
	}
}

func (c *Client) makeSeries() []handlers.Series {
	_, week := time.Now().ISOWeek()

	ui, err := c.tables.Updates().By().Week(week).Build().Query()
	if err != nil {
		c.log.Error().Err(err).Msg("failed to query database")
		return nil
	}

	c.log.Info().Interface("ui", ui).Msg("ui")
	if ui.Empty() {
		c.log.Trace().Msg("no results")
		return nil
	}

	series := make([]handlers.Series, 0, len(ui))

	for i, m := range ui {
		series = append(series, handlers.Series{
			Name: m.ItemName,
			Data: []opts.LineData{
				{
					Value:      ui[i].Monday,
					XAxisIndex: 0,
					Symbol:     "circle",
				},
				{
					Value:      ui[i].Tuesday,
					XAxisIndex: 1,
					Symbol:     "circle",
				},
				{
					Value:      ui[i].Wednesday,
					XAxisIndex: 2,
					Symbol:     "circle",
				},
				{
					Value:      ui[i].Thursday,
					XAxisIndex: 3,
					Symbol:     "circle",
				},
				{
					Value:      ui[i].Friday,
					XAxisIndex: 4,
					Symbol:     "circle",
				},
			},
		})
	}

	return series
}

func (c *Client) Start() {
	_, week := time.Now().ISOWeek()
	_, err := c.tables.Admin().Add().
		WithOptions(admin.Info{
			Week:           week,
			ActiveCampaign: true,
			ActiveTask:     c.config.Flags.TaskImplementation,
		}).Execute()
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("failed to update admin table")
	}
}

func (c *Client) Stop() {
	c.config.LifeCycle.Stop()
}

func (c *Client) Update() {
	c.config.LifeCycle.Update()
}

func (c *Client) StopSubmit(ctx *slacker.InteractionContext) {
	c.shared.StopSubmit(ctx)
}
