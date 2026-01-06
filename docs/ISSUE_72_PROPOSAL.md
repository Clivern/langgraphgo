# Checkpoint 数据隔离优化方案

## Issue #72 - 关于 checkpoint 的数据隔离问题

### 问题描述

在面向问答的 agent 应用场景中，不同会话（thread）的 state 需要数据隔离。目前的实现需要遍历所有 checkpoints 来查找特定 `thread_id` 的最新 checkpoint：

```go
// 找到 thread 的最新 checkpoint
var latestCP *graph.Checkpoint
for _, cp := range checkpoints {
    if cp.Metadata["thread_id"] == threadID {
        if latestCP == nil || cp.Version > latestCP.Version {
            latestCP = cp
        }
    }
}
```

当会话数量增长时，这种全表扫描的方式会导致性能线性下降。

### 现状分析

| Store | 现有实现 | 时间复杂度 | 问题 |
|-------|---------|-----------|-----|
| **memory** | 遍历所有 checkpoints，匹配 metadata | O(n) | 全表扫描 |
| **file** | 读取所有 JSON 文件，解析后过滤 | O(n) | 文件 I/O + 解析开销 |
| **postgres** | 有 `execution_id` 索引，但 `thread_id` 在 JSONB 中 | O(n) | JSONB 字段未索引 |
| **redis** | 使用 `execution:{id}:checkpoints` Set | O(1) | 仅支持 execution_id 查询 |

**关键发现**：
- `thread_id` 存储在 `metadata` map/JSONB 中，无法高效查询
- PostgreSQL 和 Redis 有索引机制，但未利用于 `thread_id`
- Memory/File store 完全依赖线性扫描

---

## 优化方案

### 1. 扩展 CheckpointStore 接口

在 `store/checkpoint.go` 中添加新的查询方法：

```go
// CheckpointStore defines the interface for checkpoint persistence
type CheckpointStore interface {
    // Save stores a checkpoint
    Save(ctx context.Context, checkpoint *Checkpoint) error

    // Load retrieves a checkpoint by ID
    Load(ctx context.Context, checkpointID string) (*Checkpoint, error)

    // List returns all checkpoints for a given execution
    List(ctx context.Context, executionID string) ([]*Checkpoint, error)

    // === 新增方法 ===

    // ListByThread returns all checkpoints for a specific thread_id
    // Returns checkpoints sorted by version (ascending)
    ListByThread(ctx context.Context, threadID string) ([]*Checkpoint, error)

    // GetLatestByThread returns the latest checkpoint for a thread_id
    // Returns the checkpoint with the highest version
    GetLatestByThread(ctx context.Context, threadID string) (*Checkpoint, error)

    // Delete removes a checkpoint
    Delete(ctx context.Context, checkpointID string) error

    // Clear removes all checkpoints for an execution
    Clear(ctx context.Context, executionID string) error
}
```

**优点**：
- 向后兼容：现有实现继续使用 `List` 方法
- 语义明确：`GetLatestByThread` 直接返回最新 checkpoint
- 专用优化：各 store 可为 `thread_id` 查询提供最优实现

---

### 2. 各 Store 实现优化

#### 2.1 MemoryCheckpointStore

**优化方式**：使用嵌套 map 建立 `thread_id` 索引

```go
type MemoryCheckpointStore struct {
    checkpoints    map[string]*store.Checkpoint  // id -> checkpoint
    threadIndex    map[string][]string           // thread_id -> []checkpoint IDs
    executionIndex map[string][]string           // execution_id -> []checkpoint IDs
    mutex          sync.RWMutex
}

// GetLatestByThread - O(1) 索引查找 + O(k) 遍历 (k = 该线程的 checkpoint 数)
func (m *MemoryCheckpointStore) GetLatestByThread(_ context.Context, threadID string) (*store.Checkpoint, error) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()

    ids, exists := m.threadIndex[threadID]
    if !exists || len(ids) == 0 {
        return nil, fmt.Errorf("no checkpoints found for thread: %s", threadID)
    }

    var latest *store.Checkpoint
    for _, id := range ids {
        cp := m.checkpoints[id]
        if latest == nil || cp.Version > latest.Version {
            latest = cp
        }
    }

    return latest, nil
}
```

**时间复杂度**：O(k)，其中 k 是单个 thread 的 checkpoint 数量

---

#### 2.2 FileCheckpointStore

**优化方式**：维护 `thread_id` -> `checkpoint_ids` 索引文件

```go
type FileCheckpointStore struct {
    path       string
    mutex      sync.RWMutex
    threadIndex map[string][]string  // 内存缓存，启动时加载
}

// 保存索引到 JSON 文件
// checkpoints/by_thread/{thread_id}.json -> ["checkpoint_id1", "checkpoint_id2", ...]
```

**时间复杂度**：O(k) + 文件 I/O

---

#### 2.3 PostgresCheckpointStore

**优化方式**：添加 `thread_id` 列和索引

```sql
-- Schema 迁移
ALTER TABLE checkpoints ADD COLUMN IF NOT EXISTS thread_id TEXT;
CREATE INDEX IF NOT EXISTS idx_checkpoints_thread_id ON checkpoints (thread_id);

-- 组合索引优化（可选）
CREATE INDEX IF NOT EXISTS idx_checkpoints_execution_thread
    ON checkpoints (execution_id, thread_id);
```

