package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2"
)

type App struct {
	printerManager PrinterManager
	airprintServer *AirPrintServer
	window         fyne.Window
	printerList    *widget.List
	refreshBtn     *widget.Button
	setDefaultBtn  *widget.Button
	serviceBtn     *widget.Button
	statusLabel    *widget.Label
	serviceLabel   *widget.Label
	
	printers        []PrinterInfo
	selectedPrinter string
	serviceRunning  bool
}

func main() {
	// 创建应用
	myApp := app.New()
	myApp.SetIcon(nil)
	
	// 创建主窗口
	myWindow := myApp.NewWindow("AirPrint Service Manager")
	myWindow.Resize(fyne.NewSize(600, 400))
	
	// 创建应用实例
	printerManager := NewPrinterManager()
	appInstance := &App{
		printerManager: printerManager,
		airprintServer: NewAirPrintServer(printerManager),
	}
	
	// 初始化 UI
	appInstance.window = myWindow
	appInstance.setupUI(myWindow)
	
	// 初始加载打印机列表
	appInstance.refreshPrinters()
	
	// 显示窗口并运行
	myWindow.ShowAndRun()
}

func (a *App) setupUI(myWindow fyne.Window) {
	// 状态标签
	a.statusLabel = widget.NewLabel("Loading printers...")
	
	// AirPrint 服务状态标签
	a.serviceLabel = widget.NewLabel("AirPrint Service: Stopped")
	
	// 刷新按钮
	a.refreshBtn = widget.NewButton("Refresh Printers", func() {
		a.refreshPrinters()
	})
	
	// 设置默认打印机按钮
	a.setDefaultBtn = widget.NewButton("Set as Default", func() {
		a.setAsDefault()
	})
	a.setDefaultBtn.Disable() // 初始禁用
	
	// AirPrint 服务控制按钮
	a.serviceBtn = widget.NewButton("Start AirPrint Service", func() {
		a.toggleAirPrintService()
	})
	
	// 打印机列表
	a.printerList = widget.NewList(
		func() int {
			return len(a.printers)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.DocumentIcon()),
				widget.NewLabel("Printer Name"),
				widget.NewLabel("Status"),
			)
		},
		func(i widget.ListItemID, obj fyne.CanvasObject) {
			if i >= len(a.printers) {
				return
			}
			
			printer := a.printers[i]
			hbox := obj.(*fyne.Container)
			
			// 图标
			icon := hbox.Objects[0].(*widget.Icon)
			if printer.IsDefault {
				icon.SetResource(theme.ConfirmIcon())
			} else {
				icon.SetResource(theme.DocumentIcon())
			}
			
			// 打印机名称
			nameLabel := hbox.Objects[1].(*widget.Label)
			displayName := printer.Name
			if printer.IsDefault {
				displayName += " (Default)"
			}
			nameLabel.SetText(displayName)
			
			// 状态
			statusLabel := hbox.Objects[2].(*widget.Label)
			statusLabel.SetText(printer.Status)
		},
	)
	
	// 列表选择事件
	a.printerList.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(a.printers) {
			a.selectedPrinter = a.printers[id].Name
			a.setDefaultBtn.Enable()
			a.statusLabel.SetText(fmt.Sprintf("Selected: %s", a.selectedPrinter))
		}
	}
	
	// 按钮容器
	buttonContainer := container.NewHBox(
		a.refreshBtn,
		a.setDefaultBtn,
		a.serviceBtn,
	)
	
	// 顶部状态容器
	topContainer := container.NewVBox(
		a.statusLabel,
		a.serviceLabel,
	)
	
	// 主内容容器
	content := container.NewBorder(
		topContainer,           // top
		buttonContainer,        // bottom
		nil,                   // left
		nil,                   // right
		a.printerList,         // center
	)
	
	myWindow.SetContent(content)
}

func (a *App) refreshPrinters() {
	a.statusLabel.SetText("Refreshing printer list...")
	a.setDefaultBtn.Disable()
	
	// 异步刷新
	go func() {
		printers, err := a.printerManager.GetPrinters()
		if err != nil {
			a.statusLabel.SetText(fmt.Sprintf("Refresh failed: %v", err))
			log.Printf("Failed to refresh printer list: %v", err)
			return
		}
		
		// 更新UI需要在主线程
		a.printers = printers
		a.printerList.Refresh()
		
		defaultPrinter, _ := a.printerManager.GetDefault()
		statusText := fmt.Sprintf("Found %d printer(s)", len(printers))
		if defaultPrinter != "" {
			statusText += fmt.Sprintf(", Default: %s", defaultPrinter)
		}
		a.statusLabel.SetText(statusText)
	}()
}

func (a *App) setAsDefault() {
	if a.selectedPrinter == "" {
		return
	}
	
	// 检查是否已经是默认打印机
	currentDefault, _ := a.printerManager.GetDefault()
	if currentDefault == a.selectedPrinter {
		dialog.ShowInformation("Information", 
			fmt.Sprintf("%s is already the default printer", a.selectedPrinter), 
			a.window)
		return
	}
	
	// 确认对话框
	dialog.ShowConfirm("Confirm", 
		fmt.Sprintf("Set %s as default printer?", a.selectedPrinter),
		func(confirmed bool) {
			if confirmed {
				err := a.printerManager.SetDefault(a.selectedPrinter)
				if err != nil {
					dialog.ShowError(fmt.Errorf("Failed to set default printer: %v", err), a.window)
				} else {
					dialog.ShowInformation("Success", 
						fmt.Sprintf("%s has been set as default printer", a.selectedPrinter), 
						a.window)
					a.refreshPrinters() // 刷新列表
				}
			}
		}, a.window)
}

// toggleAirPrintService 切换 AirPrint 服务状态
func (a *App) toggleAirPrintService() {
	if a.serviceRunning {
		// 停止服务
		err := a.airprintServer.Stop()
		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to stop AirPrint service: %v", err), a.window)
			return
		}
		
		a.serviceRunning = false
		a.serviceBtn.SetText("Start AirPrint Service")
		a.serviceLabel.SetText("AirPrint Service: Stopped")
		
		dialog.ShowInformation("Service Stopped", "AirPrint service has been stopped", a.window)
	} else {
		// 启动服务
		err := a.airprintServer.Start()
		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to start AirPrint service: %v", err), a.window)
			return
		}
		
		a.serviceRunning = true
		a.serviceBtn.SetText("Stop AirPrint Service")
		a.serviceLabel.SetText("AirPrint Service: Running (Port: 8082)")
		
		// 获取本机 IP
		if localIP, err := getLocalIP(); err == nil {
			dialog.ShowInformation("Service Started", 
				fmt.Sprintf("AirPrint service started\nAccess: http://%s:8082", localIP.String()), 
				a.window)
		} else {
			dialog.ShowInformation("Service Started", "AirPrint service has been started", a.window)
		}
	}
}