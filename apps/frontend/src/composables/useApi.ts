import { ref, type Ref } from 'vue'
import type { AxiosError } from 'axios'

export interface UseApiOptions<T, P extends any[] = any[]> {
  immediate?: boolean
  onSuccess?: (data: T, ...args: P) => void
  onError?: (error: string) => void
  resetOnExecute?: boolean
}

export function useApi<T, F extends (...args: any[]) => Promise<any>>(
  apiCall: F,
  options: UseApiOptions<T, Parameters<F>> = {}
) {
  type P = Parameters<F>
  const data: Ref<T | null> = ref(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  const reset = () => {
    data.value = null
    error.value = null
    isLoading.value = false
  }

  const execute = async (...args: P): Promise<T | null> => {
    if (options.resetOnExecute !== false) {
      reset()
    }

    isLoading.value = true
    error.value = null

    try {
      const response = await apiCall(...args)
      const responseBody = response?.data as any
      // Handle both { data: T } and direct T response payloads
      const payload: T = responseBody && typeof responseBody === 'object' && 'data' in responseBody
        ? responseBody.data
        : responseBody

      data.value = payload
      options.onSuccess?.(payload, ...args)
      return payload
    } catch (err) {
      const axiosError = err as AxiosError<{ error?: string; message?: string }>
      const errorMessage =
        axiosError.response?.data?.error ||
        axiosError.response?.data?.message ||
        axiosError.message ||
        'An unexpected error occurred'

      error.value = errorMessage
      options.onError?.(errorMessage)
      return null
    } finally {
      isLoading.value = false
    }
  }

  if (options.immediate) {
    execute(...([] as unknown as P))
  }

  return { data, isLoading, error, execute, reset }
}
