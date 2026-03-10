import { useEffect, useState, useCallback } from "react"
import Editor from "@monaco-editor/react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { Label } from "@/components/ui/label"
import { api } from "@/lib/api"

interface FileInfo {
  path: string
  exists: string
  lang: string
}

interface FileEditorProps {
  clusterId: string
}

function langFromPath(path: string): string {
  if (path.endsWith(".lua")) return "lua"
  if (path.endsWith(".ini")) return "ini"
  return "plaintext"
}

export function FileEditor({ clusterId }: FileEditorProps) {
  const [files, setFiles] = useState<FileInfo[]>([])
  const [selectedFile, setSelectedFile] = useState("")
  const [content, setContent] = useState("")
  const [originalContent, setOriginalContent] = useState("")
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null)
  const [showNewFile, setShowNewFile] = useState(false)
  const [newFileName, setNewFileName] = useState("")
  const [newFileDir, setNewFileDir] = useState(".")

  const currentLang = langFromPath(selectedFile)
  const hasChanges = content !== originalContent

  // Load file list from backend
  const loadFileList = useCallback(async () => {
    try {
      const data = await api.listFiles(clusterId)
      setFiles(data)
      if (!selectedFile && data.length > 0) {
        setSelectedFile(data[0].path)
      }
    } catch {
      // fallback to empty
    }
  }, [clusterId, selectedFile])

  useEffect(() => {
    loadFileList()
  }, [loadFileList])

  const loadFile = useCallback(async (path: string) => {
    if (!path) return
    setLoading(true)
    setMessage(null)
    try {
      const data = await api.readFile(clusterId, path)
      setContent(data.content)
      setOriginalContent(data.content)
    } catch (err) {
      setMessage({ type: "error", text: err instanceof Error ? err.message : "Failed to load file" })
    } finally {
      setLoading(false)
    }
  }, [clusterId])

  useEffect(() => {
    if (selectedFile) loadFile(selectedFile)
  }, [selectedFile, loadFile])

  const handleSave = async () => {
    setSaving(true)
    setMessage(null)
    try {
      await api.writeFile(clusterId, selectedFile, content)
      setOriginalContent(content)
      setMessage({ type: "success", text: "Saved. Restart server to apply changes." })
      loadFileList()
    } catch (err) {
      setMessage({ type: "error", text: err instanceof Error ? err.message : "Failed to save" })
    } finally {
      setSaving(false)
    }
  }

  const handleCreateFile = () => {
    if (!newFileName.trim()) return
    const path = newFileDir === "." ? newFileName.trim() : `${newFileDir}/${newFileName.trim()}`
    // Add to list and select
    if (!files.find(f => f.path === path)) {
      setFiles(prev => [...prev, { path, exists: "false", lang: langFromPath(path) }])
    }
    setSelectedFile(path)
    setContent("")
    setOriginalContent("")
    setShowNewFile(false)
    setNewFileName("")
  }

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Select value={selectedFile} onValueChange={(v) => v && setSelectedFile(v)}>
            <SelectTrigger className="w-[300px]">
              <SelectValue placeholder="Select a file..." />
            </SelectTrigger>
            <SelectContent>
              {files.map((f) => (
                <SelectItem key={f.path} value={f.path}>
                  {f.path}
                  {f.exists === "false" && " (new)"}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Badge variant="outline">{currentLang}</Badge>
          {hasChanges && (
            <Badge variant="secondary">Unsaved</Badge>
          )}
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowNewFile(true)}
          >
            New File
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => loadFile(selectedFile)}
            disabled={loading}
          >
            Reload
          </Button>
          <Button
            size="sm"
            onClick={handleSave}
            disabled={saving || !hasChanges}
          >
            {saving ? "Saving..." : "Save"}
          </Button>
        </div>
      </div>

      {message && (
        <p className={`text-sm ${message.type === "error" ? "text-destructive" : "text-green-600"}`}>
          {message.text}
        </p>
      )}

      <div className="border rounded-md overflow-hidden">
        {loading ? (
          <div className="h-[500px] flex items-center justify-center text-muted-foreground">
            Loading...
          </div>
        ) : (
          <Editor
            height="500px"
            language={currentLang}
            value={content}
            onChange={(v) => setContent(v || "")}
            theme="vs-dark"
            options={{
              minimap: { enabled: false },
              fontSize: 13,
              lineNumbers: "on",
              scrollBeyondLastLine: false,
              wordWrap: "on",
              tabSize: 4,
              automaticLayout: true,
            }}
          />
        )}
      </div>

      <Dialog open={showNewFile} onOpenChange={setShowNewFile}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create New File</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>Directory</Label>
              <Select value={newFileDir} onValueChange={(v) => v && setNewFileDir(v)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value=".">Root (cluster level)</SelectItem>
                  <SelectItem value="Master">Master</SelectItem>
                  <SelectItem value="Caves">Caves</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>File Name</Label>
              <Input
                value={newFileName}
                onChange={(e) => setNewFileName(e.target.value)}
                placeholder="e.g. myconfig.lua"
              />
              <p className="text-xs text-muted-foreground">
                Allowed extensions: .lua, .ini, .txt
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowNewFile(false)}>Cancel</Button>
            <Button onClick={handleCreateFile} disabled={!newFileName.trim()}>Create</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
