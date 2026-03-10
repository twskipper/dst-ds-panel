import { useEffect, useState, useCallback } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardHeader, CardTitle } from "@/components/ui/card"
import { api } from "@/lib/api"

interface AdminListProps {
  clusterId: string
}

export function AdminList({ clusterId }: AdminListProps) {
  const [admins, setAdmins] = useState<string[]>([])
  const [newAdmin, setNewAdmin] = useState("")
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState("")

  const loadAdmins = useCallback(async () => {
    try {
      const data = await api.readFile(clusterId, "adminlist.txt")
      const lines = data.content
        .split("\n")
        .map((l: string) => l.trim())
        .filter((l: string) => l && !l.startsWith("#"))
      setAdmins(lines)
    } catch {
      setAdmins([])
    }
  }, [clusterId])

  useEffect(() => {
    loadAdmins()
  }, [loadAdmins])

  const saveAdmins = async (list: string[]) => {
    setSaving(true)
    setMessage("")
    try {
      await api.writeFile(clusterId, "adminlist.txt", list.join("\n") + "\n")
      setMessage("Saved. Restart server to apply.")
    } catch (err) {
      setMessage(err instanceof Error ? err.message : "Save failed")
    } finally {
      setSaving(false)
    }
  }

  const addAdmin = () => {
    const id = newAdmin.trim()
    if (!id || admins.includes(id)) return
    const updated = [...admins, id]
    setAdmins(updated)
    setNewAdmin("")
    saveAdmins(updated)
  }

  const removeAdmin = (id: string) => {
    const updated = admins.filter((a) => a !== id)
    setAdmins(updated)
    saveAdmins(updated)
  }

  return (
    <div className="space-y-4">
      <div className="flex gap-2">
        <Input
          placeholder="Klei User ID (e.g. KU_abc123)"
          value={newAdmin}
          onChange={(e) => setNewAdmin(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && addAdmin()}
        />
        <Button onClick={addAdmin} disabled={saving}>
          Add Admin
        </Button>
      </div>

      <p className="text-xs text-muted-foreground">
        Find your Klei User ID in-game: press ~ and run{" "}
        <code className="bg-muted px-1 rounded">TheNet:GetUserID()</code>
        , or check your profile on the{" "}
        <a
          href="https://accounts.klei.com/account/info"
          target="_blank"
          rel="noopener noreferrer"
          className="underline text-primary"
        >
          Klei Account page
        </a>
      </p>

      {message && (
        <p className="text-sm text-green-600">{message}</p>
      )}

      {admins.length === 0 ? (
        <p className="text-sm text-muted-foreground py-2">
          No admins configured. Add Klei User IDs above.
        </p>
      ) : (
        <div className="space-y-2">
          {admins.map((admin) => (
            <Card key={admin}>
              <CardHeader className="py-2 px-4">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-sm font-mono">{admin}</CardTitle>
                  <Button
                    size="sm"
                    variant="ghost"
                    className="text-destructive"
                    onClick={() => removeAdmin(admin)}
                    disabled={saving}
                  >
                    Remove
                  </Button>
                </div>
              </CardHeader>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
