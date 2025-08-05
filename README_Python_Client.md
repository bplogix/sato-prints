# Python IPP 客户端使用指南

这里提供了三个 Python 脚本来与我们的 AirPrint 服务进行交互。

## 📋 文件说明

### 1. `simple_print_test.py` - 简单测试
最基础的打印请求示例，适合快速测试服务连接。

**使用方法：**
```bash
python3 simple_print_test.py
```

**功能：**
- 发送简单的文本打印请求
- 验证服务连接
- 基本的 IPP 协议演示

### 2. `test_print.py` - 完整 IPP 客户端
包含完整的 IPP 客户端类，支持多种操作。

**主要功能：**
- `get_printer_attributes()` - 获取打印机属性
- `print_job()` - 发送打印任务
- `validate_job()` - 验证打印任务
- 完整的 IPP 响应解析

**使用方法：**
```bash
python3 test_print.py
```

### 3. `advanced_print_example.py` - 高级功能演示
交互式的高级功能展示，包含多种打印场景。

**功能包括：**
- PDF 文档打印
- 图像数据打印
- 批量打印测试
- 打印机状态监控

**使用方法：**
```bash
python3 advanced_print_example.py
```

## 🔧 配置说明

在使用前，请确保修改脚本中的打印机 URL：

```python
printer_url = "http://192.168.0.100:8082/ipp/print"
#                   ^^^^^^^^^^^ 
#                   改为你的实际 IP 地址
```

## 📚 API 使用示例

### 基本使用
```python
from test_print import IPPClient

# 创建客户端
client = IPPClient("http://192.168.0.100:8082/ipp/print")

# 获取打印机信息
attrs = client.get_printer_attributes()

# 发送打印任务
result = client.print_job(
    document_data="Hello World!",
    document_name="test-doc",
    document_format="text/plain"
)
```

### 打印不同格式的文档

#### 文本文档
```python
client.print_job(
    document_data="纯文本内容\nSecond line",
    document_name="text-document",
    document_format="text/plain"
)
```

#### HTML 文档
```python
html_content = "<html><body><h1>标题</h1><p>内容</p></body></html>"
client.print_job(
    document_data=html_content,
    document_name="html-document", 
    document_format="text/html"
)
```

#### PDF 文件
```python
with open("document.pdf", "rb") as f:
    pdf_data = f.read()

client.print_job(
    document_data=pdf_data,
    document_name="pdf-document",
    document_format="application/pdf"
)
```

#### 图像文件
```python
with open("image.jpg", "rb") as f:
    image_data = f.read()

client.print_job(
    document_data=image_data,
    document_name="image-document",
    document_format="image/jpeg"
)
```

## 🔍 故障排除

### 连接失败
```
❌ 连接失败！请检查：
1. AirPrint 服务是否正在运行
2. IP 地址是否正确  
3. 端口 8082 是否可访问
```

**解决方法：**
1. 确保 Go 服务正在运行：`go run .`
2. 检查服务状态：`netstat -an | grep 8082`
3. 验证网络连接：`curl http://192.168.0.100:8082/ipp/print`

### IPP 协议错误
如果收到非 0x0000 状态码，表示 IPP 协议层面的错误：

- `0x0400` - 客户端请求错误
- `0x0401` - 未授权
- `0x0500` - 服务器内部错误
- `0x0501` - 操作不支持

### 中文字符问题
所有脚本都使用 UTF-8 编码，支持中文字符。如果遇到编码问题：

```python
# 确保文本使用正确编码
text = "中文测试".encode('utf-8')
```

## 📊 支持的文档格式

我们的 AirPrint 服务支持以下格式：

- `text/plain` - 纯文本
- `text/html` - HTML 文档
- `application/pdf` - PDF 文档
- `application/postscript` - PostScript
- `image/jpeg` - JPEG 图像
- `image/png` - PNG 图像
- `application/json` - JSON 数据

## 🚀 扩展开发

### 自定义 IPP 操作
基于 `IPPClient` 类，你可以轻松添加新的 IPP 操作：

```python
def custom_operation(self):
    buffer = io.BytesIO()
    header = self._build_ipp_header(0x0010)  # 自定义操作码
    buffer.write(header)
    
    # 添加属性...
    
    response = requests.post(self.printer_url, 
                           data=buffer.getvalue(),
                           headers={'Content-Type': 'application/ipp'})
    return self._parse_ipp_response(response.content)
```

### 异步打印
```python
import asyncio
import aiohttp

async def async_print_job(document_data):
    async with aiohttp.ClientSession() as session:
        async with session.post(
            printer_url,
            data=request_data,
            headers={'Content-Type': 'application/ipp'}
        ) as response:
            return await response.read()
```

## 📞 支持

如果遇到问题，请检查：
1. Go AirPrint 服务的日志输出
2. 网络连接状态
3. IPP 请求格式是否正确

祝使用愉快！🎉