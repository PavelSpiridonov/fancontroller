package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/andlabs/ui"
	"github.com/getlantern/systray"

	"./icon"
)

type AppGUI struct {
	mainWindow               *ui.Window
	selectPortWindow         *ui.Window
	portEdit                 *ui.EditableCombobox
	autoStart                *ui.Checkbox
	applyButton, resetButton *ui.Button

	statusPage                             StatusPage
	sensorPage                             SensorPage
	fan1Page, fan2Page, fan3Page, fan4Page FanPage

	showAppMenu, quitMenu *systray.MenuItem

	serial    *Serial
	appConfig *AppConfig
}

type StatusPage struct {
	TempA, TempB, TempC, TempD                                                                     *ui.ProgressBar
	Fan1A, Fan1B, Fan2A, Fan2B, Fan3A, Fan3B, Fan4A, Fan4B                                         *ui.ProgressBar
	Output1, Output2, Output3, Output4                                                             *ui.ProgressBar
	TempALabel, TempBLabel, TempCLabel, TempDLabel                                                 *ui.Label
	Fan1ALabel, Fan1BLabel, Fan2ALabel, Fan2BLabel, Fan3ALabel, Fan3BLabel, Fan4ALabel, Fan4BLabel *ui.Label
	Output1Label, Output2Label, Output3Label, Output4Label                                         *ui.Label
}

type SensorPage struct {
	SensorA, SensorB, SensorC, SensorD *ui.Combobox
}

type FanPage struct {
	FanTypeA, FanTypeB *ui.Combobox
	Control            *ui.Combobox
	Power              *ui.Spinbox
	MinTemp            *ui.Spinbox
	MaxTemp            *ui.Spinbox
	AllowStop          *ui.Checkbox
}

func NewAppGUI(serial *Serial, appConfig *AppConfig) *AppGUI {
	appGUI := AppGUI{
		serial:    serial,
		appConfig: appConfig,
	}
	serial.appGUI = &appGUI
	return &appGUI
}

func (app *AppGUI) ShowError(err error, main bool) {
	ui.QueueMain(func() {
		var window *ui.Window
		if main {
			window = app.mainWindow
		} else {
			window = app.selectPortWindow
		}
		if window != nil {
			ui.MsgBoxError(window, "Error", err.Error())
		}
	})
}

func (app *AppGUI) ShowMessage(msg string) {
	ui.QueueMain(func() {
		var window *ui.Window
		if app.mainWindow.Visible() {
			window = app.mainWindow
		} else {
			window = app.selectPortWindow
		}
		if window != nil {
			ui.MsgBox(window, "Message", msg)
		}
	})
}

func (app *AppGUI) UpdateConfig(portName string) {
	notExist := true
	ports := make([]string, 0)
	ports = append(ports, portName)
	for _, v := range app.appConfig.Ports {
		if v != portName {
			ports = append(ports, v)
		} else {
			notExist = false
		}
	}
	app.appConfig.Ports = ports
	if notExist {
		app.portEdit.Append(portName)
	}
	app.appConfig.AutoStartInSystray = app.autoStart.Checked()
	writeAppConfig(app.appConfig)
}

func (app *AppGUI) SetupUI() {
	app.makeSelectPortWindow()
	app.makeMainWindow()

	ui.OnShouldQuit(func() bool {
		if app.mainWindow != nil {
			app.mainWindow.Destroy()
		}
		app.selectPortWindow.Destroy()
		return true
	})

	if !(app.appConfig.AutoStartInSystray && app.serial.ConnectToController(app.portEdit.Text())) {
		app.showSelectPortWindow()
	}

	systray.Run(app.onSystrayReady, app.onSystrayExit)
}

func getAppTitle() string {
	return fmt.Sprintf("Fan Controller v%s", APP_VERSION)
}

