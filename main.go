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
    // input_str      [] string
    cursor_x       int
    cursor_y       int
    ready          bool
    processor      *JSONProcessor
}

var (
    // layout
    align_right = lipgloss.NewStyle().
        Width(5).
        Align(lipgloss.Right)

    // vim like cursor
    cursor_style = lipgloss.NewStyle().
        Reverse(true)

    key_style = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#7aa2f7"))
    string_style = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#9ece6a"))
    number_style = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#ff9e64"))

    input_str []string
)

func (m model) Init() tea.Cmd {
    input_str = m.processor.lines
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
                case "t": // just to test quering fields
                    query, err := m.processor.QueryField("merchant")
                    if err != nil {
                        log.Printf("Error running query: %v", err)
                    } else {
                        m.processor.data = query
                        m.processor.updateLines()
                        input_str = m.processor.lines
                        m.vp_content.SetContent(strings.Join(input_str, "\n"))
                        m.cursor_y = 0
                        m.vp_content.YOffset = 0
                    }
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
                    m.move_cursor(0, -min(len(input_str), m.vp_content.Height)/2)
                    return m, nil
                }
                case "pgdown", "J":
                {
                    m.move_cursor(0, min(len(input_str), m.vp_content.Height)/2)
                    return m, nil
                }
            }
        }

        case tea.WindowSizeMsg:
        {
            if !m.ready {
                m.vp_content = viewport.New(msg.Width, msg.Height)
                m.vp_content.SetContent(strings.Join(input_str, "\n"))
                m.ready = true
            } else {
                m.vp_content.Width = msg.Width
                m.vp_content.Height = msg.Height

                // adjust viewport offset when the user increases the size
                // of the window to prevent adding line number beyond the
                // json content
                if msg.Height + m.vp_content.YOffset >= len(input_str) &&
                    m.vp_content.YOffset > 0 {
                    total := msg.Height + m.vp_content.YOffset
                    // reduce the offset by the amount that exeeds the
                    // length of the json content
                    m.vp_content.YOffset = or_zero(
                        m.vp_content.YOffset - total - len(input_str))
                }

                // adjust viewport offset when the height of the window
                // is smaller that the length of the json
                if msg.Height < len(input_str) {
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
	if new_y >= len(input_str) {
		new_y = len(input_str) - 1
	}
	
	// Boundary checks for X
	if new_x < 0 {
		new_x = 0
	}
	maxX := len(input_str[new_y])
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

	// Now colorize the lines (this preserves the cursor markers)
	lines = colorize(lines)

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
    if relative_y < len(input_str) - 1 {
        for i := relative_y; i < min(len(lines), len(input_str)) - 1; i++ {
            if i+1 < len(input_str) {
                lines[i+1] = fmt.Sprintf("%5s  %s", strconv.Itoa(count), lines[i+1])
                count++
            }
        }
    }

    vp_content := (strings.Join(lines, "\n"))

    s := lipgloss.JoinHorizontal(lipgloss.Top, vp_content)
    return s
}

func colorize(lines []string) []string {
    for i, line := range lines {
        colored_line := line

        // Remove cursor markers temporarily to detect patterns correctly
        line_clean := strings.ReplaceAll(line, "§CURSOR§", "")
        line_clean = strings.ReplaceAll(line_clean, "§/CURSOR§", "")

        // Color keys (text before colon)
        if strings.Contains(line_clean, ":") {
            parts := strings.SplitN(line_clean, ":", 2)
            key := parts[0]
            value := strings.TrimSpace(parts[1])

            // Apply colors to value
            if strings.HasPrefix(value, `"`) {
                value = string_style.Render(value)
            } else if is_number(value) {
                value = number_style.Render(value)
            }

            // For the actual line with markers, apply key styling carefully
            if strings.Contains(line, "§CURSOR§") {
                // Handle cursor markers in key
                colored_line = " " + applyKeyStyleWithCursor(key, line) + ": " + value
            } else {
                colored_line = " " + key_style.Render(key) + ": " + value
            }
        } else if strings.Contains(line_clean, `"`) {
            colored_line = " " + string_style.Render(line)
        } else if is_number(line_clean) {
            colored_line = " " + number_style.Render(line)
        } else {
            colored_line = " " + line
        }

        lines[i] = colored_line
    }

    return lines
}

func applyKeyStyleWithCursor(key string, lineWithMarkers string) string {
    // Find cursor markers in the original line
    if strings.Contains(lineWithMarkers, "§CURSOR§") {
        // Split around cursor markers
        parts := strings.Split(lineWithMarkers, "§CURSOR§")
        if len(parts) >= 2 {
            beforeCursor := strings.TrimSpace(parts[0])
            remainder := parts[1]

            endParts := strings.Split(remainder, "§/CURSOR§")
            if len(endParts) >= 2 {
                cursorChar := endParts[0]
                afterCursor := endParts[1]

                // Apply key style to parts before and after cursor, preserve cursor markers
                return key_style.Render(beforeCursor) + "§CURSOR§" + cursorChar + "§/CURSOR§" + key_style.Render(afterCursor)
            }
        }
    }
    // Fallback
    return key_style.Render(key)
}

func is_number(s string) bool {
    s = strings.TrimSpace(s)
    s = strings.TrimSuffix(s, ",")
    s = strings.TrimSpace(s)

    if len(s) == 0 {
      return false
    }
    _, err := strconv.ParseFloat(s, 64)
    return err == nil
}

func main() {
    json_str := `{
  "auth_code": "000001",
  "other": {
      "true": true,
      "false": false,
      "null": null
  },
  "numbers": [1, 2, 3],
  "mixed-array": ["hello", 200, true, null, false],
  "billing_amount": 10.90,
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
  "industries": [
    {
      "name": "ecommerce",
      "code": 4123
    },
    {
      "name": "retail",
      "code": 5234
    },
    {
      "name": "electronics",
      "code": 1111
    }
  ],
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

	processor, err := NewJSONProcessor([]byte(json_str))
	if err != nil {
		log.Fatal(err)
	}

    // industries, err := processor.QueryField("industries")
    // if err != nil {
	// 	log.Printf("Error querying industries: %v", err)
	// } else {
    //     processor.data = industries
    //     processor.updateLines()
	// }

    p := tea.NewProgram(
        // model{input_str: strings.Split(json_str, "\n")}, tea.WithAltScreen())
        model{processor: processor}, tea.WithAltScreen())

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
