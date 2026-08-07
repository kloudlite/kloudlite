import * as React from 'react'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
  Button,
} from '@kloudlite/ui'

export const TextContent = () => (
  <Accordion type="single" collapsible defaultValue="what" className="w-96">
    <AccordionItem value="what">
      <AccordionTrigger>What is a workmachine?</AccordionTrigger>
      <AccordionContent className="text-muted-foreground">
        A workmachine is a remote development container scheduled onto your cluster. It
        mounts your workspace volume and joins the environment network.
      </AccordionContent>
    </AccordionItem>
  </Accordion>
)

export const RichContent = () => (
  <Accordion type="single" collapsible defaultValue="details" className="w-96">
    <AccordionItem value="details">
      <AccordionTrigger>Installation details</AccordionTrigger>
      <AccordionContent className="space-y-2">
        <div className="flex justify-between">
          <span className="text-muted-foreground">Region</span>
          <span className="font-medium">ap-south-1</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">Kubernetes</span>
          <span className="font-medium">v1.31.4</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">Namespace</span>
          <span className="font-medium">kloudlite-dev</span>
        </div>
        <Button size="sm" variant="outline" className="mt-2">
          Download kubeconfig
        </Button>
      </AccordionContent>
    </AccordionItem>
  </Accordion>
)

export const MultipleOpen = () => (
  <Accordion type="multiple" defaultValue={['a', 'b']} className="w-96">
    <AccordionItem value="a">
      <AccordionTrigger>Rollout history</AccordionTrigger>
      <AccordionContent className="text-muted-foreground">
        payments-api:2026.8.3 promoted to production 2 hours ago.
      </AccordionContent>
    </AccordionItem>
    <AccordionItem value="b">
      <AccordionTrigger>Recent events</AccordionTrigger>
      <AccordionContent className="text-muted-foreground">
        Node pool spot-a scaled from 2 to 4 nodes.
      </AccordionContent>
    </AccordionItem>
  </Accordion>
)
