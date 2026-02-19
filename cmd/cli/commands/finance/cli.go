package finance

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ExpenseCadence string

const (
	ExpenseCadenceWeekly  ExpenseCadence = "WEEKLY"
	ExpenseCadenceMonthly ExpenseCadence = "MONTHLY"
	ExpenseCadenceYearly  ExpenseCadence = "YEARLY"
)

type ExpenseInput struct {
	Name    string         `json:"name"`
	Amount  float64        `json:"amount"`
	Cadence ExpenseCadence `json:"cadence"`
}

type GoalInput struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

type BudgetInput struct {
	TaxRate  float64         `json:"taxRate"`
	Expenses []*ExpenseInput `json:"expenses"`
	Years    int             `json:"years"`
	Goals    []*GoalInput    `json:"goals"`
}

// Style definitions
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1).
			Bold(true)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#2D9EE0"))

	resultStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)
)

type FinanceModel struct {
	inputs         []textinput.Model
	resultAnnual   float64
	resultMonthly  float64
	focused        int
	submitted      bool
	step           int
	expenses       []*ExpenseInput
	currentExpense ExpenseInput
	addingExpense  bool
	goals          []*GoalInput
	currentGoal    GoalInput
	addingGoal     bool
	budget         BudgetInput
	years          int
}

func initialModel() FinanceModel {
	m := FinanceModel{
		inputs:        make([]textinput.Model, 3),
		focused:       0,
		submitted:     false,
		step:          0,
		expenses:      []*ExpenseInput{},
		goals:         []*GoalInput{},
		addingExpense: false,
		addingGoal:    false,
	}

	// Tax rate input
	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = "Tax rate (e.g. 0.3 for 30%)"
	m.inputs[0].Focus()
	m.inputs[0].Width = 30

	// Expense name and amount
	m.inputs[1] = textinput.New()
	m.inputs[1].Placeholder = "Expense name"
	m.inputs[1].Width = 30

	m.inputs[2] = textinput.New()
	m.inputs[2].Placeholder = "Expense amount"
	m.inputs[2].Width = 30

	return m
}

func (m FinanceModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m FinanceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab":
			if m.step == 0 {
				// Tax rate step, move to expenses step
				tax, err := strconv.ParseFloat(m.inputs[0].Value(), 64)
				if err == nil && tax > 0 && tax < 1 {
					m.budget.TaxRate = tax
					m.step = 1
					m.focused = 1
					cmd = m.inputs[1].Focus()
				}
			} else if m.step == 1 && m.addingExpense {
				// Cycling through expense inputs
				m.focused = (m.focused + 1) % 3
				if m.focused < 1 {
					m.focused = 1
				}
				for i := range m.inputs {
					if i == m.focused {
						cmd = m.inputs[i].Focus()
					} else {
						m.inputs[i].Blur()
					}
				}
			} else if m.step == 2 && m.addingGoal {
				// Cycling through goal inputs
				m.focused = (m.focused + 1) % 3
				if m.focused < 1 {
					m.focused = 1
				}
				for i := range m.inputs {
					if i == m.focused {
						cmd = m.inputs[i].Focus()
					} else {
						m.inputs[i].Blur()
					}
				}
			}

		case "enter":
			if m.step == 0 {
				// Tax rate step
				tax, err := strconv.ParseFloat(m.inputs[0].Value(), 64)
				if err == nil && tax > 0 && tax < 1 {
					m.budget.TaxRate = tax
					m.step = 1
					m.focused = 1
					cmd = m.inputs[1].Focus()
				}
			} else if m.step == 1 {
				if !m.addingExpense {
					// Start adding an expense
					m.addingExpense = true
					m.inputs[1].SetValue("")
					m.inputs[2].SetValue("")
					m.focused = 1
					cmd = m.inputs[1].Focus()
				} else if m.focused == 2 {
					// Add the expense
					name := m.inputs[1].Value()
					amountStr := m.inputs[2].Value()
					amount, err := strconv.ParseFloat(amountStr, 64)

					if err == nil && name != "" && amount > 0 {
						m.expenses = append(m.expenses, &ExpenseInput{
							Name:    name,
							Amount:  amount,
							Cadence: ExpenseCadenceMonthly,
						})
						m.inputs[1].SetValue("")
						m.inputs[2].SetValue("")
						m.focused = 1
						cmd = m.inputs[1].Focus()
					}
				} else {
					// Move to amount input
					m.focused = 2
					cmd = m.inputs[2].Focus()
				}
			} else if m.step == 2 {
				if !m.addingGoal {
					// Start adding a goal
					m.addingGoal = true
					m.inputs[1].SetValue("")
					m.inputs[2].SetValue("")
					m.focused = 1
					cmd = m.inputs[1].Focus()
				} else if m.focused == 2 {
					// Add the goal
					name := m.inputs[1].Value()
					amountStr := m.inputs[2].Value()
					amount, err := strconv.ParseFloat(amountStr, 64)

					if err == nil && name != "" && amount > 0 {
						m.goals = append(m.goals, &GoalInput{
							Name:   name,
							Amount: amount,
						})
						m.inputs[1].SetValue("")
						m.inputs[2].SetValue("")
						m.focused = 1
						cmd = m.inputs[1].Focus()
					}
				} else {
					// Move to amount input
					m.focused = 2
					cmd = m.inputs[2].Focus()
				}
			} else if m.step == 3 {
				// Years input
				yearsStr := m.inputs[0].Value()
				years, err := strconv.Atoi(yearsStr)
				if err == nil && years > 0 {
					m.years = years
					m.step = 4 // Go to results
					m.budget.Expenses = m.expenses

					// Calculate results
					resolver := &FinanceResolver{}
					m.resultAnnual, m.resultMonthly, _ = resolver.CalculateGoal(nil, m.budget, m.years, m.goals)
					m.submitted = true
				}
			}

		case "n":
			if m.step == 1 && !m.addingExpense {
				// Move to next step (goals)
				m.step = 2
				m.inputs[1].SetValue("")
				m.inputs[2].SetValue("")
				m.inputs[0].SetValue("")
				m.inputs[0].Placeholder = "Number of years for goals"
			} else if m.step == 2 && !m.addingGoal {
				// Move to final step (years)
				m.step = 3
				m.inputs[0].SetValue("")
				m.inputs[0].Placeholder = "Number of years for goals"
				m.focused = 0
				cmd = m.inputs[0].Focus()
			}

		case "d":
			if m.step == 1 && m.addingExpense {
				// Cancel adding expense
				m.addingExpense = false
				m.inputs[1].SetValue("")
				m.inputs[2].SetValue("")
			} else if m.step == 2 && m.addingGoal {
				// Cancel adding goal
				m.addingGoal = false
				m.inputs[1].SetValue("")
				m.inputs[2].SetValue("")
			}
		}
	}

	// Handle input updates
	for i := range m.inputs {
		m.inputs[i], _ = m.inputs[i].Update(msg)
	}

	return m, cmd
}

