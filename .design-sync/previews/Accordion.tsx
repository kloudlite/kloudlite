import * as React from 'react'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@kloudlite/ui'

export const Single = () => (
  <Accordion type="single" collapsible defaultValue="networking" className="w-96">
    <AccordionItem value="networking">
      <AccordionTrigger>Networking</AccordionTrigger>
      <AccordionContent className="text-muted-foreground">
        Each environment gets a private wireguard tunnel. Intercepts route traffic from
        the cluster back to your workmachine.
      </AccordionContent>
    </AccordionItem>
    <AccordionItem value="storage">
      <AccordionTrigger>Storage</AccordionTrigger>
      <AccordionContent className="text-muted-foreground">
        Persistent volumes are provisioned per namespace and retained for 7 days after
        an environment is deleted.
      </AccordionContent>
    </AccordionItem>
    <AccordionItem value="node-pools">
      <AccordionTrigger>Node pools</AccordionTrigger>
      <AccordionContent className="text-muted-foreground">
        Spot node pools scale to zero when no workspaces are running in ap-south-1.
      </AccordionContent>
    </AccordionItem>
  </Accordion>
)

export const Multiple = () => (
  <Accordion
    type="multiple"
    defaultValue={['kubeconfig', 'registry']}
    className="w-96"
  >
    <AccordionItem value="kubeconfig">
      <AccordionTrigger>How do I get a kubeconfig?</AccordionTrigger>
      <AccordionContent className="text-muted-foreground">
        Run <span className="font-medium">kl cluster kubeconfig kloudlite-dev</span> and
        the credentials are written to your local config.
      </AccordionContent>
    </AccordionItem>
    <AccordionItem value="registry">
      <AccordionTrigger>Which image tags are pulled?</AccordionTrigger>
      <AccordionContent className="text-muted-foreground">
        Deployments pin an immutable digest. The tag shown in the console is resolved at
        the time of the last successful rollout.
      </AccordionContent>
    </AccordionItem>
  </Accordion>
)

export const AllCollapsed = () => (
  <Accordion type="single" collapsible className="w-96">
    <AccordionItem value="billing">
      <AccordionTrigger>Billing &amp; credits</AccordionTrigger>
      <AccordionContent>Usage is billed per workmachine hour.</AccordionContent>
    </AccordionItem>
    <AccordionItem value="access">
      <AccordionTrigger>Team access</AccordionTrigger>
      <AccordionContent>Invite members from the installation settings.</AccordionContent>
    </AccordionItem>
    <AccordionItem value="support">
      <AccordionTrigger>Support</AccordionTrigger>
      <AccordionContent>Reach us from the in-console help menu.</AccordionContent>
    </AccordionItem>
  </Accordion>
)
