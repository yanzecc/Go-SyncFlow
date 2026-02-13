#!/bin/bash
# Go-SyncFlow 统一身份同步与管理平台 - 一键启动脚本
# 用法: ./start.sh
# 如果包含预编译二进制和静态文件，无需任何外部依赖即可启动

set -e

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
SERVICE_NAME="go-syncflow"

echo "=========================================="
echo "  Go-SyncFlow 统一身份同步与管理平台"
echo "            一键启动 v3.0"
echo "=========================================="

# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then
    echo "[警告] 建议使用root用户运行以支持systemd服务"
fi

# 进入项目目录
cd "$PROJECT_DIR"

# ============================================================
# 判断是否已有预编译产物（二进制 + 静态文件都齐全则免编译）
# ============================================================
HAS_BINARY=false
BINARY_PATH=""
if [ -f "$PROJECT_DIR/backend/go-syncflow" ]; then
    HAS_BINARY=true
    BINARY_PATH="$PROJECT_DIR/backend/go-syncflow"
elif [ -f "$PROJECT_DIR/backend/server" ]; then
    HAS_BINARY=true
    BINARY_PATH="$PROJECT_DIR/backend/server"
fi

HAS_STATIC=false
if [ -d "$PROJECT_DIR/backend/static" ] && [ -f "$PROJECT_DIR/backend/static/index.html" ]; then
    HAS_STATIC=true
fi

NEED_BUILD=false
if [ "$HAS_BINARY" = false ] || [ "$HAS_STATIC" = false ]; then
    NEED_BUILD=true
fi

# ============================================================
# 仅在需要编译时才检查/安装编译依赖
# ============================================================
if [ "$NEED_BUILD" = true ]; then
    echo "[*] 未检测到预编译产物，需要编译..."
    echo ""

    # ---------- 检查 Go ----------
    if [ "$HAS_BINARY" = false ]; then
        if command -v go &> /dev/null; then
            echo "[OK] Go已安装: $(go version)"
        elif [ -f "$PROJECT_DIR/tooling/go1.22.6.linux-amd64.tar.gz" ]; then
            echo "[*] 正在从tooling安装Go环境..."
            tar -C /usr/local -xzf "$PROJECT_DIR/tooling/go1.22.6.linux-amd64.tar.gz"
            export PATH=$PATH:/usr/local/go/bin
            echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
            echo "[OK] Go安装完成: $(go version)"
        else
            echo "[X] 未找到Go环境且无离线安装包，请先安装Go 1.22+"
            exit 1
        fi

        echo "[*] 编译后端..."
        cd "$PROJECT_DIR/backend"
        go mod download 2>/dev/null || true
        CGO_ENABLED=1 go build -o go-syncflow .
        BINARY_PATH="$PROJECT_DIR/backend/go-syncflow"
        echo "[OK] 后端编译完成"
        cd "$PROJECT_DIR"
    fi

    # ---------- 检查 Node.js ----------
    if [ "$HAS_STATIC" = false ]; then
        if command -v node &> /dev/null; then
            echo "[OK] Node.js已安装: $(node --version)"
        elif [ -f "$PROJECT_DIR/tooling/node-v18.20.2-linux-x64.tar.xz" ]; then
            echo "[*] 正在从tooling安装Node.js..."
            tar -xJf "$PROJECT_DIR/tooling/node-v18.20.2-linux-x64.tar.xz" -C /usr/local --strip-components=1
            echo "[OK] Node.js安装完成: $(node --version)"
        else
            echo "[X] 未找到Node.js环境，请先安装Node.js 18+"
            exit 1
        fi

        echo "[*] 构建前端..."
        cd "$PROJECT_DIR/frontend"
        if [ ! -d "node_modules" ]; then
            echo "[*] 安装前端依赖..."
            npm install --prefer-offline
        fi
        npm run build
        rm -rf "$PROJECT_DIR/backend/static"
        cp -r dist "$PROJECT_DIR/backend/static"
        echo "[OK] 前端构建完成"
        cd "$PROJECT_DIR"
    fi