func (m FinanceModel) View() string {
	var b strings.Builder

	title := titleStyle.Render(" Finance Calculator ")
	b.WriteString(title + "\n\n")

	if m.submitted {
		// Show results
		b.WriteString(resultStyle.Render("Results:") + "\n\n")
		b.WriteString(fmt.Sprintf("Annual pre-tax income needed: $%.2f\n", m.resultAnnual))
		b.WriteString(fmt.Sprintf("Monthly pre-tax income needed: $%.2f\n", m.resultMonthly))
		b.WriteString("\nPress Esc to exit.")
		return b.String()
	}

	if m.step == 0 {
		// Tax rate step
		b.WriteString("Enter your tax rate (e.g., 0.3 for 30%):\n")
		b.WriteString(m.inputs[0].View() + "\n")
		b.WriteString("\nPress Enter to continue.")
	} else if m.step == 1 {
		// Expenses step
		b.WriteString(resultStyle.Render("Expenses:") + "\n\n")

		// List existing expenses
		if len(m.expenses) > 0 {
			for i, exp := range m.expenses {
				b.WriteString(fmt.Sprintf("%d. %s: $%.2f per month\n", i+1, exp.Name, exp.Amount))
			}
			b.WriteString("\n")
		}

		if m.addingExpense {
			b.WriteString("Add expense:\n")
			b.WriteString("Name: " + m.inputs[1].View() + "\n")
			b.WriteString("Amount: " + m.inputs[2].View() + "\n")
			b.WriteString("\nPress Enter to add, 'd' to cancel.")
		} else {
			b.WriteString("\nPress Enter to add expense, 'n' for next step.")
		}
	} else if m.step == 2 {
		// Goals step
		b.WriteString(resultStyle.Render("Financial Goals:") + "\n\n")

		// List existing goals
		if len(m.goals) > 0 {
			for i, goal := range m.goals {
				b.WriteString(fmt.Sprintf("%d. %s: $%.2f\n", i+1, goal.Name, goal.Amount))
			}
			b.WriteString("\n")
		}

		if m.addingGoal {
			b.WriteString("Add goal:\n")
			b.WriteString("Name: " + m.inputs[1].View() + "\n")
			b.WriteString("Amount: " + m.inputs[2].View() + "\n")
			b.WriteString("\nPress Enter to add, 'd' to cancel.")
		} else {
			b.WriteString("\nPress Enter to add goal, 'n' for next step.")
		}
	} else if m.step == 3 {
		// Years step
		b.WriteString(resultStyle.Render("Timeframe:") + "\n\n")
		b.WriteString("How many years do you have to reach your goals?\n")
		b.WriteString(m.inputs[0].View() + "\n")
		b.WriteString("\nPress Enter to calculate results.")
	}

	return b.String()
}

func RunCLI() error {
	p := tea.NewProgram(initialModel())
	_, err := p.Run()
	return err
}
