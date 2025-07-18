// AWS Regions configuration for team creation and management

export interface Region {
  value: string
  label: string
  flag: string
  continent: 'North America' | 'Europe' | 'Asia Pacific' | 'South America'
}

export const AWS_REGIONS: Region[] = [
  // North America
  { value: 'us-east-1', label: 'US East (N. Virginia)', flag: '🇺🇸', continent: 'North America' },
  { value: 'us-east-2', label: 'US East (Ohio)', flag: '🇺🇸', continent: 'North America' },
  { value: 'us-west-1', label: 'US West (N. California)', flag: '🇺🇸', continent: 'North America' },
  { value: 'us-west-2', label: 'US West (Oregon)', flag: '🇺🇸', continent: 'North America' },
  { value: 'ca-central-1', label: 'Canada (Central)', flag: '🇨🇦', continent: 'North America' },
  
  // Europe
  { value: 'eu-central-1', label: 'Europe (Frankfurt)', flag: '🇩🇪', continent: 'Europe' },
  { value: 'eu-west-1', label: 'Europe (Ireland)', flag: '🇮🇪', continent: 'Europe' },
  { value: 'eu-west-2', label: 'Europe (London)', flag: '🇬🇧', continent: 'Europe' },
  { value: 'eu-west-3', label: 'Europe (Paris)', flag: '🇫🇷', continent: 'Europe' },
  { value: 'eu-north-1', label: 'Europe (Stockholm)', flag: '🇸🇪', continent: 'Europe' },
  
  // Asia Pacific
  { value: 'ap-south-1', label: 'Asia Pacific (Mumbai)', flag: '🇮🇳', continent: 'Asia Pacific' },
  { value: 'ap-southeast-1', label: 'Asia Pacific (Singapore)', flag: '🇸🇬', continent: 'Asia Pacific' },
  { value: 'ap-southeast-2', label: 'Asia Pacific (Sydney)', flag: '🇦🇺', continent: 'Asia Pacific' },
  { value: 'ap-northeast-1', label: 'Asia Pacific (Tokyo)', flag: '🇯🇵', continent: 'Asia Pacific' },
  { value: 'ap-northeast-2', label: 'Asia Pacific (Seoul)', flag: '🇰🇷', continent: 'Asia Pacific' },
  
  // South America
  { value: 'sa-east-1', label: 'South America (São Paulo)', flag: '🇧🇷', continent: 'South America' },
]

// Group regions by continent for better organization
export const REGIONS_BY_CONTINENT = AWS_REGIONS.reduce((acc, region) => {
  if (!acc[region.continent]) {
    acc[region.continent] = []
  }
  acc[region.continent].push(region)
  return acc
}, {} as Record<string, Region[]>)

// Helper functions
export function getRegionByValue(value: string): Region | undefined {
  return AWS_REGIONS.find(region => region.value === value)
}

export function getRegionLabel(value: string): string {
  const region = getRegionByValue(value)
  return region ? `${region.flag} ${region.label}` : value
}

// Default region for new teams
export const DEFAULT_REGION = 'us-east-1'