import { describe, it, expect } from 'vitest'
import * as fc from 'fast-check'
import { getSpeedTestColorClass, canTestSpeed, isValidSpeedTestUrl } from './speedTest'

describe('Speed Test Utils', () => {
  /**
   * **Feature: api-speed-test, Property 2: Status code color mapping consistency**
   * **Validates: Requirements 2.1, 2.2**
   * 
   * For any HTTP status code, the color class assignment SHALL follow:
   * 2xx → green (success), 3xx → blue (redirect), 4xx → orange (client error), 5xx → red (server error)
   */
  describe('getSpeedTestColorClass', () => {
    it('should return speed-success for all 2xx status codes', () => {
      fc.assert(
        fc.property(fc.integer({ min: 200, max: 299 }), (statusCode) => {
          expect(getSpeedTestColorClass(statusCode)).toBe('speed-success')
        }),
        { numRuns: 100 }
      )
    })

    it('should return speed-redirect for all 3xx status codes', () => {
      fc.assert(
        fc.property(fc.integer({ min: 300, max: 399 }), (statusCode) => {
          expect(getSpeedTestColorClass(statusCode)).toBe('speed-redirect')
        }),
        { numRuns: 100 }
      )
    })

    it('should return speed-client-error for all 4xx status codes', () => {
      fc.assert(
        fc.property(fc.integer({ min: 400, max: 499 }), (statusCode) => {
          expect(getSpeedTestColorClass(statusCode)).toBe('speed-client-error')
        }),
        { numRuns: 100 }
      )
    })

    it('should return speed-server-error for all 5xx status codes', () => {
      fc.assert(
        fc.property(fc.integer({ min: 500, max: 599 }), (statusCode) => {
          expect(getSpeedTestColorClass(statusCode)).toBe('speed-server-error')
        }),
        { numRuns: 100 }
      )
    })

    it('should return speed-error for status codes outside valid HTTP range', () => {
      fc.assert(
        fc.property(fc.integer({ min: 0, max: 199 }), (statusCode) => {
          expect(getSpeedTestColorClass(statusCode)).toBe('speed-error')
        }),
        { numRuns: 100 }
      )
    })
  })

  /**
   * **Feature: api-speed-test, Property 3: Button disabled state**
   * **Validates: Requirements 1.5**
   * 
   * For any empty or whitespace-only API URL input, the speed test button SHALL be disabled.
   */
  describe('canTestSpeed', () => {
    it('should return false for empty strings', () => {
      expect(canTestSpeed('', false)).toBe(false)
    })

    it('should return false for whitespace-only strings', () => {
      const whitespaceStrings = ['   ', '\t\t', '\n\n', '  \t\n  ', '\r\n']
      whitespaceStrings.forEach(ws => {
        expect(canTestSpeed(ws, false)).toBe(false)
      })
    })

    it('should return false when testing is in progress regardless of URL', () => {
      fc.assert(
        fc.property(fc.string({ minLength: 1 }), (url) => {
          expect(canTestSpeed(url, true)).toBe(false)
        }),
        { numRuns: 100 }
      )
    })

    it('should return true for non-empty URLs when not testing', () => {
      fc.assert(
        fc.property(
          fc.string({ minLength: 1 }).filter(s => s.trim().length > 0),
          (url) => {
            expect(canTestSpeed(url, false)).toBe(true)
          }
        ),
        { numRuns: 100 }
      )
    })
  })

  describe('isValidSpeedTestUrl', () => {
    it('should return true for valid http URLs', () => {
      const validUrls = [
        'http://example.com',
        'https://api.example.com',
        'https://api.example.com/v1/chat',
        'http://localhost:8080',
      ]
      validUrls.forEach(url => {
        expect(isValidSpeedTestUrl(url)).toBe(true)
      })
    })

    it('should return false for invalid URLs', () => {
      const invalidUrls = [
        '',
        '   ',
        'not-a-url',
        'ftp://example.com',
        'file:///path/to/file',
      ]
      invalidUrls.forEach(url => {
        expect(isValidSpeedTestUrl(url)).toBe(false)
      })
    })
  })
})
