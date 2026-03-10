import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { api } from "@/lib/api"
import type { ClusterConfig } from "@/types"

export function ClusterCreatePage() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState("")
  const [name, setName] = useState("")
  const [config, setConfig] = useState<ClusterConfig>({
    gameMode: "survival",
    maxPlayers: 6,
    pvp: false,
    clusterName: "",
    clusterDescription: "",
    clusterPassword: "",
    token: "",
  })
  const [enableCaves, setEnableCaves] = useState(true)
  const [importFile, setImportFile] = useState<File | null>(null)
  const [importName, setImportName] = useState("")

  const handleCreate = async () => {
    if (!name) {
      setError("Name is required")
      return
    }
    setLoading(true)
    setError("")
    try {
      const cluster = await api.createCluster({
        name,
        config: { ...config, clusterName: config.clusterName || name },
        enableCaves,
      })
      navigate(`/clusters/${cluster.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create cluster")
    } finally {
      setLoading(false)
    }
  }

  const handleImport = async () => {
    if (!importFile) {
      setError("Please select a zip file")
      return
    }
    setLoading(true)
    setError("")
    try {
      const cluster = await api.importCluster(importName || importFile.name, importFile)
      navigate(`/clusters/${cluster.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to import cluster")
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-2xl mx-auto">
      <Button variant="ghost" onClick={() => navigate("/")} className="mb-4">
        &larr; Back
      </Button>

      <Tabs defaultValue="new">
        <TabsList>
          <TabsTrigger value="new">New Cluster</TabsTrigger>
          <TabsTrigger value="import">Import</TabsTrigger>
        </TabsList>

        <TabsContent value="new">
          <Card>
            <CardHeader>
              <CardTitle>Create New Cluster</CardTitle>
              <CardDescription>
                Set up a new DST dedicated server cluster
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="name">Cluster Name</Label>
                <Input
                  id="name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="My DST Server"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="clusterName">Server Display Name</Label>
                <Input
                  id="clusterName"
                  value={config.clusterName}
                  onChange={(e) =>
                    setConfig({ ...config, clusterName: e.target.value })
                  }
                  placeholder="Shown in server browser"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="description">Description</Label>
                <Input
                  id="description"
                  value={config.clusterDescription}
                  onChange={(e) =>
                    setConfig({ ...config, clusterDescription: e.target.value })
                  }
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>Game Mode</Label>
                  <Select
                    value={config.gameMode}
                    onValueChange={(v) =>
                      setConfig({ ...config, gameMode: v ?? "survival" })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="survival">Survival</SelectItem>
                      <SelectItem value="endless">Endless</SelectItem>
                      <SelectItem value="wilderness">Wilderness</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="maxPlayers">Max Players</Label>
                  <Input
                    id="maxPlayers"
                    type="number"
                    min={1}
                    max={64}
                    value={config.maxPlayers}
                    onChange={(e) =>
                      setConfig({
                        ...config,
                        maxPlayers: parseInt(e.target.value) || 6,
                      })
                    }
                  />
                </div>
              </div>

              <div className="flex items-center gap-4">
                <div className="flex items-center gap-2">
                  <Switch
                    checked={config.pvp}
                    onCheckedChange={(v) => setConfig({ ...config, pvp: v })}
                  />
                  <Label>PVP</Label>
                </div>
                <div className="flex items-center gap-2">
                  <Switch
                    checked={enableCaves}
                    onCheckedChange={(v) => setEnableCaves(v)}
                  />
                  <Label>Enable Caves</Label>
                </div>
              </div>

              <p className="text-xs text-muted-foreground">
                New clusters use default Survival world settings. For custom world settings, use the Import tab to upload your own save.
              </p>

              <div className="space-y-2">
                <Label htmlFor="password">Server Password (optional)</Label>
                <Input
                  id="password"
                  type="password"
                  value={config.clusterPassword}
                  onChange={(e) =>
                    setConfig({ ...config, clusterPassword: e.target.value })
                  }
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="token">Cluster Token</Label>
                <Input
                  id="token"
                  value={config.token}
                  onChange={(e) =>
                    setConfig({ ...config, token: e.target.value })
                  }
                  placeholder="Paste your cluster token from Klei"
                />
                <p className="text-xs text-muted-foreground">
                  Get your token from{" "}
                  <a
                    href="https://accounts.klei.com/account/game/servers?game=DontStarveTogether"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="underline text-primary"
                  >
                    Klei Account &rarr; Game Servers
                  </a>
                  . Click "Add New Server", copy the token. Or in-game press ~ and run{" "}
                  <code className="bg-muted px-1 rounded text-xs">TheNet:GenerateClusterToken()</code>
                </p>
              </div>

              {error && <p className="text-destructive text-sm">{error}</p>}

              <Button onClick={handleCreate} disabled={loading} className="w-full">
                {loading ? "Creating..." : "Create Cluster"}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="import">
          <Card>
            <CardHeader>
              <CardTitle>Import Cluster</CardTitle>
              <CardDescription>
                Import an existing cluster save from a zip file
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="importName">Name</Label>
                <Input
                  id="importName"
                  value={importName}
                  onChange={(e) => setImportName(e.target.value)}
                  placeholder="Optional name for imported cluster"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="importFile">Cluster Zip File</Label>
                <Input
                  id="importFile"
                  type="file"
                  accept=".zip"
                  onChange={(e) =>
                    setImportFile(e.target.files?.[0] || null)
                  }
                />
              </div>

              {error && <p className="text-destructive text-sm">{error}</p>}

              <Button onClick={handleImport} disabled={loading} className="w-full">
                {loading ? "Importing..." : "Import Cluster"}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
