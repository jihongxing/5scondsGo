#!/bin/bash
# 5SecondsGo 部署脚本

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 检查必要的环境变量
check_env() {
    log_info "检查环境变量..."
    
    if [ ! -f .env ]; then
        log_error ".env 文件不存在，请复制 .env.example 并配置"
        exit 1
    fi
    
    source .env
    
    required_vars=("DB_PASSWORD" "REDIS_PASSWORD" "JWT_SECRET" "GRAFANA_PASSWORD")
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            log_error "环境变量 $var 未设置"
            exit 1
        fi
    done
    
    log_info "环境变量检查通过"
}

# 创建必要的目录
create_dirs() {
    log_info "创建必要的目录..."
    mkdir -p deploy/alertmanager
    mkdir -p deploy/loki
    mkdir -p deploy/grafana/provisioning/datasources
    mkdir -p deploy/grafana/provisioning/dashboards
}

# 构建服务
build_services() {
    log_info "构建 Docker 镜像..."
    docker-compose -f docker-compose.prod.yml build --no-cache server
}

# 启动服务
start_services() {
    log_info "启动服务..."
    docker-compose -f docker-compose.prod.yml up -d
    
    log_info "等待服务启动..."
    sleep 10
    
    # 检查服务状态
    docker-compose -f docker-compose.prod.yml ps
}

# 停止服务
stop_services() {
    log_info "停止服务..."
    docker-compose -f docker-compose.prod.yml down
}

# 查看日志
view_logs() {
    local service=${1:-server}
    docker-compose -f docker-compose.prod.yml logs -f "$service"
}

# 备份数据库
backup_db() {
    log_info "备份数据库..."
    source .env
    
    local backup_dir="backups/$(date +%Y%m%d)"
    mkdir -p "$backup_dir"
    
    docker exec fiveseconds-postgres pg_dump -U "$DB_USER" "$DB_NAME" | gzip > "$backup_dir/db_$(date +%H%M%S).sql.gz"
    
    log_info "备份完成: $backup_dir"
}

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    # 检查 API
    if curl -sf http://localhost:8080/api/health > /dev/null; then
        log_info "API 服务: 正常"
    else
        log_error "API 服务: 异常"
    fi
    
    # 检查 Prometheus
    if curl -sf http://localhost:9090/-/healthy > /dev/null; then
        log_info "Prometheus: 正常"
    else
        log_warn "Prometheus: 异常"
    fi
    
    # 检查 Grafana
    if curl -sf http://localhost:3000/api/health > /dev/null; then
        log_info "Grafana: 正常"
    else
        log_warn "Grafana: 异常"
    fi
}

# 显示帮助
show_help() {
    echo "5SecondsGo 部署脚本"
    echo ""
    echo "用法: $0 <命令>"
    echo ""
    echo "命令:"
    echo "  check     检查环境配置"
    echo "  build     构建 Docker 镜像"
    echo "  start     启动所有服务"
    echo "  stop      停止所有服务"
    echo "  restart   重启所有服务"
    echo "  logs      查看日志 (可指定服务名)"
    echo "  backup    备份数据库"
    echo "  health    健康检查"
    echo "  help      显示帮助"
}

# 主函数
main() {
    case "${1:-help}" in
        check)
            check_env
            ;;
        build)
            check_env
            create_dirs
            build_services
            ;;
        start)
            check_env
            create_dirs
            start_services
            health_check
            ;;
        stop)
            stop_services
            ;;
        restart)
            stop_services
            start_services
            health_check
            ;;
        logs)
            view_logs "$2"
            ;;
        backup)
            backup_db
            ;;
        health)
            health_check
            ;;
        help|*)
            show_help
            ;;
    esac
}

main "$@"
