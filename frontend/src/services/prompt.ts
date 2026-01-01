import { Call } from '@wailsio/runtime'

export type Prompt = {
    id: string
    name: string
    content: string
    platform: string
    is_active: boolean
    created_at: string
    updated_at: string
}

export type CreatePromptPayload = {
    name: string
    content: string
    platform: string
}

export const fetchPrompts = async (platform: string = ''): Promise<Prompt[]> => {
    const response = await Call.ByName('coderelay/services.PromptService.ListPrompts', platform)
    return (response as Prompt[]) ?? []
}

export const getPrompt = async (id: string): Promise<Prompt> => {
    const response = await Call.ByName('coderelay/services.PromptService.GetPrompt', id)
    return response as Prompt
}

export const createPrompt = async (payload: CreatePromptPayload): Promise<Prompt> => {
    const response = await Call.ByName('coderelay/services.PromptService.CreatePrompt', payload)
    return response as Prompt
}

export const updatePrompt = async (prompt: Prompt): Promise<void> => {
    await Call.ByName('coderelay/services.PromptService.UpdatePrompt', prompt)
}

export const deletePrompt = async (id: string): Promise<void> => {
    await Call.ByName('coderelay/services.PromptService.DeletePrompt', id)
}

export const activatePrompt = async (id: string): Promise<void> => {
    await Call.ByName('coderelay/services.PromptService.ActivatePrompt', id)
}

export const deactivatePrompt = async (id: string): Promise<void> => {
    await Call.ByName('coderelay/services.PromptService.DeactivatePrompt', id)
}

export const getActivePrompt = async (platform: string): Promise<Prompt | null> => {
    const response = await Call.ByName('coderelay/services.PromptService.GetActivePrompt', platform)
    return response as Prompt | null
}
