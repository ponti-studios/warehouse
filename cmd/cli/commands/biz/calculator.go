package biz

import (
	"fmt"
	"math"

	"github.com/charmbracelet/lipgloss"
)

type Projection struct {
	Month          int
	Users          int
	NewUsers       int
	MRR            float64
	ARR            float64
	Revenue        float64
	Costs          float64
	COGS           float64
	Profit         float64
	CumulativeProf float64
	CashBalance    float64
	BurnRate       float64
	Runway         float64
	MRRGrowth      float64
	LTV            float64
	LTVCAC         float64
	PaybackPeriod  float64
	GrossMargin    float64
}

func Calculate(input *Input, overrideMonths int) []Projection {
	months := input.GetMonths()
	if overrideMonths > 0 {
		months = overrideMonths
	}

	projections := make([]Projection, months)

	users := input.Model.Users.Initial
	cashBalance := input.GetCashOnHand()
	var prevMRR float64
	breakEvenMonth := 0

	for i := 1; i <= months; i++ {
		users = calculateUsers(users, input.Model.Users, i)

		var currentNewUsers int
		if i == 1 {
			currentNewUsers = input.Model.Users.Initial + input.Model.Users.MonthlyNew
		} else {
			prevUsers := calculateUsers(users, input.Model.Users, i-1)
			currentNewUsers = users - prevUsers
			if currentNewUsers < 0 {
				currentNewUsers = input.Model.Users.MonthlyNew
			}
		}

		var mrr, arr, revenue, cogs, costs, profit, burnRate, runway, mrrGrowth, ltv, ltvCAC, payback, grossMargin float64

		if input.Model.Type == "subscription" {
			mrr = float64(users) * input.Model.Pricing.Monthly
			arr = mrr * 12
			revenue = arr

			if i > 1 && prevMRR > 0 {
				mrrGrowth = ((mrr - prevMRR) / prevMRR) * 100
			}
			prevMRR = mrr
		} else {
			revenue = float64(users) * input.Model.Transaction.RevenuePerUser * float64(input.Model.Transaction.TxPerUser)
			mrr = revenue / 12
			arr = revenue

			if i > 1 && prevMRR > 0 {
				mrrGrowth = ((mrr - prevMRR) / prevMRR) * 100
			}
			prevMRR = mrr
		}

		cogs = float64(users) * input.Model.Costs.PerUser
		costs = input.Model.Costs.FixedMonthly + cogs + input.Model.Costs.SalesMarketing
		profit = revenue - costs
		burnRate = costs - revenue
		grossMargin = (revenue - cogs) / revenue * 100

		if burnRate > 0 {
			runway = cashBalance / burnRate
		} else {
			runway = math.MaxFloat64
		}

		lifetimeMonths := 1.0
		if input.Model.Users.ChurnRate > 0 {
			lifetimeMonths = 1.0 / input.Model.Users.ChurnRate
		}
		ltv = mrr * lifetimeMonths

		if currentNewUsers > 0 {
			cac := input.Model.Costs.SalesMarketing / float64(currentNewUsers)
			if cac > 0 {
				ltvCAC = ltv / cac
			}
			marginalRevenue := mrr - cogs
			if marginalRevenue > 0 {
				payback = cac / marginalRevenue
			}
		}

		cashBalance = cashBalance + profit

		if breakEvenMonth == 0 && profit > 0 {
			breakEvenMonth = i
		}

		projections[i-1] = Projection{
			Month:          i,
			Users:          users,
			NewUsers:       currentNewUsers,
			MRR:            mrr,
			ARR:            arr,
			Revenue:        revenue,
			Costs:          costs,
			COGS:           cogs,
			Profit:         profit,
			CumulativeProf: cashBalance - input.GetCashOnHand(),
			CashBalance:    cashBalance,
			BurnRate:       burnRate,
			Runway:         runway,
			MRRGrowth:      mrrGrowth,
			LTV:            ltv,
			LTVCAC:         ltvCAC,
			PaybackPeriod:  payback,
			GrossMargin:    grossMargin,
		}
	}

	return projections
}

