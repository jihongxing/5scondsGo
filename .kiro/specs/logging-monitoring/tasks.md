# Implementation Plan

- [x] 1. Set up trace context and structured logger foundation





  - [ ] 1.1 Create trace context package with TraceID generation and context propagation
    - Create `server/pkg/trace/context.go` with NewTraceID, WithTraceID, GetTraceID functions
    - Implement TraceContext struct with TraceID, SessionID, UserID fields
    - Use UUID v4 for trace ID generation

    - _Requirements: 1.1, 8.1, 8.3_
  - [x]* 1.2 Write property test for trace ID uniqueness


    - **Property 1: Trace ID Uniqueness**
    - **Validates: Requirements 1.1**
  - [ ] 1.3 Create structured logger wrapper with context support
    - Create `server/pkg/logger/logger.go` with Logger struct
    - Implement WithContext method to extract trace info from context
    - Configure JSON output format with consistent field names

    - Add service name and environment labels

    - _Requirements: 1.2, 3.1, 3.3_



  - [ ]* 1.4 Write property test for log entry structure
    - **Property 2: Log Entry Structure Completeness**
    - **Validates: Requirements 1.2, 3.1, 3.3**

- [ ] 2. Implement HTTP request logging middleware
  - [ ] 2.1 Create request logging middleware
    - Create `server/internal/middleware/logging.go`
    - Extract or generate trace ID from X-Trace-ID header


    - Inject trace ID into request context
    - Set X-Trace-ID response header

    - Log request completion with method, path, status, latency, trace_id
    - _Requirements: 1.1, 1.3, 8.1_
  - [-]* 2.2 Write property test for request log completeness


    - **Property 3: Request Log Field Completeness**

    - **Validates: Requirements 1.3**
  - [x] 2.3 Integrate logging middleware into main.go

    - Add middleware to Gin router
    - Replace existing gin.Recovery with custom recovery that includes trace ID
    - _Requirements: 1.3, 1.5_


- [ ] 3. Implement database query logging
  - [ ] 3.1 Create database query logger
    - Create `server/internal/repository/db_logger.go`

    - Implement LogQuery method with operation, duration, error tracking
    - Add slow query detection (>100ms threshold)
    - _Requirements: 2.1, 2.2_
  - [-]* 3.2 Write property test for slow query detection


    - **Property 4: Slow Query Detection Threshold**


    - **Validates: Requirements 2.2**
  - [x] 3.3 Integrate query logging into repository layer

    - Wrap database operations with timing and logging
    - Update existing repository methods to use DBLogger
    - _Requirements: 2.1, 2.4_

- [ ] 4. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.


- [x] 5. Implement Prometheus metrics collection

  - [x] 5.1 Create comprehensive metrics package

    - Create `server/pkg/metrics/metrics.go`
    - Define HTTP request histogram (method, path labels)

    - Define WebSocket connections gauge
    - Define database query histogram (operation label)
    - Define game rounds counter (room_id, status labels)
    - Define business metrics gauges (online_players, active_rooms)

    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_
  - [ ]* 5.2 Write property test for HTTP metrics recording
    - **Property 6: HTTP Metrics Recording**

    - **Validates: Requirements 4.1**
  - [ ]* 5.3 Write property test for WebSocket gauge accuracy
    - **Property 7: WebSocket Connection Gauge Accuracy**
    - **Validates: Requirements 4.2**
  - [ ] 5.4 Integrate metrics recording into middleware and services
    - Update HTTP middleware to record request metrics


    - Update WebSocket handler to record connection metrics



    - Update game manager to record game round metrics
    - _Requirements: 4.1, 4.2, 4.3_

- [x] 6. Implement P95 latency calculation


  - [x] 6.1 Update monitoring service with proper P95 calculation

    - Refactor calculateP95 method for correctness
    - Ensure thread-safe access to latency samples
    - _Requirements: 2.3_
  - [ ]* 6.2 Write property test for P95 calculation
    - **Property 5: P95 Latency Calculation Correctness**
    - **Validates: Requirements 2.3**

- [x] 7. Implement log level runtime adjustment

  - [-] 7.1 Create log level API handler


    - Create `server/internal/handler/admin_log_handler.go`


    - Implement GetLogLevel and SetLogLevel endpoints
    - Validate log level input (debug, info, warn, error)
    - Log level change events
    - _Requirements: 7.1, 7.2, 7.3, 7.4_
  - [x]* 7.2 Write property test for log level change

    - **Property 9: Log Level Change Validity**
    - **Validates: Requirements 7.1, 7.3**
  - [ ] 7.3 Add log level API routes to admin endpoints
    - Add GET /api/admin/log-level
    - Add PUT /api/admin/log-level
    - _Requirements: 7.1_




- [x] 8. Checkpoint - Ensure all tests pass


  - Ensure all tests pass, ask the user if questions arise.

- [x] 9. Implement room activity logging

  - [ ] 9.1 Create room activity logger service
    - Create `server/internal/service/room_logger.go`
    - Implement LogRoomCreated, LogPlayerJoined, LogPlayerLeft

    - Implement LogRoundStarted, LogRoundSettled, LogRoundFailed
    - Implement LogRoomStatusChanged
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6_
  - [ ]* 9.2 Write property test for room activity log completeness
    - **Property 12: Room Activity Log Completeness**

    - **Validates: Requirements 9.1, 9.2, 9.3, 9.4, 9.5, 9.6**
  - [ ] 9.3 Integrate room activity logger into game manager
    - Add logging calls to room creation flow
    - Add logging calls to player join/leave flow
    - Add logging calls to game round lifecycle

    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6_





