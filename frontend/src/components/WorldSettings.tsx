import { useEffect, useState, useCallback } from "react"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Separator } from "@/components/ui/separator"
import { api } from "@/lib/api"

// Types for the external config
interface OptionDef {
  value: string
  label: string
}

interface SettingDef {
  key: string
  label: string
  type: string
  locked?: boolean
}

interface CategoryDef {
  name: string
  settings: SettingDef[]
}

interface WorldSettingsConfig {
  optionTypes: Record<string, OptionDef[]>
  categories: CategoryDef[]
  presets: Record<string, Record<string, string>>
}

// Cache the loaded config
let cachedConfig: WorldSettingsConfig | null = null

async function loadConfig(): Promise<WorldSettingsConfig> {
  if (cachedConfig) return cachedConfig
  const res = await fetch("/world-settings.json")
  cachedConfig = await res.json()
  return cachedConfig!
}

// Parse overrides from leveldataoverride.lua content
function parseOverrides(content: string): Record<string, string> {
  const overrides: Record<string, string> = {}
  const overridesMatch = content.match(/overrides\s*=\s*\{([\s\S]*?)\}/)
  if (!overridesMatch) return overrides

  const block = overridesMatch[1]
  const entryRegex = /(\w+)\s*=\s*"?([^",\n]+)"?/g
  let m
  while ((m = entryRegex.exec(block)) !== null) {
    overrides[m[1]] = m[2].trim()
  }
  return overrides
}

function applyOverrides(original: string, overrides: Record<string, string>): string {
  let content = original
  if (!content.includes("overrides=") && !content.includes("overrides =")) {
    return content
  }
  for (const [key, value] of Object.entries(overrides)) {
    const regex = new RegExp(`(${key}\\s*=\\s*)"?[^",\\n]+"?`, "g")
    if (regex.test(content)) {
      content = content.replace(
        new RegExp(`(${key}\\s*=\\s*)"?[^",\\n]+"?`),
        typeof value === "string" && value !== "true" && value !== "false"
          ? `${key}="${value}"`
          : `${key}=${value}`
      )
    }
  }
  return content
}

interface WorldSettingsProps {
  clusterId: string
  shard: "Master" | "Caves"
}

export function WorldSettings({ clusterId, shard }: WorldSettingsProps) {
  const [config, setConfig] = useState<WorldSettingsConfig | null>(null)
  const [overrides, setOverrides] = useState<Record<string, string>>({})
  const [originalContent, setOriginalContent] = useState("")
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null)

  const filePath = `${shard}/leveldataoverride.lua`

  const loadSettings = useCallback(async () => {
    setLoading(true)
    try {
      const [cfg, fileData] = await Promise.all([
        loadConfig(),
        api.readFile(clusterId, filePath).catch(() =>
          api.readFile(clusterId, `${shard}/worldgenoverride.lua`).catch(() => ({ content: "", path: "" }))
        ),
      ])
      setConfig(cfg)
      setOriginalContent(fileData.content)
      setOverrides(parseOverrides(fileData.content))
    } catch {
      setOverrides({})
    } finally {
      setLoading(false)
    }
  }, [clusterId, shard, filePath])

  useEffect(() => {
    loadSettings()
  }, [loadSettings])

  const handleChange = (key: string, value: string) => {
    setOverrides((prev) => ({ ...prev, [key]: value }))
  }

  const applyPreset = (preset: Record<string, string>) => {
    setOverrides((prev) => ({ ...prev, ...preset }))
  }

  const handleSave = async () => {
    setSaving(true)
    setMessage(null)
    try {
      const updated = applyOverrides(originalContent, overrides)
      await api.writeFile(clusterId, filePath, updated)
      setOriginalContent(updated)
      setMessage({ type: "success", text: "Saved. Restart server to apply." })
    } catch (err) {
      setMessage({ type: "error", text: err instanceof Error ? err.message : "Save failed" })
    } finally {
      setSaving(false)
    }
  }

  if (loading || !config) return <p className="text-sm text-muted-foreground">Loading...</p>

  if (!originalContent) {
    return <p className="text-sm text-muted-foreground">No world settings file found for {shard}.</p>
  }

  return (
    <div className="space-y-4">
      {message && (
        <Alert variant={message.type === "error" ? "destructive" : "default"}>
          <AlertDescription>{message.text}</AlertDescription>
        </Alert>
      )}

      {/* Presets */}
      <div>
        <h3 className="font-medium mb-2">Difficulty Presets</h3>
        <div className="flex flex-wrap gap-2">
          {Object.entries(config.presets).map(([name, preset]) => (
            <Button
              key={name}
              size="sm"
              variant={name === "challenge" ? "destructive" : "outline"}
              onClick={() => {
                if (name === "challenge" && !confirm("Apply Challenge preset? This makes the world very difficult.")) return
                applyPreset(preset)
              }}
            >
              {name.charAt(0).toUpperCase() + name.slice(1)}
            </Button>
          ))}
        </div>
        <Separator className="mt-3" />
      </div>

      {/* Categories */}
      {config.categories.map((cat) => (
        <div key={cat.name}>
          <h3 className="font-medium mb-2">{cat.name}</h3>
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
            {cat.settings.map((setting) => {
              const options = config.optionTypes[setting.type] || []
              return (
                <div key={setting.key} className="space-y-1">
                  <Label className="text-xs">
                    {setting.label}
                    {setting.locked && (
                      <span className="text-muted-foreground ml-1" title="Cannot change after world generation">
                        (new world only)
                      </span>
                    )}
                  </Label>
                  <Select
                    value={overrides[setting.key] || "default"}
                    onValueChange={(v) => v && handleChange(setting.key, v)}
                  >
                    <SelectTrigger className="h-8 text-xs">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {options.map((opt) => (
                        <SelectItem key={opt.value} value={opt.value}>
                          {opt.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              )
            })}
          </div>
          <Separator className="mt-3" />
        </div>
      ))}

      <Button onClick={handleSave} disabled={saving}>
        {saving ? "Saving..." : "Save World Settings"}
      </Button>
    </div>
  )
}
