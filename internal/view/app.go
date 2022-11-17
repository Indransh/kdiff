package view

import (
	"kdiff/internal/config"
	"kdiff/internal/kube"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	kdiffLogo = `[#f5bd07::b]__     [green::-]   .___.__  _____  _____ 
[#f5bd07::b]|  | __[green::-] __| _/|__|/ ____\/ ____\
[#f5bd07::b]|  |/ /[green::-]/ __ | |  \   __\\   __\ 
[#f5bd07::b]|    <[green::-]/ /_/ | |  ||  |   |  |   
[#f5bd07::b]|__|_ [green::-]\____ | |__||__|   |__|   
[#f5bd07::b]     \/[green::-]    \/                   `
)

var (
	app     *tview.Application
	kconfig kube.KubeConfig
)

func App(config config.ConfigFlags) {
	// Parse kubeconfig and Initialize kubernetes client for each context.
	kconfig.ParseConfig(config.KubeConfig)
	kconfig.InitializeClients()

	// Create new tview app & run it.
	app = tview.NewApplication()
	buildAppUI()
	if err := app.Run(); err != nil {
		panic(err)
	}
}

// temp for debugging - should be inside buildAppUI()
var ui uiElements

func buildAppUI() {
	ui.initialize()

	// Create the layout.
	grid := tview.NewGrid().
		SetRows(6, 0, 1).
		SetBorders(false).
		// Header Grid
		AddItem(tview.NewGrid().
			SetColumns(-3, -2, -4, -3).
			SetBorders(false).
			AddItem(ui.headerLeft, 0, 0, 1, 1, 0, 0, false).
			AddItem(ui.headerCenter1, 0, 1, 1, 1, 0, 0, false).
			AddItem(ui.headerCenter2, 0, 2, 1, 1, 0, 0, false).
			AddItem(ui.headerRight, 0, 3, 1, 1, 0, 0, false),
			0, 0, 1, 1, 0, 0, false,
		).
		// Body Grid
		AddItem(tview.NewGrid().
			SetRows(-2, -2, -1).
			SetColumns(30, 0).
			SetBorders(false).
			AddItem(ui.contextList, 0, 0, 1, 1, 0, 0, true).
			AddItem(ui.namespaceList, 1, 0, 1, 1, 0, 0, false).
			AddItem(ui.resourceTypeList, 2, 0, 1, 1, 0, 0, false).
			AddItem(ui.displayArea, 0, 1, 3, 1, 0, 0, false),
			1, 0, 1, 1, 0, 0, false).
		// Footer Grid
		AddItem(tview.NewGrid().
			SetColumns(-18, -2).
			SetBorders(false).
			AddItem(ui.footerLeft, 0, 0, 1, 1, 0, 0, false).
			// AddItem(ui.footerCenter, 0, 1, 1, 1, 0, 0, false).
			AddItem(ui.footerRight, 0, 2, 1, 1, 0, 0, false),
			2, 0, 1, 1, 0, 0, false)

	// Update focused element
	x := tview.Primitive(ui.contextList)
	ui.focusedElement = &x
	ui.updateUI(false, false, false, false, false, true)

	// Setup the pages
	pages := tview.NewPages().
		AddPage(" kdiff ", grid, true, true)
	app.SetRoot(pages, true).
		SetFocus(ui.contextList).

		// Setup navigation
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			navOrder := []tview.Primitive{
				ui.contextList,
				ui.namespaceList,
				ui.resourceTypeList,
				ui.displayArea,
			}

			for i, element := range navOrder {
				// Number based navigation.
				if event.Rune() == rune(48+i+1) {
					app.SetFocus(element)
					ui.focusedElement = &element
					break
				}

				// TAB & BackTAB (Shift + TAB) navigation.
				if app.GetFocus() == element {
					if event.Key() == tcell.KeyTAB {
						nextElementIndex := (len(navOrder) + i + 1) % len(navOrder)
						app.SetFocus(navOrder[nextElementIndex])
						ui.focusedElement = &navOrder[nextElementIndex]
						break
					} else if event.Key() == tcell.KeyBacktab {
						lastElementIndex := (len(navOrder) + i - 1) % len(navOrder)
						app.SetFocus(navOrder[lastElementIndex])
						ui.focusedElement = &navOrder[lastElementIndex]
						break
					}
				}
			}
			ui.updateUI(false, false, false, false, false, true)

			// UI display options.
			switch event.Rune() {
			case 'h':
				ui.options.showImageHash = !ui.options.showImageHash
				ui.updateUI(true, false, false, false, true, false)
			case 't':
				ui.options.showImageTag = !ui.options.showImageTag
				ui.updateUI(true, false, false, false, true, false)
			case 'n':
				ui.options.showImageName = !ui.options.showImageName
				ui.updateUI(true, false, false, false, true, false)
			case 'r':
				ui.options.showImageRegistryName = !ui.options.showImageRegistryName
				ui.updateUI(true, false, false, false, true, false)
			case 'd':
				ui.options.showDifferencesOnly = !ui.options.showDifferencesOnly
				ui.updateUI(true, false, false, false, true, false)
			}

			// Enable display area table selection only if its in focus.
			if app.GetFocus() == ui.displayArea {
				ui.displayArea.SetSelectable(true, false)
			} else {
				ui.displayArea.SetSelectable(false, false)
			}

			return event
		})
}
