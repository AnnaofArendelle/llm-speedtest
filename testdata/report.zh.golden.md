# LLM API 测速报告

- 生成时间：2026-11-01 22:57:58
- 测试窗口：2026-11-01 22:12:58 ～ 2026-11-01 22:41:58
- 端点：`http://localhost:20128/v1`（识别为 9router）
- 范围：共 **163** 个模型，10 个提供商，归并为 **78** 个模型系列（excluded 14 combo models）
- 测试轮次：第 1 轮 100 次（成功 20）；第 2 轮 10 次（成功 2）
- 工具版本：llm-speedtest test

## 1. 结论摘要

| 指标 | 数值 |
|---|---|
| 测试模型数 | 163 |
| 可用（至少一次成功返回） | **34**（20.9%） |
| 不可用 | 129（79.1%） |
| 其中因提供商账号问题不可用（无凭证/认证失败/余额或额度用尽/套餐限制） | 78（60.5%） |
| 模型系列数 | 78（有可用成员 21，全员不可用 57） |
| 可用模型的稳定性（A 多轮全部成功，B 成功率 ≥ 2/3，C 不稳定） | A 27，B 4，C 3 |
| 有可用成员的系列的稳定性 | A 15，B 6，C 0 |
| 可用模型延迟中位数 / P90 / 最快 / 最慢 | 1.05s / 10.06s / 0.21s / 20.04s |

延迟口径：非流式 `chat/completions` 从发出请求到收到完整响应的耗时（`max_tokens` 很小），约等于首 token 时间加路由开销，反映的是客户端实际可感知的响应速度，不是提供商的纯推理延迟。

### 1.1 最稳定且最快的 20 个可用模型

| # | 模型 ID | 提供商 | 系列 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 |
|---|---|---|---|---|---|---|---|
| 1 | `amd/MiniCPM-V-4.7-1B` | amd | minicpm-vision-4 | A 稳定（3/3） | 0.21s | 0.21s | 0.29s |
| 2 | `cf/@cf/meta/llama-3.2-1b-instruct` | cf | llama-3 | A 稳定（3/3） | 0.41s | 0.40s | 0.42s |
| 3 | `groq/allam-2-7b` | groq | allam-2 | A 稳定（3/3） | 0.43s | 0.42s | 1.12s |
| 4 | `amd/MiniCPM5-2B` | amd | minicpm5 | A 稳定（3/3） | 0.62s | 0.16s | 8.70s |
| 5 | `groq/qwen/qwen3.8-27b` | groq | qwen3 | A 稳定（3/3） | 0.85s | 0.24s | 0.93s |
| 6 | `groq/meta-llama/llama-prompt-guard-2-86m` | groq | llama-prompt-guard-2 | A 稳定（3/3） | 0.89s | 0.26s | 1.00s |
| 7 | `groq/meta-llama/llama-prompt-guard-2-22m` | groq | llama-prompt-guard-2 | A 稳定（3/3） | 0.94s | 0.89s | 0.99s |
| 8 | `cf/@cf/meta/llama-3.3-70b-instruct-fp8-fast` | cf | llama-3 | A 稳定（3/3） | 0.98s | 0.34s | 1.01s |
| 9 | `groq/openai/gpt-oss-safeguard-20b` | groq | gpt-oss-guard | A 稳定（3/3） | 0.99s | 0.95s | 1.03s |
| 10 | `cf/@cf/zai-org/glm-4.7-flash` | cf | glm-4-flash | A 稳定（3/3） | 1.00s | 0.76s | 1.12s |
| 11 | `cf/@cf/meta/llama-3.1-8b-instruct-fp8-fast` | cf | llama-3 | A 稳定（3/3） | 1.01s | 0.98s | 1.29s |
| 12 | `amd/Qwen3.8-Flash-Next` | amd | qwen3-flash | A 稳定（3/3） | 1.02s | 0.85s | 1.43s |
| 13 | `groq/openai/gpt-oss-120b` | groq | gpt-oss | A 稳定（3/3） | 1.03s | 1.00s | 1.14s |
| 14 | `cf/@cf/meta/llama-3.2-3b-instruct` | cf | llama-3 | A 稳定（3/3） | 1.04s | 0.95s | 1.06s |
| 15 | `groq/openai/gpt-oss-20b` | groq | gpt-oss | A 稳定（3/3） | 1.05s | 0.39s | 1.23s |
| 16 | `cf/@cf/qwen/qwq-32b` | cf | qwq | A 稳定（3/3） | 1.05s | 0.95s | 1.11s |
| 17 | `cf/@cf/qwen/qwen2.5-coder-32b-instruct` | cf | qwen2-coder | A 稳定（3/3） | 1.08s | 0.43s | 1.81s |
| 18 | `cf/@cf/deepseek-ai/deepseek-r1-distill-qwen-32b` | cf | deepseek-r1 | A 稳定（3/3） | 1.17s | 1.08s | 1.24s |
| 19 | `cf/@cf/mistralai/mistral-small-3.1-24b-instruct` | cf | mistral-small-3 | A 稳定（3/3） | 1.18s | 0.97s | 1.34s |
| 20 | `amd/Qwen3.8-27B` | amd | qwen3 | A 稳定（3/3） | 1.30s | 0.68s | 2.34s |

### 1.2 可用模型延迟分布

| 中位延迟区间 | 模型数 | 占可用 |
|---|---|---|
| < 1 s | 12 | 35.3% |
| 1 – 2 s | 12 | 35.3% |
| 2 – 5 s | 4 | 11.8% |
| 5 – 10 s | 2 | 5.9% |
| 10 – 30 s | 4 | 11.8% |
| ≥ 30 s | 0 | 0.0% |

### 1.3 不可用原因分布

| 原因 | 模型数 | 占不可用 | 层级 |
|---|---|---|---|
| 余额不足，需充值 | 71 | 55.0% | 提供商/账号级 |
| 免费额度已用尽 | 5 | 3.9% | 提供商/账号级 |
| 需先接受模型条款 | 2 | 1.6% | 提供商/账号级 |
| 限流 (429) | 2 | 1.6% | 模型级/网络级 |
| 模型不存在/不可用 (404/400) | 19 | 14.7% | 模型级/网络级 |
| 模型已下线 (410) | 6 | 4.7% | 模型级/网络级 |
| 非聊天模型 | 7 | 5.4% | 模型级/网络级 |
| 接口不支持 (405) | 3 | 2.3% | 模型级/网络级 |
| 上游返回空内容 | 14 | 10.9% | 模型级/网络级 |

### 1.4 综合排名前 15 的模型系列（可用提供商多、稳定、快 → 少、不稳定、慢）

综合分 = 稳定性分 × 0.4 + 提供商广度分 × 0.3 + 速度分 × 0.3，详见第 4 节。

| # | 系列 | 综合分 | 可用提供商数 | 可用成员/总成员 | 稳定性（成功/尝试） | 系列中位 | 最稳定且最快的成员 @ 提供商 |
|---|---|---|---|---|---|---|---|
| 1 | **qwen3** | 84.4 | 2（稳定 2） | 2/6 | A 稳定（6/6） | 1.08s | `groq/qwen/qwen3.8-27b` @ groq：0.85s |
| 2 | **allam-2** | 80.0 | 1（稳定 1） | 1/1 | A 稳定（3/3） | 0.43s | `groq/allam-2-7b` @ groq：0.43s |
| 3 | **minicpm5** | 78.4 | 1（稳定 1） | 1/1 | A 稳定（3/3） | 0.62s | `amd/MiniCPM5-2B` @ amd：0.62s |
| 4 | **llama-prompt-guard-2** | 75.6 | 1（稳定 1） | 2/2 | A 稳定（6/6） | 0.91s | `groq/meta-llama/llama-prompt-guard-2-86m` @ groq：0.89s |
| 5 | **gpt-oss-guard** | 75.0 | 1（稳定 1） | 1/1 | A 稳定（3/3） | 0.99s | `groq/openai/gpt-oss-safeguard-20b` @ groq：0.99s |
| 6 | **glm-4-flash** | 74.9 | 1（稳定 1） | 1/1 | A 稳定（3/3） | 1.00s | `cf/@cf/zai-org/glm-4.7-flash` @ cf：1.00s |
| 7 | **qwen3-flash** | 74.7 | 1（稳定 1） | 1/4 | A 稳定（3/3） | 1.02s | `amd/Qwen3.8-Flash-Next` @ amd：1.02s |
| 8 | **mimo-v2-flash** | 74.7 | 2（稳定 2） | 2/4 | A 稳定（6/6） | 4.05s | `amd/MiMo-V2.6-Flash` @ amd：1.63s |
| 9 | **gpt-oss** | 74.6 | 1（稳定 1） | 2/2 | A 稳定（6/6） | 1.04s | `groq/openai/gpt-oss-120b` @ groq：1.03s |
| 10 | **qwq** | 74.5 | 1（稳定 1） | 1/1 | A 稳定（3/3） | 1.05s | `cf/@cf/qwen/qwq-32b` @ cf：1.05s |
| 11 | **qwen2-coder** | 74.4 | 1（稳定 1） | 1/1 | A 稳定（3/3） | 1.08s | `cf/@cf/qwen/qwen2.5-coder-32b-instruct` @ cf：1.08s |
| 12 | **minicpm-vision-4** | 74.3 | 1（稳定 1） | 2/2 | B 基本稳定（6/7） | 0.29s | `amd/MiniCPM-V-4.7-1B` @ amd：0.21s |
| 13 | **deepseek-r1** | 73.8 | 1（稳定 1） | 1/2 | A 稳定（3/3） | 1.17s | `cf/@cf/deepseek-ai/deepseek-r1-distill-qwen-32b` @ cf：1.17s |
| 14 | **mistral-small-3** | 73.7 | 1（稳定 1） | 1/1 | A 稳定（3/3） | 1.18s | `cf/@cf/mistralai/mistral-small-3.1-24b-instruct` @ cf：1.18s |
| 15 | **deepseek-v4-flash** | 72.3 | 3（稳定 2） | 5/13 | B 基本稳定（11/16） | 4.00s | `thai/deepseek-v4.1-flash:free` @ thai：3.64s |

## 2. 测试方法

1. 调用 `GET /models` 取得完整模型列表（含 `owned_by` 与上下文元数据）。
2. **第 1 轮**：对每个模型发送一次最小化 `chat/completions` 请求，记录 HTTP 状态、嵌套的上游状态码、报错原文、延迟与实际返回的模型名；全局与每提供商并发均受限。
3. **复测轮**：对瞬时性失败（超时、限流、上游 5xx、无可用渠道/容量、空响应、网络错误）以更低并发、更长超时再测一次；若网关给出 `reset after`/`retry-after` 提示则等待到期后再试。认证失败、无凭证、余额不足、模型不存在等确定性错误不复测。
4. **补采样轮**：对所有成功过的模型再各测若干次，排序使用延迟**中位数**，并给出最小/最大值。
5. **最终确认轮**：对复测后仍属瞬时性失败的模型再测一次；仍失败才判定为不可用。
6. 成功判定：HTTP 200 且响应体包含 `choices`。失败原因按状态码优先、关键词其次归类，报错原文全部保留。
7. **提供商归属**：不使用任何内置名称表，而是根据本端点返回数据动态选择最能区分模型且与错误中连接标识一致的字段（`owned_by` / id 前缀），可在界面中重命名或合并。
8. **系列归并**：去掉提供商/组织前缀与 `:free` 类后缀、参数规模、日期戳、量化标记与修饰词，版本只保留主版本号，保留区分产品线的档位/特长词；别名通过上游实际返回的模型名与本端点自身的词表学习，不使用内置别名字典。
9. **排名**：先按稳定性等级（A = 多轮尝试全部成功，B = 成功率 ≥ 2/3，C = 成功率 < 2/3），同级内按中位延迟；系列另给出综合分 = 稳定性 × 权重 + 提供商广度 × 权重 + 速度（对数刻度）× 权重。

本端点的提供商字段：`owned_by`（owned_by discriminates between models (10 distinct values); consistent with observed connections for 100% of 88 models）

## 3. 提供商总览