func (app *AppGUI) makeSelectPortWindow() {
	app.selectPortWindow = ui.NewWindow(getAppTitle(), 320, 200, false)
	app.selectPortWindow.SetMargined(true)
	app.selectPortWindow.OnClosing(func(*ui.Window) bool {
		systray.Quit()
		return true
	})

	grid := ui.NewGrid()
	grid.SetPadded(true)
	app.selectPortWindow.SetChild(grid)

	grid1 := ui.NewGrid()
	grid1.SetPadded(true)
	grid.Append(grid1, 0, 0, 1, 1, true, ui.AlignCenter, true, ui.AlignCenter)

	grid1.Append(ui.NewLabel("Port:"), 0, 0, 1, 1, false, ui.AlignEnd, false, ui.AlignCenter)
	app.portEdit = ui.NewEditableCombobox()
	for _, v := range app.appConfig.Ports {
		app.portEdit.Append(v)
	}
	if len(app.appConfig.Ports) > 0 {
		app.portEdit.SetText(app.appConfig.Ports[0])
	}
	grid1.Append(app.portEdit, 1, 0, 2, 1, false, ui.AlignFill, false, ui.AlignCenter)

	app.autoStart = ui.NewCheckbox("Auto start in systray")
	grid1.Append(app.autoStart, 1, 1, 1, 1, false, ui.AlignEnd, false, ui.AlignCenter)
	app.autoStart.SetChecked(app.appConfig.AutoStartInSystray)

	connectButton := ui.NewButton("Connect")
	connectButton.OnClicked(func(*ui.Button) {
		if app.serial.ConnectToController(app.portEdit.Text()) {
			app.showMainWindow()
		}
	})
	grid1.Append(connectButton, 2, 1, 1, 1, false, ui.AlignEnd, false, ui.AlignCenter)

}

func setVisibleMenu(visible bool, menu *systray.MenuItem) {
	if runtime.GOOS == "windows" {
		if visible {
			menu.Enable()

		} else {
			menu.Disable()
		}
	} else {
		if visible {
			menu.Show()
		} else {
			menu.Hide()
		}
	}
}

func (app *AppGUI) onSystrayReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle(getAppTitle())
	systray.SetTooltip(getAppTitle())

	app.showAppMenu = systray.AddMenuItem("Show App", "Show App")
	if !app.appConfig.AutoStartInSystray {
		setVisibleMenu(false, app.showAppMenu)
	}
	go func() {
		for {
			<-app.showAppMenu.ClickedCh
			app.showMainWindow()
		}
	}()
	app.quitMenu = systray.AddMenuItem("Quit", "Quit")
	go func() {
		<-app.quitMenu.ClickedCh
		systray.Quit()
	}()
}

func (app *AppGUI) onSystrayExit() {
	app.CloseMainWindow(false)
}

func (app *AppGUI) makeMainWindow() {
	app.mainWindow = ui.NewWindow(getAppTitle(), 320, 200, false)
	app.mainWindow.SetMargined(false)
	app.mainWindow.OnClosing(func(*ui.Window) bool {
		app.hideMainWindow()
		return false
	})

	grid := ui.NewGrid()
	grid.SetPadded(true)

	app.mainWindow.SetChild(grid)
	app.mainWindow.SetMargined(true)

	tab := ui.NewTab()
	grid.Append(tab, 0, 0, 1, 1, true, ui.AlignFill, true, ui.AlignFill)

	tab.Append("Status", app.makeStatusPage())
	tab.SetMargined(0, true)
	tab.Append("Sensors", app.makeSensorsPage())
	tab.SetMargined(1, true)
	tab.Append("Fans 1A and 1B", app.makeFansPage(&app.fan1Page, 1))
	tab.SetMargined(2, true)
	tab.Append("Fans 2A and 2B", app.makeFansPage(&app.fan2Page, 2))
	tab.SetMargined(3, true)
	tab.Append("Fans 3A and 3B", app.makeFansPage(&app.fan3Page, 3))
	tab.SetMargined(4, true)
	tab.Append("Fans 4A and 4B", app.makeFansPage(&app.fan4Page, 4))
	tab.SetMargined(5, true)

	gridBtns := ui.NewGrid()
	gridBtns.SetPadded(true)
	grid.Append(gridBtns, 1, 0, 1, 1, false, ui.AlignFill, false, ui.AlignEnd)

	app.applyButton = ui.NewButton("Apply")
	app.applyButton.OnClicked(func(*ui.Button) {
		app.disableActionButtons()
		app.serial.ApplyConfig(app.getConfig())
	})
	gridBtns.Append(app.applyButton, 0, 0, 1, 1, false, ui.AlignFill, false, ui.AlignEnd)
	app.resetButton = ui.NewButton("Reset")
	app.resetButton.OnClicked(func(*ui.Button) {
		app.disableActionButtons()
		app.UpdateConfigPages()
	})
	gridBtns.Append(app.resetButton, 0, 1, 1, 1, false, ui.AlignFill, false, ui.AlignEnd)
	cancelButton := ui.NewButton("Cancel")
	cancelButton.OnClicked(func(*ui.Button) {
		app.CloseMainWindow(true)
	})
	gridBtns.Append(cancelButton, 0, 2, 1, 1, false, ui.AlignFill, false, ui.AlignEnd)

	app.UpdateActionButtons(false)
}

