import { Call } from '@wailsio/runtime'

/**
 * 获取通用配置（返回 JSON 字符串）
 */
export const getCommonConfigJSON = async (kind: string): Promise<string> => {
  try {
    console.log('[commonConfig] Calling GetCommonConfigJSON with kind:', kind)
    const result = await Call.ByName('coderelay/services.CommonConfigService.GetCommonConfigJSON', kind)
    console.log('[commonConfig] GetCommonConfigJSON result:', result)
    return result as string
  } catch (error) {
    console.error('[commonConfig] GetCommonConfigJSON error:', error)
    throw error
  }
}

/**
 * 保存通用配置（接收 JSON 字符串）
 */
export const saveCommonConfigJSON = async (kind: string, jsonStr: string): Promise<void> => {
  try {
    console.log('[commonConfig] Calling SaveCommonConfigJSON with kind:', kind)
    await Call.ByName('coderelay/services.CommonConfigService.SaveCommonConfigJSON', kind, jsonStr)
    console.log('[commonConfig] SaveCommonConfigJSON success')
  } catch (error) {
    console.error('[commonConfig] SaveCommonConfigJSON error:', error)
    throw error
  }
}
