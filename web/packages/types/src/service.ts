export interface ServicePort {
  name: string
  protocol: string
  port: number
  targetPort: string
}

export interface K8sService {
  name: string
  namespace: string
  type: string
  clusterIP: string
  ports: ServicePort[]
  selector?: Record<string, string>
  replicas: number
  image?: string
}

export interface ListServicesResponse {
  services: K8sService[]
  count: number
}