func calculateUsers(prevUsers int, users UsersConfig, month int) int {
	if month == 1 {
		return users.Initial + users.MonthlyNew
	}

	churned := float64(prevUsers) * users.ChurnRate
	growth := float64(prevUsers) * users.GrowthRate

	newUsers := float64(users.MonthlyNew)
	if users.GrowthRate > 0 {
		newUsers += growth
	}

	return int(float64(prevUsers) - churned + newUsers)
}

type Summary struct {
	TotalRevenue     float64
	TotalCosts       float64
	TotalProfit      float64
	AvgUsers         float64
	PeakUsers        int
	PeakMRR          float64
	CashBalance      float64
	StartingCash     float64
	BreakEvenMonth   int
	UsersAtBreakEven int
	FinalRunway      float64
	FinalLTV         float64
	FinalLTVCAC      float64
	FinalPayback     float64
	FinalGrossMargin float64
}

func Summarize(projections []Projection, input *Input) Summary {
	if len(projections) == 0 {
		return Summary{}
	}

	var totalRevenue, totalCosts, totalProfit, avgUsers float64
	peakUsers := 0
	peakMRR := 0.0
	var breakEvenMonth, usersAtBreakEven int

	for _, p := range projections {
		totalRevenue += p.Revenue
		totalCosts += p.Costs
		totalProfit += p.Profit
		avgUsers += float64(p.Users)

		if p.Users > peakUsers {
			peakUsers = p.Users
		}
		if p.MRR > peakMRR {
			peakMRR = p.MRR
		}
		if breakEvenMonth == 0 && p.CumulativeProf > 0 {
			breakEvenMonth = p.Month
			usersAtBreakEven = p.Users
		}
	}

	avgUsers /= float64(len(projections))
	lastProj := projections[len(projections)-1]

	return Summary{
		TotalRevenue:     totalRevenue,
		TotalCosts:       totalCosts,
		TotalProfit:      totalProfit,
		AvgUsers:         avgUsers,
		PeakUsers:        peakUsers,
		PeakMRR:          peakMRR,
		CashBalance:      lastProj.CashBalance,
		StartingCash:     input.GetCashOnHand(),
		BreakEvenMonth:   breakEvenMonth,
		UsersAtBreakEven: usersAtBreakEven,
		FinalRunway:      lastProj.Runway,
		FinalLTV:         lastProj.LTV,
		FinalLTVCAC:      lastProj.LTVCAC,
		FinalPayback:     lastProj.PaybackPeriod,
		FinalGrossMargin: lastProj.GrossMargin,
	}
}

func FormatCurrency(value float64) string {
	prefix := ""
	if value < 0 {
		prefix = "-"
		value = -value
	}
	return fmt.Sprintf("%s$%.2f", prefix, value)
}

func FormatCurrencyColored(value float64) string {
	prefix := ""
	if value < 0 {
		prefix = "-"
		value = -value
		color := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		return color.Render(fmt.Sprintf("%s$%.2f", prefix, value))
	}
	if value > 0 {
		color := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
		return color.Render(fmt.Sprintf("%s$%.2f", prefix, value))
	}
	return fmt.Sprintf("%s$%.2f", prefix, value)
}

func FormatPercent(value float64) string {
	return fmt.Sprintf("%.1f%%", value)
}

func FormatPercentColored(value float64, goodThreshold, badThreshold float64) string {
	color := ""
	if value >= goodThreshold {
		color = "82"
	} else if value <= badThreshold {
		color = "196"
	} else {
		color = "255"
	}
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	return style.Render(fmt.Sprintf("%.1f%%", value))
}

func FormatRunway(value float64) string {
	if value >= math.MaxFloat64 || value > 9999 {
		return "∞"
	}
	if value < 0 {
		return "0"
	}
	return fmt.Sprintf("%.1f mo", value)
}
