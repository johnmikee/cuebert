package shared

func (s *Shared) Format(st string) string {
	return s.constFormat(st)
}

func (s *Shared) constFormat(c string) string {
	return c + "_" + s.flags.Task + "_" + s.flags.TaskImplementation
}

const (
	ApproveExclusion     = "approve_exclusion"
	ExclusionApprover    = "exclusion_approver"
	DatePicker           = "datePicker"
	DenyExclusion        = "deny_exclusion"
	ExclusionInput       = "exclusion_input"
	ExclusionReason      = "exclusion_reason"
	RemindMe             = "remind_me"
	ReminderPicker       = "reminder_picker"
	ReminderPickerHeader = "reminder_picker_header"
	RemindMeQuestion     = "remind_me_question"
	DontRemindMe         = "dont_remind_me"
	SerialInput          = "serial_input"
	SerialNumber         = "serial_number"
	TimePicker           = "timePicker"
	UpdateUserYes        = "yes_update_user"
	UpdateUserNo         = "no_update_user"
	UsersSelect          = "users_select"
	UserUpdateModal      = "user_update_modal"
	UpdateUser           = "update_user"
	UpdateUserQuestion   = "update_user_question"
	UpdateUserSelector   = "user_update_selector"
	UpdateUserSubmit     = "update_user_submit"
)

const (
	ExclusionAddQuestion  = "exclusion_add_question"
	ExclusionAddYes       = "exclusion_add_yes"
	ExclusionAddNo        = "exclusion_add_no"
	ExclusionDatePicker   = "exclusion_date_picker"
	ExclusionNo           = "exclusion_no"
	ExclusionReasonHeader = "exclusion_reason_header"
	AdminExclusionModal   = "admin_exclusion_modal"
	ExclusionModal        = "exclusion_modal"
	ExclusionQuestion     = "exclusion_question"
	YesExclusion          = "yes_exclusion"
	NoExclusion           = "no_exclusion"
	YesAddExclusion       = "yes_add_exclusion"
)

// modal and attachment fields.
const (
	Accept             = "accept"
	AckIT              = "ack_it"
	FinalAck           = "final_ack"
	DeviceBox          = "device_box"
	Start              = "start"
	Reload             = "reload"
	UserDevices        = "user_devices"
	UpdateBeforeLaunch = "update_before_launch"
	UpdateConfigYes    = "yes_update_config"
	UpdateConfigNo     = "no_update_config"
)

const (
	ApproveStop         = "approve_stop"
	BypassOverrideNo    = "bypass_override_no"
	BypassOverrideYes   = "bypass_override_yes"
	DenyStop            = "deny_stop"
	MessageTestRequest  = "message_test_request"
	MessageRequestYes   = "message_request_yes"
	MessageRequestNo    = "message_request_no"
	StopCuebertApproval = "stop_cuebert_approve"
	StopCuebertRequest  = "stop_cuebert_request"
	StopCuebertNo       = "stop_cuebert_no"
	StopCuebertOverride = "stop_cuebert_override"
	StopCuebertYes      = "stop_cuebert_yes"
)

const (
	YesCheck              = "yes_check"
	NoCheck               = "no_check"
	CheckComplianceStatus = "check_compliance_status"
	UserReminder          = "user_reminder"
	GroupDM               = "group_dm"
)

const (
	Email                    = "email"
	FirstMessageACK          = "first_message_ack"
	FirstMessageACKAt        = "first_message_ack_at"
	FirstMessageACKAtPicker  = "first_message_ack_at_picker"
	FirstMessageACKOption    = "first_message_ack_option"
	FirstMessageSentAt       = "first_message_sent_at"
	FirstMessageSentAtPicker = "first_message_sent_at_picker"
	FirstMessageSentOption   = "first_message_sent_option"
	FirstSent                = "first_message_sent"
	TruthyNo                 = "false"
	TruthyYes                = "true"
	TZOffset                 = "tz_offset"
)
