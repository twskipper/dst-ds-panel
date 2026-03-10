import { useEffect, useState } from "react"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { api } from "@/lib/api"

interface PlayerEvent {
  time: string
  player: string
  event: string
  detail?: string
}

const eventStyles: Record<string, { label: string; color: string }> = {
  join: { label: "Joined", color: "bg-green-500" },
  leave: { label: "Left", color: "bg-gray-400" },
  chat: { label: "Chat", color: "bg-blue-500" },
  death: { label: "Death", color: "bg-red-500" },
}

interface PlayerActivityProps {
  clusterId: string
}

export function PlayerActivity({ clusterId }: PlayerActivityProps) {
  const [events, setEvents] = useState<PlayerEvent[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetch = async () => {
      try {
        const data = await api.getPlayerActivity(clusterId)
        setEvents(data || [])
      } catch {
        // ignore
      } finally {
        setLoading(false)
      }
    }
    fetch()
    const interval = setInterval(fetch, 30000)
    return () => clearInterval(interval)
  }, [clusterId])

  if (loading) return <p className="text-sm text-muted-foreground">Loading...</p>

  if (events.length === 0) {
    return <p className="text-sm text-muted-foreground">No player activity recorded yet.</p>
  }

  return (
    <ScrollArea className="h-[300px]">
      <div className="space-y-1">
        {events.map((e, i) => {
          const style = eventStyles[e.event] || { label: e.event, color: "bg-gray-400" }
          return (
            <div key={i} className="flex items-center gap-2 text-sm py-1">
              <span className="text-xs text-muted-foreground w-16 shrink-0">{e.time}</span>
              <Badge variant="outline" className="gap-1 shrink-0">
                <span className={`h-1.5 w-1.5 rounded-full ${style.color}`} />
                {style.label}
              </Badge>
              <span className="font-medium">{e.player}</span>
              {e.detail && <span className="text-muted-foreground truncate">{e.detail}</span>}
            </div>
          )
        })}
      </div>
    </ScrollArea>
  )
}