func (app *AppGUI) addProgressBarOnStatusPage(index int, name, value string, grid *ui.Grid) (*ui.ProgressBar, *ui.Label) {
	progressBar := ui.NewProgressBar()
	progressBar.SetValue(0)
	label := ui.NewLabel(value)
	grid.Append(ui.NewLabel(name), 0, index, 1, 1, false, ui.AlignEnd, false, ui.AlignCenter)
	grid.Append(progressBar, 1, index, 1, 1, false, ui.AlignFill, false, ui.AlignCenter)
	grid.Append(label, 2, index, 1, 1, false, ui.AlignStart, false, ui.AlignCenter)
	return progressBar, label
}

func (app *AppGUI) makeStatusPage() ui.Control {
	grid := ui.NewGrid()
	grid.SetPadded(true)

	hbox := ui.NewVerticalBox()
	hbox.SetPadded(true)
	grid.Append(hbox, 0, 0, 1, 1, true, ui.AlignCenter, true, ui.AlignCenter)

	grid1 := ui.NewGrid()
	grid1.SetPadded(true)
	hbox.Append(grid1, false)
	hbox.Append(ui.NewHorizontalSeparator(), false)

	grid2 := ui.NewGrid()
	grid2.SetPadded(true)
	hbox.Append(grid2, false)
	hbox.Append(ui.NewHorizontalSeparator(), false)

	grid3 := ui.NewGrid()
	grid3.SetPadded(true)
	hbox.Append(grid3, false)

	app.statusPage.TempA, app.statusPage.TempALabel = app.addProgressBarOnStatusPage(0, "Sensor A:", "0 °C", grid1)
	app.statusPage.TempB, app.statusPage.TempBLabel = app.addProgressBarOnStatusPage(1, "Sensor B:", "0 °C", grid1)
	app.statusPage.TempC, app.statusPage.TempCLabel = app.addProgressBarOnStatusPage(2, "Sensor C:", "0 °C", grid1)
	app.statusPage.TempD, app.statusPage.TempDLabel = app.addProgressBarOnStatusPage(3, "Sensor D:", "0 °C", grid1)

	app.statusPage.Fan1A, app.statusPage.Fan1ALabel = app.addProgressBarOnStatusPage(0, "Fan 1A:", "0 RPM", grid2)
	app.statusPage.Fan1B, app.statusPage.Fan1BLabel = app.addProgressBarOnStatusPage(1, "Fan 1B:", "0 RPM", grid2)
	app.statusPage.Fan2A, app.statusPage.Fan2ALabel = app.addProgressBarOnStatusPage(2, "Fan 2A:", "0 RPM", grid2)
	app.statusPage.Fan2B, app.statusPage.Fan2BLabel = app.addProgressBarOnStatusPage(3, "Fan 2B:", "0 RPM", grid2)
	app.statusPage.Fan3A, app.statusPage.Fan3ALabel = app.addProgressBarOnStatusPage(4, "Fan 3A:", "0 RPM", grid2)
	app.statusPage.Fan3B, app.statusPage.Fan3BLabel = app.addProgressBarOnStatusPage(5, "Fan 3B:", "0 RPM", grid2)
	app.statusPage.Fan4A, app.statusPage.Fan4ALabel = app.addProgressBarOnStatusPage(6, "Fan 4A:", "0 RPM", grid2)
	app.statusPage.Fan4B, app.statusPage.Fan4BLabel = app.addProgressBarOnStatusPage(7, "Fan 4B:", "0 RPM", grid2)

	app.statusPage.Output1, app.statusPage.Output1Label = app.addProgressBarOnStatusPage(0, "Output Fans 1A, 1B:", "0 %", grid3)
	app.statusPage.Output2, app.statusPage.Output2Label = app.addProgressBarOnStatusPage(1, "Output Fans 2A, 2B:", "0 %", grid3)
	app.statusPage.Output3, app.statusPage.Output3Label = app.addProgressBarOnStatusPage(2, "Output Fans 3A, 3B:", "0 %", grid3)
	app.statusPage.Output4, app.statusPage.Output4Label = app.addProgressBarOnStatusPage(3, "Output Fans 4A, 4B:", "0 %", grid3)

	return grid
}