| 提供商 | 模型数 | 系列数 | 可用 | 可用率 | 可用模型中位延迟 | 主要失败原因 | 诊断 |
|---|---|---|---|---|---|---|---|
| cf | 11 | 7 | 10 | 90.9% | 1.05s | 模型不存在/不可用 (404/400) | 模型不存在/不可用 (404/400)：1/1 个失败模型 |
| amd | 16 | 14 | 9 | 56.2% | 1.30s | 非聊天模型 | 非聊天模型：5/7 个失败模型 |
| groq | 13 | 10 | 7 | 53.8% | 0.94s | 模型不存在/不可用 (404/400) | 模型不存在/不可用 (404/400)：2/6 个失败模型 |
| thai | 75 | 33 | 4 | 5.3% | 5.23s | 余额不足，需充值 | 余额不足，需充值：71/71 个失败模型 |
| stkj | 8 | 6 | 4 | 50.0% | 2.22s | 模型不存在/不可用 (404/400) | 模型不存在/不可用 (404/400)：3/4 个失败模型 |
| cu | 14 | 8 | 0 | 0.0% | - | 上游返回空内容 | 上游返回空内容：14/14 个失败模型 |
| oc | 12 | 11 | 0 | 0.0% | - | 模型不存在/不可用 (404/400) | 模型不存在/不可用 (404/400)：7/12 个失败模型 |
| nvidia | 9 | 8 | 0 | 0.0% | - | 模型已下线 (410) | 模型已下线 (410)：6/9 个失败模型 |
| cohere | 3 | 3 | 0 | 0.0% | - | 接口不支持 (405) | 接口不支持 (405)：3/3 个失败模型 |
| ps | 2 | 1 | 0 | 0.0% | - | 模型不存在/不可用 (404/400) | 模型不存在/不可用 (404/400)：2/2 个失败模型 |

## 4. 模型系列测速总排名

**4.1 综合排名**：从“可用提供商最多、响应最快、最稳定”排到“最少、最慢、最不稳定”。

| 分项 | 计算方式 | 权重 |
|---|---|---|
| 稳定性分 | 该系列可用成员全部尝试的成功率 × 100 | 0.40 |
| 提供商广度分 | 有可用成员的提供商数 ÷ 全部系列中的最大值 × 100 | 0.30 |
| 速度分 | 系列中位延迟按对数刻度映射：≤ 0.5 s 得 100 分，≥ 30 s 得 0 分 | 0.30 |

稳定性只统计“有过成功”的成员的全部尝试；提供商账号层面的失败不反映模型本身，不计入系列稳定性，但在“可用成员/总成员”和“不可用成员主因”中体现。

### 4.1 综合排名

| 排名 | 系列 | 综合分 | 稳定性分 | 广度分 | 速度分 | 可用提供商数（稳定） | 可用成员/总成员 | 稳定性（成功/尝试） | 系列中位 | 最快 | 最慢 | 稳定可用的成员（模型 @ 提供商：中位延迟，成功/尝试） |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | **qwen3** | **84.4** | 100 | 67 | 81 | 2（2） | 2/6 | A 稳定（6/6） | 1.08s | 0.24s | 2.34s | `groq/qwen/qwen3.8-27b` @ groq：0.85s（3/3）<br>`amd/Qwen3.8-27B` @ amd：1.30s（3/3） |
| 2 | **allam-2** | **80.0** | 100 | 33 | 100 | 1（1） | 1/1 | A 稳定（3/3） | 0.43s | 0.42s | 1.12s | `groq/allam-2-7b` @ groq：0.43s（3/3） |
| 3 | **minicpm5** | **78.4** | 100 | 33 | 95 | 1（1） | 1/1 | A 稳定（3/3） | 0.62s | 0.16s | 8.70s | `amd/MiniCPM5-2B` @ amd：0.62s（3/3） |
| 4 | **llama-prompt-guard-2** | **75.6** | 100 | 33 | 85 | 1（1） | 2/2 | A 稳定（6/6） | 0.91s | 0.26s | 1.00s | `groq/meta-llama/llama-prompt-guard-2-86m` @ groq：0.89s（3/3）<br>`groq/meta-llama/llama-prompt-guard-2-22m` @ groq：0.94s（3/3） |
| 5 | **gpt-oss-guard** | **75.0** | 100 | 33 | 83 | 1（1） | 1/1 | A 稳定（3/3） | 0.99s | 0.95s | 1.03s | `groq/openai/gpt-oss-safeguard-20b` @ groq：0.99s（3/3） |
| 6 | **glm-4-flash** | **74.9** | 100 | 33 | 83 | 1（1） | 1/1 | A 稳定（3/3） | 1.00s | 0.76s | 1.12s | `cf/@cf/zai-org/glm-4.7-flash` @ cf：1.00s（3/3） |
| 7 | **qwen3-flash** | **74.7** | 100 | 33 | 82 | 1（1） | 1/4 | A 稳定（3/3） | 1.02s | 0.85s | 1.43s | `amd/Qwen3.8-Flash-Next` @ amd：1.02s（3/3） |
| 8 | **mimo-v2-flash** | **74.7** | 100 | 67 | 49 | 2（2） | 2/4 | A 稳定（6/6） | 4.05s | 0.47s | 10.76s | `amd/MiMo-V2.6-Flash` @ amd：1.63s（3/3）<br>`thai/mimo-v2.6-flash:free` @ thai：6.47s（3/3） |
| 9 | **gpt-oss** | **74.6** | 100 | 33 | 82 | 1（1） | 2/2 | A 稳定（6/6） | 1.04s | 0.39s | 1.23s | `groq/openai/gpt-oss-120b` @ groq：1.03s（3/3）<br>`groq/openai/gpt-oss-20b` @ groq：1.05s（3/3） |
| 10 | **qwq** | **74.5** | 100 | 33 | 82 | 1（1） | 1/1 | A 稳定（3/3） | 1.05s | 0.95s | 1.11s | `cf/@cf/qwen/qwq-32b` @ cf：1.05s（3/3） |
| 11 | **qwen2-coder** | **74.4** | 100 | 33 | 81 | 1（1） | 1/1 | A 稳定（3/3） | 1.08s | 0.43s | 1.81s | `cf/@cf/qwen/qwen2.5-coder-32b-instruct` @ cf：1.08s（3/3） |
| 12 | **minicpm-vision-4** | **74.3** | 86 | 33 | 100 | 1（1） | 2/2 | B 基本稳定（6/7） | 0.29s | 0.21s | 0.43s | `amd/MiniCPM-V-4.7-1B` @ amd：0.21s（3/3） |
| 13 | **deepseek-r1** | **73.8** | 100 | 33 | 79 | 1（1） | 1/2 | A 稳定（3/3） | 1.17s | 1.08s | 1.24s | `cf/@cf/deepseek-ai/deepseek-r1-distill-qwen-32b` @ cf：1.17s（3/3） |
| 14 | **mistral-small-3** | **73.7** | 100 | 33 | 79 | 1（1） | 1/1 | A 稳定（3/3） | 1.18s | 0.97s | 1.34s | `cf/@cf/mistralai/mistral-small-3.1-24b-instruct` @ cf：1.18s（3/3） |
| 15 | **deepseek-v4-flash** | **72.3** | 69 | 100 | 49 | 3（2） | 5/13 | B 基本稳定（11/16） | 4.00s | 0.56s | 15.22s | `thai/deepseek-v4.1-flash:free` @ thai：3.64s（3/3）<br>`thai/deepseek-v4-flash:free` @ thai：4.00s（3/3）<br>`amd/DeepSeek-V4-Flash` @ amd：6.04s（3/3） |
| 16 | **sensenova-6-flash-lite** | **69.7** | 100 | 33 | 66 | 1（1） | 1/2 | A 稳定（3/3） | 2.04s | 1.53s | 2.41s | `stkj/sensenova-6.8-flash-lite` @ stkj：2.04s（3/3） |
| 17 | **llama-3** | **69.5** | 87 | 33 | 83 | 1（1） | 5/5 | B 基本稳定（13/15） | 1.01s | 0.34s | 1.29s | `cf/@cf/meta/llama-3.2-1b-instruct` @ cf：0.41s（3/3）<br>`cf/@cf/meta/llama-3.3-70b-instruct-fp8-fast` @ cf：0.98s（3/3）<br>`cf/@cf/meta/llama-3.1-8b-instruct-fp8-fast` @ cf：1.01s（3/3）<br>…另有 2 个可用成员见第 6 节 |
| 18 | **glm-5** | **62.6** | 67 | 33 | 87 | 1（0） | 1/6 | B 基本稳定（2/3） | 0.87s | 0.80s | 0.93s | `stkj/glm-5.2` @ stkj：0.87s（2/3） |
| 19 | **mimo-v2** | **55.3** | 100 | 33 | 18 | 1（1） | 1/2 | A 稳定（3/3） | 14.57s | 1.54s | 41.49s | `thai/mimo-v2.5:free` @ thai：14.57s（3/3） |
| 20 | **deepseek-flash** | **43.9** | 75 | 33 | 13 | 1（0） | 1/1 | B 基本稳定（3/4） | 17.72s | 13.69s | 21.05s | `stkj/deepseek-flash` @ stkj：17.72s（3/4） |
| 21 | **gemma-4** | **39.6** | 67 | 33 | 10 | 1（0） | 1/1 | B 基本稳定（2/3） | 20.04s | 2.29s | 37.79s | `amd/Gemma-4-31B-IT` @ amd：20.04s（2/3） |

### 4.2 按稳定性等级与延迟排序的系列

| 排名 | 系列 | 稳定性（成功/尝试） | 成员 | 可用成员 | 系列中位 | 可用的提供商 | 不可用成员主因 |
|---|---|---|---|---|---|---|---|
| 1 | **allam-2** | A 稳定（3/3） | 1 | 1 | 0.43s | groq(1) | - |
| 2 | **minicpm5** | A 稳定（3/3） | 1 | 1 | 0.62s | amd(1) | - |
| 3 | **llama-prompt-guard-2** | A 稳定（6/6） | 2 | 2 | 0.91s | groq(2) | - |
| 4 | **gpt-oss-guard** | A 稳定（3/3） | 1 | 1 | 0.99s | groq(1) | - |
| 5 | **glm-4-flash** | A 稳定（3/3） | 1 | 1 | 1.00s | cf(1) | - |
| 6 | **qwen3-flash** | A 稳定（3/3） | 4 | 1 | 1.02s | amd(1) | 余额不足，需充值(3) |
| 7 | **gpt-oss** | A 稳定（6/6） | 2 | 2 | 1.04s | groq(2) | - |
| 8 | **qwq** | A 稳定（3/3） | 1 | 1 | 1.05s | cf(1) | - |
| 9 | **qwen3** | A 稳定（6/6） | 6 | 2 | 1.08s | amd(1), groq(1) | 余额不足，需充值(3), 模型不存在/不可用 (404/400)(1) |
| 10 | **qwen2-coder** | A 稳定（3/3） | 1 | 1 | 1.08s | cf(1) | - |
| 11 | **deepseek-r1** | A 稳定（3/3） | 2 | 1 | 1.17s | cf(1) | 余额不足，需充值(1) |
| 12 | **mistral-small-3** | A 稳定（3/3） | 1 | 1 | 1.18s | cf(1) | - |
| 13 | **sensenova-6-flash-lite** | A 稳定（3/3） | 2 | 1 | 2.04s | stkj(1) | 模型不存在/不可用 (404/400)(1) |
| 14 | **mimo-v2-flash** | A 稳定（6/6） | 4 | 2 | 4.05s | amd(1), thai(1) | 余额不足，需充值(1), 免费额度已用尽(1) |
| 15 | **mimo-v2** | A 稳定（3/3） | 2 | 1 | 14.57s | thai(1) | 余额不足，需充值(1) |
| 16 | **minicpm-vision-4** | B 基本稳定（6/7） | 2 | 2 | 0.29s | amd(2) | - |
| 17 | **glm-5** | B 基本稳定（2/3） | 6 | 1 | 0.87s | stkj(1) | 余额不足，需充值(4), 模型已下线 (410)(1) |
| 18 | **llama-3** | B 基本稳定（13/15） | 5 | 5 | 1.01s | cf(5) | - |
| 19 | **deepseek-v4-flash** | B 基本稳定（11/16） | 13 | 5 | 4.00s | amd(2), thai(2), stkj(1) | 余额不足，需充值(4), 模型已下线 (410)(2), 模型不存在/不可用 (404/400)(1) |
| 20 | **deepseek-flash** | B 基本稳定（3/4） | 1 | 1 | 17.72s | stkj(1) | - |
| 21 | **gemma-4** | B 基本稳定（2/3） | 1 | 1 | 20.04s | amd(1) | - |

