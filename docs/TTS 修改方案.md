# TTS 修改方案

## 背景与目标

当前 `xiaozhi-server` 的 TTS 处理在单连接内是“单实例 + 单消费链路”，当 LLM 输出较长文本或高并发会话出现时，TTS 往往成为瓶颈。  
目标是在尽量不破坏现有协议与发送逻辑的前提下，引入“多实例并行合成 + 保序播放”，提升吞吐与首包速度。

---

## 一、当前调度逻辑（现状）

### 1. 主链路

1. `ConnectionHandler.chat()` 将内容按 `FIRST/MIDDLE/LAST` 放入 `conn.tts.tts_text_queue`
2. `TTSProviderBase.tts_text_priority_thread()` 单线程消费 `tts_text_queue`
3. 文本分段后调用 `to_tts_stream()` -> `text_to_speak()`（外部 TTS）
4. 合成后的音频帧写入 `tts_audio_queue`
5. `TTSProviderBase._audio_play_priority_thread()` 单线程消费 `tts_audio_queue`
6. 调用 `sendAudioMessage()` / `sendAudio()` 向客户端发送（含流控）

### 2. 现状关键特征

- **单连接仅 1 个 TTS Provider 实例**：`ConnectionHandler._initialize_tts()`
- **单文本消费线程**：`tts_text_priority_thread()` 串行执行外部 TTS 请求
- **单播放线程**：`_audio_play_priority_thread()` 串行下发音频
- **天然保序**：因为串行处理，顺序简单可靠

### 3. 现状瓶颈

- TTS 外部调用是串行，网络抖动会直接拉长整体播放延迟
- LLM 持续产出文本时，后续段落要排队等待，吞吐受限
- 无法利用多核与多连接并行能力（尤其是云端 TTS API）

---

## 二、为什么“多实例”必须配合“调度改造”

仅仅“初始化多个 TTS 实例”并不能自动提速，原因是：

- 当前入口仍然只有一条 `tts_text_queue` 且仅一个消费线程
- 该线程每次只会阻塞式执行一个 `to_tts_stream()`
- 即使有多个实例，也没有分发器把任务并行派发给它们

结论：**多实例要生效，必须重构调度层（任务分发 + 结果聚合 + 保序播放）。**

---

## 三、修改后的调度逻辑（建议方案）

## 1. 总体架构

引入三层调度：

1. **文本切片层（Producer）**  
   把 `MIDDLE` 文本分段并编号 `seq_id`
2. **并行合成层（Dispatcher + Worker Pool）**  
   按策略把分段派发到多个 TTS 实例并发合成
3. **保序聚合层（Reorderer）**  
   按 `seq_id` 重新排序后写入现有 `tts_audio_queue`

保留现有播放线程 `_audio_play_priority_thread()`，减少入侵。

## 2. 新调度流程（单会话）

1. 收到 `FIRST`：初始化会话级调度状态（`next_seq=0`、结果缓存、并发控制）
2. 收到 `TEXT`：分段并生成 `TTSTask(seq_id, text, sentence_id)`
3. Dispatcher 将任务分发到实例池（RoundRobin 或 LeastLatency）
4. Worker 实例并发执行 `text_to_speak`
5. 返回 `TTSResult(seq_id, opus_chunks or bytes, text)`
6. Reorderer 按 `next_seq` 连续出队，写入 `tts_audio_queue`
7. 收到 `LAST` 后等待所有未完成任务，flush 剩余结果，再投递 `LAST`

## 3. 调度策略建议

- **实例选择**：默认 `round_robin`，后续可扩展 `least_latency`
- **并发上限**：`asyncio.Semaphore(max_concurrency)`，避免触发供应商限流
- **失败重试**：同段失败后换实例重试 1~2 次，仍失败则降级文本提示
- **超时控制**：每段设置 `tts_timeout`，避免尾段长期阻塞
- **背压控制**：任务队列上限（如 100），超限时触发降级策略

