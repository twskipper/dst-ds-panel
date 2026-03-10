import { useState } from "react"
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
import { api } from "@/lib/api"
import type { ClusterConfig } from "@/types"

interface ClusterFormProps {
  clusterId: string
  initialConfig: ClusterConfig
  onSaved: (config: ClusterConfig) => void
}

export function ClusterForm({ clusterId, initialConfig, onSaved }: ClusterFormProps) {
  const [config, setConfig] = useState<ClusterConfig>(initialConfig)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")

  const handleSave = async () => {
    setSaving(true)
    setError("")
    try {
      const updated = await api.updateClusterConfig(clusterId, config)
      onSaved(updated.config)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Save failed")
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label>Server Display Name</Label>
        <Input
          value={config.clusterName}
          onChange={(e) => setConfig({ ...config, clusterName: e.target.value })}
        />
      </div>

      <div className="space-y-2">
        <Label>Description</Label>
        <Input
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
            onValueChange={(v) => setConfig({ ...config, gameMode: v ?? "survival" })}
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
          <Label>Max Players</Label>
          <Input
            type="number"
            min={1}
            max={64}
            value={config.maxPlayers}
            onChange={(e) =>
              setConfig({ ...config, maxPlayers: parseInt(e.target.value) || 6 })
            }
          />
        </div>
      </div>

      <div className="flex items-center gap-2">
        <Switch
          checked={config.pvp}
          onCheckedChange={(v) => setConfig({ ...config, pvp: v })}
        />
        <Label>PVP</Label>
      </div>

      <div className="space-y-2">
        <Label>Password</Label>
        <Input
          type="password"
          value={config.clusterPassword}
          onChange={(e) =>
            setConfig({ ...config, clusterPassword: e.target.value })
          }
        />
      </div>

      <div className="space-y-2">
        <Label>Cluster Token</Label>
        <Input
          value={config.token}
          onChange={(e) => setConfig({ ...config, token: e.target.value })}
          placeholder="Paste from Klei account"
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

      <Button onClick={handleSave} disabled={saving}>
        {saving ? "Saving..." : "Save Configuration"}
      </Button>
    </div>
  )
}
