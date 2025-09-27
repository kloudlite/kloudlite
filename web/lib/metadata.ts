export const siteConfig = {
  name: "Kloudlite",
  description: "Development environments as a Service",
  url: "https://kloudlite.io",
}

export function generateTitle(title?: string): string {
  if (!title) {return siteConfig.name}
  return `${title} | ${siteConfig.name}`
}