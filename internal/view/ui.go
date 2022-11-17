package view

import (
	"fmt"
	"kdiff/internal/helpers"
	"kdiff/internal/kube"
	"regexp"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"k8s.io/utils/strings/slices"
)

var unreachableError = make(map[string]error)

type getKubeResourceResult struct {
	rt        string
	ctx       string
	resources []*kube.AppsV1Resource
}

type uiOptions struct {
	showImageRegistryName bool
	showImageName         bool
	showImageTag          bool
	showImageHash         bool

	showDifferencesOnly bool
}

type uiElements struct {
	// Options
	options *uiOptions

	// Page Headers.
	headerLeft    *tview.TextView
	headerCenter1 *tview.TextView
	headerCenter2 *tview.TextView
	headerRight   *tview.TextView

	// Page Body.
	contextList      *multiSelectList
	namespaceList    *multiSelectList
	resourceTypeList *multiSelectList
	displayArea      *tview.Table

	// Page Footers.
	footerLeft   *tview.TextView
	footerCenter *tview.TextView
	footerRight  *tview.TextView

	focusedElement *tview.Primitive
}

// Initializes all UI elements with initial content.
func (u *uiElements) initialize() {
	// Initialize default options
	u.options = &uiOptions{
		showImageRegistryName: true,
		showImageName:         true,
		showImageTag:          true,
		showImageHash:         false,
		showDifferencesOnly:   false,
	}

	// Intialize header elements.
	u.headerLeft = tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)
	u.headerCenter1 = tview.NewTextView().
		SetText(getMenuText("shortcutKeys")).
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)
	u.headerCenter2 = tview.NewTextView().
		SetText(getMenuText("focusKeys")).
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)
	u.headerRight = tview.NewTextView().
		SetText(kdiffLogo).
		SetTextAlign(tview.AlignRight).
		SetDynamicColors(true)

	// Intialize body elements.
	u.contextList = createMultiSelectList("contexts")
	// u.contextList = createTviewList(u, "contexts")
	u.namespaceList = createMultiSelectList("namespaces")
	u.resourceTypeList = createMultiSelectList("resource types")
	u.displayArea = tview.NewTable().
		SetFixed(1, 0)
	u.displayArea.SetBorder(true)

	// Initialize footer elements.
	u.footerLeft = tview.NewTextView().SetTextAlign(tview.AlignLeft).SetDynamicColors(true)
	u.footerCenter = tview.NewTextView().SetTextAlign(tview.AlignCenter).SetDynamicColors(true)
	u.footerRight = tview.NewTextView().SetTextAlign(tview.AlignRight).SetDynamicColors(true).
		SetText(fmt.Sprintf("[gray]Version: %s", "dev"))

	// Update all elements.
	u.updateUI(true, true, true, true, false, false)

	// Set selection change behavior.
	u.contextList.SetSelectedFunc(func(index int, text string) {
		u.updateUI(false, false, true, false, true, false)
	}).SetHighlightedFunc(func(index int, text string) {
		if u.contextList.IsItemDisabled(index) {
			cleanName, err := cleanContextName(text)
			helpers.HandleError(err)

			re := regexp.MustCompile(`^Get "(.*)": (.*)$`)
			// Only print the error message if possible.
			if matches := re.FindAllStringSubmatch(fmt.Sprint(unreachableError[cleanName]), -1); len(matches) > 0 {
				u.footerLeft.SetText(fmt.Sprintf("[red]%v[-]", matches[0][2]))
			} else {
				u.footerLeft.SetText(fmt.Sprintf("[red]%v[-]", unreachableError[cleanName]))
			}
		} else {
			u.footerLeft.SetText("")
		}
	})
	u.namespaceList.SetSelectedFunc(func(index int, text string) {
		u.updateUI(false, false, false, false, true, false)
	})
	u.resourceTypeList.SetSelectedFunc(func(index int, text string) {
		u.updateUI(false, false, false, false, true, false)
	})
}

func (u *uiElements) updateUI(headerLeft, contextList, namespaceList, resourceTypeList, displayArea, footerLeft bool) {
	if headerLeft {
		u.updateToggles()
	}
	if contextList {
		u.updateContextList()
	}
	if namespaceList {
		u.updateNamespaceList()
	}
	if resourceTypeList {
		u.updateResourceTypeList()
	}
	if displayArea {
		u.updateDisplayArea()
	}
	if footerLeft {
		u.updateFooterLeft()
	}
}

func (u *uiElements) updateFooterLeft() {
	if *u.focusedElement == tview.Primitive(u.namespaceList) {
		u.footerLeft.SetText("> Tip: Either select 1-3 namespaces or all of them to reduce API calls.")
	} else {
		u.footerLeft.SetText("")
	}
}

// updateContextList updates the context tview list.
func (u *uiElements) updateContextList() {
	for _, ctx := range kconfig.GetContextInfo() {
		if ctx.Reachable {
			u.contextList.addItem(fmt.Sprintf("%s [%s]", ctx.Name, ctx.ServerVersion), false, nil)
		} else {
			u.contextList.addItem(fmt.Sprintf("%s (Unreachable)", ctx.Name), true, nil)
			unreachableError[ctx.Name] = ctx.UnreachableError
		}
	}
}

