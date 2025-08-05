//go:build windows

package main

import (
	"fmt"

	"github.com/alexbrainman/printer"
)

// WindowsPrinterManager Windows 打印机管理器实现
type WindowsPrinterManager struct {
	printers []PrinterInfo
}

// GetPrinters Windows 获取所有打印机
func (w *WindowsPrinterManager) GetPrinters() ([]PrinterInfo, error) {
	// 获取所有打印机名称
	printerNames, err := printer.ReadNames()
	if err != nil {
		return nil, fmt.Errorf("failed to get printer list: %v", err)
	}

	// 获取默认打印机
	defaultPrinter := printer.Default()

	var printers []PrinterInfo
	for _, name := range printerNames {
		printerInfo := PrinterInfo{
			Name:        name,
			Description: name,
			IsDefault:   name == defaultPrinter,
			Status:      "Available",
		}

		// 尝试获取打印机详细信息
		if p, err := printer.Open(name); err == nil {
			// 获取驱动信息
			if driverInfo, err := p.DriverInfo(); err == nil {
				printerInfo.Description = fmt.Sprintf("%s - %s", name, driverInfo.Name)
			}
			
			// 检查打印机状态
			if jobs, err := p.Jobs(); err == nil {
				if len(jobs) > 0 {
					printerInfo.Status = fmt.Sprintf("%d job(s)", len(jobs))
				}
			}
			
			p.Close()
		}

		printers = append(printers, printerInfo)
	}

	w.printers = printers
	return printers, nil
}

// GetDefault Windows 获取默认打印机
func (w *WindowsPrinterManager) GetDefault() (string, error) {
	defaultPrinter := printer.Default()
	if defaultPrinter == "" {
		return "", fmt.Errorf("no default printer set")
	}
	return defaultPrinter, nil
}

// SetDefault Windows 设置默认打印机 
func (w *WindowsPrinterManager) SetDefault(name string) error {
	// alexbrainman/printer 库不直接支持设置默认打印机
	// 需要调用 Windows API 或使用系统命令
	return fmt.Errorf("set default printer functionality not implemented yet")
}

// Refresh Windows 刷新打印机列表
func (w *WindowsPrinterManager) Refresh() error {
	_, err := w.GetPrinters()
	return err
}