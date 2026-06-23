import { describe, expect, test } from 'bun:test'
import { parseComposeServices } from './compose-services'

describe('parseComposeServices', () => {
  test('parses only service definitions and keeps ports separate from mounted volumes', () => {
    const compose = `version: "3.8"
services:
  web:
    image: nginx
    ports:
      - "80:80"
  mongo:
    image: mongo
    ports:
      - "27017:27017"
    volumes:
      - mdata:/data/db
volumes:
  mdata:
`

    const services = parseComposeServices(compose, 'env-demo')

    expect(services.map((svc) => svc.name)).toEqual(['web', 'mongo'])
    expect(services.find((svc) => svc.name === 'mongo')?.ports).toEqual([
      { port: 27017, targetPort: 27017, protocol: 'TCP' },
    ])
    expect(services.find((svc) => svc.name === 'mongo')?.volumes).toEqual([
      { name: 'mdata', mountPath: '/data/db', type: 'persistent' },
    ])
  })

  test('parses single-value port syntax as service and target port', () => {
    const compose = `services:
  mongo:
    image: mongo
    ports:
      - "27017"
    volumes:
      - mdata:/data/db
volumes:
  mdata:
`

    const services = parseComposeServices(compose, 'env-demo')

    expect(services.find((svc) => svc.name === 'mongo')?.ports).toEqual([
      { port: 27017, targetPort: 27017, protocol: 'TCP' },
    ])
    expect(services.find((svc) => svc.name === 'mongo')?.volumes).toEqual([
      { name: 'mdata', mountPath: '/data/db', type: 'persistent' },
    ])
  })
})
