import * as React from 'react'
import {
  Card,
  CardContent,
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
} from '@kloudlite/ui'

const pools = ['spot-workers', 'on-demand-workers', 'gpu-workers']

const Slide = ({ name }: { name: string }) => (
  <Card>
    <CardContent className="flex h-32 flex-col justify-center p-6">
      <div className="font-semibold">{name}</div>
      <div className="text-sm text-muted-foreground">Node pool &middot; ap-south-1</div>
    </CardContent>
  </Card>
)

// At the first slide embla cannot scroll back, so the control renders disabled.
export const DisabledAtStart = () => (
  <div style={{ padding: '0 3.5rem' }}>
    <Carousel className="w-72">
      <CarouselContent>
        {pools.map((p) => (
          <CarouselItem key={p}>
            <Slide name={p} />
          </CarouselItem>
        ))}
      </CarouselContent>
      <CarouselPrevious />
      <CarouselNext />
    </Carousel>
  </div>
)

export const LoopingEnabled = () => (
  <div style={{ padding: '0 3.5rem' }}>
    <Carousel className="w-72" opts={{ loop: true }}>
      <CarouselContent>
        {pools.map((p) => (
          <CarouselItem key={p}>
            <Slide name={p} />
          </CarouselItem>
        ))}
      </CarouselContent>
      <CarouselPrevious />
      <CarouselNext />
    </Carousel>
  </div>
)

export const VerticalPlacement = () => (
  <div style={{ padding: '3.5rem 0' }}>
    <Carousel orientation="vertical" className="w-72" opts={{ loop: true }}>
      <CarouselContent className="h-40">
        {pools.map((p) => (
          <CarouselItem key={p}>
            <Slide name={p} />
          </CarouselItem>
        ))}
      </CarouselContent>
      <CarouselPrevious />
      <CarouselNext />
    </Carousel>
  </div>
)
