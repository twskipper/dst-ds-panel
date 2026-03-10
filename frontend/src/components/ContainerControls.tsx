import { useState } from "react"
import { Button } from "@/components/ui/button"
import { api } from "@/lib/api"
import { toast } from "sonner"
import type { Cluster } from "@/types"

interface ContainerControlsProps {
  cluster: Cluster
  onUpdate: (cluster: Cluster) => void
}

export function ContainerControls({ cluster, onUpdate }: ContainerControlsProps) {
  const [loading, setLoading] = useState<string | null>(null)

  const handleAction = async (action: "start" | "stop" | "restart") => {
    setLoading(action)
    try {
      let updated: Cluster
      switch (action) {
        case "start":
          updated = await api.startCluster(cluster.id)
          toast.success(`Server ${cluster.name} started`)
          break
        case "stop":
          updated = await api.stopCluster(cluster.id)
          toast.success(`Server ${cluster.name} stopped`)
          break
        case "restart":
          updated = await api.restartCluster(cluster.id)
          toast.success(`Server ${cluster.name} restarted`)
          break
      }
      onUpdate(updated)
    } catch (err) {
      const msg = err instanceof Error ? err.message : `Failed to ${action}`
      toast.error(msg)
    } finally {
      setLoading(null)
    }
  }

  const isStopped = cluster.status === "stopped"
  const isRunning = cluster.status === "running"

  return (
    <div className="flex gap-2">
      {isStopped && (
        <Button
          onClick={() => handleAction("start")}
          disabled={loading !== null}
        >
          {loading === "start" ? "Starting..." : "Start Server"}
        </Button>
      )}
      {isRunning && (
        <>
          <Button
            variant="destructive"
            onClick={() => handleAction("stop")}
            disabled={loading !== null}
          >
            {loading === "stop" ? "Stopping..." : "Stop"}
          </Button>
          <Button
            variant="outline"
            onClick={() => handleAction("restart")}
            disabled={loading !== null}
          >
            {loading === "restart" ? "Restarting..." : "Restart"}
          </Button>
        </>
      )}
    </div>
  )
}
