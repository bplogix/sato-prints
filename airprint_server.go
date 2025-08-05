package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
)

// PrintJob 打印任务
type PrintJob struct {
	ID           int
	Name         string
	Format       string
	Data         []byte
	Status       string
	CreatedAt    time.Time
	PrinterName  string
}

// AirPrintServer AirPrint 服务器
type AirPrintServer struct {
	printerManager PrinterManager
	server         *zeroconf.Server
	universalServer *zeroconf.Server
	httpServer     *http.Server
	port           int
	jobCounter     int
	jobs           map[int]*PrintJob
}

// NewAirPrintServer 创建新的 AirPrint 服务器
func NewAirPrintServer(printerManager PrinterManager) *AirPrintServer {
	return &AirPrintServer{
		printerManager: printerManager,
		port:           8082, // 默认端口
		jobCounter:     0,
		jobs:           make(map[int]*PrintJob),
	}
}

// Start 启动 AirPrint 服务
func (a *AirPrintServer) Start() error {
	// 启动 HTTP 服务器
	if err := a.startHTTPServer(); err != nil {
		return fmt.Errorf("启动 HTTP 服务器失败: %v", err)
	}

	// 注册 Bonjour/mDNS 服务
	if err := a.registerMDNSService(); err != nil {
		return fmt.Errorf("注册 mDNS 服务失败: %v", err)
	}

	log.Printf("AirPrint 服务已启动，端口: %d", a.port)
	return nil
}

// Stop 停止 AirPrint 服务
func (a *AirPrintServer) Stop() error {
	// 停止 mDNS 服务
	if a.server != nil {
		a.server.Shutdown()
		a.server = nil
	}
	
	// 停止 universal subtype 服务
	if a.universalServer != nil {
		a.universalServer.Shutdown()
		a.universalServer = nil
	}

	// 停止 HTTP 服务器
	if a.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return a.httpServer.Shutdown(ctx)
	}

	return nil
}

// startHTTPServer 启动 HTTP 服务器处理 IPP 请求
func (a *AirPrintServer) startHTTPServer() error {
	mux := http.NewServeMux()
	
	// IPP 端点
	mux.HandleFunc("/ipp/print", a.handleIPPRequest)
	mux.HandleFunc("/ipp/", a.handleIPPRequest)
	
	// 根目录处理
	mux.HandleFunc("/", a.handleRootRequest)

	a.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.port),
		Handler: mux,
	}

	// 在单独的 goroutine 中启动服务器
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP 服务器错误: %v", err)
		}
	}()

	return nil
}

// registerMDNSService 注册 mDNS 服务发现
func (a *AirPrintServer) registerMDNSService() error {
	// 获取本机 IP 地址
	localIP, err := getLocalIP()
	if err != nil {
		return fmt.Errorf("获取本机 IP 失败: %v", err)
	}

	// 获取默认打印机信息
	defaultPrinter, err := a.printerManager.GetDefault()
	if err != nil {
		defaultPrinter = "AirPrint Service"
	}

	// 服务名称 - iOS 需要特定格式
	hostname, _ := os.Hostname()
	serviceName := fmt.Sprintf("%s @ %s", defaultPrinter, hostname)

	// mDNS 服务属性 - iOS AirPrint 兼容格式
	txtRecords := []string{
		"txtvers=1",
		"qtotal=1",
		"rp=ipp/print",
		"ty=" + defaultPrinter,
		"adminurl=http://" + localIP.String() + ":" + fmt.Sprintf("%d", a.port) + "/",
		"note=",
		"priority=0",
		"product=(" + defaultPrinter + ")",
		"printer-state=3",
		"printer-type=0x809046",
		// 关键：iOS 设备需要这些特定的格式支持
		"pdl=application/octet-stream,application/pdf,application/postscript,image/urf,image/jpeg,image/png",
		"URF=W8,SRGB24,CP1,RS300-600,V1.4,DM1",
		"UUID=9c85edb1-1234-5678-9abc-123456789abc",
		"Color=T",
		"Duplex=F",
		"Staple=F",
		"Sort=T",
		"Collate=T",
		"Punch=F",
		"Copies=T",
		"Bind=F",
		"PaperMax=legal-A4",
		"Kind=document,photo",
		"PaperCustom=T",
		// iOS 特定属性
		"air=username,password",
		"mopria-certified=1.3",
		"printer-location=",
		"printer-make-and-model=" + defaultPrinter,
	}

	// 注册主要的 IPP 服务
	server, err := zeroconf.Register(serviceName, "_ipp._tcp", "local.", a.port, txtRecords, nil)
	if err != nil {
		return fmt.Errorf("注册 IPP 服务失败: %v", err)
	}
	a.server = server
	
	// 尝试注册 universal subtype（iOS AirPrint 发现需要）
	// 注意：某些 mDNS 库可能不支持 subtype，这是正常的
	universalServer, err := zeroconf.Register(serviceName, "_universal._sub._ipp._tcp", "local.", a.port, txtRecords, nil)
	if err != nil {
		log.Printf("注意：无法注册 universal subtype（这在某些系统上是正常的）: %v", err)
		// 尝试替代方案：添加额外的 TXT 记录
		log.Printf("使用备用方案进行 iOS 兼容性设置")
	} else {
		a.universalServer = universalServer
		log.Printf("已注册 universal subtype 服务")
	}

	log.Printf("mDNS 服务已注册: %s", serviceName)
	return nil
}

