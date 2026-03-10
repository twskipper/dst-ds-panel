import { createContext, useContext, useState, useCallback, useEffect, type ReactNode } from "react"

interface AuthContextType {
  token: string | null
  isAuthenticated: boolean
  login: (username: string, password: string) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextType | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() =>
    localStorage.getItem("dst_token")
  )

  // Check token expiry on load
  useEffect(() => {
    const expiresAt = localStorage.getItem("dst_token_expires")
    if (expiresAt && Date.now() / 1000 > Number(expiresAt)) {
      setToken(null)
      localStorage.removeItem("dst_token")
      localStorage.removeItem("dst_token_expires")
    }
  }, [])

  const login = useCallback(async (username: string, password: string) => {
    const res = await fetch("/api/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password }),
    })
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: "Login failed" }))
      throw new Error(err.error || "Login failed")
    }
    const data = await res.json()
    setToken(data.token)
    localStorage.setItem("dst_token", data.token)
    localStorage.setItem("dst_token_expires", String(data.expiresAt))
  }, [])

  const logout = useCallback(() => {
    setToken(null)
    localStorage.removeItem("dst_token")
    localStorage.removeItem("dst_token_expires")
  }, [])

  return (
    <AuthContext.Provider value={{ token, isAuthenticated: !!token, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error("useAuth must be used within AuthProvider")
  return ctx
}
