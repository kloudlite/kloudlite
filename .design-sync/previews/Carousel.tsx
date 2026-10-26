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

const regions = [
  { id: 'ap-south-1', city: 'Mumbai', nodes: 12 },
  { id: 'eu-west-1', city: 'Ireland', nodes: 8 },
  { id: 'us-east-1', city: 'N. Virginia', nodes: 20 },
]

const Slide = ({ id, city, nodes }: (typeof regions)[number]) => (
  <Card>
    <CardContent className="flex h-32 flex-col justify-center p-6">
      <div className="text-lg font-semibold">{id}</div>
      <div className="text-sm text-muted-foreground">
        {city} &middot; {nodes} nodes
      </div>
    </CardContent>
  </Card>
)

export const Basic = () => (
  <div style={{ padding: '0 3.5rem' }}>
    <Carousel className="w-72">
      <CarouselContent>
        {regions.map((r) => (
          <CarouselItem key={r.id}>
            <Slide {...r} />
          </CarouselItem>
        ))}
      </CarouselContent>
      <CarouselPrevious />
      <CarouselNext />
    </Carousel>
  </div>
)

export const MultipleVisible = () => (
  <div style={{ padding: '0 3.5rem' }}>
    <Carousel className="w-96" opts={{ align: 'start' }}>
      <CarouselContent>
        {regions.map((r) => (
          <CarouselItem key={r.id} style={{ flexBasis: '50%' }}>
            <Slide {...r} />
          </CarouselItem>
        ))}
      </CarouselContent>
      <CarouselPrevious />
      <CarouselNext />
    </Carousel>
  </div>
)

export const WithoutControls = () => (
  <Carousel className="w-72">
    <CarouselContent>
      {regions.map((r) => (
        <CarouselItem key={r.id}>
          <Slide {...r} />
        </CarouselItem>
      ))}
    </CarouselContent>
  </Carousel>
)

export const Vertical = () => (
  <div style={{ padding: '3.5rem 0' }}>
    <Carousel orientation="vertical" className="w-72">
      <CarouselContent className="h-40">
        {regions.map((r) => (
          <CarouselItem key={r.id}>
            <Slide {...r} />
          </CarouselItem>
        ))}
      </CarouselContent>
      <CarouselPrevious />
      <CarouselNext />
    </Carousel>
  </div>
)