func (app *AppGUI) makeFansPage(fanPage *FanPage, fan int) ui.Control {
	grid := ui.NewGrid()
	grid.SetPadded(true)

	grid1 := ui.NewGrid()
	grid1.SetPadded(true)
	grid.Append(grid1, 0, 0, 1, 1, true, ui.AlignCenter, true, ui.AlignCenter)

	fanTypes := []string{"Not Used", "2 Wire", "3 Wire - x1 tacho", "3 Wire - x2 tacho", "3 Wire - x4 tacho", "4 Wire"}
	fanPage.FanTypeA = app.addComboBoxOnFanPage(1, 0, fmt.Sprintf("Fan %dA:", fan), fanTypes, grid1)
	fanPage.FanTypeB = app.addComboBoxOnFanPage(3, 0, fmt.Sprintf("Fan %dA:", fan), fanTypes, grid1)

	controlTypes := []string{"Sensor A", "Sensor B", "Sensor C", "Sensor D", "Sensor A - Sensor D", "Sensor B - Sensor D", "Sensor C - Sensor D", "Manual control"}
	fanPage.Control = app.addComboBoxOnFanPage(2, 1, "Control:", controlTypes, grid1)

	grid2 := ui.NewGrid()
	grid2.SetPadded(true)
	grid1.Append(grid2, 1, 2, 4, 1, false, ui.AlignStart, false, ui.AlignCenter)

	fanPage.Power = app.addSpinBox(0, 100)
	fanPage.MinTemp = app.addSpinBox(0, app.appConfig.MaxTemp)
	fanPage.MaxTemp = app.addSpinBox(0, app.appConfig.MaxTemp)
	grid2.Append(ui.NewLabel("Vary the fan power from"), 0, 0, 1, 1, false, ui.AlignEnd, false, ui.AlignCenter)
	grid2.Append(fanPage.Power, 1, 0, 1, 1, false, ui.AlignFill, false, ui.AlignFill)
	grid2.Append(ui.NewLabel("%"), 2, 0, 1, 1, false, ui.AlignStart, false, ui.AlignCenter)
	grid2.Append(ui.NewLabel("at"), 3, 0, 1, 1, false, ui.AlignEnd, false, ui.AlignCenter)
	grid2.Append(fanPage.MinTemp, 4, 0, 1, 1, false, ui.AlignFill, false, ui.AlignCenter)
	grid2.Append(ui.NewLabel("°C"), 5, 0, 1, 1, false, ui.AlignStart, false, ui.AlignCenter)
	grid2.Append(ui.NewLabel("to 100% at"), 6, 0, 1, 1, false, ui.AlignFill, false, ui.AlignCenter)
	grid2.Append(fanPage.MaxTemp, 7, 0, 1, 1, false, ui.AlignFill, false, ui.AlignCenter)
	grid2.Append(ui.NewLabel("°C"), 8, 0, 1, 1, false, ui.AlignEnd, false, ui.AlignCenter)

	grid3 := ui.NewGrid()
	grid3.SetPadded(false)
	grid1.Append(grid3, 2, 3, 2, 1, false, ui.AlignStart, false, ui.AlignCenter)

	fanPage.AllowStop = app.addCheckBox("Completely stop the fan when the temperature is below the minimum")
	grid3.Append(fanPage.AllowStop, 0, 0, 1, 1, false, ui.AlignFill, false, ui.AlignCenter)

	return grid
}

