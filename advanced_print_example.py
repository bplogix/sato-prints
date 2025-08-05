#!/usr/bin/env python3
"""
高级打印示例 - 展示更多 IPP 功能
"""

from test_print import IPPClient
import json
import time

def print_pdf_file():
    """打印 PDF 文件示例"""
    printer_url = "http://192.168.0.100:8082/ipp/print"
    client = IPPClient(printer_url)
    
    # 模拟 PDF 数据（实际使用中应该读取真实的 PDF 文件）
    pdf_header = b'%PDF-1.4\n'
    pdf_content = b'''1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
2 0 obj
<<
/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj
3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
/Contents 4 0 R
>>
endobj
4 0 obj
<<
/Length 44
>>
stream
BT
/F1 12 Tf
100 700 Td
(Hello from Python PDF!) Tj
ET
endstream
endobj
xref
0 5
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
0000000206 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
299
%%EOF'''
    
    fake_pdf_data = pdf_header + pdf_content
    
    result = client.print_job(
        document_data=fake_pdf_data,
        document_name="Python-Generated-PDF",
        document_format="application/pdf"
    )
    
    return result

def print_image_data():
    """打印图像数据示例"""
    printer_url = "http://192.168.0.100:8082/ipp/print"
    client = IPPClient(printer_url)
    
    # 模拟 JPEG 头部（实际使用中应该读取真实的图像文件）
    fake_jpeg_data = b'\xff\xd8\xff\xe0\x00\x10JFIF\x00\x01\x01\x01\x00H\x00H\x00\x00\xff\xdb\x00C\x00'
    fake_jpeg_data += b'This is fake JPEG data for testing purposes' * 10
    fake_jpeg_data += b'\xff\xd9'  # JPEG 结束标记
    
    result = client.print_job(
        document_data=fake_jpeg_data,
        document_name="Python-Test-Image",
        document_format="image/jpeg"
    )
    
    return result

def batch_print_test():
    """批量打印测试"""
    printer_url = "http://192.168.0.100:8082/ipp/print"
    client = IPPClient(printer_url)
    
    documents = [
        {
            "name": "Document-1-Text",
            "content": "这是第一个文档\nThis is document number 1\n测试批量打印功能",
            "format": "text/plain"
        },
        {
            "name": "Document-2-HTML",
            "content": "<html><body><h1>HTML Document</h1><p>这是一个HTML文档测试</p></body></html>",
            "format": "text/html"
        },
        {
            "name": "Document-3-JSON",
            "content": json.dumps({
                "title": "测试文档",
                "type": "JSON数据",
                "content": ["项目1", "项目2", "项目3"],
                "timestamp": time.time()
            }, ensure_ascii=False, indent=2),
            "format": "application/json"
        }
    ]
    
    print(f"📚 开始批量打印 {len(documents)} 个文档...")
    
    for i, doc in enumerate(documents, 1):
        print(f"\n📄 打印文档 {i}/{len(documents)}: {doc['name']}")
        
        result = client.print_job(
            document_data=doc['content'],
            document_name=doc['name'],
            document_format=doc['format']
        )
        
        if result and result['success']:
            print(f"  ✅ {doc['name']} 提交成功")
        else:
            print(f"  ❌ {doc['name']} 提交失败")
        
        # 添加小延迟避免过快请求
        time.sleep(0.5)
    
    print("\n🎉 批量打印完成！")

def printer_status_monitor():
    """打印机状态监控示例"""
    printer_url = "http://192.168.0.100:8082/ipp/print"
    client = IPPClient(printer_url)
    
    print("🔍 获取打印机详细状态...")
    
    # 多次获取状态，模拟监控
    for i in range(3):
        print(f"\n--- 状态检查 {i+1} ---")
        result = client.get_printer_attributes()
        
        if result and result['success']:
            print("✅ 打印机状态正常")
            print(f"📊 响应时间: {time.time()}")
        else:
            print("❌ 无法获取打印机状态")
        
        if i < 2:  # 不在最后一次检查后等待
            time.sleep(2)

def main():
    """主函数 - 高级功能演示"""
    print("🚀 高级 IPP 打印功能演示")
    print("=" * 50)
    
    while True:
        print("\n请选择要测试的功能：")
        print("1. 打印 PDF 文档")
        print("2. 打印图像数据") 
        print("3. 批量打印测试")
        print("4. 打印机状态监控")
        print("5. 退出")
        
        choice = input("\n请输入选择 (1-5): ").strip()
        
        try:
            if choice == '1':
                print("\n" + "="*30)
                print_pdf_file()
            elif choice == '2':
                print("\n" + "="*30)
                print_image_data()
            elif choice == '3':
                print("\n" + "="*30)
                batch_print_test()
            elif choice == '4':
                print("\n" + "="*30)
                printer_status_monitor()
            elif choice == '5':
                print("\n👋 再见！")
                break
            else:
                print("❌ 无效选择，请输入 1-5")
                
        except KeyboardInterrupt:
            print("\n\n👋 用户中断，再见！")
            break
        except Exception as e:
            print(f"\n❌ 发生错误: {e}")

if __name__ == "__main__":
    main()