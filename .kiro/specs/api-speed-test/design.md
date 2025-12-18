# Design Document: API Speed Test Feature

## Overview

为供应商配置弹窗中的 API 地址输入框添加测速按钮功能。用户点击按钮后，系统向 API 端点发送 HTTP HEAD 请求，测量响应时间并显示延迟和状态码。

## Architecture

该功能采用前端实现方案，直接在 Vue 组件中发起 HTTP 请求进行测速：

```
┌─────────────────────────────────────────────────────┐
│                   Index.vue                          │
│  ┌─────────────────────────────────────────────┐    │
│  │           Provider Modal Form                │    │
│  │  ┌─────────────────────┬──────────────────┐ │    │
│  │  │   API URL Input     │  Speed Test Btn  │ │    │
│  │  └─────────────────────┴──────────────────┘ │    │
│  │                         │                    │    │
│  │                         ▼                    │    │
│  │              ┌──────────────────┐           │    │
│  │              │  Speed Test      │           │    │
│  │              │  Result Display  │           │    │
│  │              │  (latency + code)│           │    │
│  │              └──────────────────┘           │    │
│  └─────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. Speed Test State (modalState 扩展)

```typescript
interface SpeedTestState {
  testing: boolean      // 是否正在测速
  latency: number | null  // 延迟毫秒数
  statusCode: number | null  // HTTP 状态码
  error: string | null  // 错误信息
}
```

### 2. Speed Test Function

```typescript
async function testApiSpeed(url: string): Promise<{
  latency: number
  statusCode: number
} | { error: string }>
```

### 3. UI Components

- **Speed Test Button**: 闪电图标按钮，位于 API URL 输入框右侧
- **Result Display**: 显示延迟和状态码，位于按钮右侧

## Data Models

### SpeedTestResult

```typescript
type SpeedTestResult = 
  | { success: true; latency: number; statusCode: number }
  | { success: false; error: string }
```

### Status Code Color Mapping

| Status Code Range | Color Class | Visual |
|-------------------|-------------|--------|
| 2xx | `speed-success` | 绿色 |
| 3xx | `speed-redirect` | 蓝色 |
| 4xx | `speed-client-error` | 橙色 |
| 5xx | `speed-server-error` | 红色 |
| Error/Timeout | `speed-error` | 红色 |

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Latency calculation accuracy
*For any* successful HTTP response, the calculated latency SHALL be the difference between response time and request start time in milliseconds, and SHALL be a non-negative number.
**Validates: Requirements 1.1, 1.2**

### Property 2: Status code color mapping consistency
*For any* HTTP status code, the color class assignment SHALL follow: 2xx → green (success), 3xx → blue (redirect), 4xx → orange (client error), 5xx → red (server error).
**Validates: Requirements 2.1, 2.2**

### Property 3: Button disabled state
*For any* empty or whitespace-only API URL input, the speed test button SHALL be disabled.
**Validates: Requirements 1.5**

### Property 4: Error handling for network failures
*For any* network error or timeout, the system SHALL display an error message and SHALL NOT display latency or status code values.
**Validates: Requirements 1.4**

## Error Handling

| Error Type | User Message | Behavior |
|------------|--------------|----------|
| Empty URL | Button disabled | 不允许点击 |
| Invalid URL | "无效的 URL 格式" | 显示错误 |
| Network Error | "网络错误" | 显示错误 |
| Timeout (5s) | "请求超时" | 显示错误 |
| CORS Error | "无法测速 (CORS)" | 显示错误 |

## Testing Strategy

### Unit Tests
- 测试 URL 验证逻辑
- 测试状态码颜色映射函数
- 测试按钮禁用状态逻辑

### Property-Based Tests
使用 fast-check 库进行属性测试：

1. **Property 1**: 生成随机时间戳对，验证延迟计算
2. **Property 2**: 生成随机状态码 (100-599)，验证颜色映射
3. **Property 3**: 生成随机字符串（包括空白字符串），验证按钮状态
4. **Property 4**: 模拟各种错误类型，验证错误处理

### Integration Tests
- 测试完整的测速流程（使用 mock HTTP）
- 测试 UI 状态转换（idle → loading → result/error）
