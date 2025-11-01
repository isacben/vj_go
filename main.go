package main

import (
	"fmt"
	"log"
	"strings"
    "strconv"
    "os"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
    vp_content     viewport.Model
    input_str      [] string
    cursor_x       int
    cursor_y       int
    ready          bool
}

var (
    // layout
    align_right = lipgloss.NewStyle().
        Width(5).
        Align(lipgloss.Right)

    // vim like cursor
    cursor_style = lipgloss.NewStyle().
        Reverse(true)
)

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		contentCmd  tea.Cmd
        numbersCmd  tea.Cmd
	)

    switch msg := msg.(type) {
        case tea.KeyMsg:
        {
            switch msg.String() {
                case "q":
                    return m, tea.Quit
                case "right", "l":
                    m.vp_content.ScrollRight(1)
                case "left", "h": 
                    m.vp_content.ScrollLeft(1)
                case "up", "k":
                {
                    m.move_cursor(0, -1)
                    return m, nil
                }
                case "down", "j":
                {
                    m.move_cursor(0, 1)
                    return m, nil
                }
                case "pgup", "K":
                {
                    m.move_cursor(0, -min(len(m.input_str), m.vp_content.Height)/2)
                    return m, nil
                }
                case "pgdown", "J":
                {
                    m.move_cursor(0, min(len(m.input_str), m.vp_content.Height)/2)
                    return m, nil
                }
            }
        }

        case tea.WindowSizeMsg:
        {
            if !m.ready {
                m.vp_content = viewport.New(msg.Width, msg.Height)
                m.vp_content.SetContent(strings.Join(m.input_str, "\n"))
                m.ready = true
            } else {
                m.vp_content.Width = msg.Width
                m.vp_content.Height = msg.Height

                // adjust viewport offset when the user increases the size
                // of the window to prevent adding line number beyond the
                // json content
                if msg.Height + m.vp_content.YOffset >= len(m.input_str) &&
                    m.vp_content.YOffset > 0 {
                    total := msg.Height + m.vp_content.YOffset
                    // reduce the offset by the amount that exeeds the
                    // length of the json content
                    m.vp_content.YOffset = or_zero(
                        m.vp_content.YOffset - total - len(m.input_str))
                }

                // adjust viewport offset when the height of the window
                // is smaller that the length of the json
                if msg.Height < len(m.input_str) {
                    if m.cursor_y >=  msg.Height + m.vp_content.YOffset {
                        // adjust the offset to put the cursor at the very
                        // bottom of the screenn
                        m.vp_content.YOffset = m.cursor_y - msg.Height + 1
                    }
                }
                log.Println("offset:", m.vp_content.YOffset)
            }
        }
    }

	m.vp_content, contentCmd = m.vp_content.Update(msg)
    return m, tea.Batch(contentCmd, numbersCmd)
}

func (m *model) move_cursor(dx int, dy int) {
	// Calculate new cursor position
	new_y := m.cursor_y + dy
	new_x := m.cursor_x + dx
	
	// Boundary checks for Y
	if new_y < 0 {
		new_y = 0
	}
	if new_y >= len(m.input_str) {
		new_y = len(m.input_str) - 1
	}
	
	// Boundary checks for X
	if new_x < 0 {
		new_x = 0
	}
	maxX := len(m.input_str[new_y])
	if new_x > maxX {
		new_x = maxX
	}
	
	m.cursor_y = new_y
	m.cursor_x = new_x
	
	// Adjust viewport if cursor moves outside visible area
	m.adjust_viewport()
}

func (m *model) adjust_viewport() {
	// Vertical scrolling
	visible_top := m.vp_content.YOffset
	visible_bottom := m.vp_content.YOffset + m.vp_content.Height - 1
	
	if m.cursor_y < visible_top {
		// Cursor is above viewport, scroll up
		m.vp_content.YOffset = m.cursor_y
	} else if m.cursor_y > visible_bottom {
		// Cursor is below viewport, scroll down
		m.vp_content.YOffset = m.cursor_y - m.vp_content.Height + 1
	}
}

func (m model) View() string {
    if !m.ready {
		return "Loading..."
	}

    // Create a custom view with cursor overlay
	lines := strings.Split(m.vp_content.View(), "\n")
	
	// Calculate cursor position relative to viewport
	relative_y := m.cursor_y - m.vp_content.YOffset
	
	// Highlight the cursor line if it's visible
	if relative_y >= 0 && relative_y < len(lines) {
		if relative_y < len(lines) {
			line := lines[relative_y]
			
			// Add cursor indicator
			if m.cursor_x <= len(line) {
				before := ""
				cursor_char := " "
				after := ""
				
				if m.cursor_x > 0 {
					before = line[:m.cursor_x]
				}
				if m.cursor_x < len(line) {
					cursor_char = string(line[m.cursor_x])
					if m.cursor_x+1 < len(line) {
						after = line[m.cursor_x+1:]
					}
				}
				
                // print current line
                current_line_number := fmt.Sprintf("%4s   ", strconv.Itoa(m.cursor_y + 1))

				lines[relative_y] = current_line_number +
                    before +
                    cursor_style.Render(cursor_char) +
                    after
			}
		}
	}

    // add the line numbers
    // top
    count := 1
    if relative_y > 0 {
        for i := relative_y; i > 0; i-- {
            lines[i-1] = fmt.Sprintf("%5s  %s", strconv.Itoa(count), lines[i - 1])
            count++
        }
    }

    // botom
    count = 1
    if relative_y < len(m.input_str) - 1 {
        for i := relative_y; i < min(len(lines), len(m.input_str)) - 1; i++ {
            if i+1 < len(m.input_str) {
                lines[i+1] = fmt.Sprintf("%5s  %s", strconv.Itoa(count), lines[i+1])
                count++
            }
        }
    }

    vp_content := (strings.Join(lines, "\n"))

    s := lipgloss.JoinHorizontal(lipgloss.Top, vp_content)
    return s
}

func main() {
    json_str := `{
  "auth_code": "000001",
  "billing_amount": 100,
  "billing_currency": "USD",
  "card_id": "7f687fe6-dcf4-4462-92fa-80335301d9d2",
  "card_nickname": "string",
  "client_data": "Some client data",
  "failure_reason": "INSUFFICIENT_FUNDS",
  "masked_card_number": "************4242",
  "merchant": {
    "category_code": "4829",
    "city": "Melbourne",
    "country": "Australia",
    "name": "Merchant A"
  },
  "network_transaction_id": "3951729271768745",
  "posted_date": "2018-03-22T16:08:02+00:00",
  "retrieval_ref": "909916088001",
  "status": "APPROVED",
  "transaction_amount": 100,
  "transaction_currency": "USD",
  "transaction_date": "2018-03-21T16:08:02+00:00",
  "transaction_id": "6c2dc266-09ad-4235-b61a-767c7cd6d6ea",
  "transaction_type": "REFUND"
}`

    if len(os.Getenv("DEBUG")) > 0 {
        f, err := tea.LogToFile("debug.log", "debug")
        if err != nil {
            fmt.Println("fatal:", err)
            os.Exit(1)
        }
        defer f.Close()
    }

    p := tea.NewProgram(
        model{input_str: strings.Split(json_str, "\n")}, tea.WithAltScreen())
    
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func or_zero(a int) int {
    if a < 0 {
        a = 0
    }
    return a
}
