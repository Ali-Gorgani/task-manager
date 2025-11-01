# Task Manager Architecture

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                             │
│                (Web, Mobile, CLI, Other Services)                │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ HTTP/REST
                             │
┌────────────────────────────▼────────────────────────────────────┐
│                      API GATEWAY LAYER                           │
│                                                                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Swagger    │  │ Prometheus   │  │   Health     │          │
│  │     Docs     │  │   Metrics    │  │    Check     │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    GIN FRAMEWORK (HTTP)                          │
│                                                                   │
│  ┌──────────────────────────────────────────────────┐           │
│  │         Middleware Pipeline                       │           │
│  │  • CORS                                           │           │
│  │  • Prometheus Metrics                             │           │
│  │  • Request Logging                                │           │
│  │  • Recovery                                       │           │
│  └──────────────────────────────────────────────────┘           │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                       HANDLERS LAYER                             │
│                   (HTTP Request/Response)                        │
│                                                                   │
│  ┌──────────────────────────────────────────────────┐           │
│  │  TaskHandler                                      │           │
│  │  • CreateTask(c *gin.Context)                    │           │
│  │  • GetTask(c *gin.Context)                       │           │
│  │  • ListTasks(c *gin.Context)                     │           │
│  │  • UpdateTask(c *gin.Context)                    │           │
│  │  • DeleteTask(c *gin.Context)                    │           │
│  └──────────────────────────────────────────────────┘           │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                       SERVICE LAYER                              │
│                     (Business Logic)                             │
│                                                                   │
│  ┌──────────────────────────────────────────────────┐           │
│  │  TaskService                                      │           │
│  │  • CreateTask(req) -> Task                       │           │
│  │  • GetTask(id) -> Task                           │           │
│  │  • ListTasks(filter) -> TaskListResponse         │           │
│  │  • UpdateTask(id, req) -> Task                   │           │
│  │  • DeleteTask(id) -> error                       │           │
│  │  • GetTaskCount() -> int                         │           │
│  └──────────────────────────────────────────────────┘           │
└──────────────┬──────────────────────────────┬───────────────────┘
               │                              │
               ▼                              ▼
┌──────────────────────────┐   ┌──────────────────────────────────┐
│    CACHE LAYER           │   │    REPOSITORY LAYER              │
│    (Optional)            │   │    (Data Access)                 │
│                          │   │                                  │
│  ┌────────────────────┐ │   │  ┌────────────────────────────┐ │
│  │   RedisCache       │ │   │  │  TaskRepository Interface  │ │
│  │                    │ │   │  │  • Create(task)            │ │
│  │  • GetTask()       │ │   │  │  • GetByID(id)             │ │
│  │  • SetTask()       │ │   │  │  • GetAll(filter)          │ │
│  │  • DeleteTask()    │ │   │  │  • Update(task)            │ │
│  │  • GetTaskList()   │ │   │  │  • Delete(id)              │ │
│  │  • SetTaskList()   │ │   │  │  • Count()                 │ │
│  │  • Invalidate()    │ │   │  └────────────────────────────┘ │
│  └────────────────────┘ │   │             │                    │
│           │              │   │             ▼                    │
│           ▼              │   │  ┌────────────────────────────┐ │
│  ┌────────────────────┐ │   │  │  PostgresTaskRepository    │ │
│  │   Redis Server     │ │   │  │  (Implementation)          │ │
│  │   (In-Memory)      │ │   │  └────────────────────────────┘ │
│  └────────────────────┘ │   └──────────────┬───────────────────┘
└──────────────────────────┘                  │
                                              ▼
                             ┌──────────────────────────────────┐
                             │     PostgreSQL Database          │
                             │                                  │
                             │  Tables:                         │
                             │  • tasks                         │
                             │    - id (PK, UUID)               │
                             │    - title                       │
                             │    - description                 │
                             │    - status (indexed)            │
                             │    - assignee (indexed)          │
                             │    - created_at (indexed)        │
                             │    - updated_at                  │
                             └──────────────────────────────────┘
```

## Data Flow

### 1. Create Task Flow
```
Client → POST /api/v1/tasks → Handler → Service → Repository → PostgreSQL
                                   │
                                   └──→ Cache: Invalidate list cache
