import { useState, useEffect, useCallback } from "react"

export function useTheme() {
  const [isDark, setIsDark] = useState(() => {
    const saved = localStorage.getItem("dst_theme")
    if (saved) return saved === "dark"
    return window.matchMedia("(prefers-color-scheme: dark)").matches
  })

  useEffect(() => {
    document.documentElement.classList.toggle("dark", isDark)
    localStorage.setItem("dst_theme", isDark ? "dark" : "light")
  }, [isDark])

  const toggle = useCallback(() => setIsDark((prev) => !prev), [])

  return { isDark, toggle }
}
