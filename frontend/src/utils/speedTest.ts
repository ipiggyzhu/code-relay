/**
 * Speed test utility functions
 * Feature: api-speed-test
 */

/**
 * Get CSS color class based on HTTP status code
 * **Feature: api-speed-test, Property 2: Status code color mapping consistency**
 * **Validates: Requirements 2.1, 2.2**
 */
export const getSpeedTestColorClass = (statusCode: number): string => {
  if (statusCode >= 200 && statusCode < 300) return 'speed-success'
  if (statusCode >= 300 && statusCode < 400) return 'speed-redirect'
  if (statusCode >= 400 && statusCode < 500) return 'speed-client-error'
  if (statusCode >= 500) return 'speed-server-error'
  return 'speed-error'
}

/**
 * Check if speed test button should be enabled
 * **Feature: api-speed-test, Property 3: Button disabled state**
 * **Validates: Requirements 1.5**
 */
export const canTestSpeed = (apiUrl: string, isTesting: boolean): boolean => {
  return apiUrl.trim().length > 0 && !isTesting
}

/**
 * Validate URL format for speed test
 */
export const isValidSpeedTestUrl = (url: string): boolean => {
  if (!url.trim()) return false
  try {
    const parsed = new URL(url)
    return /^https?:/.test(parsed.protocol)
  } catch {
    return false
  }
}
