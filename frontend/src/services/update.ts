import { CheckForUpdates, DownloadUpdate, InstallUpdate, GetCurrentExePath } from '../../bindings/coderelay/services/updateservice'

export interface UpdateInfo {
  hasUpdate: boolean
  currentVersion: string
  latestVersion: string
  downloadUrl: string
  releaseUrl: string
  fileName: string
  fileSize: number
}

export interface PendingUpdateInfo {
  hasPendingUpdate: boolean
  version: string
  filePath: string
}

export const checkForUpdates = async (): Promise<UpdateInfo | null> => {
  try {
    const result = await CheckForUpdates()
    return result as UpdateInfo
  } catch (error) {
    console.error('Failed to check for updates:', error)
    return null
  }
}

export const downloadUpdate = async (downloadUrl: string): Promise<string | null> => {
  try {
    return await DownloadUpdate(downloadUrl)
  } catch (error) {
    console.error('Failed to download update:', error)
    return null
  }
}

export const installUpdate = async (downloadedPath: string): Promise<boolean> => {
  try {
    await InstallUpdate(downloadedPath)
    return true
  } catch (error) {
    console.error('Failed to install update:', error)
    return false
  }
}

export const getCurrentExePath = async (): Promise<string | null> => {
  try {
    return await GetCurrentExePath()
  } catch (error) {
    console.error('Failed to get exe path:', error)
    return null
  }
}

// 这些函数需要在 wails generate bindings 后才能使用
// export const getPendingUpdate = async (): Promise<PendingUpdateInfo | null> => { ... }
// export const setPendingUpdate = async (path: string, version: string): Promise<boolean> => { ... }

export const formatFileSize = (bytes: number): string => {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}