### 4.3 全员不可用的系列

| 系列 | 成员 | 主要原因 | 涉及提供商 |
|---|---|---|---|
| **gpt-6** | 10 | 余额不足，需充值(10) | thai(10) |
| **gpt-5** | 7 | 余额不足，需充值(6), 上游返回空内容(1) | thai(6), cu(1) |
| **claude-4-opus** | 5 | 上游返回空内容(4), 余额不足，需充值(1) | cu(4), thai(1) |
| **muse-spark-1** | 5 | 余额不足，需充值(3), 模型不存在/不可用 (404/400)(2) | thai(3), oc(2) |
| **claude-4-sonnet** | 4 | 上游返回空内容(3), 余额不足，需充值(1) | cu(3), thai(1) |
| **claude-opus-5** | 4 | 余额不足，需充值(4) | thai(4) |
| **gemini-3-flash** | 4 | 余额不足，需充值(3), 上游返回空内容(1) | thai(3), cu(1) |
| **grok-4** | 3 | 余额不足，需充值(3) | thai(3) |
| **kimi-k2** | 3 | 上游返回空内容(1), 模型不存在/不可用 (404/400)(1), 余额不足，需充值(1) | cu(1), nvidia(1), thai(1) |
| **kimi-k3** | 3 | 余额不足，需充值(2), 限流 (429)(1) | thai(2), stkj(1) |
| **claude-fable-5** | 2 | 余额不足，需充值(2) | thai(2) |
| **claude-sonnet-5** | 2 | 余额不足，需充值(2) | thai(2) |
| **deepseek-v3** | 2 | 非聊天模型(1), 余额不足，需充值(1) | amd(1), thai(1) |
| **deepseek-v4-pro** | 2 | 模型已下线 (410)(1), 余额不足，需充值(1) | nvidia(1), thai(1) |
| **glm-5-flash** | 2 | 余额不足，需充值(1), 限流 (429)(1) | amd(1), thai(1) |
| **gpt-5-codex** | 2 | 上游返回空内容(2) | cu(2) |
| **laguna-2** | 2 | 模型不存在/不可用 (404/400)(2) | ps(2) |
| **mimo-v2-pro** | 2 | 余额不足，需充值(2) | thai(2) |
| **qwen3-max** | 2 | 余额不足，需充值(2) | thai(2) |
| **big-pickle** | 1 | 免费额度已用尽(1) | oc(1) |
| **cat-jam** | 1 | 模型不存在/不可用 (404/400)(1) | oc(1) |
| **claude-4-haiku** | 1 | 上游返回空内容(1) | cu(1) |
| **claude-haiku-5** | 1 | 余额不足，需充值(1) | thai(1) |
| **clef** | 1 | 非聊天模型(1) | amd(1) |
| **clef-flash** | 1 | 模型不存在/不可用 (404/400)(1) | cf(1) |
| **command-a** | 1 | 接口不支持 (405)(1) | cohere(1) |
| **command-r** | 1 | 接口不支持 (405)(1) | cohere(1) |
| **command-r-plus** | 1 | 接口不支持 (405)(1) | cohere(1) |
| **compound-mini** | 1 | 模型不存在/不可用 (404/400)(1) | groq(1) |
| **d3-flash** | 1 | 非聊天模型(1) | amd(1) |
| **deepseek-v4-flash-vision** | 1 | 模型不存在/不可用 (404/400)(1) | amd(1) |
| **default** | 1 | 上游返回空内容(1) | cu(1) |
| **gemini-4** | 1 | 余额不足，需充值(1) | thai(1) |
| **glm-5-flashx** | 1 | 余额不足，需充值(1) | thai(1) |
| **jev-1** | 1 | 免费额度已用尽(1) | oc(1) |
| **kimi-k3-pro** | 1 | 余额不足，需充值(1) | thai(1) |
| **ling-3-flash** | 1 | 余额不足，需充值(1) | thai(1) |
| **longcat-2** | 1 | 模型不存在/不可用 (404/400)(1) | oc(1) |
| **mineru2-pro** | 1 | 非聊天模型(1) | amd(1) |
| **minimax-m2** | 1 | 模型已下线 (410)(1) | nvidia(1) |
| **minimax-m3** | 1 | 模型已下线 (410)(1) | nvidia(1) |
| **minimax-m3-flash** | 1 | 余额不足，需充值(1) | thai(1) |
| **mistral-large** | 1 | 余额不足，需充值(1) | thai(1) |
| **mistral-large-4** | 1 | 余额不足，需充值(1) | thai(1) |
| **muse-wonder-1** | 1 | 余额不足，需充值(1) | thai(1) |
| **nemotron-3-lightning** | 1 | 免费额度已用尽(1) | oc(1) |
| **nemotron-3-ultra** | 1 | 模型不存在/不可用 (404/400)(1) | nvidia(1) |
| **omnivoice** | 1 | 非聊天模型(1) | amd(1) |
| **orpheus-arabic-saudi** | 1 | 需先接受模型条款(1) | groq(1) |
| **orpheus-v1** | 1 | 需先接受模型条款(1) | groq(1) |
| **parakeet-ctc-asr** | 1 | 模型不存在/不可用 (404/400)(1) | nvidia(1) |
| **sensenova-u1-lite** | 1 | 模型不存在/不可用 (404/400)(1) | stkj(1) |
| **space-bunny** | 1 | 模型不存在/不可用 (404/400)(1) | oc(1) |
| **step-5** | 1 | 模型不存在/不可用 (404/400)(1) | oc(1) |
| **union** | 1 | 模型不存在/不可用 (404/400)(1) | oc(1) |
| **whisper-large-v3** | 1 | 非聊天模型(1) | groq(1) |
| **whisper-large-v3-turbo** | 1 | 非聊天模型(1) | groq(1) |

### 4.4 全部可用模型排名（可用性高→低，快→慢）

| 排名 | 模型 ID | 提供商 | 系列 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 波动（最大/最小） | 实际返回模型名 | 备注 |
|---|---|---|---|---|---|---|---|---|---|---|
| 1 | `amd/MiniCPM-V-4.7-1B` | amd | minicpm-vision-4 | A 稳定（3/3） | 0.21s | 0.21s | 0.29s | 1.4× | self-dploy/MiniCPM-V-4.7-1B |  |
| 2 | `cf/@cf/meta/llama-3.2-1b-instruct` | cf | llama-3 | A 稳定（3/3） | 0.41s | 0.40s | 0.42s | 1.1× | @cf/meta/llama-3.2-1b-instruct |  |
| 3 | `groq/allam-2-7b` | groq | allam-2 | A 稳定（3/3） | 0.43s | 0.42s | 1.12s | 2.7× | allam-2-7b |  |
| 4 | `amd/MiniCPM5-2B` | amd | minicpm5 | A 稳定（3/3） | 0.62s | 0.16s | 8.70s | 54.7× | self-dploy/MiniCPM5-2B |  |
| 5 | `groq/qwen/qwen3.8-27b` | groq | qwen3 | A 稳定（3/3） | 0.85s | 0.24s | 0.93s | 3.9× | qwen/qwen3.8-27b |  |
| 6 | `groq/meta-llama/llama-prompt-guard-2-86m` | groq | llama-prompt-guard-2 | A 稳定（3/3） | 0.89s | 0.26s | 1.00s | 3.8× | meta-llama/llama-prompt-guard-2-86m |  |
| 7 | `groq/meta-llama/llama-prompt-guard-2-22m` | groq | llama-prompt-guard-2 | A 稳定（3/3） | 0.94s | 0.89s | 0.99s | 1.1× | meta-llama/llama-prompt-guard-2-22m |  |
| 8 | `cf/@cf/meta/llama-3.3-70b-instruct-fp8-fast` | cf | llama-3 | A 稳定（3/3） | 0.98s | 0.34s | 1.01s | 3.0× | @cf/meta/llama-3.3-70b-instruct-fp8-fast |  |
| 9 | `groq/openai/gpt-oss-safeguard-20b` | groq | gpt-oss-guard | A 稳定（3/3） | 0.99s | 0.95s | 1.03s | 1.1× | openai/gpt-oss-safeguard-20b |  |
| 10 | `cf/@cf/zai-org/glm-4.7-flash` | cf | glm-4-flash | A 稳定（3/3） | 1.00s | 0.76s | 1.12s | 1.5× | @cf/zai-org/glm-4.7-flash |  |
| 11 | `cf/@cf/meta/llama-3.1-8b-instruct-fp8-fast` | cf | llama-3 | A 稳定（3/3） | 1.01s | 0.98s | 1.29s | 1.3× | @cf/meta/llama-3.1-8b-instruct-fp8-fast |  |
| 12 | `amd/Qwen3.8-Flash-Next` | amd | qwen3-flash | A 稳定（3/3） | 1.02s | 0.85s | 1.43s | 1.7× | self-dploy/Qwen3.8-Flash-Next |  |
| 13 | `groq/openai/gpt-oss-120b` | groq | gpt-oss | A 稳定（3/3） | 1.03s | 1.00s | 1.14s | 1.1× | openai/gpt-oss-120b |  |
| 14 | `cf/@cf/meta/llama-3.2-3b-instruct` | cf | llama-3 | A 稳定（3/3） | 1.04s | 0.95s | 1.06s | 1.1× | @cf/meta/llama-3.2-3b-instruct |  |
| 15 | `groq/openai/gpt-oss-20b` | groq | gpt-oss | A 稳定（3/3） | 1.05s | 0.39s | 1.23s | 3.1× | openai/gpt-oss-20b |  |
| 16 | `cf/@cf/qwen/qwq-32b` | cf | qwq | A 稳定（3/3） | 1.05s | 0.95s | 1.11s | 1.2× | @cf/qwen/qwq-32b |  |
| 17 | `cf/@cf/qwen/qwen2.5-coder-32b-instruct` | cf | qwen2-coder | A 稳定（3/3） | 1.08s | 0.43s | 1.81s | 4.2× | @cf/qwen/qwen2.5-coder-32b-instruct |  |
| 18 | `cf/@cf/deepseek-ai/deepseek-r1-distill-qwen-32b` | cf | deepseek-r1 | A 稳定（3/3） | 1.17s | 1.08s | 1.24s | 1.1× | @cf/deepseek-ai/deepseek-r1-distill-qwe… |  |
| 19 | `cf/@cf/mistralai/mistral-small-3.1-24b-instruct` | cf | mistral-small-3 | A 稳定（3/3） | 1.18s | 0.97s | 1.34s | 1.4× | @cf/mistralai/mistral-small-3.1-24b-ins… |  |
| 20 | `amd/Qwen3.8-27B` | amd | qwen3 | A 稳定（3/3） | 1.30s | 0.68s | 2.34s | 3.5× | self-dploy/Qwen3.8-27B |  |
| 21 | `amd/MiMo-V2.6-Flash` | amd | mimo-v2-flash | A 稳定（3/3） | 1.63s | 0.63s | 2.19s | 3.5× | self-dploy/MiMo-V2.6-Flash |  |
| 22 | `stkj/sensenova-6.8-flash-lite` | stkj | sensenova-6-flash-lite | A 稳定（3/3） | 2.04s | 1.53s | 2.41s | 1.6× | sensenova-6.8-flash-lite |  |
| 23 | `thai/deepseek-v4.1-flash:free` | thai | deepseek-v4-flash | A 稳定（3/3） | 3.64s | 0.56s | 4.35s | 7.7× | deepseek-v4.1-flash |  |
| 24 | `thai/deepseek-v4-flash:free` | thai | deepseek-v4-flash | A 稳定（3/3） | 4.00s | 1.58s | 4.09s | 2.6× | deepseek-v4-flash |  |
| 25 | `amd/DeepSeek-V4-Flash` | amd | deepseek-v4-flash | A 稳定（3/3） | 6.04s | 1.02s | 15.22s | 15.0× | self-dploy/DeepSeek-V4-Flash |  |
| 26 | `thai/mimo-v2.6-flash:free` | thai | mimo-v2-flash | A 稳定（3/3） | 6.47s | 0.47s | 10.76s | 23.1× | mimo-v2.6-flash |  |
| 27 | `thai/mimo-v2.5:free` | thai | mimo-v2 | A 稳定（3/3） | 14.57s | 1.54s | 41.49s | 27.0× | mimo-v2.5 |  |
| 28 | `amd/MiniCPM-V-4.7-35B-A3B` | amd | minicpm-vision-4 | B 基本稳定（3/4） | 0.37s | 0.29s | 0.43s | 1.5× | self-dploy/MiniCPM-V-4.7-35B-A3B | 曾失败 限流 (429)×1，复测成功 |
| 29 | `stkj/glm-5.2` | stkj | glm-5 | B 基本稳定（2/3） | 0.87s | 0.80s | 0.93s | 1.2× | glm-5.2 | 曾失败 限流 (429)×1，复测成功 |
| 30 | `stkj/deepseek-flash` | stkj | deepseek-flash | B 基本稳定（3/4） | 17.72s | 13.69s | 21.05s | 1.5× | deepseek-flash | 曾失败 限流 (429)×1，复测成功 |
| 31 | `amd/Gemma-4-31B-IT` | amd | gemma-4 | B 基本稳定（2/3） | 20.04s | 2.29s | 37.79s | 16.5× | self-dploy/Gemma-4-31B-IT | 曾失败 超时×1，复测成功 |
| 32 | `cf/@cf/meta/llama-3.1-70b-instruct-fp8-fast` | cf | llama-3 | C 不稳定（1/3） | 1.25s | 1.25s | 1.25s | 1.0× | @cf/meta/llama-3.1-70b-instruct-fp8-fast | 曾失败 超时×2，复测成功 |
| 33 | `stkj/deepseek-v4-flash` | stkj | deepseek-v4-flash | C 不稳定（1/3） | 2.40s | 2.40s | 2.40s | 1.0× | deepseek-v4-flash | 曾失败 限流 (429)×2，复测成功 |
| 34 | `amd/DeepSeek-V4.1-Flash` | amd | deepseek-v4-flash | C 不稳定（1/4） | 11.59s | 11.59s | 11.59s | 1.0× | self-dploy/DeepSeek-V4.1-Flash | 曾失败 限流 (429)×3，复测成功 |

