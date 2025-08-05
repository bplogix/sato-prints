#!/usr/bin/env python3
"""
简单的打印测试示例
"""

import requests

def send_simple_print_request():
    """发送一个简单的打印请求"""
    
    # 打印机 URL（请根据你的实际情况修改IP地址）
    printer_url = "http://192.168.0.100:8082/ipp/print"
    
    # 构建简单的 IPP Print-Job 请求
    # IPP 头部: 版本(2) + 操作码(2) + 请求ID(4) = 8 字节
    ipp_header = b'\x01\x01'  # IPP version 1.1
    ipp_header += b'\x00\x02'  # Print-Job 操作
    ipp_header += b'\x00\x00\x00\x01'  # Request ID = 1
    
    # 操作属性组
    ipp_data = b'\x01'  # operation-attributes-tag
    
    # attributes-charset
    ipp_data += b'\x47\x00\x12attributes-charset\x00\x05utf-8'
    
    # attributes-natural-language
    ipp_data += b'\x48\x00\x1battributes-natural-language\x00\x05en-us'
    
    # printer-uri
    printer_uri = printer_url.encode('utf-8')
    ipp_data += b'\x45\x00\x0bprinter-uri'
    ipp_data += len(printer_uri).to_bytes(2, 'big')
    ipp_data += printer_uri
    
    # requesting-user-name
    ipp_data += b'\x42\x00\x14requesting-user-name\x00\x06python'
    
    # job-name
    ipp_data += b'\x42\x00\x08job-name\x00\x0bSimple Test'
    
    # document-format
    ipp_data += b'\x49\x00\x0fdocument-format\x00\ntext/plain'
    
    # 结束属性标记
    ipp_data += b'\x03'  # end-of-attributes-tag
    
    # 文档内容
    document_content = """Hello from Python!

这是一个简单的打印测试。
This is a simple print test.

时间: 2025-08-06
来源: Python IPP Client
""".encode('utf-8')
    
    # 完整的请求数据
    request_data = ipp_header + ipp_data + document_content
    
    try:
        print("🚀 发送打印请求...")
        response = requests.post(
            printer_url,
            data=request_data,
            headers={
                'Content-Type': 'application/ipp',
                'User-Agent': 'Simple-Python-Client/1.0'
            },
            timeout=10
        )
        
        print(f"📊 HTTP 响应代码: {response.status_code}")
        
        if response.status_code == 200:
            # 解析 IPP 响应状态
            if len(response.content) >= 4:
                status_code = int.from_bytes(response.content[2:4], 'big')
                if status_code == 0x0000:
                    print("✅ 打印任务提交成功！")
                else:
                    print(f"⚠️  服务器状态码: 0x{status_code:04x}")
            else:
                print("✅ 请求已发送（响应格式未知）")
        else:
            print(f"❌ HTTP 错误: {response.status_code}")
            print(f"响应内容: {response.text}")
            
    except requests.exceptions.ConnectionError:
        print("❌ 连接失败！请检查：")
        print("  1. AirPrint 服务是否正在运行")
        print("  2. IP 地址是否正确")
        print("  3. 端口 8082 是否可访问")
    except Exception as e:
        print(f"❌ 错误: {e}")

if __name__ == "__main__":
    print("📄 简单 IPP 打印测试")
    print("=" * 30)
    send_simple_print_request()