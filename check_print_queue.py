#!/usr/bin/env python3
"""
检查打印队列状态
"""

import subprocess
import sys

def check_print_queue():
    """检查系统打印队列"""
    print("📋 检查打印队列状态...")
    
    try:
        # macOS 使用 lpq 检查打印队列
        result = subprocess.run(['lpq'], capture_output=True, text=True)
        
        if result.returncode == 0:
            output = result.stdout.strip()
            if output:
                print("📊 打印队列状态:")
                print(output)
            else:
                print("✅ 打印队列为空")
        else:
            print(f"❌ 检查打印队列失败: {result.stderr}")
            
    except FileNotFoundError:
        print("❌ 找不到 lpq 命令")
    except Exception as e:
        print(f"❌ 检查打印队列时发生错误: {e}")

def check_printer_status():
    """检查打印机状态"""
    print("\n🖨️  检查打印机状态...")
    
    try:
        # 检查打印机状态
        result = subprocess.run(['lpstat', '-p'], capture_output=True, text=True)
        
        if result.returncode == 0:
            print("📊 打印机状态:")
            print(result.stdout)
        else:
            print(f"❌ 检查打印机状态失败: {result.stderr}")
            
    except Exception as e:
        print(f"❌ 检查打印机状态时发生错误: {e}")

def check_recent_jobs():
    """检查最近的打印任务"""
    print("\n📄 检查最近的打印任务...")
    
    try:
        # 检查最近的打印任务
        result = subprocess.run(['lpstat', '-o'], capture_output=True, text=True)
        
        if result.returncode == 0:
            output = result.stdout.strip()
            if output:
                print("📊 最近的打印任务:")
                print(output)
            else:
                print("✅ 没有等待中的打印任务")
        else:
            print(f"❌ 检查打印任务失败: {result.stderr}")
            
    except Exception as e:
        print(f"❌ 检查打印任务时发生错误: {e}")

def main():
    print("🔍 打印系统状态检查")
    print("=" * 40)
    
    check_printer_status()
    check_print_queue()
    check_recent_jobs()
    
    print("\n" + "=" * 40)
    print("💡 提示:")
    print("- 如果是 PDFwriter，文件会保存而不是打印")
    print("- 检查 ~/Desktop 或下载文件夹是否有新文件")

if __name__ == "__main__":
    main()