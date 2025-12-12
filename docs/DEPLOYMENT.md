# 部署指南

本文档介绍如何在生产环境部署 5SecondsGo。

## 目录

- [环境要求](#环境要求)
- [Docker 部署](#docker-部署)
- [手动部署](#手动部署)
- [Nginx 配置](#nginx-配置)
- [SSL 证书](#ssl-证书)
- [监控配置](#监控配置)
- [备份策略](#备份策略)
- [故障排查](#故障排查)

## 环境要求

### 硬件要求

| 组件 | 最低配置 | 推荐配置 |
|------|----------|----------|
| CPU | 2 核 | 4 核+ |
| 内存 | 4 GB | 8 GB+ |
| 磁盘 | 50 GB SSD | 100 GB SSD |
| 带宽 | 10 Mbps | 50 Mbps+ |

### 软件要求

- Docker 20.10+
- Docker Compose 2.0+
- 或者:
  - Go 1.21+
  - PostgreSQL 15+
  - Redis 7+
  - Nginx 1.20+

## Docker 部署

### 1. 准备配置文件

```bash
# 克隆项目
git clone https://github.com/jihongxing/5scondsGo.git
cd 5scondsGo

# 复制配置文件
cp server/config/config.example.yaml server/config/config.yaml

# 编辑配置
vim server/config/config.yaml
```

**重要配置项:**

```yaml
server:
  mode: release

database:
  password: "your-secure-db-password"

auth:
  jwt_secret: "your-very-long-random-secret-key"
```

### 2. 创建生产环境 docker-compose

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  server:
    build: ./server
    ports:
      - "8080:8080"
      - "9091:9091"
    environment:
      - CONFIG_PATH=/app/config/config.yaml
    volumes:
      - ./server/config:/app/config:ro
    depends_on:
      - postgres
      - redis
    restart: always

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: fiveseconds
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: fiveseconds
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./server/migrations/init.sql:/docker-entrypoint-initdb.d/init.sql
    restart: always

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    restart: always

  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    restart: always

  grafana:
    image: grafana/grafana:latest
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
    volumes:
      - grafana_data:/var/lib/grafana
      - ./deploy/grafana/dashboards:/etc/grafana/provisioning/dashboards
    ports:
      - "3000:3000"
    restart: always

volumes:
  postgres_data:
  redis_data:
  prometheus_data:
  grafana_data:
```

### 3. 创建环境变量文件

```bash
# .env
DB_PASSWORD=your-secure-db-password
REDIS_PASSWORD=your-redis-password
GRAFANA_PASSWORD=your-grafana-password
```

### 4. 启动服务

```bash
docker-compose -f docker-compose.prod.yml up -d
```

### 5. 验证部署

```bash
# 检查服务状态
docker-compose -f docker-compose.prod.yml ps

# 查看日志
docker-compose -f docker-compose.prod.yml logs -f server

# 测试 API
curl http://localhost:8080/api/health
```

## 手动部署

### 1. 安装 PostgreSQL

```bash
# Ubuntu/Debian
sudo apt install postgresql-15

# 创建数据库
sudo -u postgres psql
CREATE USER fiveseconds WITH PASSWORD 'your-password';
CREATE DATABASE fiveseconds OWNER fiveseconds;
\q

# 初始化表结构
psql -U fiveseconds -d fiveseconds -f server/migrations/init.sql
```

### 2. 安装 Redis

```bash
# Ubuntu/Debian
sudo apt install redis-server

# 配置密码
sudo vim /etc/redis/redis.conf
# requirepass your-redis-password

sudo systemctl restart redis
```

### 3. 构建服务端

```bash
cd server
go build -o 5secondsgo-server cmd/server/main.go
```

### 4. 创建 Systemd 服务

```ini
# /etc/systemd/system/5secondsgo.service
[Unit]
Description=5SecondsGo Game Server
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/5secondsgo
ExecStart=/opt/5secondsgo/5secondsgo-server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable 5secondsgo
sudo systemctl start 5secondsgo
```

### 5. 构建管理后台

```bash
cd admin
flutter build web --release

# 部署到 Nginx
sudo cp -r build/web/* /var/www/5secondsgo-admin/
```

## Nginx 配置

```nginx
# /etc/nginx/sites-available/5secondsgo

# API 和 WebSocket
upstream backend {
    server 127.0.0.1:8080;
    keepalive 64;
}

# 主站点
server {
    listen 80;
    server_name api.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.example.com;

    ssl_certificate /etc/letsencrypt/live/api.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.example.com/privkey.pem;

    # API
    location /api {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # WebSocket
    location /ws {
        proxy_pass http://backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 86400;
    }
}

# 管理后台
server {
    listen 443 ssl http2;
    server_name admin.example.com;

    ssl_certificate /etc/letsencrypt/live/admin.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/admin.example.com/privkey.pem;

    root /var/www/5secondsgo-admin;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

## SSL 证书

使用 Let's Encrypt 免费证书:

```bash
# 安装 certbot
sudo apt install certbot python3-certbot-nginx

# 获取证书
sudo certbot --nginx -d api.example.com -d admin.example.com

# 自动续期
sudo certbot renew --dry-run
```

## 监控配置

### Prometheus

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: '5secondsgo'
    static_configs:
      - targets: ['localhost:9091']
```

### Grafana 仪表盘

导入预配置的仪表盘:
- `deploy/grafana/dashboards/system-overview.json`
- `deploy/grafana/dashboards/business-metrics.json`

### 告警规则

```yaml
# deploy/prometheus/alerts.yaml
groups:
  - name: 5secondsgo
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
```

## 备份策略

### 数据库备份

```bash
#!/bin/bash
# /opt/scripts/backup-db.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR=/opt/backups/postgres

pg_dump -U fiveseconds -d fiveseconds | gzip > $BACKUP_DIR/fiveseconds_$DATE.sql.gz

# 保留最近 7 天
find $BACKUP_DIR -name "*.sql.gz" -mtime +7 -delete
```

```bash
# Crontab
0 2 * * * /opt/scripts/backup-db.sh
```

### Redis 备份

```bash
#!/bin/bash
# /opt/scripts/backup-redis.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR=/opt/backups/redis

redis-cli BGSAVE
sleep 5
cp /var/lib/redis/dump.rdb $BACKUP_DIR/dump_$DATE.rdb

find $BACKUP_DIR -name "*.rdb" -mtime +7 -delete
```

## 故障排查

### 常见问题

#### 1. 服务无法启动

```bash
# 查看日志
journalctl -u 5secondsgo -f

# 检查端口占用
netstat -tlnp | grep 8080
```

#### 2. 数据库连接失败

```bash
# 测试连接
psql -h localhost -U fiveseconds -d fiveseconds

# 检查 pg_hba.conf
sudo vim /etc/postgresql/15/main/pg_hba.conf
```

#### 3. WebSocket 断开

- 检查 Nginx 超时配置
- 检查防火墙设置
- 查看服务端日志

#### 4. 性能问题

```bash
# 查看系统资源
htop

# 查看数据库慢查询
SELECT * FROM pg_stat_activity WHERE state = 'active';

# 查看 Redis 内存
redis-cli INFO memory
```

### 日志位置

| 组件 | 日志位置 |
|------|----------|
| 服务端 | stdout / journalctl |
| PostgreSQL | /var/log/postgresql/ |
| Redis | /var/log/redis/ |
| Nginx | /var/log/nginx/ |

## 安全建议

1. **修改默认密码** - 所有默认密码必须修改
2. **使用 HTTPS** - 生产环境必须使用 SSL
3. **防火墙** - 只开放必要端口 (80, 443)
4. **定期更新** - 保持系统和依赖更新
5. **日志审计** - 定期检查访问日志
6. **备份验证** - 定期测试备份恢复
