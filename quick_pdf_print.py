#!/usr/bin/env python3
"""
快速 PDF 打印 - 最简单的方式
"""

import sys
from test_print import IPPClient

def quick_print_pdf(pdf_file_path):
    """快速打印 PDF 文件"""
    
    # 打印机 URL（请根据你的实际情况修改）
    printer_url = "http://192.168.0.100:8082/ipp/print"
    
    try:
        # 读取 PDF 文件
        print(f"📖 读取 PDF 文件: {pdf_file_path}")
        with open(pdf_file_path, 'rb') as f:
            pdf_data = f.read()
        
        print(f"📊 文件大小: {len(pdf_data):,} 字节")
        
        # 创建 IPP 客户端并发送打印任务
        client = IPPClient(printer_url)
        
        # 获取文件名
        file_name = pdf_file_path.split('/')[-1]  # 简单获取文件名
        
        print(f"🖨️  发送打印任务...")
        result = client.print_job(
            document_data=pdf_data,
            document_name=file_name,
            document_format="application/pdf"
        )
        
        if result and result['success']:
            print("✅ PDF 打印任务提交成功！")
        else:
            print("❌ PDF 打印任务提交失败")
            
    except FileNotFoundError:
        print(f"❌ 找不到文件: {pdf_file_path}")
    except Exception as e:
        print(f"❌ 错误: {e}")

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("使用方法: python3 quick_pdf_print.py <PDF文件路径>")
        print("示例: python3 quick_pdf_print.py /path/to/document.pdf")
        sys.exit(1)
    
    pdf_path = sys.argv[1]
    quick_print_pdf(pdf_path)