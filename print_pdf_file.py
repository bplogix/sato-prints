#!/usr/bin/env python3
"""
PDF 文件打印工具
直接读取并打印 PDF 文件到 AirPrint 服务
"""

import os
import sys
from pathlib import Path
from test_print import IPPClient

def print_pdf_file(pdf_path, printer_url="http://192.168.0.100:8082/ipp/print"):
    """
    打印 PDF 文件
    
    Args:
        pdf_path: PDF 文件路径
        printer_url: 打印机 URL
    """
    
    # 检查文件是否存在
    if not os.path.exists(pdf_path):
        print(f"❌ 文件不存在: {pdf_path}")
        return False
    
    # 检查文件是否为 PDF
    if not pdf_path.lower().endswith('.pdf'):
        print(f"⚠️  警告: 文件可能不是 PDF 格式: {pdf_path}")
    
    try:
        # 读取 PDF 文件
        print(f"📖 读取 PDF 文件: {pdf_path}")
        with open(pdf_path, 'rb') as f:
            pdf_data = f.read()
        
        file_size = len(pdf_data)
        print(f"📊 文件大小: {file_size:,} 字节 ({file_size/1024:.1f} KB)")
        
        # 创建 IPP 客户端
        client = IPPClient(printer_url)
        
        # 获取文件名（不含路径）
        file_name = Path(pdf_path).name
        
        print(f"🖨️  发送打印任务: {file_name}")
        
        # 发送打印任务
        result = client.print_job(
            document_data=pdf_data,
            document_name=file_name,
            document_format="application/pdf"
        )
        
        if result and result['success']:
            print("✅ PDF 文件打印任务提交成功！")
            print(f"📋 任务状态: {result['status_text']}")
            return True
        else:
            print("❌ PDF 文件打印任务提交失败")
            if result:
                print(f"📋 错误状态: {result['status_text']}")
            return False
            
    except FileNotFoundError:
        print(f"❌ 找不到文件: {pdf_path}")
        return False
    except PermissionError:
        print(f"❌ 没有权限读取文件: {pdf_path}")
        return False
    except Exception as e:
        print(f"❌ 读取文件时发生错误: {e}")
        return False

def print_multiple_pdfs(pdf_paths, printer_url="http://192.168.0.100:8082/ipp/print"):
    """
    批量打印多个 PDF 文件
    
    Args:
        pdf_paths: PDF 文件路径列表
        printer_url: 打印机 URL
    """
    
    print(f"📚 准备批量打印 {len(pdf_paths)} 个 PDF 文件")
    print("=" * 50)
    
    success_count = 0
    
    for i, pdf_path in enumerate(pdf_paths, 1):
        print(f"\n📄 [{i}/{len(pdf_paths)}] 处理: {pdf_path}")
        
        if print_pdf_file(pdf_path, printer_url):
            success_count += 1
            print(f"  ✅ 成功")
        else:
            print(f"  ❌ 失败")
    
    print("\n" + "=" * 50)
    print(f"📊 批量打印完成: {success_count}/{len(pdf_paths)} 成功")

def find_pdf_files(directory):
    """
    在指定目录中查找所有 PDF 文件
    
    Args:
        directory: 目录路径
        
    Returns:
        PDF 文件路径列表
    """
    
    pdf_files = []
    
    if not os.path.exists(directory):
        print(f"❌ 目录不存在: {directory}")
        return pdf_files
    
    print(f"🔍 搜索目录: {directory}")
    
    for root, dirs, files in os.walk(directory):
        for file in files:
            if file.lower().endswith('.pdf'):
                pdf_path = os.path.join(root, file)
                pdf_files.append(pdf_path)
    
    print(f"📋 找到 {len(pdf_files)} 个 PDF 文件")
    return pdf_files

def interactive_mode():
    """交互式模式"""
    
    print("🖨️  PDF 文件打印工具")
    print("=" * 40)
    
    # 设置打印机 URL
    default_url = "http://192.168.0.100:8082/ipp/print"
    printer_url = input(f"打印机 URL (回车使用默认: {default_url}): ").strip()
    if not printer_url:
        printer_url = default_url
    
    print(f"🖨️  使用打印机: {printer_url}")
    
    while True:
        print("\n请选择操作:")
        print("1. 打印单个 PDF 文件")
        print("2. 批量打印多个 PDF 文件")
        print("3. 打印目录中的所有 PDF 文件")
        print("4. 退出")
        
        choice = input("\n请输入选择 (1-4): ").strip()
        
        try:
            if choice == '1':
                pdf_path = input("请输入 PDF 文件路径: ").strip()
                if pdf_path:
                    print_pdf_file(pdf_path, printer_url)
                
            elif choice == '2':
                print("请输入多个 PDF 文件路径，每行一个，输入空行结束:")
                pdf_paths = []
                while True:
                    path = input().strip()
                    if not path:
                        break
                    pdf_paths.append(path)
                
                if pdf_paths:
                    print_multiple_pdfs(pdf_paths, printer_url)
                else:
                    print("❌ 没有输入任何文件路径")
                
            elif choice == '3':
                directory = input("请输入目录路径: ").strip()
                if directory:
                    pdf_files = find_pdf_files(directory)
                    if pdf_files:
                        print("\n找到的 PDF 文件:")
                        for i, pdf_path in enumerate(pdf_files, 1):
                            print(f"  {i}. {pdf_path}")
                        
                        confirm = input(f"\n确定要打印这 {len(pdf_files)} 个文件吗? (y/N): ").strip().lower()
                        if confirm in ['y', 'yes']:
                            print_multiple_pdfs(pdf_files, printer_url)
                        else:
                            print("❌ 取消操作")
                    else:
                        print("❌ 目录中没有找到 PDF 文件")
                
            elif choice == '4':
                print("👋 再见!")
                break
                
            else:
                print("❌ 无效选择，请输入 1-4")
                
        except KeyboardInterrupt:
            print("\n\n👋 用户中断，再见!")
            break
        except Exception as e:
            print(f"❌ 发生错误: {e}")

def main():
    """主函数"""
    
    # 检查命令行参数
    if len(sys.argv) > 1:
        # 命令行模式
        pdf_path = sys.argv[1]
        printer_url = sys.argv[2] if len(sys.argv) > 2 else "http://192.168.0.100:8082/ipp/print"
        
        print(f"📄 命令行模式 - 打印文件: {pdf_path}")
        success = print_pdf_file(pdf_path, printer_url)
        sys.exit(0 if success else 1)
    else:
        # 交互式模式
        interactive_mode()

if __name__ == "__main__":
    main()