else
    echo ""
    echo "[OK] 检测到预编译二进制: $(basename $BINARY_PATH)"
    echo "[OK] 检测到前端静态文件: backend/static/"
    echo "[OK] 无需安装任何编译依赖，直接启动"
    echo ""
fi

# ============================================================
# 设置时区
# ============================================================
timedatectl set-timezone Asia/Shanghai 2>/dev/null || ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime 2>/dev/null || true
echo "[OK] 时区: $(date '+%Y-%m-%d %H:%M:%S %Z')"

# ============================================================
# 初始化运行时目录
# ============================================================
echo "[*] 初始化运行环境..."
mkdir -p "$PROJECT_DIR/backend/data"
mkdir -p "$PROJECT_DIR/backend/certs"
echo "[OK] 数据目录就绪"

# ============================================================
# 配置 systemd 服务
# ============================================================
echo "[*] 配置系统服务..."

# 确定最终二进制路径
if [ -f "$PROJECT_DIR/backend/go-syncflow" ]; then
    EXEC_PATH="$PROJECT_DIR/backend/go-syncflow"
else
    EXEC_PATH="$PROJECT_DIR/backend/server"
fi

# 兼容：如果旧服务存在则停止并移除
if systemctl is-active --quiet bi-dashboard 2>/dev/null; then
    systemctl stop bi-dashboard 2>/dev/null || true
    systemctl disable bi-dashboard 2>/dev/null || true
    rm -f /etc/systemd/system/bi-dashboard.service
fi

cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=Go-SyncFlow Unified Identity Sync Platform
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$PROJECT_DIR/backend
ExecStart=$EXEC_PATH
Restart=always
RestartSec=5
Environment=GIN_MODE=release
StandardOutput=journal
StandardError=journal
SyslogIdentifier=go-syncflow
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable ${SERVICE_NAME} 2>/dev/null
echo "[OK] 系统服务配置完成"

# ============================================================
# 启动服务
# ============================================================
echo "[*] 启动服务..."

if systemctl is-active --quiet ${SERVICE_NAME}; then
    echo "[*] 服务已在运行，重启中..."
    systemctl restart ${SERVICE_NAME}
else
    systemctl start ${SERVICE_NAME}
fi

sleep 3

if systemctl is-active --quiet ${SERVICE_NAME}; then
    echo "[OK] 服务启动成功"
else
    echo "[X] 服务启动失败"
    echo "    查看日志: journalctl -u ${SERVICE_NAME} --no-pager -n 20"
    journalctl -u ${SERVICE_NAME} --no-pager -n 10
    exit 1
fi

# ============================================================
# 显示访问信息
# ============================================================
LOCAL_IP=$(hostname -I 2>/dev/null | awk '{print $1}')
[ -z "$LOCAL_IP" ] && LOCAL_IP="127.0.0.1"

echo ""
echo "=========================================="
echo "  部署完成！"
echo "=========================================="
echo ""
echo "  访问地址:"
echo "    HTTP:  http://${LOCAL_IP}:8080"
echo "    HTTPS: https://${LOCAL_IP}:8443"
echo ""
echo "  默认管理员账号:"
echo "    用户名: admin"
echo "    密码:   Admin@2024"
echo ""
echo "  常用命令:"
echo "    查看状态: systemctl status ${SERVICE_NAME}"
echo "    查看日志: journalctl -u ${SERVICE_NAME} -f"
echo "    停止服务: $SCRIPT_DIR/stop.sh"
echo "    重启服务: $SCRIPT_DIR/restart.sh"
echo "    重置密码: $SCRIPT_DIR/reset-admin.sh"
echo ""
echo "  LDAP 服务: ldap://${LOCAL_IP}:389"
echo "  LDAPS:     ldaps://${LOCAL_IP}:636"
echo "  (Samba 属性默认启用，可直接对接群晖NAS)"
echo ""
