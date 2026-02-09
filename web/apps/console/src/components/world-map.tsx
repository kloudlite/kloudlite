'use client'

import { useRef, useMemo } from 'react'
import DottedMap from 'dotted-map'

// Region coordinates for all cloud providers
const REGION_COORDINATES: Record<string, { lat: number; lng: number; name: string }> = {
  // AWS Regions
  'us-east-1': { lat: 37.4316, lng: -78.6569, name: 'N. Virginia' },
  'us-east-2': { lat: 39.9612, lng: -82.9988, name: 'Ohio' },
  'us-west-1': { lat: 37.3541, lng: -121.9552, name: 'N. California' },
  'us-west-2': { lat: 45.5152, lng: -122.6784, name: 'Oregon' },
  'eu-west-1': { lat: 53.3498, lng: -6.2603, name: 'Ireland' },
  'eu-west-2': { lat: 51.5074, lng: -0.1278, name: 'London' },
  'eu-west-3': { lat: 48.8566, lng: 2.3522, name: 'Paris' },
  'eu-central-1': { lat: 50.1109, lng: 8.6821, name: 'Frankfurt' },
  'eu-north-1': { lat: 59.3293, lng: 18.0686, name: 'Stockholm' },
  'ap-south-1': { lat: 19.0760, lng: 72.8777, name: 'Mumbai' },
  'ap-southeast-1': { lat: 1.3521, lng: 103.8198, name: 'Singapore' },
  'ap-southeast-2': { lat: -33.8688, lng: 151.2093, name: 'Sydney' },
  'ap-northeast-1': { lat: 35.6762, lng: 139.6503, name: 'Tokyo' },
  'ap-northeast-2': { lat: 37.5665, lng: 126.9780, name: 'Seoul' },
  'sa-east-1': { lat: -23.5505, lng: -46.6333, name: 'São Paulo' },
  'ca-central-1': { lat: 45.5017, lng: -73.5673, name: 'Montreal' },

  // GCP Regions
  'us-central1': { lat: 41.2619, lng: -95.8608, name: 'Iowa' },
  'us-east1': { lat: 33.1960, lng: -80.0131, name: 'South Carolina' },
  'us-east4': { lat: 39.0438, lng: -77.4874, name: 'N. Virginia' },
  'us-west1': { lat: 45.5945, lng: -121.1787, name: 'Oregon' },
  'us-west2': { lat: 34.0522, lng: -118.2437, name: 'Los Angeles' },
  'us-west3': { lat: 40.7608, lng: -111.8910, name: 'Salt Lake City' },
  'us-west4': { lat: 36.1699, lng: -115.1398, name: 'Las Vegas' },
  'europe-west1': { lat: 50.8503, lng: 4.3517, name: 'Belgium' },
  'europe-west2': { lat: 51.5074, lng: -0.1278, name: 'London' },
  'europe-west3': { lat: 50.1109, lng: 8.6821, name: 'Frankfurt' },
  'europe-west4': { lat: 52.3676, lng: 4.9041, name: 'Netherlands' },
  'europe-north1': { lat: 60.5693, lng: 27.1878, name: 'Finland' },
  'asia-east1': { lat: 25.0330, lng: 121.5654, name: 'Taiwan' },
  'asia-east2': { lat: 22.3193, lng: 114.1694, name: 'Hong Kong' },
  'asia-southeast1': { lat: 1.3521, lng: 103.8198, name: 'Singapore' },
  'asia-southeast2': { lat: -6.2088, lng: 106.8456, name: 'Jakarta' },
  'asia-south1': { lat: 19.0760, lng: 72.8777, name: 'Mumbai' },
  'asia-northeast1': { lat: 35.6762, lng: 139.6503, name: 'Tokyo' },
  'asia-northeast2': { lat: 34.6937, lng: 135.5023, name: 'Osaka' },
  'asia-northeast3': { lat: 37.5665, lng: 126.9780, name: 'Seoul' },
  'australia-southeast1': { lat: -33.8688, lng: 151.2093, name: 'Sydney' },
  'southamerica-east1': { lat: -23.5505, lng: -46.6333, name: 'São Paulo' },

  // Azure Locations
  'eastus': { lat: 37.3719, lng: -79.8164, name: 'Virginia' },
  'eastus2': { lat: 36.6681, lng: -78.3889, name: 'Virginia' },
  'westus': { lat: 37.783, lng: -122.417, name: 'California' },
  'westus2': { lat: 47.233, lng: -119.852, name: 'Washington' },
  'westus3': { lat: 33.448, lng: -112.074, name: 'Arizona' },
  'centralus': { lat: 41.5908, lng: -93.6208, name: 'Iowa' },
  'northcentralus': { lat: 41.8819, lng: -87.6278, name: 'Illinois' },
  'southcentralus': { lat: 29.4241, lng: -98.4936, name: 'Texas' },
  'westeurope': { lat: 52.3676, lng: 4.9041, name: 'Netherlands' },
  'northeurope': { lat: 53.3498, lng: -6.2603, name: 'Ireland' },
  'uksouth': { lat: 51.5074, lng: -0.1278, name: 'London' },
  'ukwest': { lat: 51.4816, lng: -3.1791, name: 'Cardiff' },
  'francecentral': { lat: 48.8566, lng: 2.3522, name: 'Paris' },
  'germanywestcentral': { lat: 50.1109, lng: 8.6821, name: 'Frankfurt' },
  'swedencentral': { lat: 60.6749, lng: 17.1413, name: 'Gävle' },
  'southeastasia': { lat: 1.3521, lng: 103.8198, name: 'Singapore' },
  'eastasia': { lat: 22.3193, lng: 114.1694, name: 'Hong Kong' },
  'japaneast': { lat: 35.6762, lng: 139.6503, name: 'Tokyo' },
  'japanwest': { lat: 34.6937, lng: 135.5023, name: 'Osaka' },
  'koreacentral': { lat: 37.5665, lng: 126.9780, name: 'Seoul' },
  'australiaeast': { lat: -33.8688, lng: 151.2093, name: 'Sydney' },
  'australiasoutheast': { lat: -37.8136, lng: 144.9631, name: 'Melbourne' },
  'centralindia': { lat: 18.5204, lng: 73.8567, name: 'Pune' },
  'southindia': { lat: 13.0827, lng: 80.2707, name: 'Chennai' },
  'brazilsouth': { lat: -23.5505, lng: -46.6333, name: 'São Paulo' },
  'canadacentral': { lat: 43.6532, lng: -79.3832, name: 'Toronto' },
  'canadaeast': { lat: 46.8139, lng: -71.2080, name: 'Quebec' },
}

