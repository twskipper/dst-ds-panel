import { useState } from "react"
import { useNavigate } from "react-router-dom"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { api } from "@/lib/api"
import type { Cluster } from "@/types"

const statusColors: Record<string, string> = {
  running: "bg-green-500",
  stopped: "bg-gray-400",
  starting: "bg-yellow-500",
  error: "bg-red-500",
}

interface ClusterCardProps {
  cluster: Cluster
  onAction: () => void
}

export function ClusterCard({ cluster, onAction }: ClusterCardProps) {
  const navigate = useNavigate()
  const [actionLoading, setActionLoading] = useState(false)

  const [error, setError] = useState("")

  const handleStart = async (e: React.MouseEvent) => {
    e.stopPropagation()
    setActionLoading(true)
    setError("")
    try {
      await api.startCluster(cluster.id)
      onAction()
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to start")
    } finally {
      setActionLoading(false)
    }
  }

  const handleStop = async (e: React.MouseEvent) => {
    e.stopPropagation()
    setActionLoading(true)
    setError("")
    try {
      await api.stopCluster(cluster.id)
      onAction()
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to stop")
    } finally {
      setActionLoading(false)
    }
  }

  return (
    <Card
      className="cursor-pointer hover:shadow-md transition-shadow"
      onClick={() => navigate(`/clusters/${cluster.id}`)}
    >
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg">{cluster.name}</CardTitle>
          <Badge variant="outline" className="gap-1.5">
            <span
              className={`h-2 w-2 rounded-full ${statusColors[cluster.status] || "bg-gray-400"}`}
            />
            {cluster.status}
          </Badge>
        </div>
        <CardDescription>
          {cluster.config.gameMode} | Max {cluster.config.maxPlayers} players
        </CardDescription>
      </CardHeader>
      <CardContent>
        {error && (
          <Alert variant="destructive" className="mb-2">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        <div className="flex gap-2">
          {cluster.status === "stopped" ? (
            <Button
              size="sm"
              onClick={handleStart}
              disabled={actionLoading}
            >
              {actionLoading ? "Starting..." : "Start"}
            </Button>
          ) : cluster.status === "running" ? (
            <Button
              size="sm"
              variant="destructive"
              onClick={handleStop}
              disabled={actionLoading}
            >
              {actionLoading ? "Stopping..." : "Stop"}
            </Button>
          ) : null}
        </div>
      </CardContent>
    </Card>
  )
}
