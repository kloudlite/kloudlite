import * as React from 'react'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
  Badge,
} from '@kloudlite/ui'

export const OpenAndClosed = () => (
  <Accordion type="single" collapsible defaultValue="open-one" className="w-96">
    <AccordionItem value="open-one">
      <AccordionTrigger>Expanded trigger &mdash; chevron up</AccordionTrigger>
      <AccordionContent className="text-muted-foreground">
        Deployment rollout finished 12 minutes ago.
      </AccordionContent>
    </AccordionItem>
    <AccordionItem value="closed-one">
      <AccordionTrigger>Collapsed trigger &mdash; chevron down</AccordionTrigger>
      <AccordionContent>Nothing to show yet.</AccordionContent>
    </AccordionItem>
  </Accordion>
)

export const WithTrailingContent = () => (
  <Accordion type="single" collapsible defaultValue="cluster" className="w-96">
    <AccordionItem value="cluster">
      <AccordionTrigger>
        <span className="flex items-center gap-2">
          Production cluster
          <Badge variant="secondary">eu-west-1</Badge>
        </span>
      </AccordionTrigger>
      <AccordionContent className="text-muted-foreground">
        12 nodes across 3 node pools, all reporting ready.
      </AccordionContent>
    </AccordionItem>
    <AccordionItem value="staging">
      <AccordionTrigger>
        <span className="flex items-center gap-2">
          Staging cluster
          <Badge variant="outline">ap-south-1</Badge>
        </span>
      </AccordionTrigger>
      <AccordionContent className="text-muted-foreground">
        4 nodes, scaled down outside working hours.
      </AccordionContent>
    </AccordionItem>
  </Accordion>
)

export const LongLabel = () => (
  <Accordion type="single" collapsible defaultValue="long" className="w-96">
    <AccordionItem value="long">
      <AccordionTrigger>
        Why does my workmachine take a minute to become reachable after the tunnel
        reconnects?
      </AccordionTrigger>
      <AccordionContent className="text-muted-foreground">
        DNS records for the environment subdomain are refreshed once the tunnel reports
        healthy.
      </AccordionContent>
    </AccordionItem>
  </Accordion>
)
