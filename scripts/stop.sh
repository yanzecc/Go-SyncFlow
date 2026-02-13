#!/bin/bash
# Go-SyncFlow 统一身份同步与管理平台 - 一键停止脚本
# 用法: ./stop.sh

SERVICE_NAME="go-syncflow"

echo "=========================================="
echo "    Go-SyncFlow - 停止服务"
echo "=========================================="

# 兼容旧服务名
for svc in "$SERVICE_NAME" "bi-dashboard"; do
    if systemctl is-active --quiet "$svc" 2>/dev/null; then
        echo "[*] 正在停止服务 $svc ..."
        systemctl stop "$svc"
        echo "[✓] 服务 $svc 已停止"
    fi
done

# 显示状态
systemctl status ${SERVICE_NAME} --no-pager 2>/dev/null || true

echo ""
echo "如需完全卸载服务，请运行:"
echo "  systemctl disable ${SERVICE_NAME}"
echo "  rm /etc/systemd/system/${SERVICE_NAME}.service"
echo "  systemctl daemon-reload"
