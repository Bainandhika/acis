export function sanitizeString(input: string): string {
  if (!input) return ''
  return Array.from(input.replace(/[<>]/g, ''))
    .filter((char) => {
      const code = char.charCodeAt(0)
      return (code >= 32 && code !== 127) || code === 9 || code === 10 || code === 13
    })
    .join('')
    .trim()
    .slice(0, 1000)
}