func (app *AppGUI) makeSensorsPage() ui.Control {
	grid := ui.NewGrid()
	grid.SetPadded(true)

	grid1 := ui.NewGrid()
	grid1.SetPadded(true)
	grid.Append(grid1, 0, 0, 1, 1, true, ui.AlignCenter, true, ui.AlignCenter)

	sensorsTypes := []string{"Not Connected", "Use °C", "Use °F"}
	app.sensorPage.SensorA = app.addComboBoxOnSensorsPage(0, "A", sensorsTypes, grid1)
	app.sensorPage.SensorB = app.addComboBoxOnSensorsPage(1, "B", sensorsTypes, grid1)
	app.sensorPage.SensorC = app.addComboBoxOnSensorsPage(2, "C", sensorsTypes, grid1)
	app.sensorPage.SensorD = app.addComboBoxOnSensorsPage(3, "D", sensorsTypes, grid1)

	return grid
}

func (app *AppGUI) addComboBoxOnSensorsPage(index int, label string, items []string, grid *ui.Grid) *ui.Combobox {
	combobox := ui.NewCombobox()
	combobox.OnSelected(func(*ui.Combobox) {
		app.UpdateActionButtons(true)
	})
	for _, s := range items {
		combobox.Append(s)
	}
	grid.Append(ui.NewLabel("Temperature sensor "+label+":"), 0, index, 1, 1, false, ui.AlignFill, false, ui.AlignCenter)
	grid.Append(combobox, 1, index, 1, 1, false, ui.AlignFill, false, ui.AlignCenter)
	return combobox
}

func (app *AppGUI) addComboBoxOnFanPage(col, row int, label string, items []string, grid *ui.Grid) *ui.Combobox {
	comboBox := ui.NewCombobox()
	comboBox.OnSelected(func(*ui.Combobox) {
		app.UpdateActionButtons(true)
	})
	for _, s := range items {
		comboBox.Append(s)
	}
	grid.Append(ui.NewLabel(label), col, row, 1, 1, false, ui.AlignEnd, false, ui.AlignCenter)
	grid.Append(comboBox, col+1, row, 1, 1, false, ui.AlignStart, false, ui.AlignCenter)
	return comboBox
}

func (app *AppGUI) addSpinBox(min, max int) *ui.Spinbox {
	spinBox := ui.NewSpinbox(min, max)
	spinBox.OnChanged(func(*ui.Spinbox) {
		app.UpdateActionButtons(true)
	})
	return spinBox
}

func (app *AppGUI) addCheckBox(label string) *ui.Checkbox {
	checkbox := ui.NewCheckbox(label)
	checkbox.OnToggled(func(*ui.Checkbox) {
		app.UpdateActionButtons(true)
	})
	return checkbox
}

func (app *AppGUI) disableActionButtons() {
	app.applyButton.Disable()
	app.resetButton.Disable()
}

