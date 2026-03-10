import { useState } from "react"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import type { Mod } from "@/types"

interface ModConfigDialogProps {
  mod: Mod
  open: boolean
  onClose: () => void
  onSave: (config: Record<string, unknown>) => void
}

export function ModConfigDialog({ mod, open, onClose, onSave }: ModConfigDialogProps) {
  const [entries, setEntries] = useState<[string, string][]>(() =>
    Object.entries(mod.config || {}).map(([k, v]) => [k, String(v)])
  )

  const addEntry = () => {
    setEntries([...entries, ["", ""]])
  }

  const updateEntry = (index: number, field: 0 | 1, value: string) => {
    const updated = [...entries]
    updated[index] = [...updated[index]] as [string, string]
    updated[index][field] = value
    setEntries(updated)
  }

  const removeEntry = (index: number) => {
    setEntries(entries.filter((_, i) => i !== index))
  }

  const handleSave = () => {
    const config: Record<string, unknown> = {}
    for (const [key, val] of entries) {
      if (!key.trim()) continue
      // Try to parse as number or boolean
      if (val === "true") config[key] = true
      else if (val === "false") config[key] = false
      else if (!isNaN(Number(val)) && val.trim() !== "") config[key] = Number(val)
      else config[key] = val
    }
    onSave(config)
  }

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>
            Configure: {mod.name || mod.workshopId}
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-3 max-h-[400px] overflow-y-auto">
          {entries.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              No configuration options. Add key-value pairs below.
            </p>
          ) : (
            entries.map(([key, val], i) => (
              <div key={i} className="flex gap-2 items-end">
                <div className="flex-1 space-y-1">
                  <Label className="text-xs">Key</Label>
                  <Input
                    value={key}
                    onChange={(e) => updateEntry(i, 0, e.target.value)}
                    placeholder="option_name"
                  />
                </div>
                <div className="flex-1 space-y-1">
                  <Label className="text-xs">Value</Label>
                  <Input
                    value={val}
                    onChange={(e) => updateEntry(i, 1, e.target.value)}
                    placeholder="value"
                  />
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-destructive"
                  onClick={() => removeEntry(i)}
                >
                  X
                </Button>
              </div>
            ))
          )}
          <Button variant="outline" size="sm" onClick={addEntry}>
            + Add Option
          </Button>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={handleSave}>Save</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
