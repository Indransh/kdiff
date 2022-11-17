package view

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type multiSelectListItem struct {
	Text       string
	Selected   func()
	IsDisabled bool
}

type multiSelectList struct {
	*tview.Box
	items         []*multiSelectListItem
	selectedItems []int
	currentItem   int
	selectedStyle tcell.Style
	textStyle     tcell.Style
	selected      func(int, string)
	highlighted   func(int, string)
}

func newMultiSelectList() *multiSelectList {
	return &multiSelectList{
		Box:           tview.NewBox(),
		textStyle:     tcell.StyleDefault.Foreground(tview.Styles.PrimaryTextColor),
		selectedStyle: tcell.StyleDefault.Foreground(tview.Styles.PrimitiveBackgroundColor).Background(tview.Styles.PrimaryTextColor),
	}
}

func (m *multiSelectList) addItem(text string, isDisabled bool, selected func()) *multiSelectList {
	item := &multiSelectListItem{
		Text:       text,
		Selected:   selected,
		IsDisabled: isDisabled,
	}

	m.items = append(m.items, item)
	if m.currentItem == -1 {
		m.currentItem = 0
	}
	return m
}

func (m *multiSelectList) Draw(screen tcell.Screen) {
	m.Box.DrawForSubclass(screen, m)
	x, y, width, height := m.GetInnerRect()

	for index, item := range m.items {
		if index >= height {
			break
		}
		selectedIcon := "\u0020"
		if indexOf(m.selectedItems, index) >= 0 {
			selectedIcon = "\u2713"
		}

		line := fmt.Sprintf(" [green::b]%s[-::-] %s", selectedIcon, item.Text)

		var (
			color tcell.Color
			style tcell.Style
		)

		// Text color.
		color = tview.Styles.PrimitiveBackgroundColor
		if index == m.currentItem {
		} else if item.IsDisabled {
			color = tcell.ColorDarkSlateGray
		} else {
			color = tview.Styles.PrimaryTextColor
		}

		// Background color.
		if index == m.currentItem {
			textWidth := len(item.Text)
			if textWidth > (width - 3) {
				textWidth = width - 3
			}
			for bx := 0; bx < textWidth; bx++ {
				mc, c, _, _ := screen.GetContent(x+bx, y)
				if item.IsDisabled {
					style = tcell.StyleDefault.Foreground(tview.Styles.PrimitiveBackgroundColor).Background(tcell.ColorDarkSlateGray)
				} else {
					style = tcell.StyleDefault.Foreground(tview.Styles.PrimitiveBackgroundColor).Background(tview.Styles.PrimaryTextColor)
				}
				screen.SetContent(x+bx+3, y+index, mc, c, style)
			}
		}

		tview.Print(screen, line, x, y+index, width, tview.AlignLeft, color)
	}
}

func (m *multiSelectList) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return m.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		// Return if list is empty
		if len(m.items) == 0 {
			return
		}

		switch event.Key() {
		case tcell.KeyUp:
			m.currentItem--
			if m.currentItem < 0 {
				m.currentItem = 0
			}
			m.triggerHighlightedFunc()
		case tcell.KeyDown:
			m.currentItem++
			if m.currentItem >= len(m.items) {
				m.currentItem = len(m.items) - 1
			}
			m.triggerHighlightedFunc()
		case tcell.KeyEnd:
			m.currentItem = len(m.items) - 1
			m.triggerHighlightedFunc()
		case tcell.KeyHome:
			m.currentItem = 0
			m.triggerHighlightedFunc()
		case tcell.KeyRune:
			switch event.Rune() {

			case rune(tcell.KeyEnter):
				fallthrough
			case ' ':
				if !m.getItem(m.currentItem).IsDisabled {
					m.selectedItems = toggleSelection(m.selectedItems, m.currentItem)
				}
				// Call selected function at the end.
				m.triggerSelectedFunc()
			case 'a':
				fallthrough
			case 'A':
				if m.getSelectableItemCount() == len(m.selectedItems) {
					m.selectedItems = nil
				} else {
					m.SelectAllItems()
				}
				// Call selected function at the end.
				m.triggerSelectedFunc()
			}
		}
	})
}

func (m *multiSelectList) triggerSelectedFunc() {
	if m.selected != nil {
		item := m.items[m.currentItem].Text
		m.selected(m.currentItem, item)
	}
}

func (m *multiSelectList) triggerHighlightedFunc() {
	if m.highlighted != nil {
		item := m.items[m.currentItem].Text
		m.highlighted(m.currentItem, item)
	}
}

// SetSelectedFunc sets the function which is called when the user selects a
// list item by pressing Enter or Space on the current selection. The function receives
// the item's index in the list of items (starting with 0), its main text,
// secondary text, and its shortcut rune.
func (m *multiSelectList) SetSelectedFunc(handler func(int, string)) *multiSelectList {
	m.selected = handler
	return m
}

// SetHighlightedFunc sets the function which is called when user highlights a
// list item by pressing up/down arrow keys.
func (m *multiSelectList) SetHighlightedFunc(handler func(int, string)) *multiSelectList {
	m.highlighted = handler
	return m
}

// GetSelectedItems returns a list of indices of all selected items.
func (m *multiSelectList) GetSelectedItems() []int {
	return m.selectedItems
}

// GetCurrentItem returns the index of the current item.
func (m *multiSelectList) GetCurrentItem() int {
	return m.currentItem
}

// IsCurrentItemDisabled returns true if item at given index is disabled.
func (m *multiSelectList) IsItemDisabled(index int) bool {
	return m.getItem(index).IsDisabled
}

func (m *multiSelectList) Clear() *multiSelectList {
	m.items = nil
	m.selectedItems = nil
	m.currentItem = -1
	return m
}

func (m *multiSelectList) GetItemText(index int) string {
	return m.items[index].Text
}

func (m *multiSelectList) getItem(index int) *multiSelectListItem {
	return m.items[index]
}

func (m *multiSelectList) GetItemCount() int {
	return len(m.items)
}

// getSelectableItemCount returns the count of items that can be
// selected (i.e. items that are not disabled).
func (m *multiSelectList) getSelectableItemCount() int {
	count := 0
	for _, item := range m.items {
		if !item.IsDisabled {
			count++
		}
	}
	return count
}

func (m *multiSelectList) getItemTextMultiple(indices []int) []string {
	var list []string
	for _, index := range indices {
		list = append(list, m.GetItemText(index))
	}
	return list
}

func (m *multiSelectList) SelectAllItems() {
	m.selectedItems = nil
	for index, item := range m.items {
		if !item.IsDisabled {
			m.selectedItems = append(m.selectedItems, index)
		}
	}
}

func indexOf(list []int, v int) int {
	for i, n := range list {
		if n == v {
			return i
		}
	}
	return -1
}

func toggleSelection(list []int, currentItem int) []int {
	var (
		returnList []int
		keys       = make(map[int]bool, len(list))
	)

	for _, item := range list {
		keys[item] = true
	}
	if keys[currentItem] {
		keys[currentItem] = false
	} else {
		keys[currentItem] = true
	}

	for k, v := range keys {
		if v {
			returnList = append(returnList, k)
		}
	}
	return returnList
}