func (app *AppGUI) UpdateActionButtons(enable bool) {
	if enable {
		app.applyButton.Enable()
		app.resetButton.Enable()
	} else {
		app.applyButton.Disable()
		app.resetButton.Disable()
	}
}

func (app *AppGUI) updateTempOnStatusPage(progressBar *ui.ProgressBar, label *ui.Label, temp int8) {
	progressBar.SetValue(app.tempToPerc(temp))
	label.SetText(fmt.Sprintf("%d °C", temp))
}

func (app *AppGUI) updateRPMOnStatusPage(progressBar *ui.ProgressBar, label *ui.Label, rpm int16) {
	progressBar.SetValue(app.rpmToPerc(rpm))
	label.SetText(fmt.Sprintf("%d RPM", rpm))
}

func (app *AppGUI) updateOutputOnStatusPage(progressBar *ui.ProgressBar, label *ui.Label, output int8) {
	progressBar.SetValue(int(output))
	label.SetText(fmt.Sprintf("%d %s", output, "%"))
}

func (app *AppGUI) UpdateStatusPage() {
	status := app.serial.GetStatus()
	config := app.serial.GetConfig()

	if config.SensorTypes.SensorTypeA != SENSOR_NOT_CONNECTED {
		app.updateTempOnStatusPage(app.statusPage.TempA, app.statusPage.TempALabel, status.Temperatures.SensorA)
	} else {
		app.updateTempOnStatusPage(app.statusPage.TempA, app.statusPage.TempALabel, 0)
	}
	if config.SensorTypes.SensorTypeB != SENSOR_NOT_CONNECTED {
		app.updateTempOnStatusPage(app.statusPage.TempB, app.statusPage.TempBLabel, status.Temperatures.SensorB)
	} else {
		app.updateTempOnStatusPage(app.statusPage.TempB, app.statusPage.TempBLabel, 0)
	}
	if config.SensorTypes.SensorTypeC != SENSOR_NOT_CONNECTED {
		app.updateTempOnStatusPage(app.statusPage.TempC, app.statusPage.TempCLabel, status.Temperatures.SensorC)
	} else {
		app.updateTempOnStatusPage(app.statusPage.TempC, app.statusPage.TempCLabel, 0)
	}
	if config.SensorTypes.SensorTypeD != SENSOR_NOT_CONNECTED {
		app.updateTempOnStatusPage(app.statusPage.TempD, app.statusPage.TempDLabel, status.Temperatures.SensorD)
	} else {
		app.updateTempOnStatusPage(app.statusPage.TempD, app.statusPage.TempDLabel, 0)
	}
	if config.Fan1Config.FanTypeA != FAN_NOT_CONNECTED {
		app.updateRPMOnStatusPage(app.statusPage.Fan1A, app.statusPage.Fan1ALabel, status.RPMS.Fan1A)
	} else {
		app.updateRPMOnStatusPage(app.statusPage.Fan1A, app.statusPage.Fan1ALabel, 0)
	}
	if config.Fan1Config.FanTypeB != FAN_NOT_CONNECTED {
		app.updateRPMOnStatusPage(app.statusPage.Fan1B, app.statusPage.Fan1BLabel, status.RPMS.Fan1B)
	} else {
		app.updateRPMOnStatusPage(app.statusPage.Fan1B, app.statusPage.Fan1BLabel, 0)
	}
	if config.Fan2Config.FanTypeA != FAN_NOT_CONNECTED {
		app.updateRPMOnStatusPage(app.statusPage.Fan2A, app.statusPage.Fan2ALabel, status.RPMS.Fan2A)
	} else {
		app.updateRPMOnStatusPage(app.statusPage.Fan2A, app.statusPage.Fan2ALabel, 0)
	}
	if config.Fan2Config.FanTypeB != FAN_NOT_CONNECTED {
		app.updateRPMOnStatusPage(app.statusPage.Fan2B, app.statusPage.Fan2BLabel, status.RPMS.Fan2B)
	} else {
		app.updateRPMOnStatusPage(app.statusPage.Fan2B, app.statusPage.Fan2BLabel, 0)
	}
	if config.Fan3Config.FanTypeA != FAN_NOT_CONNECTED {
		app.updateRPMOnStatusPage(app.statusPage.Fan3A, app.statusPage.Fan3ALabel, status.RPMS.Fan3A)
	} else {
		app.updateRPMOnStatusPage(app.statusPage.Fan3A, app.statusPage.Fan3ALabel, 0)
	}
	if config.Fan3Config.FanTypeB != FAN_NOT_CONNECTED {
		app.updateRPMOnStatusPage(app.statusPage.Fan3B, app.statusPage.Fan3BLabel, status.RPMS.Fan3B)
	} else {
		app.updateRPMOnStatusPage(app.statusPage.Fan3B, app.statusPage.Fan3BLabel, 0)
	}
	if config.Fan4Config.FanTypeA != FAN_NOT_CONNECTED {
		app.updateRPMOnStatusPage(app.statusPage.Fan4A, app.statusPage.Fan4ALabel, status.RPMS.Fan4A)
	} else {
		app.updateRPMOnStatusPage(app.statusPage.Fan4A, app.statusPage.Fan4ALabel, 0)
	}
	if config.Fan4Config.FanTypeB != FAN_NOT_CONNECTED {
		app.updateRPMOnStatusPage(app.statusPage.Fan4B, app.statusPage.Fan4BLabel, status.RPMS.Fan4B)
	} else {
		app.updateRPMOnStatusPage(app.statusPage.Fan4B, app.statusPage.Fan4BLabel, 0)
	}
	app.updateOutputOnStatusPage(app.statusPage.Output1, app.statusPage.Output1Label, status.Outputs.Fan1)
	app.updateOutputOnStatusPage(app.statusPage.Output2, app.statusPage.Output2Label, status.Outputs.Fan2)
	app.updateOutputOnStatusPage(app.statusPage.Output3, app.statusPage.Output3Label, status.Outputs.Fan3)
	app.updateOutputOnStatusPage(app.statusPage.Output4, app.statusPage.Output4Label, status.Outputs.Fan4)
}

