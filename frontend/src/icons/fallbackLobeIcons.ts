const fallbackIcons: Record<string, string> = {
  aicoding: `<svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <rect x="4" y="4" width="16" height="16" rx="5" stroke="currentColor" stroke-width="1.6" />
    <path d="M9 15l3-6 3 6" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" />
    <circle cx="12" cy="15" r="1.2" fill="currentColor" />
  </svg>`,
  kimi: `<svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <circle cx="8" cy="12" r="3.2" stroke="currentColor" stroke-width="1.5" />
    <circle cx="16" cy="12" r="3.2" stroke="currentColor" stroke-width="1.5" />
    <path d="M5 8l3.5-3.5M19 8l-3.5-3.5M5 16l3.5 3.5M19 16l-3.5 3.5" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" />
  </svg>`,
  deepseek: `<svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M4 12l8.5-8.5L21 12l-8.5 8.5L4 12z" stroke="currentColor" stroke-width="1.4" stroke-linejoin="round" />
    <path d="M8.5 12L12 8.5 15.5 12 12 15.5 8.5 12z" fill="currentColor" />
  </svg>`,
  google: `<svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/>
    <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-1 .67-2.28 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
    <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z" fill="#FBBC05"/>
    <path d="M12 5.38c1.62 0 3.06.56 4.21 1.66l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
  </svg>`,
}

export default fallbackIcons
