const CF_TOKEN = process.env.CLOUDFLARE_API_TOKEN!
const CF_ZONE = process.env.CLOUDFLARE_ZONE_ID!
export const CLOUDFLARE_DNS_DOMAIN = process.env.CLOUDFLARE_DNS_DOMAIN!

const BASE = `https://api.cloudflare.com/client/v4/zones/${CF_ZONE}/dns_records`
const headers = { Authorization: `Bearer ${CF_TOKEN}`, 'Content-Type': 'application/json' }

async function cf<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, { headers, ...init })
  const body = await res.json()
  if (!body.success) throw new Error(body.errors?.[0]?.message || 'Cloudflare API error')
  return body.result
}

async function del(name: string, type: string) {
  const records: any[] = await cf(`?name=${encodeURIComponent(name)}&type=${type}`)
  for (const r of records) await cf(`/${r.id}`, { method: 'DELETE' })
  return records.length
}

export async function createDnsRecord(name: string, content: string, type = 'A', proxied = false) {
  await del(name, type)
  return cf('', { method: 'POST', body: JSON.stringify({ type, name, content, ttl: 120, proxied }) })
}

export async function createCnameRecord(name: string, target: string) {
  return createDnsRecord(name, target, 'CNAME', true)
}

export async function updateDnsRecord(name: string, content: string, type = 'A', proxied = false) {
  const existing: any[] = await cf(`?name=${encodeURIComponent(name)}&type=${type}`)
  if (!existing.length) return createDnsRecord(name, content, type, proxied)
  return cf(`/${existing[0].id}`, { method: 'PATCH', body: JSON.stringify({ content, proxied, ttl: 120 }) })
}

export async function deleteDnsRecord(recordId: string) {
  return cf(`/${recordId}`, { method: 'DELETE' })
}

export async function getDnsRecord(name: string, type = 'A') {
  const records: any[] = await cf(`?name=${encodeURIComponent(name)}&type=${type}`)
  return records[0] || null
}

export async function getAllDnsRecords(name: string) {
  return cf(`?name=${encodeURIComponent(name)}`)
}

export async function createInstallationDnsRecords(name: string, ip: string) {
  return [
    await createDnsRecord(name, ip),
    await createCnameRecord(`*.${name}`, name),
  ]
}

export async function createWorkmachineDnsRecords(name: string, ip: string) {
  return createDnsRecord(name, ip)
}

export async function updateDnsRecords(names: string[], ip: string) {
  return Promise.all(names.map(n => updateDnsRecord(n, ip)))
}

export async function deleteDnsRecords(names: string[]) {
  return Promise.all(names.map(n => del(n, 'A')))
}

export async function createDomainRouteCnameRecords(domain: string) {
  return createCnameRecord(domain, CLOUDFLARE_DNS_DOMAIN)
}

export async function createDomainRequestDnsRecords(subdomain: string, ip: string) {
  return [
    await createDnsRecord(subdomain, ip),
    await createCnameRecord(`*.${subdomain}`, subdomain),
  ]
}