// getActiveContexts returns a list of cleaned context names from activeContexts.
func (u *uiElements) getActiveContexts() []string {
	activeContexts := u.contextList.getItemTextMultiple(u.contextList.GetSelectedItems())

	var cleanedContextNames []string
	for _, ctx := range activeContexts {
		cleanName, err := cleanContextName(ctx)
		helpers.HandleError(err)
		cleanedContextNames = append(cleanedContextNames, cleanName)
	}
	return cleanedContextNames
}

// updateNamespaceList updates the namespace tview list based on the active context.
func (u *uiElements) updateNamespaceList() {
	activeContexts := u.getActiveContexts()

	u.namespaceList.Clear()
	for _, ns := range kube.GetNamespacesForContextList(activeContexts) {
		u.namespaceList.addItem(ns, false, nil)
	}
}

// updateResourceTypeList updates the resourceType tview list.
func (u *uiElements) updateResourceTypeList() {
	resourceTypes := []string{
		"DaemonSet", "Deployment", "StatefulSet",
	}

	for _, rt := range resourceTypes {
		u.resourceTypeList.addItem(rt, false, nil)
	}
}

func getKubeResources(rt, ctx, ns string, out chan<- getKubeResourceResult) {
	var resources []*kube.AppsV1Resource

	if rt == "Deployment" {
		resources = kube.GetDeployments(ctx, ns)
	} else if rt == "StatefulSet" {
		resources = kube.GetStatefulSets(ctx, ns)
	} else if rt == "DaemonSet" {
		resources = kube.GetDaemonSets(ctx, ns)
	}
	out <- getKubeResourceResult{
		rt:        rt,
		ctx:       ctx,
		resources: resources,
	}
}

// updateDisplayArea updates the tview.Table element based on current selections.
func (u *uiElements) updateDisplayArea() {
	// Get current selections.
	activeContexts := u.getActiveContexts()
	activeResourceTypes := u.resourceTypeList.getItemTextMultiple(u.resourceTypeList.GetSelectedItems())
	activeNamespaces := u.namespaceList.getItemTextMultiple(u.namespaceList.GetSelectedItems())
	// If all namespaces are selected, only a single API call is needed with ns=""
	if len(u.namespaceList.GetSelectedItems()) == u.namespaceList.GetItemCount() {
		activeNamespaces = []string{""}
	}

	// Clear table.
	u.displayArea.Clear()

	var (
		allResources  = make(map[string]map[string]map[string]map[string]*kube.AppsV1Resource)
		contextIndex  = make(map[string]int)
		mistmatches   = make(map[string][]string)
		chanResources = make(chan getKubeResourceResult)
	)

	// Concurrently get resources.
	for i, ctx := range activeContexts {
		contextIndex[ctx] = i
		for _, rt := range activeResourceTypes {
			for _, ns := range activeNamespaces {
				go getKubeResources(rt, ctx, ns, chanResources)
			}
		}
	}

	// Collect all results.
	for i := 0; i < (len(activeContexts) * len(activeResourceTypes) * len(activeNamespaces)); i++ {
		result := <-chanResources

		for _, res := range result.resources {
			resourceName := res.GetName()
			if _, exists := allResources[result.rt]; !exists {
				allResources[result.rt] = make(map[string]map[string]map[string]*kube.AppsV1Resource)
			}
			if _, exists := allResources[result.rt][resourceName]; !exists {
				allResources[result.rt][resourceName] = make(map[string]map[string]*kube.AppsV1Resource)
			}
			containers := res.GetContainers()

			for _, containerName := range containers {
				if _, exists := allResources[result.rt][resourceName][containerName]; !exists {
					allResources[result.rt][resourceName][containerName] = make(map[string]*kube.AppsV1Resource)
				}
				allResources[result.rt][resourceName][containerName][result.ctx] = res
			}
		}
	}

	// Identify mismatching images
	for rt, resourceMap := range allResources {
		for resourceName, containerMap := range resourceMap {
			for containerName, contextMap := range containerMap {
				var allImages []string
				for _, res := range contextMap {
					fullImageName, err := res.GetImage(containerName, true, true, true, true)
					helpers.HandleError(err)
					allImages = append(allImages, fullImageName)
				}
				if len(helpers.GetUniqueStrings(allImages)) != 1 {
					key := fmt.Sprintf("%s-%s", rt, resourceName)
					mistmatches[key] = append(mistmatches[key], containerName)
				}
			}
		}
	}

	// Fill table.
	row, column := 1, 1
	u.displayArea.SetCell(0, 0, tview.NewTableCell("").
		SetAttributes(tcell.AttrBold).
		SetExpansion(1).
		SetTextColor(tcell.GetColor("#f5bd07")))
	u.displayArea.SetCell(0, 1, tview.NewTableCell("Name").
		SetAttributes(tcell.AttrBold).
		SetExpansion(3).
		SetTextColor(tcell.GetColor("#f5bd07")))
	for _, rt := range activeResourceTypes {
		// Set header only if there are any resources to be displayed.
		if _, exists := allResources[rt]; exists {
			u.displayArea.SetCell(row, 0, tview.NewTableCell(rt).
				SetAttributes(tcell.AttrBold).
				SetTextColor(tcell.GetColor("#f5bd07")))
		}
		for resourceName, containerMap := range allResources[rt] {
			mistmatchesKey := fmt.Sprintf("%s-%s", rt, resourceName)
			if u.options.showDifferencesOnly {
				if _, exists := mistmatches[mistmatchesKey]; exists {
					setTableCell(u, row, 1, resourceName)
				} else {
					continue
				}
			} else {
				setTableCell(u, row, 1, resourceName)
			}

			for containerName := range containerMap {
				// Set empty cell at column 1 if it doesn't already contain some text.
				if len(u.displayArea.GetCell(row, 1).Text) < 1 {
					setTableCell(u, row, 1, "")
				}

				for _, ctx := range activeContexts {
					column = contextIndex[ctx] + 2

					if _, exists := containerMap[containerName][ctx]; exists {
						// Lookup errors should be ignored here just in case user disables all 4 options.
						imageDisplayName, _ := containerMap[containerName][ctx].GetImage(
							containerName,
							u.options.showImageRegistryName,
							u.options.showImageName,
							u.options.showImageTag,
							u.options.showImageHash,
						)
						u.displayArea.SetCell(0, column, tview.NewTableCell(ctx).
							SetAttributes(tcell.AttrBold).
							SetExpansion(6).
							SetTextColor(tcell.GetColor("#f5bd07")))
						if _, exists := mistmatches[mistmatchesKey]; exists {
							if slices.Contains(mistmatches[mistmatchesKey], containerName) {
								setTableCellWithBackgroundColor(u, row, column, imageDisplayName, tcell.ColorRed)
							}
						} else if !u.options.showDifferencesOnly {
							setTableCell(u, row, column, imageDisplayName)
						}
					} else {
						// Set empty cell since there's nothing to display
						setTableCell(u, row, column, "")
					}
				}
				row++
			}
		}
	}
}

