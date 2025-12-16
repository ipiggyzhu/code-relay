import { CurrentVersion } from '../../bindings/coderelay/versionservice'

export const fetchCurrentVersion = async (): Promise<string> => {
  const version = await CurrentVersion()
  return version ?? ''
}
