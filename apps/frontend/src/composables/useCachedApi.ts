const cache = new Map<string, { data: any; timestamp: number }>()

export function useCachedApi<T>(key: string, fetcher: () => Promise<T>, ttl: number = 60000): Promise<T> {
  const cached = cache.get(key)
  if (cached && Date.now() - cached.timestamp < ttl) {
    return Promise.resolve(cached.data as T)
  }
  return fetcher().then((data) => {
    cache.set(key, { data, timestamp: Date.now() })
    return data
  })
}

export function clearApiCache(key?: string) {
  if (key) {
    cache.delete(key)
  } else {
    cache.clear()
  }
}