```go
// GetLatestByThread - 使用索引查询
func (s *PostgresCheckpointStore) GetLatestByThread(ctx context.Context, threadID string) (*graph.Checkpoint, error) {
    query := fmt.Sprintf(`
        SELECT id, node_name, state, metadata, timestamp, version
        FROM %s
        WHERE thread_id = $1
        ORDER BY version DESC
        LIMIT 1
    `, s.tableName)

    var cp graph.Checkpoint
    // ... scan and return
}
```

**时间复杂度**：O(log n)（索引查找）

---

#### 2.4 RedisCheckpointStore

**优化方式**：扩展索引模式，添加 `thread:{id}:checkpoints` Set

```go
func (s *RedisCheckpointStore) threadKey(id string) string {
    return fmt.Sprintf("%sthread:%s:checkpoints", s.prefix, id)
}

// Save 时维护索引
func (s *RedisCheckpointStore) Save(ctx context.Context, checkpoint *graph.Checkpoint) error {
    // ... existing code ...

    // Index by thread_id if present
    if threadID, ok := checkpoint.Metadata["thread_id"].(string); ok && threadID != "" {
        threadKey := s.threadKey(threadID)
        pipe.SAdd(ctx, threadKey, checkpoint.ID)
        if s.ttl > 0 {
            pipe.Expire(ctx, threadKey, s.ttl)
        }
    }

    // ... existing code ...
}

// GetLatestByThread - 使用索引
func (s *RedisCheckpointStore) GetLatestByThread(ctx context.Context, threadID string) (*graph.Checkpoint, error) {
    threadKey := s.threadKey(threadID)
    checkpointIDs, err := s.client.SMembers(ctx, threadKey).Result()
    // ... fetch checkpoints and find latest by version
}
```

**时间复杂度**：O(k)（网络往返）

---

### 3. 更新上层调用

在 `graph/checkpointing.go` 中优先使用新方法：

```go
// GetState - 优先使用 GetLatestByThread
func (cr *CheckpointableRunnable[S]) GetState(ctx context.Context, config *Config) (*StateSnapshot, error) {
    var threadID string
    var checkpointID string

    if config != nil && config.Configurable != nil {
        if tid, ok := config.Configurable["thread_id"].(string); ok {
            threadID = tid
        }
        if cid, ok := config.Configurable["checkpoint_id"].(string); ok {
            checkpointID = cid
        }
    }

    var checkpoint *store.Checkpoint
    var err error

    if checkpointID != "" {
        checkpoint, err = cr.config.Store.Load(ctx, checkpointID)
    } else if threadID != "" {
        // 优先使用新的高效方法
        if latestGetter, ok := cr.config.Store.(interface {
            GetLatestByThread(ctx context.Context, threadID string) (*store.Checkpoint, error)
        }); ok {
            checkpoint, err = latestGetter.GetLatestByThread(ctx, threadID)
        } else {
            // Fallback 到 List 方法
            checkpoints, err := cr.config.Store.List(ctx, threadID)
            if err == nil && len(checkpoints) > 0 {
                checkpoint = checkpoints[0]
                for _, cp := range checkpoints {
                    if cp.Version > checkpoint.Version {
                        checkpoint = cp
                    }
                }
            }
        }
    }

    // ... rest of the method
}
```

---

## 实现计划

| 阶段 | 任务 | 优先级 |
|-----|------|-------|
| 1 | 更新 `CheckpointStore` 接口定义 | 高 |
| 2 | 实现 `MemoryCheckpointStore` 优化 | 高 |
| 3 | 实现 `PostgresCheckpointStore` 优化 | 高 |
| 4 | 实现 `RedisCheckpointStore` 优化 | 中 |
| 5 | 实现 `FileCheckpointStore` 优化 | 中 |
| 6 | 更新 `graph/checkpointing.go` 使用新方法 | 高 |
| 7 | 更新测试用例 | 高 |
| 8 | 运行 `make test` 验证 | 高 |

---

## 兼容性说明

- ✅ **向后兼容**：保留所有现有方法，不破坏现有代码
- ✅ **渐进式迁移**：新方法优先，回退到旧方法作为 fallback
- ✅ **可选实现**：Store 可选择性地实现新方法

---

## 性能对比

| 场景 | 优化前 | 优化后 |
|-----|-------|--------|
| 1000 threads，每 thread 10 checkpoints | O(10000) 全表扫描 | O(10) 索引查找 |
| 10000 threads，每 thread 10 checkpoints | O(100000) 全表扫描 | O(10) 索引查找 |

**理论提升**：约 100-1000 倍（取决于 thread 数量）

---

## 测试策略

1. **单元测试**：每个 store 的新方法单独测试
2. **集成测试**：验证 `CheckpointableRunnable.GetState` 使用新方法
3. **性能测试**：对比优化前后的查询时间
4. **兼容性测试**：验证 fallback 逻辑正常工作

---

## 后续优化方向

1. **复合查询**：支持 `thread_id + checkpoint_ns` 等复合条件查询
2. **分页支持**：`ListByThread` 添加 `limit` 和 `offset` 参数
3. **缓存层**：在 memory 中缓存热点 thread 的最新 checkpoint
4. **TTL 管理**：Redis 自动过期，其他 store 需要手动清理
