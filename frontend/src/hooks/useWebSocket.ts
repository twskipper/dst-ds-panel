import { useEffect, useRef, useState, useCallback } from "react"

export function useWebSocket(url: string | null) {
  const [isConnected, setIsConnected] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)
  const listenersRef = useRef<((data: string) => void)[]>([])

  const addListener = useCallback((fn: (data: string) => void) => {
    listenersRef.current.push(fn)
    return () => {
      listenersRef.current = listenersRef.current.filter((l) => l !== fn)
    }
  }, [])

  useEffect(() => {
    if (!url) return

    let reconnectTimer: ReturnType<typeof setTimeout>
    let ws: WebSocket

    function connect() {
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:"
      const token = localStorage.getItem("dst_token")
      const sep = url!.includes("?") ? "&" : "?"
      const wsUrl = `${protocol}//${window.location.host}${url}${token ? sep + "token=" + token : ""}`
      ws = new WebSocket(wsUrl)
      wsRef.current = ws

      ws.onopen = () => setIsConnected(true)
      ws.onclose = () => {
        setIsConnected(false)
        reconnectTimer = setTimeout(connect, 3000)
      }
      ws.onmessage = (e) => {
        listenersRef.current.forEach((fn) => fn(e.data))
      }
    }

    connect()

    return () => {
      clearTimeout(reconnectTimer)
      ws?.close()
      wsRef.current = null
    }
  }, [url])

  return { isConnected, addListener }
}
