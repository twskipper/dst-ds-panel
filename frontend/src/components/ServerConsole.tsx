import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { api } from "@/lib/api"

interface ServerConsoleProps {
  clusterId: string
  shard: string
  isRunning: boolean
}

export function ServerConsole({ clusterId, shard, isRunning }: ServerConsoleProps) {
  const [command, setCommand] = useState("")
  const [announceMsg, setAnnounceMsg] = useState("")
  const [history, setHistory] = useState<Array<{ cmd: string; time: string; ok: boolean }>>([])
  const [sending, setSending] = useState(false)

  const sendCommand = async (cmd: string) => {
    if (!cmd.trim()) return
    setSending(true)
    const time = new Date().toLocaleTimeString()
    try {
      await api.sendCommand(clusterId, shard, cmd)
      setHistory((prev) => [{ cmd, time, ok: true }, ...prev].slice(0, 20))
      setCommand("")
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed"
      setHistory((prev) => [{ cmd: `${cmd} — ${msg}`, time, ok: false }, ...prev].slice(0, 20))
    } finally {
      setSending(false)
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    sendCommand(command)
  }

  const handleAnnounce = () => {
    if (!announceMsg.trim()) return
    const escaped = announceMsg.replace(/'/g, "\\'")
    sendCommand(`c_announce('${escaped}')`)
    setAnnounceMsg("")
  }

  if (!isRunning) {
    return (
      <p className="text-sm text-muted-foreground">
        Server not running. Start the server to use the console.
      </p>
    )
  }

  return (
    <div className="space-y-3">
      {/* Announce */}
      <div className="flex gap-2">
        <Input
          value={announceMsg}
          onChange={(e) => setAnnounceMsg(e.target.value)}
          placeholder="Type a message to announce to all players..."
          disabled={sending}
          onKeyDown={(e) => e.key === "Enter" && handleAnnounce()}
        />
        <Button
          onClick={handleAnnounce}
          disabled={sending || !announceMsg.trim()}
          variant="outline"
        >
          Announce
        </Button>
      </div>

      {/* Quick actions */}
      <div className="flex flex-wrap gap-2">
        <Button size="sm" variant="outline" onClick={() => sendCommand("c_save()")} disabled={sending}>
          Save
        </Button>
        <Button size="sm" variant="outline" onClick={() => sendCommand("c_rollback(1)")} disabled={sending}>
          Rollback
        </Button>
        <Button
          size="sm"
          variant="destructive"
          onClick={() => confirm("Regenerate the entire world? All progress will be lost.") && sendCommand("c_regenerateworld()")}
          disabled={sending}
        >
          Regenerate
        </Button>
        <Button
          size="sm"
          variant="destructive"
          onClick={() => confirm("Shutdown the server?") && sendCommand("c_shutdown(true)")}
          disabled={sending}
        >
          Shutdown
        </Button>
      </div>

      {/* Raw command */}
      <form onSubmit={handleSubmit} className="flex gap-2">
        <Input
          value={command}
          onChange={(e) => setCommand(e.target.value)}
          placeholder="Raw Lua command, e.g. c_countprefabs('beefalo')"
          disabled={sending}
          className="font-mono text-sm"
        />
        <Button type="submit" disabled={sending || !command.trim()} size="sm">
          Run
        </Button>
      </form>

      {/* History */}
      {history.length > 0 && (
        <div className="space-y-1 max-h-[200px] overflow-y-auto">
          {history.map((entry, i) => (
            <div
              key={i}
              className={`text-xs font-mono px-2 py-1 rounded ${
                entry.ok
                  ? "bg-muted text-muted-foreground"
                  : "bg-destructive/10 text-destructive"
              }`}
            >
              <span className="text-muted-foreground">[{entry.time}]</span>{" "}
              {entry.cmd}
            </div>
          ))}
        </div>
      )}

      <Alert>
        <AlertDescription className="text-xs">
          Commands run on the {shard} shard. Common:{" "}
          <code className="bg-muted px-1 rounded">c_save()</code>{" "}
          <code className="bg-muted px-1 rounded">c_rollback(N)</code>{" "}
          <code className="bg-muted px-1 rounded">c_countprefabs("name")</code>{" "}
          <code className="bg-muted px-1 rounded">c_listallplayers()</code>
        </AlertDescription>
      </Alert>
    </div>
  )
}
