# 5SecondsGo 生产环境部署指南

本文档详细介绍如何在生产环境部署 5SecondsGo 游戏平台。

## 目录

- [架构概览](#架构概览)
- [服务器要求](#服务器要求)
- [快速部署](#快速部署)
- [详细部署步骤](#详细部署步骤)
- [SSL 证书配置](#ssl-证书配置)
- [监控配置](#监控配置)
- [备份策略](#备份策略)
- [运维操作](#运维操作)
- [故障排查](#故障排查)
- [安全加固](#安全加固)

---

## 架构概览

```
                                    ┌─────────────────┐
                                    │   用户/客户端    │
                                    └────────┬────────┘
                                             │
                                    ┌────────▼────────┐
                                    │     Nginx       │
                                    │  (SSL + 反向代理) │
                                    └────────┬────────┘
                         ┌───────────────────┼───────────────────┐
                         │                   │                   │
                ┌────────▼────────┐ ┌────────▼────────┐ ┌────────▼────────┐
                │   API 服务      │ │   管理后台      │ │   监控面板      │
                │ api.domain.com  │ │ admin.domain.com│ │monitor.domain.com│
                └────────┬────────┘ └─────────────────┘ └────────┬────────┘
                         │                                       │
                ┌────────▼────────┐                     ┌────────▼────────┐
                │   Go Server     │                     │    Grafana      │
                │   :8080/:9091   │                     │     :3000       │
                └────────┬────────┘                     └────────┬────────┘
         ┌───────────────┼───────────────┐                       │
         │               │               │              ┌────────┴────────┐
┌────────▼────┐ ┌────────▼────┐ ┌────────▼────┐ ┌───────▼───────┐ ┌───────▼───────┐
│ PostgreSQL  │ │    Redis    │ │  Prometheus │ │     Loki      │ │  Alertmanager │
│    :5432    │ │    :6379    │ │    :9090    │ │    :3100      │ │    :9093      │
└─────────────┘ └─────────────┘ └─────────────┘ └───────────────┘ └───────────────┘
```

### 服务说明

| 服务 | 端口 | 用途 |
|------|------|------|
| Go Server | 8080 | API + WebSocket |
| Go Metrics | 9091 | Prometheus 指标 |
| PostgreSQL | 5432 | 主数据库 |
| Redis | 6379 | 缓存 + 会话 |
| Prometheus | 9090 | 指标收集 |
| Grafana | 3000 | 监控仪表盘 |
| Loki | 3100 | 日志存储 |
| Alertmanager | 9093 | 告警管理 |

---

## 服务器要求

### 最低配置

| 资源 | 要求 |
|------|------|
| CPU | 2 核 |
| 内存 | 4 GB |
| 磁盘 | 50 GB SSD |
| 带宽 | 10 Mbps |

### 推荐配置

| 资源 | 要求 |
|------|------|
| CPU | 4 核+ |
| 内存 | 8 GB+ |
| 磁盘 | 100 GB SSD |
| 带宽 | 50 Mbps+ |

### 软件要求

- Ubuntu 20.04/22.04 LTS 或 CentOS 8+
- Docker 20.10+
- Docker Compose 2.0+
- Nginx 1.20+

---

## 快速部署

```bash
# 1. 克隆项目
git clone https://github.com/jihongxing/5scondsGo.git
cd 5scondsGo

# 2. 配置环境变量
cp .env.example .env
vim .env  # 修改配置

# 3. 启动服务
chmod +x scripts/deploy.sh
./scripts/deploy.sh start

# 4. 检查状态
./scripts/deploy.sh health
```

---

## 详细部署步骤

### 1. 安装 Docker

```bash
# Ubuntu
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# 安装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 验证安装
docker --version
docker-compose --version
```

### 2. 安装 Nginx

```bash
# Ubuntu
sudo apt update
sudo apt install nginx -y

# 启动并设置开机自启
sudo systemctl start nginx
sudo systemctl enable nginx
```

### 3. 克隆项目

```bash
cd /opt
sudo git clone https://github.com/jihongxing/5scondsGo.git
sudo chown -R $USER:$USER 5scondsGo
cd 5scondsGo
```

### 4. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env` 文件：

```bash
# 数据库配置 - 使用强密码
DB_USER=fiveseconds
DB_PASSWORD=YourSecureDBPassword123!
DB_NAME=fiveseconds

# Redis 配置
REDIS_PASSWORD=YourSecureRedisPassword456!

# JWT 密钥 - 生成随机字符串
# 生成方法: openssl rand -base64 32
JWT_SECRET=your-very-long-random-jwt-secret-key-at-least-32-chars

# Grafana 配置
GRAFANA_USER=admin
GRAFANA_PASSWORD=YourGrafanaPassword789!
GRAFANA_ROOT_URL=https://monitor.your-domain.com

# 域名配置
API_DOMAIN=api.your-domain.com
ADMIN_DOMAIN=admin.your-domain.com
MONITOR_DOMAIN=monitor.your-domain.com
```

### 5. 配置服务端

编辑 `server/config/config.yaml`：

```yaml
server:
  host: 0.0.0.0
  port: 8080
  mode: release  # 生产环境必须设为 release

database:
  host: postgres  # Docker 网络中的服务名
  port: 5432
  sslmode: disable

redis:
  host: redis
  port: 6379

auth:
  jwt_secret: ${JWT_SECRET}  # 从环境变量读取
  jwt_expire: 24h

logging:
  level: info
  format: json
```

### 6. 构建并启动服务

```bash
# 构建镜像
./scripts/deploy.sh build

# 启动服务
./scripts/deploy.sh start

# 查看状态
docker-compose -f docker-compose.prod.yml ps
```

### 7. 配置 Nginx

```bash
# 复制配置文件
sudo cp deploy/nginx/nginx.conf /etc/nginx/sites-available/fiveseconds

# 修改域名
sudo sed -i 's/your-domain.com/your-actual-domain.com/g' /etc/nginx/sites-available/fiveseconds

# 启用配置
sudo ln -s /etc/nginx/sites-available/fiveseconds /etc/nginx/sites-enabled/

# 测试配置
sudo nginx -t

# 重载 Nginx
sudo systemctl reload nginx
```

### 8. 构建管理后台

```bash
cd admin
flutter build web --release

# 部署到 Nginx
sudo mkdir -p /var/www/fiveseconds-admin
sudo cp -r build/web/* /var/www/fiveseconds-admin/
sudo chown -R www-data:www-data /var/www/fiveseconds-admin
```

---

## SSL 证书配置

### 使用 Let's Encrypt

```bash
# 安装 Certbot
sudo apt install certbot python3-certbot-nginx -y

# 获取证书
sudo certbot --nginx -d api.your-domain.com -d admin.your-domain.com -d monitor.your-domain.com

# 测试自动续期
sudo certbot renew --dry-run
```

### 配置自动续期

```bash
# 添加 crontab
sudo crontab -e

# 添加以下行（每天凌晨 3 点检查续期）
0 3 * * * /usr/bin/certbot renew --quiet --post-hook "systemctl reload nginx"
```

---

## 监控配置

### 访问 Grafana

1. 打开 `https://monitor.your-domain.com`
2. 使用配置的用户名密码登录
3. 数据源已自动配置（Prometheus + Loki）

### 导入仪表盘

仪表盘已自动加载，包括：
- **系统概览**: 在线人数、活跃房间、交易量
- **业务指标**: 游戏轮次、胜率分布、资金流水
- **基础设施**: API 延迟、错误率、数据库连接

### 配置告警通知

编辑 `deploy/alertmanager/alertmanager.yml`：

```yaml
# 钉钉告警示例
receivers:
  - name: 'dingtalk'
    webhook_configs:
      - url: 'https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN'
        send_resolved: true

# 企业微信告警示例
receivers:
  - name: 'wechat'
    webhook_configs:
      - url: 'https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR_KEY'
        send_resolved: true
```

重启 Alertmanager：

```bash
docker-compose -f docker-compose.prod.yml restart alertmanager
```

---

## 备份策略

### 数据库备份

```bash
# 手动备份
./scripts/deploy.sh backup

# 自动备份 (添加到 crontab)
0 2 * * * /opt/5scondsGo/scripts/deploy.sh backup
```

### 备份脚本

```bash
#!/bin/bash
# /opt/5scondsGo/scripts/backup-full.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR=/opt/backups/$DATE

mkdir -p $BACKUP_DIR

# 备份数据库
docker exec fiveseconds-postgres pg_dump -U fiveseconds fiveseconds | gzip > $BACKUP_DIR/db.sql.gz

# 备份 Redis
docker exec fiveseconds-redis redis-cli -a $REDIS_PASSWORD BGSAVE
sleep 5
docker cp fiveseconds-redis:/data/dump.rdb $BACKUP_DIR/redis.rdb

# 备份配置
cp -r /opt/5scondsGo/.env $BACKUP_DIR/
cp -r /opt/5scondsGo/server/config $BACKUP_DIR/

# 清理 7 天前的备份
find /opt/backups -type d -mtime +7 -exec rm -rf {} +

echo "备份完成: $BACKUP_DIR"
```

### 恢复数据库

```bash
# 恢复数据库
gunzip -c backup/db.sql.gz | docker exec -i fiveseconds-postgres psql -U fiveseconds fiveseconds

# 恢复 Redis
docker cp backup/redis.rdb fiveseconds-redis:/data/dump.rdb
docker-compose -f docker-compose.prod.yml restart redis
```

---

## 运维操作

### 常用命令

```bash
# 查看服务状态
docker-compose -f docker-compose.prod.yml ps

# 查看日志
./scripts/deploy.sh logs server
./scripts/deploy.sh logs postgres

# 重启服务
docker-compose -f docker-compose.prod.yml restart server

# 停止所有服务
./scripts/deploy.sh stop

# 启动所有服务
./scripts/deploy.sh start

# 健康检查
./scripts/deploy.sh health
```

### 更新部署

```bash
# 拉取最新代码
git pull origin main

# 重新构建
./scripts/deploy.sh build

# 重启服务
./scripts/deploy.sh restart
```

### 扩容

```bash
# 增加 Go Server 实例 (需要负载均衡)
docker-compose -f docker-compose.prod.yml up -d --scale server=3
```

---

## 故障排查

### 服务无法启动

```bash
# 查看详细日志
docker-compose -f docker-compose.prod.yml logs --tail=100 server

# 检查端口占用
netstat -tlnp | grep -E '8080|5432|6379'

# 检查磁盘空间
df -h
```

### 数据库连接失败

```bash
# 检查 PostgreSQL 状态
docker exec fiveseconds-postgres pg_isready

# 检查连接
docker exec -it fiveseconds-postgres psql -U fiveseconds -d fiveseconds -c "SELECT 1"
```

### WebSocket 断开

1. 检查 Nginx 超时配置
2. 检查防火墙设置
3. 查看服务端日志

```bash
# 检查 WebSocket 连接数
curl -s http://localhost:9091/metrics | grep websocket
```

### 性能问题

```bash
# 查看容器资源使用
docker stats

# 查看数据库慢查询
docker exec fiveseconds-postgres psql -U fiveseconds -c "SELECT * FROM pg_stat_activity WHERE state = 'active'"

# 查看 Redis 内存
docker exec fiveseconds-redis redis-cli -a $REDIS_PASSWORD INFO memory
```

---

## 安全加固

### 1. 防火墙配置

```bash
# UFW (Ubuntu)
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

### 2. 限制数据库访问

数据库端口只监听本地：
```yaml
# docker-compose.prod.yml 中已配置
ports:
  - "127.0.0.1:5432:5432"
```

### 3. 监控面板 IP 白名单

编辑 Nginx 配置：
```nginx
# 只允许特定 IP 访问监控面板
location / {
    allow 1.2.3.4;  # 你的 IP
    deny all;
    proxy_pass http://grafana;
}
```

### 4. 定期更新

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 更新 Docker 镜像
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d
```

### 5. 日志审计

- 定期检查 Nginx 访问日志
- 监控异常登录尝试
- 设置资金异常告警

---

## 检查清单

部署前确认：

- [ ] 修改所有默认密码
- [ ] 配置 SSL 证书
- [ ] 配置防火墙
- [ ] 配置备份策略
- [ ] 配置告警通知
- [ ] 测试恢复流程
- [ ] 记录所有密码到安全位置

部署后确认：

- [ ] API 服务正常响应
- [ ] WebSocket 连接正常
- [ ] 管理后台可访问
- [ ] 监控面板可访问
- [ ] 告警通知正常
- [ ] 备份任务正常执行

---

## 联系支持

- 项目地址: https://github.com/jihongxing/5scondsGo
- 问题反馈: https://github.com/jihongxing/5scondsGo/issues
