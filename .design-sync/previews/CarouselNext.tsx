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

const builds = [
  { tag: 'v1.9.2', duration: '2m 41s' },
  { tag: 'v1.9.1', duration: '3m 04s' },
  { tag: 'v1.9.0', duration: '2m 58s' },
]

const Slide = ({ tag, duration }: (typeof builds)[number]) => (
  <Card>
    <CardContent className="flex h-32 flex-col justify-center p-6">
      <div className="font-mono font-semibold">{tag}</div>
      <div className="text-sm text-muted-foreground">Build finished in {duration}</div>
    </CardContent>
  </Card>
)

export const Enabled = () => (
  <div style={{ padding: '0 3.5rem' }}>
    <Carousel className="w-72">
      <CarouselContent>
        {builds.map((b) => (
          <CarouselItem key={b.tag}>
            <Slide {...b} />
          </CarouselItem>
        ))}
      </CarouselContent>
      <CarouselPrevious />
      <CarouselNext />
    </Carousel>
  </div>
)

// A single slide leaves nothing to scroll to, so the control renders disabled.
export const DisabledAtEnd = () => (
  <div style={{ padding: '0 3.5rem' }}>
    <Carousel className="w-72">
      <CarouselContent>
        <CarouselItem>
          <Slide {...builds[0]} />
        </CarouselItem>
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
        {builds.map((b) => (
          <CarouselItem key={b.tag}>
            <Slide {...b} />
          </CarouselItem>
        ))}
      </CarouselContent>
      <CarouselPrevious />
      <CarouselNext />
    </Carousel>
  </div>
)
