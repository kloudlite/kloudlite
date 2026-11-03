import * as React from 'react'
import {
  Badge,
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableFooter,
  TableHead,
  TableHeader,
  TableRow,
} from '@kloudlite/ui'

const deployments = [
  { name: 'api-gateway', ns: 'kloudlite-dev', replicas: '3/3', status: 'Running', age: '4d' },
  { name: 'auth-service', ns: 'kloudlite-dev', replicas: '2/2', status: 'Running', age: '11d' },
  { name: 'console-web', ns: 'kloudlite-prod', replicas: '1/2', status: 'Degraded', age: '2h' },
  { name: 'iot-console', ns: 'kloudlite-prod', replicas: '2/2', status: 'Running', age: '27d' },
]

export const Basic = () => (
  <div className="w-full">
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Deployment</TableHead>
          <TableHead>Namespace</TableHead>
          <TableHead>Replicas</TableHead>
          <TableHead>Status</TableHead>
          <TableHead className="text-right">Age</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {deployments.map((d) => (
          <TableRow key={d.name}>
            <TableCell className="font-medium">{d.name}</TableCell>
            <TableCell className="text-muted-foreground">{d.ns}</TableCell>
            <TableCell className="tabular-nums">{d.replicas}</TableCell>
            <TableCell>
              <Badge variant={d.status === 'Running' ? 'default' : 'destructive'}>{d.status}</Badge>
            </TableCell>
            <TableCell className="text-right tabular-nums">{d.age}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  </div>
)

export const WithFooter = () => (
  <div className="w-full">
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Environment</TableHead>
          <TableHead className="text-right">Pods</TableHead>
          <TableHead className="text-right">vCPU</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow>
          <TableCell className="font-medium">development</TableCell>
          <TableCell className="text-right tabular-nums">18</TableCell>
          <TableCell className="text-right tabular-nums">6</TableCell>
        </TableRow>
        <TableRow>
          <TableCell className="font-medium">staging</TableCell>
          <TableCell className="text-right tabular-nums">12</TableCell>
          <TableCell className="text-right tabular-nums">4</TableCell>
        </TableRow>
        <TableRow>
          <TableCell className="font-medium">production</TableCell>
          <TableCell className="text-right tabular-nums">42</TableCell>
          <TableCell className="text-right tabular-nums">24</TableCell>
        </TableRow>
      </TableBody>
      <TableFooter>
        <TableRow>
          <TableCell>Total</TableCell>
          <TableCell className="text-right tabular-nums">72</TableCell>
          <TableCell className="text-right tabular-nums">34</TableCell>
        </TableRow>
      </TableFooter>
    </Table>
  </div>
)

export const WithCaption = () => (
  <div className="w-full">
    <Table>
      <TableCaption>Installations across all your Kloudlite organisations.</TableCaption>
      <TableHeader>
        <TableRow>
          <TableHead>Installation</TableHead>
          <TableHead>Region</TableHead>
          <TableHead>Status</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow>
          <TableCell className="font-medium">kloudlite-dev</TableCell>
          <TableCell>ap-south-1</TableCell>
          <TableCell>Active</TableCell>
        </TableRow>
        <TableRow>
          <TableCell className="font-medium">acme-eu</TableCell>
          <TableCell>eu-west-1</TableCell>
          <TableCell>Installing</TableCell>
        </TableRow>
      </TableBody>
    </Table>
  </div>
)

export const SelectedRow = () => (
  <div className="w-full">
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Deployment</TableHead>
          <TableHead>Namespace</TableHead>
          <TableHead className="text-right">Age</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow>
          <TableCell className="font-medium">api-gateway</TableCell>
          <TableCell className="text-muted-foreground">kloudlite-dev</TableCell>
          <TableCell className="text-right tabular-nums">4d</TableCell>
        </TableRow>
        <TableRow data-state="selected">
          <TableCell className="font-medium">auth-service</TableCell>
          <TableCell className="text-muted-foreground">kloudlite-dev</TableCell>
          <TableCell className="text-right tabular-nums">11d</TableCell>
        </TableRow>
        <TableRow>
          <TableCell className="font-medium">console-web</TableCell>
          <TableCell className="text-muted-foreground">kloudlite-prod</TableCell>
          <TableCell className="text-right tabular-nums">2h</TableCell>
        </TableRow>
      </TableBody>
    </Table>
  </div>
)
