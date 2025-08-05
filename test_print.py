#!/usr/bin/env python3
"""
Python IPP 打印客户端测试脚本
用于向我们的 AirPrint 服务发送打印请求
"""

import requests
import struct
import io

class IPPClient:
    def __init__(self, printer_url):
        """
        初始化 IPP 客户端
        
        Args:
            printer_url: 打印机的 IPP URL，例如 "http://192.168.0.100:8082/ipp/print"
        """
        self.printer_url = printer_url
        self.request_id = 1
    
    def _build_ipp_header(self, operation_id):
        """构建 IPP 请求头"""
        header = struct.pack('!HHI', 
                           0x0101,           # IPP 版本 1.1
                           operation_id,     # 操作码
                           self.request_id)  # 请求 ID
        self.request_id += 1
        return header
    
    def _add_attribute(self, buffer, tag, name, value):
        """添加 IPP 属性到缓冲区"""
        if isinstance(value, str):
            value_bytes = value.encode('utf-8')
        elif isinstance(value, int):
            value_bytes = struct.pack('!I', value)
        else:
            value_bytes = value
        
        name_bytes = name.encode('utf-8') if name else b''
        
        # 属性格式: tag(1) + name_length(2) + name + value_length(2) + value
        buffer.write(struct.pack('!BH', tag, len(name_bytes)))
        buffer.write(name_bytes)
        buffer.write(struct.pack('!H', len(value_bytes)))
        buffer.write(value_bytes)
    
    def get_printer_attributes(self):
        """获取打印机属性"""
        print("🔍 获取打印机属性...")
        
        # 构建 IPP 请求
        buffer = io.BytesIO()
        
        # IPP 头部
        header = self._build_ipp_header(0x000B)  # Get-Printer-Attributes
        buffer.write(header)
        
        # 操作属性组
        buffer.write(b'\x01')  # operation-attributes-tag
        
        # attributes-charset
        self._add_attribute(buffer, 0x47, "attributes-charset", "utf-8")
        
        # attributes-natural-language  
        self._add_attribute(buffer, 0x48, "attributes-natural-language", "en-us")
        
        # printer-uri
        self._add_attribute(buffer, 0x45, "printer-uri", self.printer_url)
        
        # 结束属性标记
        buffer.write(b'\x03')  # end-of-attributes-tag
        
        # 发送请求
        response = requests.post(
            self.printer_url,
            data=buffer.getvalue(),
            headers={
                'Content-Type': 'application/ipp',
                'User-Agent': 'Python-IPP-Client/1.0'
            }
        )
        
        if response.status_code == 200:
            print("✅ 获取打印机属性成功")
            return self._parse_ipp_response(response.content)
        else:
            print(f"❌ 获取打印机属性失败: {response.status_code}")
            return None
    
    def print_job(self, document_data, document_name="test-document", document_format="application/pdf"):
        """发送打印任务"""
        print(f"🖨️  发送打印任务: {document_name}")
        
        # 构建 IPP 请求
        buffer = io.BytesIO()
        
        # IPP 头部
        header = self._build_ipp_header(0x0002)  # Print-Job
        buffer.write(header)
        
        # 操作属性组
        buffer.write(b'\x01')  # operation-attributes-tag
        
        # attributes-charset
        self._add_attribute(buffer, 0x47, "attributes-charset", "utf-8")
        
        # attributes-natural-language
        self._add_attribute(buffer, 0x48, "attributes-natural-language", "en-us")
        
        # printer-uri
        self._add_attribute(buffer, 0x45, "printer-uri", self.printer_url)
        
        # requesting-user-name
        self._add_attribute(buffer, 0x42, "requesting-user-name", "python-client")
        
        # job-name
        self._add_attribute(buffer, 0x42, "job-name", document_name)
        
        # document-format
        self._add_attribute(buffer, 0x49, "document-format", document_format)
        
        # 任务属性组
        buffer.write(b'\x02')  # job-attributes-tag
        
        # copies
        self._add_attribute(buffer, 0x21, "copies", 1)
        
        # 结束属性标记
        buffer.write(b'\x03')  # end-of-attributes-tag
        
        # 添加文档数据
        if isinstance(document_data, str):
            buffer.write(document_data.encode('utf-8'))
        else:
            buffer.write(document_data)
        
        # 发送请求
        response = requests.post(
            self.printer_url,
            data=buffer.getvalue(),
            headers={
                'Content-Type': 'application/ipp',
                'User-Agent': 'Python-IPP-Client/1.0'
            }
        )
        
        if response.status_code == 200:
            print("✅ 打印任务提交成功")
            return self._parse_ipp_response(response.content)
        else:
            print(f"❌ 打印任务提交失败: {response.status_code}")
            print(f"响应内容: {response.text}")
            return None
    
    def validate_job(self, document_format="application/pdf"):
        """验证打印任务"""
        print("🔍 验证打印任务...")
        
        # 构建 IPP 请求
        buffer = io.BytesIO()
        
        # IPP 头部  
        header = self._build_ipp_header(0x0004)  # Validate-Job
        buffer.write(header)
        
        # 操作属性组
        buffer.write(b'\x01')  # operation-attributes-tag
        
        # attributes-charset
        self._add_attribute(buffer, 0x47, "attributes-charset", "utf-8")
        
        # attributes-natural-language
        self._add_attribute(buffer, 0x48, "attributes-natural-language", "en-us")
        
        # printer-uri
        self._add_attribute(buffer, 0x45, "printer-uri", self.printer_url)
        
        # requesting-user-name
        self._add_attribute(buffer, 0x42, "requesting-user-name", "python-client")
        
        # document-format
        self._add_attribute(buffer, 0x49, "document-format", document_format)
        
        # 结束属性标记
        buffer.write(b'\x03')  # end-of-attributes-tag
        
        # 发送请求
        response = requests.post(
            self.printer_url,
            data=buffer.getvalue(),
            headers={
                'Content-Type': 'application/ipp',
                'User-Agent': 'Python-IPP-Client/1.0'
            }
        )
        
        if response.status_code == 200:
            print("✅ 任务验证成功")
            return self._parse_ipp_response(response.content)
        else:
            print(f"❌ 任务验证失败: {response.status_code}")
            return None
    
    def _parse_ipp_response(self, response_data):
        """解析 IPP 响应"""
        if len(response_data) < 8:
            return {"error": "响应数据太短"}
        
        # 解析 IPP 头部
        version = struct.unpack('!H', response_data[0:2])[0]
        status_code = struct.unpack('!H', response_data[2:4])[0]
        request_id = struct.unpack('!I', response_data[4:8])[0]
        
        status_text = {
            0x0000: "successful-ok",
            0x0001: "successful-ok-ignored-or-substituted-attributes",
            0x0002: "successful-ok-conflicting-attributes",
            0x0400: "client-error-bad-request",
            0x0401: "client-error-unauthorized",
            0x0403: "client-error-forbidden",
            0x0404: "client-error-not-found",
            0x0500: "server-error-internal-error",
            0x0501: "server-error-operation-not-supported"
        }.get(status_code, f"unknown-status-{status_code:04x}")
        
        result = {
            "version": f"{version >> 8}.{version & 0xFF}",
            "status_code": status_code,
            "status_text": status_text,
            "request_id": request_id,
            "success": status_code < 0x0300
        }
        
        print(f"📊 IPP 响应: {result['status_text']} (代码: 0x{status_code:04x})")
        return result

