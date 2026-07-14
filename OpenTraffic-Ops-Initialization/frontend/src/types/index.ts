// 用户相关类型
export interface User {
  id: string
  username: string
  role: 'admin' | 'user'
  created_at: string
  updated_at: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: User
}

// 组件相关类型
export type ComponentType = 'postgresql' | 'redis' | 'nats' | 'nginx' | 'app' | 'frontend'
export type ComponentStatus = 'installing' | 'running' | 'stopped' | 'error'

export interface ComponentConfig {
  [key: string]: any
  port?: string
  env?: { [key: string]: string }
  volumes?: string[]
  command?: string[]
  network?: string
}

export interface Component {
  id: string
  name: string
  type: ComponentType
  image: string
  version?: string
  status: ComponentStatus
  config: ComponentConfig
  container_id?: string
  image_source?: 'pull' | 'upload'
  created_at: string
  updated_at: string
}

export interface ComponentStats {
  cpu_usage: number
  memory_usage: number
  memory_limit: number
  network_rx: number
  network_tx: number
  block_read: number
  block_write: number
}

export interface ComponentCatalogItem {
  type: ComponentType
  name: string
  description: string
  default_image: string
  default_version: string
  default_port: string
  embedded_image: string
  default_config?: ComponentConfig
  installed: boolean
  component_id?: string
  status?: ComponentStatus | 'unknown'
  container_state?: string
  container_id?: string
  docker_available: boolean
  docker_error?: string
}

// 监控相关类型
export interface Overview {
  total_components: number
  running_components: number
  stopped_components: number
  error_components: number
  components_by_type: { [key: string]: number }
}

export interface ComponentDetail {
  component: Component
  info?: ContainerInfo
  stats?: ComponentStats
}

export interface ContainerInfo {
  id: string
  name: string
  image: string
  status: string
  state: string
  running: boolean
  paused: boolean
  restarting: boolean
  oom_killed: boolean
  dead: boolean
  pid: number
  created: string
}

// 服务器相关类型
export interface Server {
  id: string
  name: string
  host: string
  port: number
  username: string
  auth_type: 'password' | 'key'
  deploy_path: string
  description: string
  created_at: string
  updated_at: string
}

export interface CreateServerRequest {
  name: string
  host: string
  port: number
  username: string
  auth_type: 'password' | 'key'
  password?: string
  private_key?: string
  passphrase?: string
  deploy_path: string
  description?: string
}

export interface DeployRequest {
  server_id: string
  binary_name: 'opentraffic-ops-proxy' | 'opentraffic-ops' | 'algo_md'
  version?: string
  config_content?: string
}

export interface DeployRecord {
  id: number
  server_id: string
  server_name: string
  binary_name: string
  remote_path: string
  status: 'pending' | 'success' | 'failed'
  log: string
  created_at: string
}

export type ServiceStatus = 'running' | 'stopped' | 'unknown'

export interface ServerServiceStatus {
  software: string
  status: ServiceStatus
  label: string
}

// 配置相关类型（已随 ConfigTemplate 移除而清空，后续可在此扩展组件配置相关类型）
