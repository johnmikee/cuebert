package timebound

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/johnmikee/cuebert/cuebert/bot"
	br "github.com/johnmikee/cuebert/db/bot"
	"github.com/slack-go/slack"
)

// PostInit implements method.Actions.
func (t *TimeBound) PostCheck(sa []string) {
	t.log.Trace().Strs("", sa).Msg("not implemented")
}

// FirstMessage implements method.Actions.
func (t *TimeBound) FirstMessage() string {
	return t.firstMessage(t.cfg.deadline)
}

func (t *TimeBound) firstMessage(day string) string {
	return fmt.Sprintf(`
	Attention, esteemed denizens of Middle-earth (and those dwelling in the realms of modern technology),

	We beseech thee, with utmost urgency, to bestow thy device with the blessings of an update by %s. Failure to do so may invoke the wrath of the mighty Ents, who shall tickle thy gadget's circuits until it malfunctions in the most perplexing manner!
	
	Imagine, noble hobbits, if the Eye of Sauron should gaze upon thy unpatched device. It shall unleash a horde of mischievous Gollums to swipe thy precious files, leaving naught but digital crumbs for thy journey.
	
	Fellowship of the Update Button, unite! For only by clicking below and embarking on the perilous path of software updates shall we evade the clutches of Nazgûl malware, dark wizards of the virtual realm. Fear not, as Gandalf himself shall guide thee through the update process, lighting the way with his mighty staff of progress bars.
	
	But, alas! Should thou dare to defy this plea, be prepared for the most peculiar of punishments. Frodo shall mysteriously rearrange thy keyboard, swapping 'F' with 'G' and 'B' with 'H', leaving thee in a state of befuddlement fit for Bilbo Baggins at his most forgetful.
	
	Furthermore, Legolas, that nimble archer, shall turn thy cursor into a mischievous squirrel, forever eluding thy attempts to click on any link or button. Thou shalt find thyself in an endless chase, as if trapped in a wild elven dance!
	
	So, hearken unto these words of jest and wisdom. Update thy device promptly, or risk thyself becoming entangled in a web of technological woes, where hobbits giggle, wizards frown, and elves refuse to lend thee their Wi-Fi passwords.
	
	May the updates be swift, thy files secure, and thy journey through the digital realm filled with mirth and merriment!
	
	Sincerely,
	The Fellowship of the IT Ring
	`, day)
}

func (t *TimeBound) ReminderMessage(rp *bot.ReminderPayload) error {
	attachment := slack.Attachment{
		Text:       t.reminderMessage(t.cfg.deadline),
		CallbackID: ReminderMessage,
		Color:      "#3AA3E3",
	}

	message := slack.MsgOptionAttachments(attachment)

	channelID, timestamp, err := t.sc.PostMessage(rp.UserSlackID, message)
	if err != nil {
		t.log.Debug().AnErr("failed to send reminder message", err).Send()
		return err
	}

	t.log.Debug().
		Str("serial", rp.Serial).
		Str("time", timestamp).
		Str("user", rp.UserName).
		Str("channel", channelID).
		Msg("manager message sent")

	return nil
}

func (t *TimeBound) reminderMessage(day string) string {
	return fmt.Sprintf(`
	Attention, valiant adventurers of Middle-earth,

	Thy device's plea for an update has not fallen upon deaf ears. We, the Fellowship of the Technological Ring, beseech thee once more to heed the call of progress.
	
	Remember, neglecting this crucial task may awaken the wrath of the Ents, who will pester thy device relentlessly until it succumbs to an unexpected and amusing malfunction on %s.
	
	Beware the lurking gaze of the Eye of Sauron, for it seeks unpatched devices to summon a legion of mischievous Gollums, eager to snatch thy precious files away.
	
	Yet, fear not, for Gandalf the Grey shall be thy guiding light, leading thee through the perilous path of software updates. His progress bars shall illuminate the way to a secure digital realm.
	
	However, dear wanderers, should ye persist in defiance, be prepared for Frodo's whimsy. He shall rearrange thy keyboard, transforming thy 'F's to 'G's and thy 'B's to 'H's, leaving thee puzzled and perplexed.
	
	And lo, Legolas the agile shall curse thy cursor, transforming it into a mischievous squirrel. It shall evade thy attempts to click, mocking thy every move in an eternal chase.
	
	Therefore, we implore thee once more: Update thy device posthaste! Shield thyself from the perils of Nazgûl malware and the scorn of the elven realm, where Wi-Fi passwords shall be denied.
	
	May thy files remain safe, thy journey through the digital realm be filled with laughter, and the updates be swift!
	
	With kindest regards,
	The Fellowship of the Technological Ring
	`, day)
}

func (t *TimeBound) waitSend(b *br.Info) {
	n, err := rand.Int(rand.Reader, big.NewInt(120))
	if err != nil {
		t.log.Err(err).Msg("could not generate random delay")
	}
	time.Sleep(time.Duration(n.Int64()) * time.Second)
	t.bot.SendReminder(3, b)
}
