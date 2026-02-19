package biz

import (
	"fmt"

	"github.com/DitzDev/gochart"
	"github.com/DitzDev/gochart/bar"
)

func RenderChartMRR(projections []Projection) string {
	data := make([]gochart.DataPoint, 0, len(projections))

	for _, p := range projections {
		data = append(data, gochart.DataPoint{
			Label: fmt.Sprintf("M%d", p.Month),
			Value: int(p.MRR),
		})
	}

	chart := bar.NewVerticalBarChart("MRR Growth", data, 12)
	return chart.RenderToString()
}

func RenderChartProfit(projections []Projection) string {
	data := make([]gochart.DataPoint, 0, len(projections))

	for _, p := range projections {
		data = append(data, gochart.DataPoint{
			Label: fmt.Sprintf("M%d", p.Month),
			Value: int(p.Profit),
		})
	}

	chart := bar.NewVerticalBarChart("Monthly Profit", data, 12)
	return chart.RenderToString()
}

func RenderChartUsers(projections []Projection) string {
	data := make([]gochart.DataPoint, 0, len(projections))

	for _, p := range projections {
		data = append(data, gochart.DataPoint{
			Label: fmt.Sprintf("M%d", p.Month),
			Value: p.Users,
		})
	}

	chart := bar.NewVerticalBarChart("User Growth", data, 12)
	return chart.RenderToString()
}
