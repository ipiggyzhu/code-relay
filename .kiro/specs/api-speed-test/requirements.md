# Requirements Document

## Introduction

为供应商配置弹窗中的 API 地址字段添加测速功能，允许用户在配置供应商时快速测试 API 端点的连通性和响应速度。

## Glossary

- **API 地址**: 供应商的 API 端点 URL
- **测速**: 向 API 端点发送请求并测量响应时间
- **延迟**: 从发送请求到收到响应的时间（毫秒）
- **状态码**: HTTP 响应状态码（如 200、404、500）

## Requirements

### Requirement 1

**User Story:** As a user, I want to test the API endpoint speed when configuring a provider, so that I can verify the endpoint is accessible and check its response time.

#### Acceptance Criteria

1. WHEN a user clicks the speed test button next to the API URL input THEN the system SHALL send a request to the API endpoint and display the response latency in milliseconds
2. WHEN the speed test completes successfully THEN the system SHALL display both the latency (e.g., "406ms") and the HTTP status code (e.g., "200")
3. WHILE the speed test is in progress THEN the system SHALL display a loading indicator on the button
4. IF the API endpoint is unreachable or times out THEN the system SHALL display an error message indicating the failure
5. WHEN the API URL input is empty THEN the system SHALL disable the speed test button

### Requirement 2

**User Story:** As a user, I want the speed test results to be visually clear, so that I can quickly understand the API endpoint status.

#### Acceptance Criteria

1. WHEN the status code is 2xx THEN the system SHALL display the result in green color indicating success
2. WHEN the status code is 4xx or 5xx THEN the system SHALL display the result in red/orange color indicating an error
3. WHEN displaying results THEN the system SHALL show latency prominently with status code below it
