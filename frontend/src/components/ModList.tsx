import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Switch } from "@/components/ui/switch"
import { Card, CardHeader, CardTitle } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { ModConfigDialog } from "@/components/ModConfigDialog"
import { api } from "@/lib/api"
import type { Mod } from "@/types"

interface ModListProps {
  clusterId: string
  mods: Mod[]
  onUpdate: (mods: Mod[]) => void
}

export function ModList({ clusterId, mods, onUpdate }: ModListProps) {
  const [newWorkshopId, setNewWorkshopId] = useState("")
  const [saving, setSaving] = useState(false)
  const [editingMod, setEditingMod] = useState<Mod | null>(null)

  const addMod = () => {
    if (!newWorkshopId.trim()) return
    const exists = mods.some((m) => m.workshopId === newWorkshopId.trim())
    if (exists) return

    const updated = [
      ...mods,
      {
        workshopId: newWorkshopId.trim(),
        name: `Workshop-${newWorkshopId.trim()}`,
        enabled: true,
        config: {},
      },
    ]
    onUpdate(updated)
    setNewWorkshopId("")
  }

  const toggleMod = (workshopId: string) => {
    const updated = mods.map((m) =>
      m.workshopId === workshopId ? { ...m, enabled: !m.enabled } : m
    )
    onUpdate(updated)
  }

  const removeMod = (workshopId: string) => {
    onUpdate(mods.filter((m) => m.workshopId !== workshopId))
  }

  const updateModConfig = (workshopId: string, config: Record<string, unknown>) => {
    const updated = mods.map((m) =>
      m.workshopId === workshopId ? { ...m, config } : m
    )
    onUpdate(updated)
    setEditingMod(null)
  }

  const saveMods = async () => {
    setSaving(true)
    try {
      const saved = await api.updateMods(clusterId, mods)
      onUpdate(saved)
    } catch (err) {
      console.error("Failed to save mods:", err)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex gap-2">
        <Input
          placeholder="Workshop ID (e.g. 378160973)"
          value={newWorkshopId}
          onChange={(e) => setNewWorkshopId(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && addMod()}
        />
        <Button onClick={addMod}>Add Mod</Button>
      </div>

      {mods.length === 0 ? (
        <p className="text-muted-foreground text-sm py-4">
          No mods configured. Add a mod by Workshop ID.
        </p>
      ) : (
        <div className="space-y-2">
          {mods.map((mod) => (
            <Card key={mod.workshopId}>
              <CardHeader className="py-3 px-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <Switch
                      checked={mod.enabled}
                      onCheckedChange={() => toggleMod(mod.workshopId)}
                    />
                    <div>
                      <CardTitle className="text-sm">
                        {mod.name || mod.workshopId}
                      </CardTitle>
                      <p className="text-xs text-muted-foreground">
                        ID: {mod.workshopId}
                      </p>
                    </div>
                  </div>
                  <div className="flex gap-1">
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => setEditingMod(mod)}
                    >
                      Config
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      className="text-destructive"
                      onClick={() => removeMod(mod.workshopId)}
                    >
                      Remove
                    </Button>
                  </div>
                </div>
              </CardHeader>
            </Card>
          ))}
        </div>
      )}

      <Separator />

      <Button onClick={saveMods} disabled={saving} className="w-full">
        {saving ? "Saving..." : "Save Mod Configuration"}
      </Button>

      {editingMod && (
        <ModConfigDialog
          mod={editingMod}
          open={true}
          onClose={() => setEditingMod(null)}
          onSave={(config) => updateModConfig(editingMod.workshopId, config)}
        />
      )}
    </div>
  )
}