func (app *AppGUI) UpdateConfigPages() {
	config := app.serial.GetConfig()
	app.sensorPage.SensorA.SetSelected(app.sensorTypeToIndex(config.SensorTypes.SensorTypeA))
	app.sensorPage.SensorB.SetSelected(app.sensorTypeToIndex(config.SensorTypes.SensorTypeB))
	app.sensorPage.SensorC.SetSelected(app.sensorTypeToIndex(config.SensorTypes.SensorTypeC))
	app.sensorPage.SensorD.SetSelected(app.sensorTypeToIndex(config.SensorTypes.SensorTypeD))

	app.updateConfigPage(&config.Fan1Config, &app.fan1Page)
	app.updateConfigPage(&config.Fan2Config, &app.fan2Page)
	app.updateConfigPage(&config.Fan3Config, &app.fan3Page)
	app.updateConfigPage(&config.Fan4Config, &app.fan4Page)

	app.UpdateActionButtons(false)
}

func (app *AppGUI) CloseMainWindow(selectPort bool) {
	app.serial.StopRead()

	if selectPort {
		app.showSelectPortWindow()
	} else {
		ui.QueueMain(func() {
			ui.Quit()
			os.Exit(0)
		})
	}
}

func (app *AppGUI) showSelectPortWindow() {
	ui.QueueMain(func() {
		if app.showAppMenu != nil {
			setVisibleMenu(false, app.showAppMenu)
		}
		if app.mainWindow != nil {
			app.mainWindow.Hide()
		}
		if app.selectPortWindow != nil {
			app.selectPortWindow.Show()
		}
	})
}

func (app *AppGUI) hideMainWindow() {
	ui.QueueMain(func() {
		if app.showAppMenu != nil {
			setVisibleMenu(true, app.showAppMenu)
		}
		if app.mainWindow != nil {
			app.mainWindow.Hide()
		}
	})
}

