# Requirements Document

## Introduction

本功能为 5SecondsGo 游戏平台实现完整的日志和监控系统。目标是提供结构化日志、分布式追踪、完整的指标采集、Loki 日志聚合、Grafana 可视化仪表盘以及告警规则配置，确保系统的可观测性和问题排查能力。

## Glossary

- **Trace ID**: 唯一标识一个请求链路的 ID，贯穿整个请求生命周期
- **Structured Logging**: 结构化日志，以 JSON 格式输出，便于日志聚合和查询
- **Loki**: Grafana 开源的日志聚合系统，用于收集和查询日志
- **Prometheus**: 开源的监控和告警系统，用于采集和存储时序指标
- **Grafana**: 开源的可视化平台，用于展示指标和日志
- **P95/P99 Latency**: 95%/99% 的请求延迟低于该值
- **Alerting Rules**: 告警规则，定义触发告警的条件

## Requirements

### Requirement 1

**User Story:** As a developer, I want structured request logging with trace IDs, so that I can trace and debug issues across the request lifecycle.

#### Acceptance Criteria

1. WHEN an HTTP request arrives THEN the System SHALL generate a unique trace ID and attach it to the request context
2. WHEN the System logs any message during request processing THEN the System SHALL include the trace ID in the log entry
3. WHEN an HTTP request completes THEN the System SHALL log the request method, path, status code, latency, and trace ID in JSON format
4. WHEN a WebSocket message is processed THEN the System SHALL log the message type, user ID, room ID, and processing time with a trace ID
5. WHEN an error occurs during request processing THEN the System SHALL log the error with stack trace and trace ID

### Requirement 2

**User Story:** As a developer, I want database query logging, so that I can identify slow queries and optimize database performance.

#### Acceptance Criteria

1. WHEN a database query executes THEN the System SHALL record the query execution time
2. WHEN a database query takes longer than 100ms THEN the System SHALL log the query as a slow query with execution time and trace ID
3. WHEN database query metrics are recorded THEN the System SHALL update the DB latency P95 metric
4. WHEN the System starts THEN the System SHALL configure the database connection pool with query logging enabled

### Requirement 3

**User Story:** As an operator, I want Loki integration for log aggregation, so that I can search and analyze logs centrally.

#### Acceptance Criteria

1. WHEN the System outputs logs THEN the System SHALL format logs in JSON with consistent field names compatible with Loki
2. WHEN configuring the logging system THEN the System SHALL support outputting logs to stdout for container log collection
3. WHEN logs are written THEN the System SHALL include labels for service name, environment, and log level
4. WHEN the System is deployed THEN the System SHALL provide a Promtail configuration for shipping logs to Loki

### Requirement 4

**User Story:** As an operator, I want complete Prometheus metrics collection, so that I can monitor system health and performance.

#### Acceptance Criteria

1. WHEN an HTTP request completes THEN the System SHALL record the request latency in a Prometheus histogram with method and path labels
2. WHEN a WebSocket connection is established or closed THEN the System SHALL update the active WebSocket connections gauge
3. WHEN a game round starts or ends THEN the System SHALL increment the game rounds counter with room ID and status labels
4. WHEN a database query executes THEN the System SHALL record the query latency in a Prometheus histogram with operation label
5. WHEN the System exposes metrics THEN the System SHALL include business metrics for online players, active rooms, and daily volume

### Requirement 5

**User Story:** As an operator, I want pre-configured Grafana dashboards, so that I can visualize system metrics without manual setup.

#### Acceptance Criteria

1. WHEN the System is deployed THEN the System SHALL provide a Grafana dashboard JSON for system overview metrics
2. WHEN viewing the system dashboard THEN the operator SHALL see panels for request rate, error rate, latency percentiles, and active connections
3. WHEN viewing the business dashboard THEN the operator SHALL see panels for online players, active rooms, games per minute, and daily volume
4. WHEN viewing the infrastructure dashboard THEN the operator SHALL see panels for CPU, memory, database connections, and Redis connections

### Requirement 6

**User Story:** As an operator, I want Prometheus alerting rules, so that I can be notified of system issues automatically.

#### Acceptance Criteria

