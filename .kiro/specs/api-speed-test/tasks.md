# Implementation Plan

- [x] 1. Add speed test state and function to Index.vue




  - [ ] 1.1 Add speedTestState to modalState reactive object
    - Add `speedTest: { testing: boolean, latency: number | null, statusCode: number | null, error: string | null }` to modalState


    - Initialize with default values when modal opens
    - _Requirements: 1.1, 1.3_
  - [ ] 1.2 Implement testApiSpeed function
    - Create async function that sends fetch request to API URL
    - Calculate latency using performance.now() or Date.now()

    - Handle success response with latency and status code
    - Handle errors (network, timeout, CORS) with appropriate messages
    - Set 5 second timeout using AbortController
    - _Requirements: 1.1, 1.2, 1.4_
  - [x] 1.3 Implement getSpeedTestColorClass helper function




    - Return 'speed-success' for 2xx status codes
    - Return 'speed-redirect' for 3xx status codes
    - Return 'speed-client-error' for 4xx status codes
    - Return 'speed-server-error' for 5xx status codes

    - _Requirements: 2.1, 2.2_

- [ ] 2. Add speed test UI to provider modal form
  - [ ] 2.1 Add speed test button next to API URL input
    - Add lightning bolt icon button after the input field


    - Disable button when apiUrl is empty or testing is in progress
    - Show loading spinner when testing
    - Wire up click handler to testApiSpeed function
    - _Requirements: 1.1, 1.3, 1.5_
  - [x] 2.2 Add speed test result display


    - Show latency in milliseconds (e.g., "406ms") with appropriate color
    - Show status code below latency (e.g., "状态码: 200")
    - Show error message when test fails


    - Position result to the right of the button





    - _Requirements: 1.2, 2.1, 2.2, 2.3_


- [ ] 3. Add CSS styles for speed test components
  - Add styles for speed test button (normal, hover, disabled, loading states)
  - Add color classes for different status code ranges
  - Add styles for result display layout
  - Ensure dark mode compatibility
  - _Requirements: 2.1, 2.2, 2.3_

- [ ] 4. Add i18n translations
  - Add Chinese translations for speed test labels and error messages
  - Add English translations for speed test labels and error messages
  - _Requirements: 1.4_

- [ ] 5. Checkpoint - Make sure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 6. Write property tests for speed test logic
  - [ ] 6.1 Write property test for status code color mapping
    - **Property 2: Status code color mapping consistency**
    - **Validates: Requirements 2.1, 2.2**
  - [ ] 6.2 Write property test for button disabled state
    - **Property 3: Button disabled state**
    - **Validates: Requirements 1.5**
