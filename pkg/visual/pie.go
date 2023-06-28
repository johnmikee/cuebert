package visual

import (
	"github.com/vicanso/go-charts/v2"
)

type PieChartOption struct {
	ValueList []float64
	XAxis     []string
	Text      string
	Subtext   string
	Query     string
}

// custom func to sho the count instead of percentage.
// from the documentation:
//
// {b}: the name of a data item.
// {c}: the value of a data item.
// {d}: the percent of a data item(pie chart)
func pieSeriesShowCount() charts.OptionFunc {
	return func(opt *charts.ChartOption) {
		for index := range opt.SeriesList {
			opt.SeriesList[index].Label.Show = true
			opt.SeriesList[index].Label.Formatter = "{b}:  {c}"
		}
	}
}

func PieChart(o *PieChartOption) (string, error) {
	p, err := charts.PieRender(
		o.ValueList,
		charts.TitleOptionFunc(
			charts.TitleOption{
				Text:    o.Text,
				Subtext: o.Subtext,
				Left:    charts.PositionRight,
			},
		),
		charts.PaddingOptionFunc(
			charts.Box{
				Top:    20,
				Right:  20,
				Bottom: 20,
				Left:   20,
			},
		),
		charts.LegendOptionFunc(
			charts.LegendOption{
				Orient: charts.OrientVertical,
				Data:   o.XAxis,
				Left:   charts.PositionLeft,
			},
		),
		charts.HeightOptionFunc(1200),
		charts.WidthOptionFunc(1600),
		pieSeriesShowCount(),
	)

	if err != nil {
		return "", err
	}

	buf, err := p.Bytes()
	if err != nil {
		return "", err
	}

	file, err := buildOut(o.Query, nil, Pie, buf)
	return file, err
}
