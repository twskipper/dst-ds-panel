import { useEffect, useState } from "react"
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  ResponsiveContainer,
  Tooltip,
} from "recharts"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { useWebSocket } from "@/hooks/useWebSocket"
import type { ContainerStats } from "@/types"

interface StatsChartProps {
  clusterId: string
  shard: string
  isRunning: boolean
}

interface StatsPoint {
  time: string
  cpu: number
  mem: number
}

export function StatsChart({ clusterId, shard, isRunning }: StatsChartProps) {
  const [data, setData] = useState<StatsPoint[]>([])

  const wsUrl = isRunning
    ? `/api/clusters/${clusterId}/shards/${shard}/stats`
    : null

  const { addListener } = useWebSocket(wsUrl)

  useEffect(() => {
    if (!isRunning) {
      setData([])
      return
    }
    const unsub = addListener((raw: string) => {
      try {
        const stats: ContainerStats = JSON.parse(raw)
        const point: StatsPoint = {
          time: new Date(stats.timestamp * 1000).toLocaleTimeString(),
          cpu: Math.round(stats.cpuPercent * 100) / 100,
          mem: Math.round(stats.memUsageMb * 10) / 10,
        }
        setData((prev) => [...prev.slice(-59), point])
      } catch {
        // ignore invalid messages
      }
    })
    return unsub
  }, [addListener, isRunning])

  if (!isRunning) {
    return (
      <div className="text-sm text-muted-foreground">
        Server not running
      </div>
    )
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm">CPU Usage (%)</CardTitle>
        </CardHeader>
        <CardContent>
          <ResponsiveContainer width="100%" height={150}>
            <LineChart data={data}>
              <XAxis dataKey="time" hide />
              <YAxis domain={[0, "auto"]} width={40} fontSize={10} />
              <Tooltip />
              <Line
                type="monotone"
                dataKey="cpu"
                stroke="#3b82f6"
                dot={false}
                strokeWidth={2}
              />
            </LineChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm">Memory (MB)</CardTitle>
        </CardHeader>
        <CardContent>
          <ResponsiveContainer width="100%" height={150}>
            <LineChart data={data}>
              <XAxis dataKey="time" hide />
              <YAxis domain={[0, "auto"]} width={50} fontSize={10} />
              <Tooltip />
              <Line
                type="monotone"
                dataKey="mem"
                stroke="#10b981"
                dot={false}
                strokeWidth={2}
              />
            </LineChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>
    </div>
  )
}