// Default region when none selected (for AWS which allows CLI default)
const DEFAULT_REGIONS: Record<string, string> = {
  aws: 'us-east-1',
  gcp: 'us-central1',
  azure: 'eastus',
}

interface WorldMapProps {
  selectedRegion: string
  provider: 'aws' | 'gcp' | 'azure' | 'oci'
  className?: string
}

export function WorldMap({ selectedRegion, provider, className = '' }: WorldMapProps) {
  const svgRef = useRef<HTMLDivElement>(null)

  const effectiveRegion = selectedRegion || DEFAULT_REGIONS[provider]

  const regionInfo = useMemo(() => {
    return REGION_COORDINATES[effectiveRegion]
  }, [effectiveRegion])

  const svgMap = useMemo(() => {
    const map = new DottedMap({ height: 60, grid: 'diagonal' })

    // Add the selected region pin
    if (regionInfo) {
      map.addPin({
        lat: regionInfo.lat,
        lng: regionInfo.lng,
        svgOptions: { color: '#22c55e', radius: 0.8 }
      })
    }

    return map.getSVG({
      radius: 0.22,
      color: '#6b7280',
      shape: 'circle',
      backgroundColor: 'transparent',
    })
  }, [regionInfo])

  const providerColors = {
    aws: { bg: 'from-orange-500/10 to-orange-600/5', text: 'text-orange-600', dot: 'bg-orange-500' },
    gcp: { bg: 'from-blue-500/10 to-blue-600/5', text: 'text-blue-600', dot: 'bg-blue-500' },
    azure: { bg: 'from-sky-500/10 to-sky-600/5', text: 'text-sky-600', dot: 'bg-sky-500' },
    oci: { bg: 'from-red-500/10 to-red-600/5', text: 'text-red-600', dot: 'bg-red-500' },
  }

  const colors = providerColors[provider]

  return (
    <div className={`relative border bg-gradient-to-br ${colors.bg} overflow-hidden ${className}`}>
      {/* Map */}
      <div
        ref={svgRef}
        className="w-full h-40 flex items-center justify-center opacity-60"
        dangerouslySetInnerHTML={{ __html: svgMap }}
      />

      {/* Location indicator */}
      {regionInfo && (
        <div className="absolute bottom-3 left-3 right-3 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className={`relative flex h-2.5 w-2.5 ${colors.dot}`}>
              <span className={`animate-ping absolute inline-flex h-full w-full ${colors.dot} opacity-75`}></span>
            </span>
            <span className={`text-base font-medium ${colors.text}`}>
              {regionInfo.name}
              {!selectedRegion && (
                <span className="ml-1.5 text-sm text-muted-foreground">(default)</span>
              )}
            </span>
          </div>
          <span className="text-sm text-muted-foreground font-mono">
            {regionInfo.lat.toFixed(2)}°, {regionInfo.lng.toFixed(2)}°
          </span>
        </div>
      )}
    </div>
  )
}