func (app *AppGUI) showMainWindow() {
	ui.QueueMain(func() {
		if app.showAppMenu != nil {
			setVisibleMenu(false, app.showAppMenu)
		}
		if app.mainWindow != nil {
			app.mainWindow.Show()
		}
		if app.selectPortWindow != nil {
			app.selectPortWindow.Hide()
		}
	})
}

func (app *AppGUI) updateConfigPage(fanConfig *FanConfig, fanPage *FanPage) {
	fanPage.FanTypeA.SetSelected(app.fanTypeToIndex(fanConfig.FanTypeA))
	fanPage.FanTypeB.SetSelected(app.fanTypeToIndex(fanConfig.FanTypeB))
	fanPage.Control.SetSelected(app.controlToIndex(fanConfig.SensorControlling))
	fanPage.Power.SetValue(app.powerToInt(fanConfig.MinimumPower))
	fanPage.MinTemp.SetValue(app.tempToInt(fanConfig.MinimumTemperature))
	fanPage.MaxTemp.SetValue(app.tempToInt(fanConfig.MaximumTemperature))
	fanPage.AllowStop.SetChecked(fanConfig.AllowStopped)
}

func (app *AppGUI) getFanConfig(fanPage *FanPage) FanConfig {
	return FanConfig{
		MinimumPower:       int8(fanPage.Power.Value()),
		SensorControlling:  int8(fanPage.Control.Selected()),
		MinimumTemperature: int16(fanPage.MinTemp.Value()),
		MaximumTemperature: int16(fanPage.MaxTemp.Value()),
		AllowStopped:       fanPage.AllowStop.Checked(),
		FanTypeA:           int8(fanPage.FanTypeA.Selected()),
		FanTypeB:           int8(fanPage.FanTypeB.Selected()),
	}
}

func (app *AppGUI) getConfig() *Config {
	config := &Config{
		SensorTypes: SensorTypes{
			SensorTypeA: int8(app.sensorPage.SensorA.Selected()),
			SensorTypeB: int8(app.sensorPage.SensorB.Selected()),
			SensorTypeC: int8(app.sensorPage.SensorC.Selected()),
			SensorTypeD: int8(app.sensorPage.SensorD.Selected()),
		},
		Fan1Config: app.getFanConfig(&app.fan1Page),
		Fan2Config: app.getFanConfig(&app.fan2Page),
		Fan3Config: app.getFanConfig(&app.fan3Page),
		Fan4Config: app.getFanConfig(&app.fan4Page),
	}
	return config
}

func (app *AppGUI) tempToPerc(temp int8) int {
	if int(temp) > app.appConfig.MaxTemp {
		return 100
	}
	return int(float64(temp) * 100.0 / float64(app.appConfig.MaxTemp))
}

func (app *AppGUI) rpmToPerc(rpm int16) int {
	if int(rpm) > app.appConfig.MaxRPM {
		return 100
	}
	return int(float64(rpm) * 100.0 / float64(app.appConfig.MaxRPM))
}

func (app *AppGUI) sensorTypeToIndex(sensorType int8) int {
	if sensorType >= 0 && sensorType <= 2 {
		return int(sensorType)
	}
	return 0
}

func (app *AppGUI) fanTypeToIndex(fanType int8) int {
	if fanType >= 0 && fanType <= 5 {
		return int(fanType)
	}
	return 0
}

func (app *AppGUI) controlToIndex(control int8) int {
	if control >= 0 && control <= 7 {
		return int(control)
	}
	return 0
}

func (app *AppGUI) powerToInt(power int8) int {
	if power >= 0 && power <= 100 {
		return int(power)
	} else if power < 0 {
		return 0
	} else {
		return 100
	}
}

func (app *AppGUI) tempToInt(temp int16) int {
	if temp >= 0 && temp < 150 {
		return int(temp)
	} else if temp < 0 {
		return 0
	} else {
		return 150
	}
}
