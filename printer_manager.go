package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// PrinterInfo 打印机信息结构
type PrinterInfo struct {
	Name        string
	Description string
	IsDefault   bool
	Status      string
}

// PrinterManager 打印机管理接口
type PrinterManager interface {
	GetPrinters() ([]PrinterInfo, error)
	GetDefault() (string, error)
	SetDefault(name string) error
	Refresh() error
}

// NewPrinterManager 根据操作系统创建对应的打印机管理器
func NewPrinterManager() PrinterManager {
	switch runtime.GOOS {
	case "windows":
		return &WindowsPrinterManager{}
	case "darwin":
		return &CUPSManager{}
	case "linux":
		return &CUPSManager{}
	default:
		return &CUPSManager{}
	}
}

// CUPSManager macOS/Linux CUPS 打印机管理器
type CUPSManager struct {
	printers []PrinterInfo
}

// GetPrinters 获取所有打印机
func (c *CUPSManager) GetPrinters() ([]PrinterInfo, error) {
	// 获取默认打印机
	defaultPrinter, _ := c.GetDefault()

	// 获取打印机列表
	cmd := exec.Command("lpstat", "-p")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute lpstat -p: %v", err)
	}

	var printers []PrinterInfo
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// 解析 lpstat -p 输出格式
		// 英文格式: "printer PrinterName is idle..."
		// 中文格式: "打印机PrinterName闲置..."
		var printerName string
		
		if strings.HasPrefix(line, "printer ") {
			// 英文格式
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				printerName = parts[1]
			}
		} else if strings.HasPrefix(line, "打印机") {
			// 中文格式: "打印机PDFwriter闲置，启用时间始于..."
			// 提取打印机名称
			text := strings.TrimPrefix(line, "打印机")
			// 找到第一个非字母数字字符的位置
			for i, r := range text {
				if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
					printerName = text[:i]
					break
				}
			}
			if printerName == "" && len(text) > 0 {
				// 如果没有找到分隔符，使用整个字符串
				printerName = text
			}
		}
		
		if printerName != "" {
			printer := PrinterInfo{
				Name:        printerName,
				Description: printerName,
				IsDefault:   printerName == defaultPrinter,
				Status:      "Available",
			}
			printers = append(printers, printer)
		}
	}

	c.printers = printers
	return printers, nil
}

// GetDefault 获取默认打印机
func (c *CUPSManager) GetDefault() (string, error) {
	cmd := exec.Command("lpstat", "-d")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute lpstat -d: %v", err)
	}

	line := strings.TrimSpace(string(output))
	if line == "" || strings.Contains(line, "no system default destination") {
		return "", fmt.Errorf("no default printer set")
	}

	// 解析输出格式
	// 英文格式: "system default destination: PrinterName"
	// 中文格式: "系统默认目的位置：PrinterName"
	
	if strings.Contains(line, ":") {
		// 英文格式或中文格式都用冒号分隔
		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1]), nil
		}
	} else if strings.Contains(line, "：") {
		// 中文全角冒号
		parts := strings.Split(line, "：")
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1]), nil
		}
	}

	return "", fmt.Errorf("failed to parse default printer: %s", line)
}

// SetDefault 设置默认打印机
func (c *CUPSManager) SetDefault(name string) error {
	cmd := exec.Command("lpoptions", "-d", name)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to set default printer: %v", err)
	}
	return nil
}

// Refresh 刷新打印机列表
func (c *CUPSManager) Refresh() error {
	_, err := c.GetPrinters()
	return err
}

// WindowsPrinterManager Windows 打印机管理器
type WindowsPrinterManager struct {
	printers []PrinterInfo
}

// GetPrinters Windows 获取所有打印机
func (w *WindowsPrinterManager) GetPrinters() ([]PrinterInfo, error) {
	// 这里需要使用 alexbrainman/printer 库
	// 由于当前在 macOS 环境，暂时返回空实现
	return []PrinterInfo{}, fmt.Errorf("Windows printer manager requires Windows OS")
}

// GetDefault Windows 获取默认打印机
func (w *WindowsPrinterManager) GetDefault() (string, error) {
	return "", fmt.Errorf("Windows printer manager requires Windows OS")
}

// SetDefault Windows 设置默认打印机
func (w *WindowsPrinterManager) SetDefault(name string) error {
	return fmt.Errorf("Windows printer manager requires Windows OS")
}

// Refresh Windows 刷新打印机列表
func (w *WindowsPrinterManager) Refresh() error {
	return fmt.Errorf("Windows printer manager requires Windows OS")
}