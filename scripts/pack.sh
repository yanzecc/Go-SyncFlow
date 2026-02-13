#!/bin/bash
# Go-SyncFlow 统一身份同步与管理平台 - 打包脚本
# 用法: ./pack.sh [输出文件名]
# 打包内容：预编译二进制 + 前端静态文件 + 源码 + 文档 + 部署脚本
# 打包结果：解压后可一键启动，无需编译

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
OUTPUT_NAME="${1:-go-syncflow-v3.0-$(date +%Y%m%d)}"
OUTPUT_FILE="/tmp/${OUTPUT_NAME}.tar.gz"

echo "=========================================="
echo "    Go-SyncFlow 统一身份同步与管理平台"
echo "              打包工具 v3.0"
echo "=========================================="

cd "$PROJECT_DIR"

# ========== 步骤 1：编译最新版本 ==========
echo ""
echo "[1/7] 编译后端..."
cd "$PROJECT_DIR/backend"
CGO_ENABLED=1 go build -o go-syncflow . 2>&1
echo "  [✓] 后端编译完成"

echo "[2/7] 构建前端..."
cd "$PROJECT_DIR/frontend"
if [ ! -d "node_modules" ]; then
    npm install --prefer-offline 2>&1
fi
npm run build 2>&1
rm -rf "$PROJECT_DIR/backend/static"
cp -r dist "$PROJECT_DIR/backend/static"
echo "  [✓] 前端构建完成"
cd "$PROJECT_DIR"

# ========== 步骤 2：创建打包目录 ==========
echo "[3/7] 准备打包目录..."
PACK_DIR="/tmp/go-syncflow-pack"
rm -rf "$PACK_DIR"
mkdir -p "$PACK_DIR/Go-SyncFlow"

# ========== 步骤 3：复制文件 ==========
echo "[4/7] 复制文件..."

# 复制后端（含预编译二进制和静态文件）
cp -r backend "$PACK_DIR/Go-SyncFlow/"

# 复制前端源码（含 node_modules 以减少新机器下载）
cp -r frontend "$PACK_DIR/Go-SyncFlow/"
# 删除前端构建产物（backend/static 已包含）
rm -rf "$PACK_DIR/Go-SyncFlow/frontend/dist"

# 复制脚本
cp -r scripts "$PACK_DIR/Go-SyncFlow/"

# 复制工具包（Go 安装包等）
if [ -d "tooling" ]; then
    cp -r tooling "$PACK_DIR/Go-SyncFlow/"
fi

# 复制文档
if [ -d "docs" ]; then
    cp -r docs "$PACK_DIR/Go-SyncFlow/"
fi
cp -f README.md "$PACK_DIR/Go-SyncFlow/" 2>/dev/null || true
cp -f 快速部署说明.txt "$PACK_DIR/Go-SyncFlow/" 2>/dev/null || true

# ========== 步骤 4：清理敏感数据 ==========
echo "[5/7] 清理敏感数据..."

DEST="$PACK_DIR/Go-SyncFlow"

# ---- 数据库文件（通知渠道、用户数据等全在里面）----
# 删除所有位置的数据库文件（backend根目录 + data子目录）
find "$DEST" \( -name "*.db" -o -name "*.sqlite" -o -name "*.sqlite3" \) -exec rm -f {} + 2>/dev/null || true
rm -rf "$DEST/backend/data"
echo "  - 数据库文件已清除（含通知渠道、短信通道、用户数据等）"

# ---- 证书和密钥 ----
rm -rf "$DEST/backend/certs"
find "$DEST" \( -name "jwt_secret" -o -name "*.pem" -o -name "*.key" -o -name "*.crt" \) -exec rm -f {} + 2>/dev/null || true
echo "  - 证书和密钥已清除"

# ---- 日志文件 ----
find "$DEST" -name "*.log" -exec rm -f {} + 2>/dev/null || true
echo "  - 日志文件已清除"

# ---- 旧编译产物 ----
rm -f "$DEST/backend/server"
rm -f "$DEST/backend/bi-dashboard"

# ---- 其他 ----
rm -rf "$DEST/.git"
rm -rf "$DEST/frontend/node_modules/.vite"
rm -rf "$DEST/frontend/node_modules/.cache"
find "$DEST/backend" -name "*_test.go" -exec rm -f {} + 2>/dev/null || true

# ---- 最终校验：确认没有残留的数据库文件 ----
REMAINING_DB=$(find "$DEST" \( -name "*.db" -o -name "*.sqlite" -o -name "*.sqlite3" \) 2>/dev/null | wc -l)
if [ "$REMAINING_DB" -gt 0 ]; then
    echo "  [!] 警告：仍有残留数据库文件，正在强制删除..."
    find "$DEST" \( -name "*.db" -o -name "*.sqlite" -o -name "*.sqlite3" \) -exec rm -f {} + 2>/dev/null
fi

echo "  [OK] 全部敏感数据已清理"

# ========== 步骤 5：确保脚本可执行 ==========
echo "[6/7] 设置文件权限..."
chmod +x "$DEST/scripts/"*.sh
chmod +x "$DEST/backend/go-syncflow"

# ========== 步骤 6：打包 ==========
echo "[7/7] 压缩打包..."

cd /tmp
tar -czf "$OUTPUT_FILE" -C go-syncflow-pack Go-SyncFlow

# 清理临时目录
rm -rf "$PACK_DIR"

# 复制到项目目录
cp "$OUTPUT_FILE" "$PROJECT_DIR/"

FILE_SIZE=$(du -h "$PROJECT_DIR/${OUTPUT_NAME}.tar.gz" | cut -f1)

echo ""
echo "=========================================="
echo "    打包完成！"
echo "=========================================="
echo ""
echo "输出文件: $PROJECT_DIR/${OUTPUT_NAME}.tar.gz"
echo "文件大小: $FILE_SIZE"
echo ""
echo "=========================================="
echo "  部署方法（一键部署）"
echo "=========================================="
echo ""
echo "  1. 上传到新服务器"
echo "  2. 解压："
echo "     tar -xzf ${OUTPUT_NAME}.tar.gz -C /opt/"
echo ""
echo "  3. 一键启动："
echo "     cd /opt/Go-SyncFlow && chmod +x scripts/*.sh"
echo "     ./scripts/start.sh"
echo ""
echo "  4. 其他命令："
echo "     ./scripts/stop.sh           # 停止服务"
echo "     ./scripts/restart.sh        # 重启服务"
echo "     ./scripts/reset-admin.sh    # 重置管理员密码"
echo ""
echo "  默认管理员: admin / Admin@2024"
echo "  HTTP:  http://服务器IP:8080"
echo "  HTTPS: https://服务器IP:8443"
echo ""
echo "  说明："
echo "  - 包含预编译二进制，无需在新机器上编译"
echo "  - LDAP 服务默认启用 Samba 属性（兼容群晖NAS）"
echo "  - 数据库初始为空，首次启动自动初始化"
echo "  - 通知渠道、同步连接器等需在界面中配置"
echo ""