## 5. 不可用模型详细清单

### 5.1 错误类别说明

| 类别 | 含义 | 处理建议 |
|---|---|---|
| **余额不足，需充值** | 上游账号余额为 0、积分不足或预扣费失败，拒绝了计费请求。属于账号级问题，与具体模型无关。 | 充值；免费额度类账号可更换账号。 |
| **免费额度已用尽** | 报错表明账号级的免费请求额度 / 免费模型配额 / 赠送 token 已用完（通常为 429），需等待窗口重置，而非瞬时限流。 | 等待额度窗口重置、充值或更换账号。 |
| **需先接受模型条款** | 提供商要求组织管理员先在控制台接受该模型的使用条款。 | 到提供商控制台接受条款后即可使用。 |
| **限流 (429)** | 上游瞬时 RPM/TPM 限制、按分钟计数的免费层请求上限，或聚合平台“所有渠道繁忙”；放慢并复测后仍失败。 | 降低调用频率、稍后重试。 |
| **模型不存在/不可用 (404/400)** | 上游不认识该模型名、账号无权访问、模型已停用或暂时禁用；也包含网关本地直接判定“不支持”的模型。属于模型级问题。 | 从列表移除，或改用上游当前有效的模型名。 |
| **模型已下线 (410)** | 上游明确告知该模型已停用/到期，多数给出了替代模型。 | 按报错提示改用替代模型。 |
| **非聊天模型** | 上游明确说明该模型不支持 /v1/chat/completions（图像生成、OCR、语音、决策等专用端点）。 | 改用对应的专用端点。 |
| **接口不支持 (405)** | 网关的该提供商适配器不支持 chat completions 调用方式。 | 改用该提供商原生接口，或等待网关适配。 |
| **上游返回空内容** | 上游连接成功但没有返回任何内容（复测仍如此）。 | 更换模型或稍后重试。 |

说明：报错原文为网关转发的上游响应（已去掉网关的 `[连接/模型] [状态码]:` 包装）。“HTTP” 列为 `端点返回 / 上游返回`。

### 5.2 余额不足，需充值（71）

#### thai（71）

| 模型 ID | HTTP | 耗时 | 报错原文 |
|---|---|---|---|
| `thai/claude-fable-5` | 503/402 | 1.26s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 1s) |
| `thai/claude-fable-5.1` | 503/402 | 2.34s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 11s) |
| `thai/claude-haiku-5.5` | 503/402 | 0.50s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 24s) |
| `thai/claude-opus-4.8` | 503/402 | 1.37s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 1m 2s) |
| `thai/claude-opus-5` | 503/402 | 1.09s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 42s) |
| `thai/claude-opus-5-fast` | 503/402 | 1.08s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 16s) |
| `thai/claude-opus-5.5` | 503/402 | 0.56s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 11s) |
| `thai/claude-opus-5.5-fast` | 503/402 | 1.25s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 16s) |
| `thai/claude-sonnet-4.6` | 503/402 | 1.14s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 9s) |
| `thai/claude-sonnet-5` | 503/402 | 1.17s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 3s) |
| `thai/claude-sonnet-5.5` | 503/402 | 1.21s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 12s) |
| `thai/deepseek-r1-0528` | 503/402 | 1.03s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 49s) |
| `thai/deepseek-v3.2` | 503/402 | 0.39s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |
| `thai/deepseek-v4-flash` | 503/402 | 0.61s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 16s) |
| `thai/deepseek-v4-pro` | 503/402 | 0.37s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 6s) |
| `thai/deepseek-v4.1-flash` | 503/402 | 1.17s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |
| `thai/deepseek-v4.1-flash-1017` | 503/402 | 1.36s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 3s) |
| `thai/deepseek-v4.1-flash-1017-fast` | 503/402 | 0.49s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 6s) |
| `thai/gemini-3.6-flash` | 503/402 | 1.08s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |
| `thai/gemini-3.7-flash` | 503/402 | 0.65s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 59s) |
| `thai/gemini-3.8-flash` | 503/402 | 1.23s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 6s) |
| `thai/gemini-4-argon` | 503/402 | 0.98s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 28s) |
| `thai/glm-5.1` | 503/402 | 1.30s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 10s) |
| `thai/glm-5.2` | 503/402 | 1.09s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |
| `thai/glm-5.3` | 503/402 | 1.12s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 1m 41s) |
| `thai/glm-5.3-fast` | 503/402 | 1.10s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 5s) |
| `thai/glm-5.3-flash` | 503/402 | 0.44s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 11s) |
| `thai/glm-5.3-flashx` | 503/402 | 0.69s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |
| `thai/gpt-5.6-luna` | 503/402 | 0.35s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |
| `thai/gpt-5.6-luna-fast` | 503/402 | 0.53s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 16s) |
| `thai/gpt-5.6-sol` | 503/402 | 1.63s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 4s) |
| `thai/gpt-5.6-sol-fast` | 503/402 | 0.43s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 9s) |
| `thai/gpt-5.6-terra` | 503/402 | 1.05s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 27s) |
| `thai/gpt-5.6-terra-fast` | 503/402 | 0.99s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 9s) |
| `thai/gpt-6-astra` | 503/402 | 1.30s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 23s) |
| `thai/gpt-6-astra-fast` | 503/402 | 1.10s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |
| `thai/gpt-6-luna` | 503/402 | 1.47s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 15s) |
| `thai/gpt-6-luna-fast` | 503/402 | 1.16s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 7s) |
| `thai/gpt-6-sol` | 503/402 | 0.69s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 26s) |
| `thai/gpt-6-sol-fast` | 503/402 | 0.41s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 9s) |
| `thai/gpt-6.1-sol` | 503/402 | 2.60s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |
| `thai/gpt-6.1-sol-fast` | 503/402 | 2.47s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 12s) |
| `thai/gpt-6.1-terra` | 503/402 | 1.17s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 9s) |
| `thai/gpt-6.1-terra-fast` | 503/402 | 1.53s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 24s) |
| `thai/grok-4.5` | 503/402 | 1.15s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 1s) |
| `thai/grok-4.6` | 503/402 | 1.22s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 29s) |
| `thai/grok-4.7` | 503/402 | 1.12s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |
| `thai/kimi-k2.6` | 503/402 | 1.09s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 48s) |
| `thai/kimi-k3` | 503/402 | 0.63s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 18s) |
| `thai/kimi-k3-fast` | 503/402 | 1.30s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |
| `thai/kimi-k3.1-pro` | 503/402 | 0.48s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 11s) |
| `thai/ling-3.1-flash` | 503/402 | 1.06s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 28s) |
| `thai/mimo-v2.5` | 503/402 | 1.12s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 5s) |
| `thai/mimo-v2.5-pro` | 503/402 | 1.61s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2m) |
| `thai/mimo-v2.6-flash` | 503/402 | 1.18s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 4s) |
| `thai/mimo-v2.6-pro` | 503/402 | 2.25s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 20s) |
| `thai/minimax-m3.1-flash` | 503/402 | 1.10s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 4s) |
| `thai/mistral-large-2512` | 503/402 | 1.07s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 3s) |
| `thai/mistral-large-4` | 503/402 | 1.54s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 20s) |
| `thai/muse-spark-1-3` | 503/402 | 1.02s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 1m 41s) |
| `thai/muse-spark-1.1` | 503/402 | 1.08s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 33s) |
| `thai/muse-spark-1.2` | 503/402 | 1.11s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 1s) |
| `thai/muse-wonder-1.0` | 503/402 | 1.38s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 16s) |
| `thai/qwen3.5-27b` | 503/402 | 1.32s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 14s) |
| `thai/qwen3.6-27b` | 503/402 | 0.43s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 1m 59s) |
| `thai/qwen3.6-flash` | 503/402 | 2.27s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 15s) |
| `thai/qwen3.7-flash` | 503/402 | 1.22s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 12s) |
| `thai/qwen3.7-max` | 503/402 | 0.51s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 23s) |
| `thai/qwen3.8-27b` | 503/402 | 1.30s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 12s) |
| `thai/qwen3.8-flash` | 503/402 | 0.39s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 10s) |
| `thai/qwen3.8-max` | 503/402 | 1.05s | Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after 2s) |

### 5.3 免费额度已用尽（5）

#### oc（5）

全部 5 个模型报错相同：`Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: https://opencode.ai/go`

| 模型 ID | HTTP | 耗时 | 报错原文 |
|---|---|---|---|
| `oc/big-pickle` | 429 | 1.10s | Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: https://opencode.ai/go |
| `oc/deepseek-v4.1-flash-free` | 429 | 1.05s | Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: https://opencode.ai/go |
| `oc/jev-1.13-free` | 429 | 1.12s | Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: https://opencode.ai/go |
| `oc/mimo-v2.6-flash-free` | 429 | 1.12s | Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: https://opencode.ai/go |
| `oc/nemotron-3.5-lightning-free` | 429 | 1.23s | Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: https://opencode.ai/go |

### 5.4 需先接受模型条款（2）

#### groq（2）

| 模型 ID | HTTP | 耗时 | 报错原文 |
|---|---|---|---|
| `groq/canopylabs/orpheus-arabic-saudi` | 400 | 0.99s | The model `canopylabs/orpheus-arabic-saudi` requires terms acceptance. Please have the org admin accept the terms at https://console.groq.com/playground?model=… |
| `groq/canopylabs/orpheus-v1-english` | 400 | 0.85s | The model `canopylabs/orpheus-v1-english` requires terms acceptance. Please have the org admin accept the terms at https://console.groq.com/playground?model=ca… |