```

### 2. Get Task Flow (with Cache)
```
Client → GET /api/v1/tasks/:id → Handler → Service → Cache (Hit?) → Return
                                                  │
                                                  └─(Miss)→ Repository → PostgreSQL
                                                                  │
                                                                  └──→ Cache: Store
```

### 3. List Tasks Flow (with Cache & Filtering)
```
Client → GET /api/v1/tasks?status=pending&page=1 → Handler → Service
                                                                │
                                                                ▼
                                                         Check Cache Key
                                                                │
                                          ┌─────────────────────┴──────────────────┐
                                          ▼                                        ▼
                                    Cache Hit                                 Cache Miss
                                    Return                                        │
                                                                                  ▼
                                                                          Repository (with filters)
                                                                                  │
                                                                                  ▼
                                                                            PostgreSQL
                                                                                  │
                                                                                  ▼
                                                                            Store in Cache
```

## Component Responsibilities

### Handlers Layer
- HTTP request/response handling
- Input validation (binding)
- Status code management
- Error formatting

### Service Layer
- Business logic
- Input validation (business rules)
- Orchestration between cache and repository
- Transaction management

### Repository Layer
- Database queries
- SQL execution
- Data mapping
- Error handling

### Cache Layer
- Cache-aside pattern implementation
- TTL management (5 minutes)
- Cache key generation
- Invalidation logic

## Design Patterns Used

1. **Repository Pattern** - Abstracts data access
2. **Dependency Injection** - Loose coupling
3. **Cache-Aside** - Lazy loading cache pattern
4. **Factory Pattern** - Model creation (NewTask)
5. **Interface Segregation** - TaskRepository interface

## Scalability Considerations

### Horizontal Scaling
- ✅ Stateless API design
- ✅ Database connection pooling
- ✅ Redis for distributed caching
- ⚠️ Need load balancer for multiple instances

### Vertical Scaling
- ✅ Efficient database queries with indexes
- ✅ Connection pooling
- ✅ Minimal memory footprint

### Performance Optimizations
1. **Database Indexes** on frequently queried fields
2. **Redis Caching** reduces database load
3. **Pagination** prevents large data transfers
4. **Connection Pooling** reuses connections
5. **Multi-stage Docker** reduces image size

## Security Considerations

### Current Implementation
- ✅ SQL injection protection (parameterized queries)
- ✅ Input validation
- ✅ Error message sanitization

### Future Enhancements
- [ ] JWT authentication
- [ ] Rate limiting
- [ ] API key authentication
- [ ] HTTPS/TLS
- [ ] CORS configuration
- [ ] Request size limits

## Monitoring & Observability

```
┌────────────────────────────────────────┐
│         Application                    │
│                                        │
│  ┌──────────────────────────────────┐ │
│  │   Metrics Collection              │ │
│  │   • requests_total                │ │
│  │   • request_latency_histogram     │ │
│  │   • tasks_count                   │ │
│  └──────────────┬───────────────────┘ │
└─────────────────┼─────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│         Prometheus                      │
│         (Metrics Storage)               │
│                                         │
│  • Scrapes /metrics every 15s          │
│  • Stores time-series data             │
│  • Query interface                     │
└─────────────────────────────────────────┘
```

## Deployment Architecture

```
┌──────────────────────────────────────────────────────────┐
│                    Docker Compose                        │
│                                                          │
│  ┌────────────┐  ┌────────────┐  ┌──────────────────┐  │
│  │   API      │  │ PostgreSQL │  │     Redis        │  │
│  │  :3000     │  │   :5432    │  │     :6379        │  │
│  └────────────┘  └────────────┘  └──────────────────┘  │
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │              Prometheus                           │  │
│  │                :9090                              │  │
│  └──────────────────────────────────────────────────┘  │
│                                                          │
│            Network: taskmanager-network                  │
└──────────────────────────────────────────────────────────┘
```

## Technology Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| HTTP Framework | Gin | High-performance HTTP routing |
| Database | PostgreSQL 15 | ACID-compliant data storage |
| Cache | Redis 7 | In-memory caching |
| Metrics | Prometheus | Monitoring and alerting |
| Documentation | Swagger/OpenAPI | API documentation |
| Testing | testify, sqlmock | Unit and integration testing |
| Containerization | Docker | Application packaging |
| Orchestration | Docker Compose | Multi-container deployment |

---

*This architecture is designed to be simple, maintainable, and production-ready for a task management microservice.*