def main():
    """主函数 - 测试示例"""
    print("🚀 Python IPP 客户端测试")
    print("=" * 50)
    
    # 配置打印机 URL（请根据你的实际情况修改）
    printer_url = "http://192.168.0.100:8082/ipp/print"
    
    # 创建 IPP 客户端
    client = IPPClient(printer_url)
    
    try:
        # 1. 获取打印机属性
        client.get_printer_attributes()
        print()
        
        # 2. 验证打印任务
        client.validate_job()
        print()
        
        # 3. 发送文本打印任务
        text_content = """
Hello from Python IPP Client!

这是一个测试打印任务。
Test document created at: 2025-08-06

功能测试：
✅ IPP 协议通信
✅ Print-Job 操作
✅ 中文字符支持

祝你使用愉快！
        """
        
        result = client.print_job(
            document_data=text_content,
            document_name="Python-Test-Document",
            document_format="text/plain"
        )
        
        if result and result['success']:
            print("🎉 打印任务完成！")
        else:
            print("⚠️  打印任务可能遇到问题")
            
    except requests.exceptions.ConnectionError:
        print("❌ 无法连接到打印服务")
        print("请确保：")
        print("1. AirPrint 服务正在运行")
        print("2. URL 地址正确")
        print("3. 网络连接正常")
    except Exception as e:
        print(f"❌ 发生错误: {e}")

if __name__ == "__main__":
    main()