func setTableCell(u *uiElements, row int, column int, text string) {
	u.displayArea.SetCell(row, column, tview.NewTableCell(text))
}

func setTableCellWithBackgroundColor(u *uiElements, row int, column int, text string, color tcell.Color) {
	u.displayArea.SetCell(row, column, tview.NewTableCell(text).SetTextColor(color))
}

// createMultiSelectList returns a tview.List with default settings and a Title.
func createMultiSelectList(title string) *multiSelectList {
	multiSelectlist := newMultiSelectList()
	multiSelectlist.SetBorder(true).
		SetBorder(true).
		SetTitle(fmt.Sprintf(" %s ", cases.Title(language.AmericanEnglish).String(title))).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			// Disable tab key since its used for navigation.
			if event.Key() == tcell.KeyTAB || event.Key() == tcell.KeyBacktab {
				return nil
			}
			return event
		})
	return multiSelectlist
}

// getMenuText returns key combinations & shortcuts for the app header.
func getMenuText(mode string) string {
	var str string

	shortcutKeys := map[string]string{
		"<TAB>":         "Cycle forward",
		"<Shift + TAB>": "Cycle backward",
		"<Space>":       "Select item",
		"<a>":           "Select all (toggle)",
	}
	focusKeys := map[string]string{
		"<1>": "Contexts",
		"<2>": "Namespaces",
		"<3>": "Resource Types",
		"<4>": "Display Area",
	}

	if mode == "shortcutKeys" {
		for _, k := range helpers.GetSortedMapKeys(shortcutKeys) {
			str += fmt.Sprintf("[yellow]%15s  [white]%s\n", k, shortcutKeys[k])
		}
	} else if mode == "focusKeys" {
		for _, k := range helpers.GetSortedMapKeys(focusKeys) {
			str += fmt.Sprintf("[blue]%8s  [white]%s\n", k, focusKeys[k])
		}
	}
	return str
}

func (u *uiElements) updateToggles() {
	options := map[string]bool{
		"<r>  Show Image Registry Name": u.options.showImageRegistryName,
		"<n>  Show Image Name":          u.options.showImageName,
		"<t>  Show Image Tag":           u.options.showImageTag,
		"<h>  Show Image Hash":          u.options.showImageHash,
		"<d>  Show Differences Only":    u.options.showDifferencesOnly,
	}
	var str string
	for _, k := range helpers.GetSortedMapKeysBool(options) {
		color := "gray"
		if options[k] {
			color = "green"
		}
		str = str + fmt.Sprintf("[%s]%s\n", color, k)
	}
	u.headerLeft.SetText(str)
}

func cleanContextName(ctx string) (string, error) {
	re := regexp.MustCompile(`^(.+?) [[(].*[])]$`)
	if matches := re.FindAllStringSubmatch(ctx, -1); len(matches) > 0 {
		return matches[0][1], nil
	}
	return "", fmt.Errorf("couldn't clean context name '%s'", ctx)
}
