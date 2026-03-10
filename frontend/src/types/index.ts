export type ClusterStatus = "stopped" | "running" | "starting" | "error"

export interface Cluster {
  id: string
  name: string
  dirName: string
  status: ClusterStatus
  shards: Shard[]
  config: ClusterConfig
  createdAt: string
}

export interface Shard {
  name: string
  containerId: string
  status: ClusterStatus
  config: ShardConfig
}

export interface ClusterConfig {
  gameMode: string
  maxPlayers: number
  pvp: boolean
  clusterName: string
  clusterDescription: string
  clusterPassword: string
  token: string
}

export interface ShardConfig {
  isMaster: boolean
  serverPort: number
  masterPort: number
}

export interface Mod {
  workshopId: string
  name: string
  enabled: boolean
  config: Record<string, unknown>
}

export interface ContainerStats {
  cpuPercent: number
  memUsageMb: number
  memLimitMb: number
  timestamp: number
}
