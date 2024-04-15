package handlers

import (
	"io"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// PieInfo represents the data needed to generate a pie chart
type PieInfo struct {
	Title    string
	Subtitle string
	Items    []Items
}

// Items represents the items contained in a pie chart
type Items struct {
	Name  string
	Value interface{}
}

// LineInfo represents the data needed to generate a line chart
type LineInfo struct {
	Title    string
	Subtitle string
	Series   []Series
	XAxis    []string
}

// Series contains the data used to generate a line chart
type Series struct {
	Name    string
	Data    []opts.LineData
	Options []charts.SeriesOpts
}

func generatePieItems(n *PieInfo) []opts.PieData {
	items := make([]opts.PieData, 0)
	for i := 0; i < len(n.Items); i++ {
		items = append(items, opts.PieData{Name: n.Items[i].Name, Value: n.Items[i].Value})
	}
	return items
}

// Lines represents the line chart
type Lines struct {
	c *charts.Line
}

func (l *Lines) addLines(s Series) *charts.Line {
	return l.c.AddSeries(s.Name, s.Data)
}

func makePie(i PieInfo, w io.Writer) error {
	pie := charts.NewPie()
	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Top: "20%", Title: i.Title}),
		charts.WithInitializationOpts(opts.Initialization{PageTitle: "Cuebert Reporting"}),
	)

	pie.AddSeries("pie", generatePieItems(&i)).
		SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{
				Show:      true,
				Formatter: "{b}: {c}",
			}),
			charts.WithPieChartOpts(opts.PieChart{
				Radius: []string{"40%", "75%"},
			}),
		)

	return pie.Render(w)
}

func makeLines(i *LineInfo, w io.Writer) error {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Subtitle: i.Subtitle}),
		charts.WithLegendOpts(opts.Legend{Show: true, Type: "scroll"}),
		charts.WithInitializationOpts(opts.Initialization{PageTitle: "Cuebert Reporting"}),
	)

	line.SetXAxis(i.XAxis)
	lc := &Lines{c: line}
	for _, x := range i.Series {
		lc.addLines(x)
	}
	line.SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))
	return line.Render(w)
}