---

## 四、与当前代码的映射改造点

## 1. 初始化层

- 文件：`main/xiaozhi-server/core/connection.py`
  - 现状：`self.tts = _initialize_tts()`
  - 改造：保留 `self.tts` 作为调度器对象，内部持有 `providers[]`

- 文件：`main/xiaozhi-server/core/utils/modules_initialize.py`
  - 新增 `initialize_tts_pool(config)` 或扩展 `initialize_tts(config)` 支持 `pool_size`

## 2. Provider 工厂层

- 文件：`main/xiaozhi-server/core/utils/tts.py`
  - 保持 `create_instance()`，由上层循环创建 N 个实例

## 3. 调度核心层（重点）

- 文件：`main/xiaozhi-server/core/providers/tts/base.py`
  - 现状：`tts_text_priority_thread()` 直接串行 `to_tts_stream`
  - 改造：拆分为
    - `task_dispatch_thread`（派发任务）
    - `result_reorder_thread`（保序回写）
    - 可复用现有 `_audio_play_priority_thread`

建议新增结构（示意）：

- `TTSTask(sentence_id, seq_id, text, is_last)`
- `TTSResult(sentence_id, seq_id, audio_packets, text, error)`
- `pending_results: Dict[seq_id, TTSResult]`
- `next_emit_seq: int`

## 4. 发送层

- 文件：`main/xiaozhi-server/core/handle/sendAudioHandle.py`
  - 基本不改；继续使用现有流控和 `start/sentence_start/stop` 状态消息
  - 前提：上游保证写入 `tts_audio_queue` 的顺序正确

---

## 五、配置设计建议

在 `TTS.<module>` 下增加：

- `pool_size`: 实例数量（默认 1，建议先 2）
- `max_concurrency`: 最大并发任务数（默认 = pool_size）
- `dispatch_policy`: `round_robin | least_latency`
- `segment_max_chars`: 段长上限（控制并发粒度）
- `segment_timeout_ms`: 单段超时
- `segment_retry`: 单段重试次数

示例（概念）：

```yaml
selected_module:
  TTS: IndexStreamTTS

TTS:
  IndexStreamTTS:
    type: index_stream
    pool_size: 2
    max_concurrency: 2
    dispatch_policy: round_robin
    segment_max_chars: 80
    segment_timeout_ms: 8000
    segment_retry: 1
```

---

## 六、两种改造路径对比

## 路径 A：最小改造（推荐先做）

- 仅在 `base.py` 增加“并发请求 + 保序回写”
- 保持 `ConnectionHandler`、`sendAudioHandle` 主逻辑不变
- 适合快速验证收益，风险低

## 路径 B：深度改造

- 抽象独立 `TTSOrchestrator`（调度器）类
- Provider 只负责“文本->音频”，调度统一由 Orchestrator 管
- 可扩展性最好，但改动面更大

---

## 七、预期收益与风险

## 收益

- 长文本场景下整体播报时间下降
- 外部 TTS 接口抖动对总时延影响降低
- 更好利用 CPU 与网络并发能力

## 风险

- 顺序错乱风险（必须保序）
- API 限流/封禁风险（需要并发上限）
- 内存上涨（任务与结果缓存）
- 异常路径更复杂（超时、重试、取消）

---

## 八、落地顺序（执行建议）

1. 先实现 `pool_size=2` + `round_robin` + `保序聚合`
2. 加入超时与重试（1 次）
3. 压测对比单实例与双实例（首包时延、总时长、错误率）
4. 再考虑 `least_latency` 自适应策略

---

## 九、结论

你判断得对：**调度逻辑必须修改**。  
在当前代码结构下，最优先是把 `TTSProviderBase` 从“串行消费器”升级为“并行分发 + 保序聚合”的调度核心，同时尽量复用现有播放与协议层，能以较低风险实现多实例提速。
