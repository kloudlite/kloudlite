// Mapping of related posts for each blog article
// This defines the logical relationships between blog posts

export const relatedPostsMap: Record<string, string[]> = {
  // Core Features
  'environment-forking': ['workspace-forking', 'environment-snapshots', 'clone-fork-environments', 'environment-switching'],
  'workspace-forking': ['environment-forking', 'workspace-snapshots', 'ai-ready-workspaces', 'team-collaboration'],
  'service-intercepts': ['real-services-vs-mocks', 'faster-feedback-loops', 'live-monitoring', 'private-network-access'],
  'nix-package-management': ['docker-compose-compatible', 'workspace-snapshots', 'zero-setup-development', 'kubernetes-native-architecture'],
  'environment-switching': ['environment-forking', 'workspace-environment-connections', 'network-isolation', 'team-collaboration'],
  'environment-snapshots': ['environment-forking', 'workspace-snapshots', 'clone-fork-environments', 'instant-deployment'],
  'workspace-snapshots': ['environment-snapshots', 'workspace-forking', 'team-collaboration', 'nix-package-management'],
  'ai-ready-workspaces': ['workspace-forking', 'ide-integration', 'gpu-enabled', 'zero-configuration'],
  'ide-integration': ['ai-ready-workspaces', 'workspace-environment-connections', 'private-network-access', 'vpn-gateway'],
  'docker-compose-compatible': ['nix-package-management', 'zero-setup-development', 'network-isolation', 'kubernetes-native-architecture'],
  'network-isolation': ['docker-compose-compatible', 'environment-switching', 'private-network-access', 'team-collaboration'],
  'team-collaboration': ['workspace-forking', 'environment-forking', 'network-isolation', 'workspace-environment-connections'],

  // Infrastructure & Platform
  'vpn-gateway': ['private-network-access', 'ide-integration', 'workspace-environment-connections', 'network-isolation'],
  'compute-storage': ['flexible-resources', 'auto-stop', 'performance-monitoring', 'high-availability'],
  'auto-stop': ['compute-storage', 'flexible-resources', 'performance-monitoring', 'sub-30s-startup'],
  'flexible-resources': ['compute-storage', 'gpu-enabled', 'performance-monitoring', 'auto-stop'],
  'gpu-enabled': ['flexible-resources', 'ai-ready-workspaces', 'performance-monitoring', 'compute-storage'],
  'performance-monitoring': ['compute-storage', 'live-monitoring', 'flexible-resources', 'high-availability'],
  'high-availability': ['compute-storage', 'performance-monitoring', 'instant-deployment', 'sub-30s-startup'],
  'sub-30s-startup': ['auto-stop', 'instant-deployment', 'speed-above-all', 'zero-setup-development'],

  // Development Workflow
  'zero-setup-development': ['distributed-apps-localhost-problem', 'docker-compose-compatible', 'sub-30s-startup', 'zero-configuration'],
  'real-services-vs-mocks': ['mocks-dont-match-production', 'service-intercepts', 'faster-feedback-loops', 'cloud-dev-environments-solution'],
  'faster-feedback-loops': ['real-services-vs-mocks', 'service-intercepts', 'live-monitoring', 'instant-deployment'],
  'private-network-access': ['vpn-gateway', 'network-isolation', 'workspace-environment-connections', 'ide-integration'],
  'workspace-environment-connections': ['private-network-access', 'environment-switching', 'service-intercepts', 'team-collaboration'],
  'clone-fork-environments': ['environment-forking', 'environment-snapshots', 'team-collaboration', 'network-isolation'],
  'live-monitoring': ['performance-monitoring', 'faster-feedback-loops', 'service-intercepts', 'instant-deployment'],
  'instant-deployment': ['sub-30s-startup', 'live-monitoring', 'faster-feedback-loops', 'high-availability'],

  // Problem/Solution Narrative
  'distributed-apps-localhost-problem': ['mocks-dont-match-production', 'cloud-dev-environments-solution', 'zero-setup-development', 'real-services-vs-mocks'],
  'mocks-dont-match-production': ['distributed-apps-localhost-problem', 'cloud-dev-environments-solution', 'real-services-vs-mocks', 'service-intercepts'],
  'cloud-dev-environments-solution': ['distributed-apps-localhost-problem', 'mocks-dont-match-production', 'zero-setup-development', 'kubernetes-native-architecture'],

  // Philosophy & Values
  'speed-above-all': ['sub-30s-startup', 'instant-deployment', 'zero-configuration', 'faster-feedback-loops'],
  'zero-configuration': ['zero-setup-development', 'speed-above-all', 'docker-compose-compatible', 'open-by-default'],
  'open-by-default': ['kubernetes-native-architecture', 'zero-configuration', 'service-mesh-integration', 'custom-resource-definitions'],

  // Architecture Deep Dive
  'kubernetes-native-architecture': ['custom-resource-definitions', 'kubernetes-controllers', 'service-mesh-integration', 'open-by-default'],
  'custom-resource-definitions': ['kubernetes-native-architecture', 'kubernetes-controllers', 'environment-forking', 'service-intercepts'],
  'kubernetes-controllers': ['kubernetes-native-architecture', 'custom-resource-definitions', 'service-mesh-integration', 'docker-compose-compatible'],
  'service-mesh-integration': ['kubernetes-native-architecture', 'service-intercepts', 'kubernetes-controllers', 'network-isolation'],
}
