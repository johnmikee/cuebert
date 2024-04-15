package update

import (
	"time"

	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/johnmikee/cuebert/cuebert/config"
	"github.com/johnmikee/cuebert/cuebert/task/update/common"
	"github.com/johnmikee/cuebert/handlers"
)

// LogFlags implements task.Rabbit.
func (c *Client) LogFlags() {
	config.LogFlags(c.log, c.config.Flags)
}

// SetupReportHandler implements task.Rabbit.
func (c *Client) SetupReportHandler() *handlers.Handler {
	return &handlers.Handler{
		Status:    c.config.StatusHandler,
		Reporting: true,
		Grahps:    c,
		Log:       c.log,
	}
}

// MakePie implements handlers.Reporting.
func (c *Client) MakePie() []handlers.PieInfo {
	osVers := c.common.OSCount()
	err := osVers.AddUpdates(c.tables)
	if err != nil {
		c.log.Info().Err(err).Msg("failed to add updates")
	}

	return c.makePies(osVers)
}

// MakeLines implements handlers.Reporting.
func (c *Client) MakeLines() *handlers.LineInfo {
	return &handlers.LineInfo{
		Title:    "OS Version Installs",
		Subtitle: "by day of the week",
		Series:   c.makeSeries(),
		XAxis:    []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
	}
}

func (c *Client) makePies(osVers common.OSCount) []handlers.PieInfo {
	mvers := osVers.ByMajorVers()
	pies := []handlers.PieInfo{}
	for _, m := range mvers {
		items := []handlers.Items{}
		for _, os := range m.OS {
			items = append(items, handlers.Items{
				Name:  os.OS,
				Value: os.Count,
			})
		}
		pies = append(pies, handlers.PieInfo{
			Title:    m.MajorVers,
			Subtitle: "OS Version Installs",
			Items:    items,
		})
	}
	return pies
}

func (c *Client) makeSeries() []handlers.Series {
	_, week := time.Now().ISOWeek()

	ui, err := c.tables.Updates().By().Week(week).Build().Query()
	if err != nil {
		c.log.Error().Err(err).Msg("failed to query database")
		return nil
	}

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
