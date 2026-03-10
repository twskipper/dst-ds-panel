import { useEffect, useRef, useState } from "react"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { useWebSocket } from "@/hooks/useWebSocket"

interface LogViewerProps {
  clusterId: string
  shard: string
  isRunning: boolean
}

export function LogViewer({ clusterId, shard, isRunning }: LogViewerProps) {
  const [lines, setLines] = useState<string[]>([])
  const [autoScroll, setAutoScroll] = useState(true)
  const bottomRef = useRef<HTMLDivElement>(null)

  const wsUrl = isRunning
    ? `/api/clusters/${clusterId}/shards/${shard}/logs`
    : null

  const { isConnected, addListener } = useWebSocket(wsUrl)

  useEffect(() => {
    if (!isRunning) {
      setLines([])
      return
    }
    const unsub = addListener((data: string) => {
      setLines((prev) => [...prev.slice(-499), data])
    })
    return unsub
  }, [addListener, isRunning])

  useEffect(() => {
    if (autoScroll) {
      bottomRef.current?.scrollIntoView({ behavior: "smooth" })
    }
  }, [lines, autoScroll])

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium">Logs</span>
          <Badge variant={isConnected ? "default" : "secondary"}>
            {isConnected ? "Connected" : "Disconnected"}
          </Badge>
        </div>
        <div className="flex gap-2">
          <Button
            size="sm"
            variant="outline"
            onClick={() => setAutoScroll(!autoScroll)}
          >
            {autoScroll ? "Pause Scroll" : "Resume Scroll"}
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={() => setLines([])}
          >
            Clear
          </Button>
        </div>
      </div>
      <ScrollArea className="h-[350px] rounded-md border bg-black p-3">
        <div className="font-mono text-xs text-green-400 whitespace-pre-wrap">
          {lines.length === 0 ? (
            <span className="text-muted-foreground">
              {isRunning ? "Waiting for logs..." : "Server not running"}
            </span>
          ) : (
            lines.map((line, i) => <div key={i}>{line}</div>)
          )}
          <div ref={bottomRef} />
        </div>
      </ScrollArea>
    </div>
  )
}
