import * as React from 'react'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@kloudlite/ui'

export const InAccordion = () => (
  <Accordion type="single" collapsible defaultValue="intercepts" className="w-96">
    <AccordionItem value="intercepts">
      <AccordionTrigger>Intercepts</AccordionTrigger>
      <AccordionContent className="text-muted-foreground">
        An intercept swaps a deployment in the cluster for the process running on your
        workmachine.
      </AccordionContent>
    </AccordionItem>
    <AccordionItem value="tunnels">
      <AccordionTrigger>Tunnels</AccordionTrigger>
      <AccordionContent className="text-muted-foreground">
        Tunnels stay connected while the environment is active.
      </AccordionContent>
    </AccordionItem>
  </Accordion>
)

export const Disabled = () => (
  <Accordion type="single" collapsible defaultValue="active" className="w-96">
    <AccordionItem value="active">
      <AccordionTrigger>payments-api</AccordionTrigger>
      <AccordionContent className="text-muted-foreground">
        3 replicas running in namespace kloudlite-dev.
      </AccordionContent>
    </AccordionItem>
    <AccordionItem value="archived" disabled>
      <AccordionTrigger>legacy-billing (archived)</AccordionTrigger>
      <AccordionContent>Archived environments cannot be expanded.</AccordionContent>
    </AccordionItem>
  </Accordion>
)

export const SingleItem = () => (
  <Accordion type="single" collapsible defaultValue="only" className="w-96">
    <AccordionItem value="only">
      <AccordionTrigger>Advanced cluster options</AccordionTrigger>
      <AccordionContent className="text-muted-foreground">
        Override the CNI, pod CIDR and the default storage class for this installation.
      </AccordionContent>
    </AccordionItem>
  </Accordion>
)