- [x] 10. Implement fund anomaly logging


  - [x] 10.1 Create fund anomaly logger service

    - Create `server/internal/service/fund_anomaly_logger.go`

    - Implement LogNegativeBalance with critical alert
    - Implement LogLargeTransaction (>10000 threshold)
    - Implement LogConsecutiveWins (>10 threshold)
    - Implement LogHighWinRate (>80% over 50 rounds)
    - Implement LogDuplicateDeviceFingerprint
    - Implement LogInsufficientCustodyQuota
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6_
  - [ ]* 10.2 Write property test for fund anomaly detection
    - **Property 13: Fund Anomaly Detection Thresholds**
    - **Validates: Requirements 10.1, 10.2**
  - [ ]* 10.3 Write property test for consecutive win detection
    - **Property 14: Consecutive Win Detection**
    - **Validates: Requirements 10.3**
  - [ ] 10.4 Integrate fund anomaly logger into services
    - Add logging to balance update operations
    - Add logging to transaction creation


    - Add logging to risk service

    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6_

- [x] 11. Implement reconciliation logging



  - [ ] 11.1 Create reconciliation logger service
    - Create `server/internal/service/reconciliation_logger.go`
    - Implement LogReconciliationStarted
    - Implement LogReconciliationResult with all balance fields
    - Implement LogOwnerReconciliation
    - Implement LogReconciliationError


    - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5_
  - [ ]* 11.2 Write property test for reconciliation log completeness
    - **Property 15: Reconciliation Log Completeness**


    - **Validates: Requirements 11.1, 11.3, 11.4**
  - [x]* 11.3 Write property test for reconciliation imbalance alert

    - **Property 16: Reconciliation Imbalance Alert**
    - **Validates: Requirements 11.2**


  - [ ] 11.4 Integrate reconciliation logger into fund service
    - Update startConservationAutoCheck to use logger

    - Update startDailyOwnerConservationCheck to use logger
    - _Requirements: 11.1, 11.3, 11.4, 11.5_


- [x] 12. Checkpoint - Ensure all tests pass


  - Ensure all tests pass, ask the user if questions arise.

- [ ] 13. Create configuration files for observability stack
  - [ ] 13.1 Create Promtail configuration
    - Create `deploy/promtail/promtail.yaml`
    - Configure log scraping from stdout


    - Add pipeline stages for JSON parsing
    - Add labels for service, environment, level

    - _Requirements: 3.2, 3.4_
  - [ ] 13.2 Create Prometheus alerting rules
    - Create `deploy/prometheus/alerts.yaml`

    - Add HighAPILatency rule (P95 > 500ms for 5min)
    - Add CriticalAPILatency rule (P95 > 1s for 2min)
    - Add HighErrorRate rule (>1% for 5min)
    - Add CriticalErrorRate rule (>5% for 2min)
    - Add HighDBConnectionUsage rule (>80% for 5min)
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_
  - [ ] 13.3 Create Grafana dashboard JSON files
    - Create `deploy/grafana/dashboards/system-overview.json`
    - Create `deploy/grafana/dashboards/business-metrics.json`
    - Create `deploy/grafana/dashboards/infrastructure.json`
    - _Requirements: 5.1, 5.2, 5.3, 5.4_
  - [ ] 13.4 Update docker-compose.yml with observability services
    - Add Promtail service configuration
    - Update Prometheus with alerting rules volume
    - Add Grafana dashboard provisioning
    - _Requirements: 3.4, 5.1_


- [x] 14. Implement WebSocket logging enhancements

  - [x] 14.1 Add session ID to WebSocket connections

    - Generate session ID on connection establishment
    - Store session ID in connection context
    - Include session ID in all WebSocket message logs
    - _Requirements: 1.4, 8.4_
  - [ ]* 14.2 Write property test for WebSocket session ID persistence
    - **Property 11: WebSocket Session ID Persistence**
    - **Validates: Requirements 8.4**
  - [x] 14.3 Add trace ID propagation for outbound requests

    - Update HTTP client to include X-Trace-ID header
    - Ensure trace ID is preserved in spawned goroutines
    - _Requirements: 8.2, 8.3_
  - [ ]* 14.4 Write property test for trace ID propagation
    - **Property 10: Trace ID Propagation**
    - **Validates: Requirements 8.1, 8.3**


- [ ] 15. Final integration and cleanup
  - [ ] 15.1 Update main.go to initialize all logging components
    - Initialize structured logger with config
    - Initialize room activity logger
    - Initialize fund anomaly logger
    - Initialize reconciliation logger
    - Wire up all components

    - _Requirements: 3.2, 7.4_
  - [ ] 15.2 Add metrics for fund anomalies and reconciliation
    - Add fund_anomaly_total counter with type label
    - Add reconciliation_total counter with status and period labels
    - _Requirements: 4.5_

- [ ] 16. Final Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.
