package main

import (
	"fmt"
	"strconv"
	"strings"
)

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
				current_line_number := fmt.Sprintf("%4s   ", strconv.Itoa(m.cursor_y+1))

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
			lines[i-1] = fmt.Sprintf("%5s  %s", strconv.Itoa(count), lines[i-1])
			count++
		}
	}

	// botom
	count = 1
	if relative_y < len(input_str)-1 {
		for i := relative_y; i < min(len(lines), len(input_str))-1; i++ {
			if i+1 < len(input_str) {
				lines[i+1] = fmt.Sprintf("%5s  %s", strconv.Itoa(count), lines[i+1])
				count++
			}
		}
	}

	vp_content := (strings.Join(lines, "\n"))
	s := vp_content
	return s
}

func colorize(lines []string) []string {
	for i, line := range lines {
		colored_line := line

		// Color keys (text before colon)
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			key := parts[0]
			value := strings.TrimSpace(parts[1])

			// Apply colors to value
			if strings.HasPrefix(value, `"`) {
				value = string_style.Render(value)
			} else if is_number(value) {
				value = number_style.Render(value)
			}

			colored_line = " " + key_style.Render(key) + ": " + value
		} else if strings.Contains(line, `"`) {
			colored_line = " " + string_style.Render(line)
		} else if is_number(line) {
			colored_line = " " + number_style.Render(line)
		} else {
			colored_line = " " + line
		}

		lines[i] = colored_line
	}

	return lines
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
