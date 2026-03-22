import type { Cluster, ClusterConfig, Mod } from "@/types"

const BASE = "/api"

function getToken(): string | null {
  return localStorage.getItem("dst_token")
}

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const token = getToken()
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  }
  if (token) {
    headers["Authorization"] = `Bearer ${token}`
  }
  const res = await fetch(BASE + url, {
    headers,
    ...options,
  })
  if (res.status === 401) {
    localStorage.removeItem("dst_token")
    localStorage.removeItem("dst_token_expires")
    window.location.href = "/login"
    throw new Error("Session expired")
  }
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || res.statusText)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  listClusters: () => request<Cluster[]>("/clusters"),

  getCluster: (id: string) => request<Cluster>(`/clusters/${id}`),

  createCluster: (data: { name: string; config: ClusterConfig; enableCaves?: boolean }) =>
    request<Cluster>("/clusters", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  getPlayerActivity: (id: string) =>
    request<Array<{ time: string; player: string; event: string; detail?: string }>>(
      `/clusters/${id}/players`
    ),

  cloneCluster: (id: string, name: string) =>
    request<Cluster>(`/clusters/${id}/clone`, {
      method: "POST",
      body: JSON.stringify({ name }),
    }),

  updateClusterConfig: (id: string, config: ClusterConfig) =>
    request<Cluster>(`/clusters/${id}/config`, {
      method: "PUT",
      body: JSON.stringify(config),
    }),

  deleteCluster: (id: string) =>
    request<void>(`/clusters/${id}`, { method: "DELETE" }),

  startCluster: (id: string) =>
    request<Cluster>(`/clusters/${id}/start`, { method: "POST" }),

  stopCluster: (id: string) =>
    request<Cluster>(`/clusters/${id}/stop`, { method: "POST" }),

  restartCluster: (id: string) =>
    request<Cluster>(`/clusters/${id}/restart`, { method: "POST" }),

  listMods: (id: string) => request<Mod[]>(`/clusters/${id}/mods`),

  updateMods: (id: string, mods: Mod[]) =>
    request<Mod[]>(`/clusters/${id}/mods`, {
      method: "PUT",
      body: JSON.stringify(mods),
    }),

  imageStatus: () => request<{ imageExists: boolean; dstInstalled: boolean; dstVersion: string; needsManualUpdate: boolean }>("/image/status"),

  buildImage: () => request<{ status: string; output: string }>("/image/build", { method: "POST" }),

  updateDST: (beta?: string) => request<{ status: string; output: string }>(`/dst/update${beta ? `?beta=${beta}` : ""}`, { method: "POST" }),

  sendCommand: (id: string, shard: string, command: string) =>
    request<{ status: string; command: string }>(
      `/clusters/${id}/shards/${shard}/console`,
      { method: "POST", body: JSON.stringify({ command }) }
    ),

  listFiles: (id: string) =>
    request<Array<{ path: string; exists: string; lang: string }>>(
      `/clusters/${id}/files/list`
    ),

  readFile: (id: string, path: string) =>
    request<{ content: string; path: string }>(
      `/clusters/${id}/files?path=${encodeURIComponent(path)}`
    ),

  writeFile: (id: string, path: string, content: string) =>
    request<{ status: string }>(
      `/clusters/${id}/files?path=${encodeURIComponent(path)}`,
      { method: "PUT", body: JSON.stringify({ content }) }
    ),

  importCluster: async (name: string, file: File): Promise<Cluster> => {
    const formData = new FormData()
    formData.append("file", file)
    formData.append("name", name)
    const token = getToken()
    const headers: Record<string, string> = {}
    if (token) headers["Authorization"] = `Bearer ${token}`
    const res = await fetch(BASE + "/clusters/import", {
      method: "POST",
      headers,
      body: formData,
    })
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: res.statusText }))
      throw new Error(err.error || res.statusText)
    }
    return res.json()
  },
}