### 5.5 限流 (429)（2）

#### amd（1）

| 模型 ID | HTTP | 耗时 | 报错原文 |
|---|---|---|---|
| `amd/GLM-5.3-Flash` | 503/429 | 0.14s | {'error': {'message': "Model 'GLM-5.3-Flash' is at its concurrency limit (32); please retry later or use another model", 'type': 'rate_limit_error', 'code': 'm… |

#### stkj（1）

| 模型 ID | HTTP | 耗时 | 报错原文 |
|---|---|---|---|
| `stkj/kimi-k3` | 503/429 | 0.19s | inference exceeds tpm/rpm limit (reset after 8s) |

### 5.6 模型不存在/不可用 (404/400)（19）

#### oc（7）

全部 7 个模型报错相同：`Upstream request failed: Model is unavailable.`

| 模型 ID | HTTP | 耗时 | 报错原文 |
|---|---|---|---|
| `oc/cat-jam-free` | 400 | 0.97s | Upstream request failed: Model is unavailable. |
| `oc/longcat-2.5-preview-free` | 400 | 1.03s | Upstream request failed: Model is unavailable. |
| `oc/muse-spark-1.2-contributor-free` | 400 | 0.95s | Upstream request failed: Model is unavailable. |
| `oc/muse-spark-1.3-contributor-free` | 400 | 0.87s | Upstream request failed: Model is unavailable. |
| `oc/space-bunny-free` | 400 | 0.89s | Upstream request failed: Model is unavailable. |
| `oc/step-5-preview-free` | 400 | 0.89s | Upstream request failed: Model is unavailable. |
| `oc/union-alpha` | 400 | 0.99s | Upstream request failed: Model is unavailable. |

#### nvidia（3）

| 模型 ID | HTTP | 耗时 | 报错原文 |
|---|---|---|---|
| `nvidia/moonshotai/kimi-k2.6` | 503/404 | 1.01s | Function '23d4f03a-b8a6-4adb-a183-7daa083a09cc': Not found for account 'W2B7Rw3HvwisCg4qUJPJzj1U1KQv4lpkBSYeuVEg4UM' (reset after 2m) |
| `nvidia/nemotron-3-ultra-550b-a55b` | 503/404 | 1.02s | 404 page not found (reset after 1m 25s) |
| `nvidia/parakeet-ctc-1.1b-asr` | 503/404 | 0.88s | 404 page not found (reset after 2m) |

#### stkj（3）

| 模型 ID | HTTP | 耗时 | 报错原文 |
|---|---|---|---|
| `stkj/deepseek-v4.1-flash` | 503/403 | 3.14s | model is not available in the current token plan (reset after 2m) |
| `stkj/sensenova-6.7-flash-lite` | 503/404 | 0.10s | model is not found (reset after 1m 18s) |
| `stkj/sensenova-u1.5-lite` | 503/404 | 0.13s | model is not found (reset after 2m) |

#### groq（2）

| 模型 ID | HTTP | 耗时 | 报错原文 |
|---|---|---|---|
| `groq/groq/compound-mini` | 503/404 | 1.00s | The model `groq/compound-mini` does not exist or you do not have access to it. (reset after 2m) |
| `groq/qwen/qwen3.6-27b` | 503/404 | 0.81s | The model `qwen/qwen3.6-27b` does not exist or you do not have access to it. (reset after 2m) |

#### ps（2）

| 模型 ID | HTTP | 耗时 | 报错原文 |
|---|---|---|---|
| `ps/laguna-s-2.1` | 503/404 | 1.24s | please check the model you provided (reset after 2m) |
| `ps/laguna-xs-2.1` | 503/404 | 1.06s | please check the model you provided (reset after 1m 25s) |

#### amd（1）

| 模型 ID | HTTP | 耗时 | 报错原文 |
|---|---|---|---|
| `amd/DeepSeek-V4-Flash-Vision-Exp` | 400 | 0.14s | Requested model DeepSeek-V4-Flash-Vision-Exp not supported |

#### cf（1）

| 模型 ID | HTTP | 耗时 | 报错原文 |
|---|---|---|---|
| `cf/@cf/cloudflare/clef-flash` | 400 | 1.60s | {"errors":[{"message":"AiError: No such model: No such model clef-flash or task (8f4d7a58-6cf3-4253-933b-39a8d38270ef)","code":5007}],"success":false,"result":… |

### 5.7 模型已下线 (410)（6）

#### nvidia（6）

| 模型 ID | HTTP | 耗时 | 报错原文 |
|---|---|---|---|
| `nvidia/deepseek-ai/deepseek-v4-flash` | 410 | 0.88s | The model 'deepseek-ai/deepseek-v4-flash' has reached its end of life on 2026-08-07T09:00:00Z and is no longer available. |
| `nvidia/deepseek-ai/deepseek-v4-flash-0731` | 410 | 0.86s | The model 'deepseek-ai/deepseek-v4-flash-0731' has reached its end of life on 2026-09-21T08:00:00Z and is no longer available. |
| `nvidia/deepseek-ai/deepseek-v4-pro` | 410 | 0.98s | The model 'deepseek-ai/deepseek-v4-pro' has reached its end of life on 2026-08-07T09:00:00Z and is no longer available. |
| `nvidia/minimaxai/minimax-m2.7` | 410 | 0.19s | The model 'minimaxai/minimax-m2.7' has reached its end of life on 2026-07-27T00:00:00Z and is no longer available. |
| `nvidia/minimaxai/minimax-m3` | 410 | 0.96s | The model 'minimaxai/minimax-m3' has reached its end of life on 2026-09-09T09:00:00Z and is no longer available. |
| `nvidia/z-ai/glm-5.2` | 410 | 0.85s | The model 'z-ai/glm-5.2' has reached its end of life on 2026-08-21T09:00:00Z and is no longer available. |

### 5.8 非聊天模型（7）

#### amd（5）

| 模型 ID | HTTP | 耗时 | 报错原文 |
|---|---|---|---|
| `amd/Clef` | 400 | 0.14s | Model Clef is not a chat model; call /v1/systemone instead of /v1/chat/completions |
| `amd/d3` | 400 | 0.15s | Model d3 is not a chat model; call /v1/systemone instead of /v1/chat/completions |
| `amd/d3-flash` | 400 | 0.13s | Model d3-flash is not a chat model; call /v1/systemone instead of /v1/chat/completions |
| `amd/MinerU2.5-Pro` | 400 | 0.17s | Model MinerU2.5-Pro is not a chat model; call /v1/ocr instead of /v1/chat/completions |
| `amd/OmniVoice` | 400 | 0.17s | Model OmniVoice is not a chat model; call /v1/audio/speech instead of /v1/chat/completions |

#### groq（2）

| 模型 ID | HTTP | 耗时 | 报错原文 |
|---|---|---|---|
| `groq/whisper-large-v3` | 400 | 0.92s | The model `whisper-large-v3` does not support chat completions |
| `groq/whisper-large-v3-turbo` | 400 | 1.01s | The model `whisper-large-v3-turbo` does not support chat completions |

### 5.9 接口不支持 (405)（3）

#### cohere（3）

全部 3 个模型报错相同：`HTTP 405`

| 模型 ID | HTTP | 耗时 | 报错原文 |
|---|---|---|---|
| `cohere/command-a-03-2025` | 405 | 0.99s | HTTP 405 |
| `cohere/command-r-08-2024` | 405 | 0.87s | HTTP 405 |
| `cohere/command-r-plus-08-2024` | 405 | 2.05s | HTTP 405 |

### 5.10 上游返回空内容（14）

#### cu（14）

全部 14 个模型报错相同：`Cursor AgentService returned an empty turn`

| 模型 ID | HTTP | 耗时 | 报错原文 |
|---|---|---|---|
| `cu/claude-4.5-haiku` | 400 | 0.95s | Cursor AgentService returned an empty turn |
| `cu/claude-4.5-opus` | 400 | 1.07s | Cursor AgentService returned an empty turn |
| `cu/claude-4.5-opus-high` | 400 | 1.06s | Cursor AgentService returned an empty turn |
| `cu/claude-4.5-opus-high-thinking` | 400 | 0.96s | Cursor AgentService returned an empty turn |
| `cu/claude-4.5-sonnet` | 400 | 1.07s | Cursor AgentService returned an empty turn |
| `cu/claude-4.5-sonnet-thinking` | 400 | 1.28s | Cursor AgentService returned an empty turn |
| `cu/claude-4.6-opus-max` | 400 | 1.14s | Cursor AgentService returned an empty turn |
| `cu/claude-4.6-sonnet-medium-thinking` | 400 | 1.05s | Cursor AgentService returned an empty turn |
| `cu/default` | 400 | 2.04s | Cursor AgentService returned an empty turn |
| `cu/gemini-3-flash-preview` | 400 | 1.01s | Cursor AgentService returned an empty turn |
| `cu/gpt-5.2` | 400 | 1.02s | Cursor AgentService returned an empty turn |
| `cu/gpt-5.2-codex` | 400 | 1.07s | Cursor AgentService returned an empty turn |
| `cu/gpt-5.3-codex` | 400 | 0.96s | Cursor AgentService returned an empty turn |
| `cu/kimi-k2.5` | 400 | 0.95s | Cursor AgentService returned an empty turn |

## 6. 模型系列明细（含每个模型的提供商与测速结果）

顺序与第 4.1 节综合排名一致：先有可用成员的系列，再全员不可用的系列。“备注”给出失败原因与报错摘要，或成功时曾失败后复测成功的情况。

### 1. qwen3

归并键 `qwen||3` ｜ 成员 6，可用 2 ｜ 提供商：thai(3), groq(2), amd(1) ｜ 综合分 84.4（稳定 100 / 广度 67 / 速度 81） ｜ 可用提供商 2 个，其中稳定 2 个 ｜ 稳定性 A 稳定（6/6） ｜ 延迟 最快 0.24s / 中位 1.08s / 最慢 2.34s

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `groq/qwen/qwen3.8-27b` | groq | ✅ 可用 | A 稳定（3/3） | 0.85s | 0.24s | 0.93s | qwen/qwen3.8-27b | - |  |
| `amd/Qwen3.8-27B` | amd | ✅ 可用 | A 稳定（3/3） | 1.30s | 0.68s | 2.34s | self-dploy/Qwen3.8-27B | - |  |
| `groq/qwen/qwen3.6-27b` | groq | ❌ 模型不存在/不可用 (404/400) | 0/1 | - | - | - | - | - | HTTP 503/404：The model `qwen/qwen3.6-27b` does not exist or you do not have access to it. (reset after 2m) |
| `thai/qwen3.5-27b` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/qwen3.6-27b` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/qwen3.8-27b` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 2. allam-2

归并键 `allam||2` ｜ 成员 1，可用 1 ｜ 提供商：groq(1) ｜ 综合分 80.0（稳定 100 / 广度 33 / 速度 100） ｜ 可用提供商 1 个，其中稳定 1 个 ｜ 稳定性 A 稳定（3/3） ｜ 延迟 最快 0.42s / 中位 0.43s / 最慢 1.12s

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `groq/allam-2-7b` | groq | ✅ 可用 | A 稳定（3/3） | 0.43s | 0.42s | 1.12s | allam-2-7b | - |  |

### 3. minicpm5

归并键 `minicpm||5` ｜ 成员 1，可用 1 ｜ 提供商：amd(1) ｜ 综合分 78.4（稳定 100 / 广度 33 / 速度 95） ｜ 可用提供商 1 个，其中稳定 1 个 ｜ 稳定性 A 稳定（3/3） ｜ 延迟 最快 0.16s / 中位 0.62s / 最慢 8.70s

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `amd/MiniCPM5-2B` | amd | ✅ 可用 | A 稳定（3/3） | 0.62s | 0.16s | 8.70s | self-dploy/MiniCPM5-2B | - |  |

### 4. llama-prompt-guard-2

归并键 `llama-prompt|guard|2` ｜ 成员 2，可用 2 ｜ 提供商：groq(2) ｜ 综合分 75.6（稳定 100 / 广度 33 / 速度 85） ｜ 可用提供商 1 个，其中稳定 1 个 ｜ 稳定性 A 稳定（6/6） ｜ 延迟 最快 0.26s / 中位 0.91s / 最慢 1.00s

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `groq/meta-llama/llama-prompt-guard-2-86m` | groq | ✅ 可用 | A 稳定（3/3） | 0.89s | 0.26s | 1.00s | meta-llama/llama-prompt-guard-2-86m | - |  |
| `groq/meta-llama/llama-prompt-guard-2-22m` | groq | ✅ 可用 | A 稳定（3/3） | 0.94s | 0.89s | 0.99s | meta-llama/llama-prompt-guard-2-22m | - |  |

### 5. gpt-oss-guard

归并键 `gpt-oss|guard|` ｜ 成员 1，可用 1 ｜ 提供商：groq(1) ｜ 综合分 75.0（稳定 100 / 广度 33 / 速度 83） ｜ 可用提供商 1 个，其中稳定 1 个 ｜ 稳定性 A 稳定（3/3） ｜ 延迟 最快 0.95s / 中位 0.99s / 最慢 1.03s

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `groq/openai/gpt-oss-safeguard-20b` | groq | ✅ 可用 | A 稳定（3/3） | 0.99s | 0.95s | 1.03s | openai/gpt-oss-safeguard-20b | - |  |

### 6. glm-4-flash

归并键 `glm|flash|4` ｜ 成员 1，可用 1 ｜ 提供商：cf(1) ｜ 综合分 74.9（稳定 100 / 广度 33 / 速度 83） ｜ 可用提供商 1 个，其中稳定 1 个 ｜ 稳定性 A 稳定（3/3） ｜ 延迟 最快 0.76s / 中位 1.00s / 最慢 1.12s

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `cf/@cf/zai-org/glm-4.7-flash` | cf | ✅ 可用 | A 稳定（3/3） | 1.00s | 0.76s | 1.12s | @cf/zai-org/glm-4.7-flash | - |  |

### 7. qwen3-flash

归并键 `qwen|flash|3` ｜ 成员 4，可用 1 ｜ 提供商：thai(3), amd(1) ｜ 综合分 74.7（稳定 100 / 广度 33 / 速度 82） ｜ 可用提供商 1 个，其中稳定 1 个 ｜ 稳定性 A 稳定（3/3） ｜ 延迟 最快 0.85s / 中位 1.02s / 最慢 1.43s

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `amd/Qwen3.8-Flash-Next` | amd | ✅ 可用 | A 稳定（3/3） | 1.02s | 0.85s | 1.43s | self-dploy/Qwen3.8-Flash-Next | - |  |
| `thai/qwen3.6-flash` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/qwen3.7-flash` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/qwen3.8-flash` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 8. mimo-v2-flash

归并键 `mimo|flash|2` ｜ 成员 4，可用 2 ｜ 提供商：thai(2), amd(1), oc(1) ｜ 综合分 74.7（稳定 100 / 广度 67 / 速度 49） ｜ 可用提供商 2 个，其中稳定 2 个 ｜ 稳定性 A 稳定（6/6） ｜ 延迟 最快 0.47s / 中位 4.05s / 最慢 10.76s

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `amd/MiMo-V2.6-Flash` | amd | ✅ 可用 | A 稳定（3/3） | 1.63s | 0.63s | 2.19s | self-dploy/MiMo-V2.6-Flash | - |  |
| `thai/mimo-v2.6-flash:free` | thai | ✅ 可用 | A 稳定（3/3） | 6.47s | 0.47s | 10.76s | mimo-v2.6-flash | - |  |
| `oc/mimo-v2.6-flash-free` | oc | ❌ 免费额度已用尽 | 0/2 | - | - | - | - | - | HTTP 429：Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: h… |
| `thai/mimo-v2.6-flash` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 9. gpt-oss

归并键 `gpt-oss||` ｜ 成员 2，可用 2 ｜ 提供商：groq(2) ｜ 综合分 74.6（稳定 100 / 广度 33 / 速度 82） ｜ 可用提供商 1 个，其中稳定 1 个 ｜ 稳定性 A 稳定（6/6） ｜ 延迟 最快 0.39s / 中位 1.04s / 最慢 1.23s

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `groq/openai/gpt-oss-120b` | groq | ✅ 可用 | A 稳定（3/3） | 1.03s | 1.00s | 1.14s | openai/gpt-oss-120b | - |  |
| `groq/openai/gpt-oss-20b` | groq | ✅ 可用 | A 稳定（3/3） | 1.05s | 0.39s | 1.23s | openai/gpt-oss-20b | - |  |

### 10. qwq

归并键 `qwq||` ｜ 成员 1，可用 1 ｜ 提供商：cf(1) ｜ 综合分 74.5（稳定 100 / 广度 33 / 速度 82） ｜ 可用提供商 1 个，其中稳定 1 个 ｜ 稳定性 A 稳定（3/3） ｜ 延迟 最快 0.95s / 中位 1.05s / 最慢 1.11s

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `cf/@cf/qwen/qwq-32b` | cf | ✅ 可用 | A 稳定（3/3） | 1.05s | 0.95s | 1.11s | @cf/qwen/qwq-32b | - |  |

### 11. qwen2-coder

归并键 `qwen|coder|2` ｜ 成员 1，可用 1 ｜ 提供商：cf(1) ｜ 综合分 74.4（稳定 100 / 广度 33 / 速度 81） ｜ 可用提供商 1 个，其中稳定 1 个 ｜ 稳定性 A 稳定（3/3） ｜ 延迟 最快 0.43s / 中位 1.08s / 最慢 1.81s

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `cf/@cf/qwen/qwen2.5-coder-32b-instruct` | cf | ✅ 可用 | A 稳定（3/3） | 1.08s | 0.43s | 1.81s | @cf/qwen/qwen2.5-coder-32b-instruct | - |  |

### 12. minicpm-vision-4

归并键 `minicpm|vision|4` ｜ 成员 2，可用 2 ｜ 提供商：amd(2) ｜ 综合分 74.3（稳定 86 / 广度 33 / 速度 100） ｜ 可用提供商 1 个，其中稳定 1 个 ｜ 稳定性 B 基本稳定（6/7） ｜ 延迟 最快 0.21s / 中位 0.29s / 最慢 0.43s

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `amd/MiniCPM-V-4.7-1B` | amd | ✅ 可用 | A 稳定（3/3） | 0.21s | 0.21s | 0.29s | self-dploy/MiniCPM-V-4.7-1B | - |  |
| `amd/MiniCPM-V-4.7-35B-A3B` | amd | ✅ 可用 | B 基本稳定（3/4） | 0.37s | 0.29s | 0.43s | self-dploy/MiniCPM-V-4.7-35B-A3B | - | 曾失败 限流 (429)×1，复测成功 |

### 13. deepseek-r1

归并键 `deepseek||1` ｜ 成员 2，可用 1 ｜ 提供商：cf(1), thai(1) ｜ 综合分 73.8（稳定 100 / 广度 33 / 速度 79） ｜ 可用提供商 1 个，其中稳定 1 个 ｜ 稳定性 A 稳定（3/3） ｜ 延迟 最快 1.08s / 中位 1.17s / 最慢 1.24s

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `cf/@cf/deepseek-ai/deepseek-r1-distill-qwen-32b` | cf | ✅ 可用 | A 稳定（3/3） | 1.17s | 1.08s | 1.24s | @cf/deepseek-ai/deepseek-r1-distill-qwe… | - |  |
| `thai/deepseek-r1-0528` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 14. mistral-small-3

归并键 `mistral|small|3` ｜ 成员 1，可用 1 ｜ 提供商：cf(1) ｜ 综合分 73.7（稳定 100 / 广度 33 / 速度 79） ｜ 可用提供商 1 个，其中稳定 1 个 ｜ 稳定性 A 稳定（3/3） ｜ 延迟 最快 0.97s / 中位 1.18s / 最慢 1.34s

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `cf/@cf/mistralai/mistral-small-3.1-24b-instruct` | cf | ✅ 可用 | A 稳定（3/3） | 1.18s | 0.97s | 1.34s | @cf/mistralai/mistral-small-3.1-24b-ins… | - |  |

### 15. deepseek-v4-flash

归并键 `deepseek|flash|4` ｜ 成员 13，可用 5 ｜ 提供商：thai(6), amd(2), nvidia(2), stkj(2), oc(1) ｜ 综合分 72.3（稳定 69 / 广度 100 / 速度 49） ｜ 可用提供商 3 个，其中稳定 2 个 ｜ 稳定性 B 基本稳定（11/16） ｜ 延迟 最快 0.56s / 中位 4.00s / 最慢 15.22s

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `thai/deepseek-v4.1-flash:free` | thai | ✅ 可用 | A 稳定（3/3） | 3.64s | 0.56s | 4.35s | deepseek-v4.1-flash | - |  |
| `thai/deepseek-v4-flash:free` | thai | ✅ 可用 | A 稳定（3/3） | 4.00s | 1.58s | 4.09s | deepseek-v4-flash | - |  |
| `amd/DeepSeek-V4-Flash` | amd | ✅ 可用 | A 稳定（3/3） | 6.04s | 1.02s | 15.22s | self-dploy/DeepSeek-V4-Flash | - |  |
| `stkj/deepseek-v4-flash` | stkj | ✅ 可用 | C 不稳定（1/3） | 2.40s | 2.40s | 2.40s | deepseek-v4-flash | - | 曾失败 限流 (429)×2，复测成功 |
| `amd/DeepSeek-V4.1-Flash` | amd | ✅ 可用 | C 不稳定（1/4） | 11.59s | 11.59s | 11.59s | self-dploy/DeepSeek-V4.1-Flash | - | 曾失败 限流 (429)×3，复测成功 |
| `nvidia/deepseek-ai/deepseek-v4-flash` | nvidia | ❌ 模型已下线 (410) | 0/1 | - | - | - | - | - | HTTP 410：The model 'deepseek-ai/deepseek-v4-flash' has reached its end of life on 2026-08-07T09:00:00Z and is no longer availabl… |
| `nvidia/deepseek-ai/deepseek-v4-flash-0731` | nvidia | ❌ 模型已下线 (410) | 0/1 | - | - | - | - | - | HTTP 410：The model 'deepseek-ai/deepseek-v4-flash-0731' has reached its end of life on 2026-09-21T08:00:00Z and is no longer ava… |
| `oc/deepseek-v4.1-flash-free` | oc | ❌ 免费额度已用尽 | 0/2 | - | - | - | - | - | HTTP 429：Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: h… |
| `stkj/deepseek-v4.1-flash` | stkj | ❌ 模型不存在/不可用 (404/400) | 0/1 | - | - | - | - | - | HTTP 503/403：model is not available in the current token plan (reset after 2m) |
| `thai/deepseek-v4-flash` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/deepseek-v4.1-flash` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/deepseek-v4.1-flash-1017` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/deepseek-v4.1-flash-1017-fast` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 16. sensenova-6-flash-lite

归并键 `sensenova|flash+lite|6` ｜ 成员 2，可用 1 ｜ 提供商：stkj(2) ｜ 综合分 69.7（稳定 100 / 广度 33 / 速度 66） ｜ 可用提供商 1 个，其中稳定 1 个 ｜ 稳定性 A 稳定（3/3） ｜ 延迟 最快 1.53s / 中位 2.04s / 最慢 2.41s

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `stkj/sensenova-6.8-flash-lite` | stkj | ✅ 可用 | A 稳定（3/3） | 2.04s | 1.53s | 2.41s | sensenova-6.8-flash-lite | - |  |
| `stkj/sensenova-6.7-flash-lite` | stkj | ❌ 模型不存在/不可用 (404/400) | 0/1 | - | - | - | - | - | HTTP 503/404：model is not found (reset after 1m 18s) |

### 17. llama-3

归并键 `llama||3` ｜ 成员 5，可用 5 ｜ 提供商：cf(5) ｜ 综合分 69.5（稳定 87 / 广度 33 / 速度 83） ｜ 可用提供商 1 个，其中稳定 1 个 ｜ 稳定性 B 基本稳定（13/15） ｜ 延迟 最快 0.34s / 中位 1.01s / 最慢 1.29s

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `cf/@cf/meta/llama-3.2-1b-instruct` | cf | ✅ 可用 | A 稳定（3/3） | 0.41s | 0.40s | 0.42s | @cf/meta/llama-3.2-1b-instruct | - |  |
| `cf/@cf/meta/llama-3.3-70b-instruct-fp8-fast` | cf | ✅ 可用 | A 稳定（3/3） | 0.98s | 0.34s | 1.01s | @cf/meta/llama-3.3-70b-instruct-fp8-fast | - |  |
| `cf/@cf/meta/llama-3.1-8b-instruct-fp8-fast` | cf | ✅ 可用 | A 稳定（3/3） | 1.01s | 0.98s | 1.29s | @cf/meta/llama-3.1-8b-instruct-fp8-fast | - |  |
| `cf/@cf/meta/llama-3.2-3b-instruct` | cf | ✅ 可用 | A 稳定（3/3） | 1.04s | 0.95s | 1.06s | @cf/meta/llama-3.2-3b-instruct | - |  |
| `cf/@cf/meta/llama-3.1-70b-instruct-fp8-fast` | cf | ✅ 可用 | C 不稳定（1/3） | 1.25s | 1.25s | 1.25s | @cf/meta/llama-3.1-70b-instruct-fp8-fast | - | 曾失败 超时×2，复测成功 |

### 18. glm-5

归并键 `glm||5` ｜ 成员 6，可用 1 ｜ 提供商：thai(4), nvidia(1), stkj(1) ｜ 综合分 62.6（稳定 67 / 广度 33 / 速度 87） ｜ 可用提供商 1 个，其中稳定 0 个 ｜ 稳定性 B 基本稳定（2/3） ｜ 延迟 最快 0.80s / 中位 0.87s / 最慢 0.93s

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `stkj/glm-5.2` | stkj | ✅ 可用 | B 基本稳定（2/3） | 0.87s | 0.80s | 0.93s | glm-5.2 | - | 曾失败 限流 (429)×1，复测成功 |
| `nvidia/z-ai/glm-5.2` | nvidia | ❌ 模型已下线 (410) | 0/1 | - | - | - | - | - | HTTP 410：The model 'z-ai/glm-5.2' has reached its end of life on 2026-08-21T09:00:00Z and is no longer available. |
| `thai/glm-5.1` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/glm-5.2` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/glm-5.3` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/glm-5.3-fast` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 19. mimo-v2

归并键 `mimo||2` ｜ 成员 2，可用 1 ｜ 提供商：thai(2) ｜ 综合分 55.3（稳定 100 / 广度 33 / 速度 18） ｜ 可用提供商 1 个，其中稳定 1 个 ｜ 稳定性 A 稳定（3/3） ｜ 延迟 最快 1.54s / 中位 14.57s / 最慢 41.49s

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `thai/mimo-v2.5:free` | thai | ✅ 可用 | A 稳定（3/3） | 14.57s | 1.54s | 41.49s | mimo-v2.5 | - |  |
| `thai/mimo-v2.5` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 20. deepseek-flash

归并键 `deepseek|flash|` ｜ 成员 1，可用 1 ｜ 提供商：stkj(1) ｜ 综合分 43.9（稳定 75 / 广度 33 / 速度 13） ｜ 可用提供商 1 个，其中稳定 0 个 ｜ 稳定性 B 基本稳定（3/4） ｜ 延迟 最快 13.69s / 中位 17.72s / 最慢 21.05s

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `stkj/deepseek-flash` | stkj | ✅ 可用 | B 基本稳定（3/4） | 17.72s | 13.69s | 21.05s | deepseek-flash | - | 曾失败 限流 (429)×1，复测成功 |

### 21. gemma-4

归并键 `gemma||4` ｜ 成员 1，可用 1 ｜ 提供商：amd(1) ｜ 综合分 39.6（稳定 67 / 广度 33 / 速度 10） ｜ 可用提供商 1 个，其中稳定 0 个 ｜ 稳定性 B 基本稳定（2/3） ｜ 延迟 最快 2.29s / 中位 20.04s / 最慢 37.79s

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `amd/Gemma-4-31B-IT` | amd | ✅ 可用 | B 基本稳定（2/3） | 20.04s | 2.29s | 37.79s | self-dploy/Gemma-4-31B-IT | - | 曾失败 超时×1，复测成功 |

### 22. gpt-6

归并键 `gpt||6` ｜ 成员 10，可用 0 ｜ 提供商：thai(10)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `thai/gpt-6-astra` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-6-astra-fast` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-6-luna` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-6-luna-fast` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-6-sol` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-6-sol-fast` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-6.1-sol` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-6.1-sol-fast` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-6.1-terra` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-6.1-terra-fast` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 23. gpt-5

归并键 `gpt||5` ｜ 成员 7，可用 0 ｜ 提供商：thai(6), cu(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `cu/gpt-5.2` | cu | ❌ 上游返回空内容 | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `thai/gpt-5.6-luna` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-5.6-luna-fast` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-5.6-sol` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-5.6-sol-fast` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-5.6-terra` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gpt-5.6-terra-fast` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 24. claude-4-opus

归并键 `claude|opus|4` ｜ 成员 5，可用 0 ｜ 提供商：cu(4), thai(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `cu/claude-4.5-opus` | cu | ❌ 上游返回空内容 | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `cu/claude-4.5-opus-high` | cu | ❌ 上游返回空内容 | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `cu/claude-4.5-opus-high-thinking` | cu | ❌ 上游返回空内容 | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `cu/claude-4.6-opus-max` | cu | ❌ 上游返回空内容 | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `thai/claude-opus-4.8` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 25. muse-spark-1

归并键 `muse-spark||1` ｜ 成员 5，可用 0 ｜ 提供商：thai(3), oc(2)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `oc/muse-spark-1.2-contributor-free` | oc | ❌ 模型不存在/不可用 (404/400) | 0/1 | - | - | - | - | - | HTTP 400：Upstream request failed: Model is unavailable. |
| `oc/muse-spark-1.3-contributor-free` | oc | ❌ 模型不存在/不可用 (404/400) | 0/1 | - | - | - | - | - | HTTP 400：Upstream request failed: Model is unavailable. |
| `thai/muse-spark-1-3` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/muse-spark-1.1` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/muse-spark-1.2` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 26. claude-4-sonnet

归并键 `claude|sonnet|4` ｜ 成员 4，可用 0 ｜ 提供商：cu(3), thai(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `cu/claude-4.5-sonnet` | cu | ❌ 上游返回空内容 | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `cu/claude-4.5-sonnet-thinking` | cu | ❌ 上游返回空内容 | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `cu/claude-4.6-sonnet-medium-thinking` | cu | ❌ 上游返回空内容 | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `thai/claude-sonnet-4.6` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 27. claude-opus-5

归并键 `claude|opus|5` ｜ 成员 4，可用 0 ｜ 提供商：thai(4)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `thai/claude-opus-5` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/claude-opus-5-fast` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/claude-opus-5.5` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/claude-opus-5.5-fast` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 28. gemini-3-flash

归并键 `gemini|flash|3` ｜ 成员 4，可用 0 ｜ 提供商：thai(3), cu(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `cu/gemini-3-flash-preview` | cu | ❌ 上游返回空内容 | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `thai/gemini-3.6-flash` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gemini-3.7-flash` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/gemini-3.8-flash` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 29. grok-4

归并键 `grok||4` ｜ 成员 3，可用 0 ｜ 提供商：thai(3)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `thai/grok-4.5` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/grok-4.6` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/grok-4.7` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 30. kimi-k2

归并键 `kimi||2` ｜ 成员 3，可用 0 ｜ 提供商：cu(1), nvidia(1), thai(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `cu/kimi-k2.5` | cu | ❌ 上游返回空内容 | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `nvidia/moonshotai/kimi-k2.6` | nvidia | ❌ 模型不存在/不可用 (404/400) | 0/1 | - | - | - | - | - | HTTP 503/404：Function '23d4f03a-b8a6-4adb-a183-7daa083a09cc': Not found for account 'W2B7Rw3HvwisCg4qUJPJzj1U1KQv4lpkBSYeuVEg4UM' (r… |
| `thai/kimi-k2.6` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 31. kimi-k3

归并键 `kimi||3` ｜ 成员 3，可用 0 ｜ 提供商：thai(2), stkj(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `stkj/kimi-k3` | stkj | ❌ 限流 (429) | 0/3 | - | - | - | - | - | HTTP 503/429：inference exceeds tpm/rpm limit (reset after 8s) |
| `thai/kimi-k3` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/kimi-k3-fast` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 32. claude-fable-5

归并键 `claude|fable|5` ｜ 成员 2，可用 0 ｜ 提供商：thai(2)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `thai/claude-fable-5` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/claude-fable-5.1` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 33. claude-sonnet-5

归并键 `claude|sonnet|5` ｜ 成员 2，可用 0 ｜ 提供商：thai(2)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `thai/claude-sonnet-5` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/claude-sonnet-5.5` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 34. deepseek-v3

归并键 `deepseek||3` ｜ 成员 2，可用 0 ｜ 提供商：amd(1), thai(1) ｜ 1 个成员通过别名学习/覆盖归入

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `amd/d3` | amd | ❌ 非聊天模型 | 0/1 | - | - | - | - | - | HTTP 400：Model d3 is not a chat model; call /v1/systemone instead of /v1/chat/completions |
| `thai/deepseek-v3.2` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 35. deepseek-v4-pro

归并键 `deepseek|pro|4` ｜ 成员 2，可用 0 ｜ 提供商：nvidia(1), thai(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `nvidia/deepseek-ai/deepseek-v4-pro` | nvidia | ❌ 模型已下线 (410) | 0/1 | - | - | - | - | - | HTTP 410：The model 'deepseek-ai/deepseek-v4-pro' has reached its end of life on 2026-08-07T09:00:00Z and is no longer available. |
| `thai/deepseek-v4-pro` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 36. glm-5-flash

归并键 `glm|flash|5` ｜ 成员 2，可用 0 ｜ 提供商：amd(1), thai(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `amd/GLM-5.3-Flash` | amd | ❌ 限流 (429) | 0/3 | - | - | - | - | - | HTTP 503/429：{'error': {'message': "Model 'GLM-5.3-Flash' is at its concurrency limit (32); please retry later or use another model"… |
| `thai/glm-5.3-flash` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 37. gpt-5-codex

归并键 `gpt|codex|5` ｜ 成员 2，可用 0 ｜ 提供商：cu(2)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `cu/gpt-5.2-codex` | cu | ❌ 上游返回空内容 | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |
| `cu/gpt-5.3-codex` | cu | ❌ 上游返回空内容 | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |

### 38. laguna-2

归并键 `laguna||2` ｜ 成员 2，可用 0 ｜ 提供商：ps(2)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `ps/laguna-s-2.1` | ps | ❌ 模型不存在/不可用 (404/400) | 0/1 | - | - | - | - | - | HTTP 503/404：please check the model you provided (reset after 2m) |
| `ps/laguna-xs-2.1` | ps | ❌ 模型不存在/不可用 (404/400) | 0/1 | - | - | - | - | - | HTTP 503/404：please check the model you provided (reset after 1m 25s) |

### 39. mimo-v2-pro

归并键 `mimo|pro|2` ｜ 成员 2，可用 0 ｜ 提供商：thai(2)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `thai/mimo-v2.5-pro` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/mimo-v2.6-pro` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 40. qwen3-max

归并键 `qwen|max|3` ｜ 成员 2，可用 0 ｜ 提供商：thai(2)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `thai/qwen3.7-max` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |
| `thai/qwen3.8-max` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 41. big-pickle

归并键 `big-pickle||` ｜ 成员 1，可用 0 ｜ 提供商：oc(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `oc/big-pickle` | oc | ❌ 免费额度已用尽 | 0/2 | - | - | - | - | - | HTTP 429：Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: h… |

### 42. cat-jam

归并键 `cat-jam||` ｜ 成员 1，可用 0 ｜ 提供商：oc(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `oc/cat-jam-free` | oc | ❌ 模型不存在/不可用 (404/400) | 0/1 | - | - | - | - | - | HTTP 400：Upstream request failed: Model is unavailable. |

### 43. claude-4-haiku

归并键 `claude|haiku|4` ｜ 成员 1，可用 0 ｜ 提供商：cu(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `cu/claude-4.5-haiku` | cu | ❌ 上游返回空内容 | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |

### 44. claude-haiku-5

归并键 `claude|haiku|5` ｜ 成员 1，可用 0 ｜ 提供商：thai(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `thai/claude-haiku-5.5` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 45. clef

归并键 `clef||` ｜ 成员 1，可用 0 ｜ 提供商：amd(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `amd/Clef` | amd | ❌ 非聊天模型 | 0/1 | - | - | - | - | - | HTTP 400：Model Clef is not a chat model; call /v1/systemone instead of /v1/chat/completions |

### 46. clef-flash

归并键 `clef|flash|` ｜ 成员 1，可用 0 ｜ 提供商：cf(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `cf/@cf/cloudflare/clef-flash` | cf | ❌ 模型不存在/不可用 (404/400) | 0/1 | - | - | - | - | - | HTTP 400：{"errors":[{"message":"AiError: No such model: No such model clef-flash or task (8f4d7a58-6cf3-4253-933b-39a8d38270ef)"… |

### 47. command-a

归并键 `command-a||` ｜ 成员 1，可用 0 ｜ 提供商：cohere(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `cohere/command-a-03-2025` | cohere | ❌ 接口不支持 (405) | 0/1 | - | - | - | - | - | HTTP 405：HTTP 405 |

### 48. command-r

归并键 `command-r||` ｜ 成员 1，可用 0 ｜ 提供商：cohere(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `cohere/command-r-08-2024` | cohere | ❌ 接口不支持 (405) | 0/1 | - | - | - | - | - | HTTP 405：HTTP 405 |

### 49. command-r-plus

归并键 `command-r|plus|` ｜ 成员 1，可用 0 ｜ 提供商：cohere(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `cohere/command-r-plus-08-2024` | cohere | ❌ 接口不支持 (405) | 0/1 | - | - | - | - | - | HTTP 405：HTTP 405 |

### 50. compound-mini

归并键 `compound|mini|` ｜ 成员 1，可用 0 ｜ 提供商：groq(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `groq/groq/compound-mini` | groq | ❌ 模型不存在/不可用 (404/400) | 0/1 | - | - | - | - | - | HTTP 503/404：The model `groq/compound-mini` does not exist or you do not have access to it. (reset after 2m) |

### 51. d3-flash

归并键 `d|flash|3` ｜ 成员 1，可用 0 ｜ 提供商：amd(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `amd/d3-flash` | amd | ❌ 非聊天模型 | 0/1 | - | - | - | - | - | HTTP 400：Model d3-flash is not a chat model; call /v1/systemone instead of /v1/chat/completions |

### 52. deepseek-v4-flash-vision

归并键 `deepseek|flash+vision|4` ｜ 成员 1，可用 0 ｜ 提供商：amd(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `amd/DeepSeek-V4-Flash-Vision-Exp` | amd | ❌ 模型不存在/不可用 (404/400) | 0/1 | - | - | - | - | - | HTTP 400：Requested model DeepSeek-V4-Flash-Vision-Exp not supported |

### 53. default

归并键 `default||` ｜ 成员 1，可用 0 ｜ 提供商：cu(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `cu/default` | cu | ❌ 上游返回空内容 | 0/3 | - | - | - | - | - | HTTP 400：Cursor AgentService returned an empty turn |

### 54. gemini-4

归并键 `gemini||4` ｜ 成员 1，可用 0 ｜ 提供商：thai(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `thai/gemini-4-argon` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 55. glm-5-flashx

归并键 `glm|flashx|5` ｜ 成员 1，可用 0 ｜ 提供商：thai(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `thai/glm-5.3-flashx` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 56. jev-1

归并键 `jev||1` ｜ 成员 1，可用 0 ｜ 提供商：oc(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `oc/jev-1.13-free` | oc | ❌ 免费额度已用尽 | 0/2 | - | - | - | - | - | HTTP 429：Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: h… |

### 57. kimi-k3-pro

归并键 `kimi|pro|3` ｜ 成员 1，可用 0 ｜ 提供商：thai(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `thai/kimi-k3.1-pro` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 58. ling-3-flash

归并键 `ling|flash|3` ｜ 成员 1，可用 0 ｜ 提供商：thai(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `thai/ling-3.1-flash` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 59. longcat-2

归并键 `longcat||2` ｜ 成员 1，可用 0 ｜ 提供商：oc(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `oc/longcat-2.5-preview-free` | oc | ❌ 模型不存在/不可用 (404/400) | 0/1 | - | - | - | - | - | HTTP 400：Upstream request failed: Model is unavailable. |

### 60. mineru2-pro

归并键 `mineru|pro|2` ｜ 成员 1，可用 0 ｜ 提供商：amd(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `amd/MinerU2.5-Pro` | amd | ❌ 非聊天模型 | 0/1 | - | - | - | - | - | HTTP 400：Model MinerU2.5-Pro is not a chat model; call /v1/ocr instead of /v1/chat/completions |

### 61. minimax-m2

归并键 `minimax||2` ｜ 成员 1，可用 0 ｜ 提供商：nvidia(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `nvidia/minimaxai/minimax-m2.7` | nvidia | ❌ 模型已下线 (410) | 0/1 | - | - | - | - | - | HTTP 410：The model 'minimaxai/minimax-m2.7' has reached its end of life on 2026-07-27T00:00:00Z and is no longer available. |

### 62. minimax-m3

归并键 `minimax||3` ｜ 成员 1，可用 0 ｜ 提供商：nvidia(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `nvidia/minimaxai/minimax-m3` | nvidia | ❌ 模型已下线 (410) | 0/1 | - | - | - | - | - | HTTP 410：The model 'minimaxai/minimax-m3' has reached its end of life on 2026-09-09T09:00:00Z and is no longer available. |

### 63. minimax-m3-flash

归并键 `minimax|flash|3` ｜ 成员 1，可用 0 ｜ 提供商：thai(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `thai/minimax-m3.1-flash` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 64. mistral-large

归并键 `mistral|large|` ｜ 成员 1，可用 0 ｜ 提供商：thai(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `thai/mistral-large-2512` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 65. mistral-large-4

归并键 `mistral|large|4` ｜ 成员 1，可用 0 ｜ 提供商：thai(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `thai/mistral-large-4` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 66. muse-wonder-1

归并键 `muse-wonder||1` ｜ 成员 1，可用 0 ｜ 提供商：thai(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `thai/muse-wonder-1.0` | thai | ❌ 余额不足，需充值 | 0/1 | - | - | - | - | - | HTTP 503/402：Your Token Harbor balance is at $0. Top up at https://tokenharbor.ai/dashboard to keep using paid models. (reset after … |

### 67. nemotron-3-lightning

归并键 `nemotron|lightning|3` ｜ 成员 1，可用 0 ｜ 提供商：oc(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `oc/nemotron-3.5-lightning-free` | oc | ❌ 免费额度已用尽 | 0/2 | - | - | - | - | - | HTTP 429：Free usage limit reached. Free models work in OpenCode (https://opencode.ai); subscribe to OpenCode Go to keep going: h… |

### 68. nemotron-3-ultra

归并键 `nemotron|ultra|3` ｜ 成员 1，可用 0 ｜ 提供商：nvidia(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `nvidia/nemotron-3-ultra-550b-a55b` | nvidia | ❌ 模型不存在/不可用 (404/400) | 0/1 | - | - | - | - | - | HTTP 503/404：404 page not found (reset after 1m 25s) |

### 69. omnivoice

归并键 `omnivoice||` ｜ 成员 1，可用 0 ｜ 提供商：amd(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `amd/OmniVoice` | amd | ❌ 非聊天模型 | 0/1 | - | - | - | - | - | HTTP 400：Model OmniVoice is not a chat model; call /v1/audio/speech instead of /v1/chat/completions |

### 70. orpheus-arabic-saudi

归并键 `orpheus-arabic-saudi||` ｜ 成员 1，可用 0 ｜ 提供商：groq(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `groq/canopylabs/orpheus-arabic-saudi` | groq | ❌ 需先接受模型条款 | 0/2 | - | - | - | - | - | HTTP 400：The model `canopylabs/orpheus-arabic-saudi` requires terms acceptance. Please have the org admin accept the terms at ht… |

### 71. orpheus-v1

归并键 `orpheus||1` ｜ 成员 1，可用 0 ｜ 提供商：groq(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `groq/canopylabs/orpheus-v1-english` | groq | ❌ 需先接受模型条款 | 0/2 | - | - | - | - | - | HTTP 400：The model `canopylabs/orpheus-v1-english` requires terms acceptance. Please have the org admin accept the terms at http… |

### 72. parakeet-ctc-asr

归并键 `parakeet-ctc|asr|` ｜ 成员 1，可用 0 ｜ 提供商：nvidia(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `nvidia/parakeet-ctc-1.1b-asr` | nvidia | ❌ 模型不存在/不可用 (404/400) | 0/1 | - | - | - | - | - | HTTP 503/404：404 page not found (reset after 2m) |

### 73. sensenova-u1-lite

归并键 `sensenova-u|lite|1` ｜ 成员 1，可用 0 ｜ 提供商：stkj(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `stkj/sensenova-u1.5-lite` | stkj | ❌ 模型不存在/不可用 (404/400) | 0/1 | - | - | - | - | - | HTTP 503/404：model is not found (reset after 2m) |

### 74. space-bunny

归并键 `space-bunny||` ｜ 成员 1，可用 0 ｜ 提供商：oc(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `oc/space-bunny-free` | oc | ❌ 模型不存在/不可用 (404/400) | 0/1 | - | - | - | - | - | HTTP 400：Upstream request failed: Model is unavailable. |

### 75. step-5

归并键 `step||5` ｜ 成员 1，可用 0 ｜ 提供商：oc(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `oc/step-5-preview-free` | oc | ❌ 模型不存在/不可用 (404/400) | 0/1 | - | - | - | - | - | HTTP 400：Upstream request failed: Model is unavailable. |

### 76. union

归并键 `union||` ｜ 成员 1，可用 0 ｜ 提供商：oc(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `oc/union-alpha` | oc | ❌ 模型不存在/不可用 (404/400) | 0/1 | - | - | - | - | - | HTTP 400：Upstream request failed: Model is unavailable. |

### 77. whisper-large-v3

归并键 `whisper|large|3` ｜ 成员 1，可用 0 ｜ 提供商：groq(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `groq/whisper-large-v3` | groq | ❌ 非聊天模型 | 0/2 | - | - | - | - | - | HTTP 400：The model `whisper-large-v3` does not support chat completions |

### 78. whisper-large-v3-turbo

归并键 `whisper|large+turbo|3` ｜ 成员 1，可用 0 ｜ 提供商：groq(1)

| 模型 ID | 提供商 | 状态 | 稳定性（成功/尝试） | 中位 | 最小 | 最大 | 实际返回模型名 | 上下文 | 备注 |
|---|---|---|---|---|---|---|---|---|---|
| `groq/whisper-large-v3-turbo` | groq | ❌ 非聊天模型 | 0/2 | - | - | - | - | - | HTTP 400：The model `whisper-large-v3-turbo` does not support chat completions |

## 7. 局限性与注意事项

- 延迟为端到端单次小请求的耗时，样本有限，用于横向比较“响应速度”，不是吞吐量或生成速度基准；长输出的 tokens/s 见延迟实验室。
- 所有流量经本机网络与可能的代理，绝对数值会随网络、节点与时段变化；同一时段内的相对排序更有参考价值。
- 大量不可用是**提供商账号层面**的问题（无凭证、token 失效、余额/免费额度用尽、套餐限制），这些提供商下的模型本身未必不可用，修复账号后应重新测试。
- 429 类错误已尽量区分“瞬时限流”与“账号额度用尽”，但依赖报错文本，可能有少量误判；原文均已附在第 5 节。
- 同一模型系列的归并基于命名规则与从本端点学到的别名，不同提供商对同名模型的实际后端可能不同（见“实际返回模型名”列）。