// handleRootRequest 处理根目录请求
func (a *AirPrintServer) handleRootRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	printers, err := a.printerManager.GetPrinters()
	if err != nil {
		http.Error(w, fmt.Sprintf("获取打印机列表失败: %v", err), http.StatusInternalServerError)
		return
	}

	html := `<!DOCTYPE html>
<html>
<head>
    <title>AirPrint 服务</title>
    <meta charset="utf-8">
</head>
<body>
    <h1>AirPrint 服务状态</h1>
    <h2>可用打印机列表:</h2>
    <ul>
`
	
	for _, printer := range printers {
		status := "可用"
		if printer.IsDefault {
			status += " (默认)"
		}
		html += fmt.Sprintf("        <li>%s - %s</li>\n", printer.Name, status)
	}
	
	html += `    </ul>
    <p>服务正在运行，可以通过 AirPrint 进行打印。</p>
</body>
</html>`
	
	w.Write([]byte(html))
}

// handleIPPRequest 处理 IPP 打印请求
func (a *AirPrintServer) handleIPPRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("收到 IPP 请求: %s %s", r.Method, r.URL.Path)
	log.Printf("Content-Type: %s", r.Header.Get("Content-Type"))
	log.Printf("Content-Length: %s", r.Header.Get("Content-Length"))
	log.Printf("User-Agent: %s", r.Header.Get("User-Agent"))

	// 设置 IPP 响应头
	w.Header().Set("Content-Type", "application/ipp")
	w.Header().Set("Server", "AirPrint/1.0")
	
	switch r.Method {
	case "POST":
		// 处理 IPP POST 请求（打印任务、获取属性等）
		a.handleIPPPost(w, r)
	case "GET":
		// 处理 IPP GET 请求
		a.handleIPPGet(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleIPPPost 处理 IPP POST 请求
func (a *AirPrintServer) handleIPPPost(w http.ResponseWriter, r *http.Request) {
	log.Printf("处理 IPP POST 请求")
	
	// 读取 IPP 请求体
	body := make([]byte, r.ContentLength)
	n, err := r.Body.Read(body)
	if err != nil && n == 0 {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	body = body[:n] // 调整到实际读取的长度
	
	log.Printf("接收到 %d 字节的 IPP 数据", n)
	
	if len(body) < 8 {
		log.Printf("IPP 请求太短: %d 字节", len(body))
		http.Error(w, "Invalid IPP request", http.StatusBadRequest)
		return
	}
	
	// 解析 IPP 请求头
	version := (uint16(body[0]) << 8) | uint16(body[1])
	operation := (uint16(body[2]) << 8) | uint16(body[3])
	requestID := (uint32(body[4]) << 24) | (uint32(body[5]) << 16) | (uint32(body[6]) << 8) | uint32(body[7])
	
	log.Printf("IPP 请求 - Version: 0x%04x, Operation: 0x%04x, RequestID: %d", version, operation, requestID)
	
	var response []byte
	
	switch operation {
	case 0x000B: // Get-Printer-Attributes
		response = a.buildGetPrinterAttributesResponse(requestID)
	case 0x0002: // Print-Job
		response = a.buildPrintJobResponse(requestID, body)
	case 0x0004: // Validate-Job
		response = a.buildValidateJobResponse(requestID)
	default:
		response = a.buildErrorResponse(requestID, 0x0500) // server-error-operation-not-supported
	}
	
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// handleIPPGet 处理 IPP GET 请求
func (a *AirPrintServer) handleIPPGet(w http.ResponseWriter, r *http.Request) {
	log.Printf("处理 IPP GET 请求: %s", r.URL.Path)
	
	// 对于 IPP 端点的 GET 请求，返回打印机信息
	if r.URL.Path == "/ipp/print" || r.URL.Path == "/ipp/" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("AirPrint Service Ready - Use POST for IPP operations"))
		return
	}
	
	// 其他 GET 请求
	http.Error(w, "Not Found", http.StatusNotFound)
}

// buildGetPrinterAttributesResponse 构建获取打印机属性响应
func (a *AirPrintServer) buildGetPrinterAttributesResponse(requestID uint32) []byte {
	// 获取本机 IP
	localIP, _ := getLocalIP()
	printerURI := fmt.Sprintf("ipp://%s:%d/ipp/print", localIP.String(), a.port)
	
	// 构建基本的 IPP 响应头
	response := []byte{
		0x01, 0x01, // IPP version 1.1
		0x00, 0x00, // successful-ok
		byte(requestID >> 24), byte(requestID >> 16), byte(requestID >> 8), byte(requestID), // request-id
		0x01, // operation-attributes-tag
	}
	
	// 添加 charset 属性
	response = append(response, 0x47, 0x00, 0x12) // textWithoutLanguage tag + name length
	response = append(response, []byte("attributes-charset")...)
	response = append(response, 0x00, 0x05) // value length
	response = append(response, []byte("utf-8")...)
	
	// 添加 natural-language 属性
	response = append(response, 0x48, 0x00, 0x1B) // naturalLanguage tag + name length
	response = append(response, []byte("attributes-natural-language")...)
	response = append(response, 0x00, 0x05) // value length
	response = append(response, []byte("en-us")...)
	
	// 打印机属性标签
	response = append(response, 0x02) // printer-attributes-tag
	
	// printer-uri-supported
	response = append(response, 0x45, 0x00, 0x13) // uri tag + name length
	response = append(response, []byte("printer-uri-supported")...)
	response = append(response, byte(len(printerURI)>>8), byte(len(printerURI))) // value length
	response = append(response, []byte(printerURI)...)
	
	// printer-name
	printerName := "AirPrint Service"
	response = append(response, 0x42, 0x00, 0x0C) // nameWithoutLanguage tag + name length
	response = append(response, []byte("printer-name")...)
	response = append(response, byte(len(printerName)>>8), byte(len(printerName))) // value length
	response = append(response, []byte(printerName)...)
	
	// printer-state (3 = idle)
	response = append(response, 0x21, 0x00, 0x0D) // enum tag + name length
	response = append(response, []byte("printer-state")...)
	response = append(response, 0x00, 0x04, 0x00, 0x00, 0x00, 0x03) // idle state
	
	// printer-state-reasons
	response = append(response, 0x44, 0x00, 0x15) // keyword tag + name length
	response = append(response, []byte("printer-state-reasons")...)
	response = append(response, 0x00, 0x04) // value length
	response = append(response, []byte("none")...)
	
	// operations-supported
	response = append(response, 0x21, 0x00, 0x13) // enum tag + name length
	response = append(response, []byte("operations-supported")...)
	response = append(response, 0x00, 0x04, 0x00, 0x00, 0x00, 0x02) // Print-Job
	response = append(response, 0x21, 0x00, 0x00) // enum tag + no name (additional value)
	response = append(response, 0x00, 0x04, 0x00, 0x00, 0x00, 0x04) // Validate-Job
	response = append(response, 0x21, 0x00, 0x00) // enum tag + no name (additional value)
	response = append(response, 0x00, 0x04, 0x00, 0x00, 0x00, 0x0B) // Get-Printer-Attributes
	
	// document-format-supported
	response = append(response, 0x49, 0x00, 0x17) // mimeMediaType tag + name length
	response = append(response, []byte("document-format-supported")...)
	response = append(response, 0x00, 0x0F) // value length
	response = append(response, []byte("application/pdf")...)
	response = append(response, 0x49, 0x00, 0x00) // mimeMediaType tag + no name (additional value)
	response = append(response, 0x00, 0x0A) // value length
	response = append(response, []byte("image/jpeg")...)
	response = append(response, 0x49, 0x00, 0x00) // mimeMediaType tag + no name (additional value)
	response = append(response, 0x00, 0x09) // value length
	response = append(response, []byte("image/png")...)
	
	// color-supported
	response = append(response, 0x22, 0x00, 0x0F) // boolean tag + name length
	response = append(response, []byte("color-supported")...)
	response = append(response, 0x00, 0x01, 0x01) // true
	
	// printer-resolution-supported
	response = append(response, 0x32, 0x00, 0x1A) // resolution tag + name length
	response = append(response, []byte("printer-resolution-supported")...)
	response = append(response, 0x00, 0x09) // value length
	response = append(response, 0x00, 0x00, 0x01, 0x2C, 0x00, 0x00, 0x01, 0x2C, 0x03) // 300x300 dpi
	
	// end-of-attributes-tag
	response = append(response, 0x03)
	
	return response
}

// buildPrintJobResponse 构建打印任务响应并实际执行打印
func (a *AirPrintServer) buildPrintJobResponse(requestID uint32, requestBody []byte) []byte {
	// 解析 IPP 请求以提取文档数据和属性
	jobName, documentFormat, documentData := a.parseIPPPrintRequest(requestBody)
	
	// 创建打印任务
	a.jobCounter++
	jobID := a.jobCounter
	
	job := &PrintJob{
		ID:          jobID,
		Name:        jobName,
		Format:      documentFormat,
		Data:        documentData,
		Status:      "pending",
		CreatedAt:   time.Now(),
		PrinterName: "",
	}
	
	// 获取默认打印机
	if defaultPrinter, err := a.printerManager.GetDefault(); err == nil {
		job.PrinterName = defaultPrinter
	}
	
	a.jobs[jobID] = job
	
	log.Printf("创建打印任务 - ID: %d, 名称: %s, 格式: %s, 大小: %d 字节", 
		jobID, jobName, documentFormat, len(documentData))
	
	// 异步执行实际打印
	go a.executePrintJob(job)
	
	// 构建 IPP 响应
	response := []byte{
		0x01, 0x01, // IPP version 1.1
		0x00, 0x00, // successful-ok
		byte(requestID >> 24), byte(requestID >> 16), byte(requestID >> 8), byte(requestID), // request-id
		0x02, // job-attributes-tag
		// job-id
		0x21, 0x00, 0x06, 'j', 'o', 'b', '-', 'i', 'd',
		0x00, 0x04, byte(jobID>>24), byte(jobID>>16), byte(jobID>>8), byte(jobID), // job-id
		// job-state
		0x21, 0x00, 0x09, 'j', 'o', 'b', '-', 's', 't', 'a', 't', 'e',
		0x00, 0x04, 0x00, 0x00, 0x00, 0x03, // pending state
		0x03, // end-of-attributes-tag
	}
	
	log.Printf("接受打印任务，Job ID: %d", jobID)
	return response
}

// buildValidateJobResponse 构建验证任务响应
func (a *AirPrintServer) buildValidateJobResponse(requestID uint32) []byte {
	response := []byte{
		0x01, 0x01, // IPP version 1.1
		0x00, 0x00, // successful-ok
		byte(requestID >> 24), byte(requestID >> 16), byte(requestID >> 8), byte(requestID), // request-id
		0x03, // end-of-attributes-tag
	}
	return response
}

// buildErrorResponse 构建错误响应
func (a *AirPrintServer) buildErrorResponse(requestID uint32, statusCode uint16) []byte {
	response := []byte{
		0x01, 0x01, // IPP version 1.1
		byte(statusCode >> 8), byte(statusCode), // status code
		byte(requestID >> 24), byte(requestID >> 16), byte(requestID >> 8), byte(requestID), // request-id
		0x03, // end-of-attributes-tag
	}
	return response
}

// parseIPPPrintRequest 解析 IPP 打印请求
func (a *AirPrintServer) parseIPPPrintRequest(body []byte) (jobName, documentFormat string, documentData []byte) {
	// 设置默认值
	jobName = "Untitled"
	documentFormat = "application/octet-stream"
	
	// 简单解析：找到文档数据的开始位置
	// 在实际的 IPP 实现中，这需要完整解析 IPP 属性
	dataStart := -1
	
	// 查找 end-of-attributes-tag (0x03)
	for i := 8; i < len(body)-1; i++ {
		if body[i] == 0x03 {
			dataStart = i + 1
			break
		}
	}
	
	if dataStart > 0 && dataStart < len(body) {
		documentData = body[dataStart:]
		log.Printf("提取文档数据: %d 字节", len(documentData))
	} else {
		log.Printf("未找到文档数据")
		documentData = []byte{}
	}
	
	// 简单解析一些常见属性（更完整的实现需要全面解析 IPP 属性）
	bodyStr := string(body)
	if strings.Contains(bodyStr, "application/pdf") {
		documentFormat = "application/pdf"
	} else if strings.Contains(bodyStr, "text/plain") {
		documentFormat = "text/plain"
	} else if strings.Contains(bodyStr, "image/jpeg") {
		documentFormat = "image/jpeg"
	} else if strings.Contains(bodyStr, "image/png") {
		documentFormat = "image/png"
	}
	
	return jobName, documentFormat, documentData
}

// executePrintJob 执行实际的打印任务
func (a *AirPrintServer) executePrintJob(job *PrintJob) {
	log.Printf("开始执行打印任务 ID: %d", job.ID)
	
	job.Status = "processing"
	
	// 创建临时文件来保存文档数据
	tempFile, err := a.createTempFile(job)
	if err != nil {
		log.Printf("创建临时文件失败: %v", err)
		job.Status = "aborted"
		return
	}
	defer os.Remove(tempFile) // 清理临时文件
	
	// 执行打印命令
	err = a.printToSystem(tempFile, job.PrinterName)
	if err != nil {
		log.Printf("打印失败: %v", err)
		job.Status = "aborted"
		return
	}
	
	job.Status = "completed"
	log.Printf("打印任务 ID: %d 完成", job.ID)
}

// createTempFile 创建临时文件
func (a *AirPrintServer) createTempFile(job *PrintJob) (string, error) {
	// 根据文档格式确定文件扩展名
	var ext string
	switch job.Format {
	case "application/pdf":
		ext = ".pdf"
	case "text/plain":
		ext = ".txt"
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	default:
		ext = ".dat"
	}
	
	// 创建临时文件
	tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("airprint_%d_%d%s", 
		job.ID, time.Now().Unix(), ext))
	
	err := ioutil.WriteFile(tempFile, job.Data, 0644)
	if err != nil {
		return "", fmt.Errorf("写入临时文件失败: %v", err)
	}
	
	log.Printf("创建临时文件: %s", tempFile)
	return tempFile, nil
}

// printToSystem 发送文件到系统打印机
func (a *AirPrintServer) printToSystem(filePath, printerName string) error {
	var cmd *exec.Cmd
	
	// 根据操作系统选择打印命令
	switch runtime.GOOS {
	case "darwin": // macOS
		if printerName != "" {
			// 使用指定打印机
			cmd = exec.Command("lpr", "-P", printerName, filePath)
		} else {
			// 使用默认打印机
			cmd = exec.Command("lpr", filePath)
		}
	case "linux":
		if printerName != "" {
			cmd = exec.Command("lp", "-d", printerName, filePath)
		} else {
			cmd = exec.Command("lp", filePath)
		}
	case "windows":
		// Windows 打印命令（需要进一步实现）
		return fmt.Errorf("Windows 打印支持正在开发中")
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
	
	log.Printf("执行打印命令: %s", cmd.String())
	
	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("打印命令执行失败: %v, 输出: %s", err, string(output))
	}
	
	log.Printf("打印命令执行成功，输出: %s", string(output))
	return nil
}

// getLocalIP 获取本机 IP 地址
func getLocalIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}