import { useEffect, useState } from "react"
import { useParams, useNavigate } from "react-router-dom"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { ContainerControls } from "@/components/ContainerControls"
import { ClusterForm } from "@/components/ClusterForm"
import { LogViewer } from "@/components/LogViewer"
import { StatsChart } from "@/components/StatsChart"
import { ModList } from "@/components/ModList"
import { FileEditor } from "@/components/FileEditor"
import { AdminList } from "@/components/AdminList"
import { ServerConsole } from "@/components/ServerConsole"
import { PlayerActivity } from "@/components/PlayerActivity"
import { WorldSettings } from "@/components/WorldSettings"
import { useTranslation } from "react-i18next"
import { api } from "@/lib/api"
import type { Cluster, Mod } from "@/types"

const statusColors: Record<string, string> = {
  running: "bg-green-500",
  stopped: "bg-gray-400",
  starting: "bg-yellow-500",
  error: "bg-red-500",
}

export function ClusterDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { t } = useTranslation()
  const [cluster, setCluster] = useState<Cluster | null>(null)
  const [mods, setMods] = useState<Mod[]>([])
  const [loading, setLoading] = useState(true)
  const [deleting, setDeleting] = useState(false)

  useEffect(() => {
    if (!id) return
    const fetchData = async () => {
      try {
        const [clusterData, modsData] = await Promise.all([
          api.getCluster(id),
          api.listMods(id),
        ])
        setCluster(clusterData)
        setMods(modsData || [])
      } catch (err) {
        console.error("Failed to fetch cluster:", err)
      } finally {
        setLoading(false)
      }
    }
    fetchData()
  }, [id])

  const handleDelete = async () => {
    if (!id || !confirm("Delete this cluster? This will remove all save data."))
      return
    setDeleting(true)
    try {
      await api.deleteCluster(id)
      navigate("/")
    } catch (err) {
      console.error("Failed to delete:", err)
      setDeleting(false)
    }
  }

  if (loading) return <p className="text-muted-foreground">Loading...</p>
  if (!cluster) return <p className="text-destructive">Cluster not found</p>

  const isRunning = cluster.status === "running"

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-4">
          <Button variant="ghost" onClick={() => navigate("/")}>
            &larr; Back
          </Button>
          <h1 className="text-2xl font-bold">{cluster.name}</h1>
          <Badge variant="outline" className="gap-1.5">
            <span
              className={`h-2 w-2 rounded-full ${statusColors[cluster.status] || "bg-gray-400"}`}
            />
            {cluster.status}
          </Badge>
        </div>
        <div className="flex gap-2">
          <ContainerControls cluster={cluster} onUpdate={setCluster} />
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              window.open(`/api/clusters/${cluster.id}/backup`, "_blank")
            }}
          >
            Backup
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={async () => {
              const name = prompt("New cluster name:")
              if (!name) return
              try {
                const cloned = await api.cloneCluster(cluster.id, name)
                navigate(`/clusters/${cloned.id}`)
              } catch (err) {
                alert(err instanceof Error ? err.message : "Clone failed")
              }
            }}
          >
            Clone
          </Button>
          <Button
            variant="destructive"
            size="sm"
            onClick={handleDelete}
            disabled={deleting}
          >
            {deleting ? "Deleting..." : "Delete"}
          </Button>
        </div>
      </div>

      <Tabs defaultValue="overview">
        <TabsList>
          <TabsTrigger value="overview">{t("tabs.overview")}</TabsTrigger>
          <TabsTrigger value="master">{t("tabs.master")}</TabsTrigger>
          <TabsTrigger value="caves">{t("tabs.caves")}</TabsTrigger>
          <TabsTrigger value="console">{t("tabs.console")}</TabsTrigger>
          <TabsTrigger value="world">{t("tabs.world")}</TabsTrigger>
          <TabsTrigger value="mods">{t("tabs.mods")}</TabsTrigger>
          <TabsTrigger value="files">{t("tabs.files")}</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="mt-4 space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>{t("overview.clusterConfig")}</CardTitle>
            </CardHeader>
            <CardContent>
              <ClusterForm
                clusterId={cluster.id}
                initialConfig={cluster.config}
                onSaved={(config) =>
                  setCluster({ ...cluster, config })
                }
              />
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>{t("overview.serverAdmins")}</CardTitle>
            </CardHeader>
            <CardContent>
              <AdminList clusterId={cluster.id} />
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>{t("overview.playerActivity")}</CardTitle>
            </CardHeader>
            <CardContent>
              <PlayerActivity clusterId={cluster.id} />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="master" className="mt-4 space-y-4">
          <ShardPanel
            clusterId={cluster.id}
            shard="Master"
            isRunning={isRunning}
          />
        </TabsContent>

        <TabsContent value="caves" className="mt-4 space-y-4">
          <ShardPanel
            clusterId={cluster.id}
            shard="Caves"
            isRunning={isRunning}
          />
        </TabsContent>

        <TabsContent value="console" className="mt-4 space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>{t("console.title")}</CardTitle>
            </CardHeader>
            <CardContent>
              <ServerConsole
                clusterId={cluster.id}
                shard="Master"
                isRunning={isRunning}
              />
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>{t("console.serverLog")}</CardTitle>
            </CardHeader>
            <CardContent>
              <LogViewer
                clusterId={cluster.id}
                shard="Master"
                isRunning={isRunning}
              />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="world" className="mt-4 space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>{t("world.overworld")}</CardTitle>
            </CardHeader>
            <CardContent>
              <WorldSettings clusterId={cluster.id} shard="Master" />
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>{t("world.caves")}</CardTitle>
            </CardHeader>
            <CardContent>
              <WorldSettings clusterId={cluster.id} shard="Caves" />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="mods" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle>{t("mods.title")}</CardTitle>
            </CardHeader>
            <CardContent>
              <ModList
                clusterId={cluster.id}
                mods={mods}
                onUpdate={setMods}
              />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="files" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle>{t("files.title")}</CardTitle>
            </CardHeader>
            <CardContent>
              <FileEditor clusterId={cluster.id} />
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}

function ShardPanel({
  clusterId,
  shard,
  isRunning,
}: {
  clusterId: string
  shard: string
  isRunning: boolean
}) {
  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>{shard} Shard</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <StatsChart
            clusterId={clusterId}
            shard={shard}
            isRunning={isRunning}
          />
          <LogViewer
            clusterId={clusterId}
            shard={shard}
            isRunning={isRunning}
          />
        </CardContent>
      </Card>
    </>
  )
}
