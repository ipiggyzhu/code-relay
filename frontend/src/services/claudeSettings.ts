import { Call } from '@wailsio/runtime'
import type { ClaudeProxyStatus } from '../../bindings/coderelay/services/models'

type Platform = 'claude' | 'codex'

const serviceNames: Record<Platform, string> = {
  claude: 'coderelay/services.ClaudeSettingsService',
  codex: 'coderelay/services.CodexSettingsService',
}

const callByPlatform = async <T = unknown>(platform: Platform, method: string, payload?: any[]): Promise<T> => {
  const service = serviceNames[platform]
  const args = payload ?? []
  return Call.ByName(`${service}.${method}`, ...args)
}

export const fetchProxyStatus = async (platform: Platform): Promise<ClaudeProxyStatus> => {
  return callByPlatform<ClaudeProxyStatus>(platform, 'ProxyStatus')
}

export const enableProxy = async (platform: Platform): Promise<void> => {
  await callByPlatform(platform, 'EnableProxy')
}

export const disableProxy = async (platform: Platform): Promise<void> => {
  await callByPlatform(platform, 'DisableProxy')
}