1. WHEN API latency P95 exceeds 500ms for 5 minutes THEN the System SHALL trigger a warning alert
2. WHEN API latency P95 exceeds 1000ms for 2 minutes THEN the System SHALL trigger a critical alert
3. WHEN error rate exceeds 1% for 5 minutes THEN the System SHALL trigger a warning alert
4. WHEN error rate exceeds 5% for 2 minutes THEN the System SHALL trigger a critical alert
5. WHEN database connection pool utilization exceeds 80% for 5 minutes THEN the System SHALL trigger a warning alert
6. WHEN the System provides alerting rules THEN the System SHALL include rules in Prometheus alerting rules format

### Requirement 7

**User Story:** As a developer, I want runtime log level adjustment, so that I can increase logging verbosity for debugging without restarting the service.

#### Acceptance Criteria

1. WHEN an admin calls the log level API with a valid level THEN the System SHALL change the log level immediately
2. WHEN the log level is changed THEN the System SHALL log the change event with the old and new levels
3. WHEN an invalid log level is provided THEN the System SHALL return an error and maintain the current level
4. WHEN the System starts THEN the System SHALL read the initial log level from configuration

### Requirement 8

**User Story:** As a developer, I want request context propagation, so that I can correlate logs across service boundaries.

#### Acceptance Criteria

1. WHEN an HTTP request includes an X-Trace-ID header THEN the System SHALL use that trace ID instead of generating a new one
2. WHEN the System makes outbound HTTP requests THEN the System SHALL include the trace ID in the X-Trace-ID header
3. WHEN logging within a goroutine spawned from a request THEN the System SHALL preserve the trace ID from the parent context
4. WHEN a WebSocket connection is established THEN the System SHALL associate a session ID that persists across messages

### Requirement 9

**User Story:** As an operator, I want room activity logging, so that I can monitor game room operations and troubleshoot issues.

#### Acceptance Criteria

1. WHEN a room is created THEN the System SHALL log the room ID, owner ID, bet amount, and configuration
2. WHEN a player joins or leaves a room THEN the System SHALL log the user ID, room ID, and timestamp
3. WHEN a game round starts THEN the System SHALL log the room ID, round number, participant count, and pool amount
4. WHEN a game round settles THEN the System SHALL log the room ID, round number, winner IDs, prize distribution, and settlement time
5. WHEN a game round fails THEN the System SHALL log the room ID, round number, failure reason, and refund details
6. WHEN a room status changes THEN the System SHALL log the room ID, old status, new status, and reason

### Requirement 10

**User Story:** As an operator, I want fund anomaly logging, so that I can detect and investigate suspicious financial activities.

#### Acceptance Criteria

1. WHEN a player balance becomes negative THEN the System SHALL log a critical alert with user ID, balance, and last transaction
2. WHEN a single transaction exceeds 10000 THEN the System SHALL log a warning with user ID, amount, and transaction type
3. WHEN a player wins more than 10 consecutive rounds THEN the System SHALL log a warning with user ID, win streak, and room ID
4. WHEN a player win rate exceeds 80 percent over 50 rounds THEN the System SHALL log a warning with user ID, win rate, and total rounds
5. WHEN multiple accounts share the same device fingerprint THEN the System SHALL log a warning with user IDs and fingerprint
6. WHEN owner custody quota is insufficient for player withdrawal THEN the System SHALL log a warning with owner ID, quota, and requested amount

### Requirement 11

**User Story:** As an operator, I want periodic reconciliation logging, so that I can audit fund conservation and detect discrepancies.

#### Acceptance Criteria

1. WHEN the 2-hour reconciliation task runs THEN the System SHALL log the total player balance, custody quota, margin, platform balance, and difference
2. WHEN the reconciliation detects an imbalance THEN the System SHALL log a critical alert with the discrepancy amount and affected accounts
3. WHEN the daily owner reconciliation runs THEN the System SHALL log each owner reconciliation result with owner ID, player count, and balance totals
4. WHEN reconciliation completes successfully THEN the System SHALL log the reconciliation period, duration, and balanced status
5. WHEN reconciliation fails due to error THEN the System SHALL log the error details and retry information

