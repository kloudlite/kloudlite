# Runtime Architecture

```mermaid
flowchart TB
  subgraph CP[Control Plane Cluster]
    PC[platform-controller\nAPI server + controllers + webhooks]
    UC[user controller]
    EC[environment controller]
    WMC[workmachine controller]
    WSC[workspace controller]
    SC[snapshot controllers]

    PC --> UC
    PC --> EC
    PC --> WMC
    PC --> WSC
    PC --> SC
  end

  subgraph WM[Per WorkMachine Namespace / Runtime]
    HM[host-manager]
    IC[wm-ingress-controller]
    DD[docker-dind]
    TS[tunnel-server]
    CA[code-analyzer]
    AUX[SSH keys / sshd_config / TLS secret / RBAC / NetworkPolicy]
  end

  subgraph WS[Per Workspace]
    WP[workspace pod\nworkspace-comprehensive]
    WSS[workspace service]
    WSH[workspace headless service]
  end

  WMC --> HM
  WMC --> IC
  WMC --> DD
  WMC --> TS
  WMC --> CA
  WMC --> AUX

  WSC --> WP
  WSC --> WSS
  WSC --> WSH

  HM --> WP
  DD --> WP
  IC --> WSS
  TS --> WSS
  TS --> WSH
```

## Placement Summary

### Control Plane Cluster

- `platform-controller`
- user controller
- environment controller
- workmachine controller
- workspace controller
- snapshot controllers
- webhook server

### Per WorkMachine Runtime

- `host-manager`
- `wm-ingress-controller`
- `docker-dind`
- `tunnel-server`
- `code-analyzer`
- support resources: network policy, TLS secret sync, SSH host keys, `sshd_config`, RBAC

### Per Workspace

- workspace pod using `workspace-comprehensive`
- workspace service
- workspace headless service
