import { useEffect, useState } from "react"
import { useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Card, CardContent } from "@/components/ui/card"
import { ClusterCard } from "@/components/ClusterCard"
import { api } from "@/lib/api"
import { useTranslation } from "react-i18next"
import type { Cluster } from "@/types"

export function DashboardPage() {
  const navigate = useNavigate()
  const { t } = useTranslation()
  const [clusters, setClusters] = useState<Cluster[]>([])
  const [loading, setLoading] = useState(true)
  const [dstStatus, setDstStatus] = useState<{ dstInstalled: boolean; dstVersion: string; needsManualUpdate: boolean } | null>(null)
  const [updating, setUpdating] = useState(false)
  const [updateMsg, setUpdateMsg] = useState<{ type: "success" | "error"; text: string } | null>(null)
  const [betaVersion, setBetaVersion] = useState("")

  const fetchClusters = async () => {
    try {
      const data = await api.listClusters()
      setClusters(data || [])
    } catch (err) {
      console.error("Failed to fetch clusters:", err)
    } finally {
      setLoading(false)
    }
  }

  const fetchStatus = async () => {
    try {
      const status = await api.imageStatus()
      setDstStatus(status)
    } catch {
      // ignore
    }
  }

  useEffect(() => {
    fetchClusters()
    fetchStatus()
    // Auto-refresh cluster status every 10 seconds
    const interval = setInterval(fetchClusters, 10000)
    return () => clearInterval(interval)
  }, [])

  const handleUpdateDST = async () => {
    setUpdating(true)
    setUpdateMsg(null)
    try {
      await api.updateDST(betaVersion.trim() || undefined)
      setUpdateMsg({ type: "success", text: "DST server updated successfully. Restart running clusters to use the new version." })
      fetchStatus()
    } catch (err) {
      setUpdateMsg({ type: "error", text: err instanceof Error ? err.message : "Update failed" })
    } finally {
      setUpdating(false)
    }
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-bold">{t("dashboard.title")}</h1>
          {dstStatus && (
            dstStatus.dstInstalled && dstStatus.dstVersion ? (
              <a
                href="https://forums.kleientertainment.com/game-updates/dst/"
                target="_blank"
                rel="noopener noreferrer"
              >
                <Badge variant="outline" className="cursor-pointer hover:bg-accent">
                  DST {dstStatus.dstVersion}
                </Badge>
              </a>
            ) : (
              <Badge variant={dstStatus.dstInstalled ? "outline" : "destructive"}>
                {dstStatus.dstInstalled ? t("dashboard.dstInstalled") : t("dashboard.dstNotInstalled")}
              </Badge>
            )
          )}
        </div>
        <div className="flex gap-2">
          {dstStatus?.needsManualUpdate && (
            <>
              <input
                type="text"
                placeholder={t("dashboard.betaPlaceholder")}
                value={betaVersion}
                onChange={(e) => setBetaVersion(e.target.value)}
                className="h-9 w-32 rounded-md border border-input bg-background px-2 text-sm"
              />
              <Button
                variant="outline"
                onClick={handleUpdateDST}
                disabled={updating}
              >
                {updating
                  ? t("dashboard.updatingDST")
                  : dstStatus?.dstInstalled
                    ? t("dashboard.updateDST")
                    : t("dashboard.installDST")}
              </Button>
            </>
          )}
          <Button onClick={() => navigate("/clusters/new")}>{t("dashboard.newCluster")}</Button>
        </div>
      </div>

      {updateMsg && (
        <Alert variant={updateMsg.type === "error" ? "destructive" : "default"} className="mb-4">
          <AlertDescription>{updateMsg.text}</AlertDescription>
        </Alert>
      )}

      {!loading && clusters.length > 0 && (
        <div className="grid gap-4 grid-cols-2 md:grid-cols-4 mb-6">
          <Card>
            <CardContent className="pt-4 pb-3">
              <p className="text-xs text-muted-foreground">{t("dashboard.totalClusters")}</p>
              <p className="text-2xl font-bold">{clusters.length}</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-4 pb-3">
              <p className="text-xs text-muted-foreground">{t("dashboard.running")}</p>
              <p className="text-2xl font-bold text-green-600">
                {clusters.filter(c => c.status === "running").length}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-4 pb-3">
              <p className="text-xs text-muted-foreground">{t("dashboard.stopped")}</p>
              <p className="text-2xl font-bold text-muted-foreground">
                {clusters.filter(c => c.status === "stopped").length}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-4 pb-3">
              <p className="text-xs text-muted-foreground">{t("dashboard.totalShards")}</p>
              <p className="text-2xl font-bold">
                {clusters.reduce((acc, c) => acc + c.shards.length, 0)}
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      {loading ? (
        <p className="text-muted-foreground">{t("common.loading")}</p>
      ) : clusters.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-muted-foreground mb-4">{t("dashboard.noClusters")}</p>
          <Button onClick={() => navigate("/clusters/new")}>
            {t("dashboard.createFirst")}
          </Button>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {clusters.map((cluster) => (
            <ClusterCard
              key={cluster.id}
              cluster={cluster}
              onAction={fetchClusters}
            />
          ))}
        </div>
      )}
    </div>
  )
}
