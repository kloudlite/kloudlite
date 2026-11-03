import * as React from 'react'
import {
  Badge,
  Card,
  CardContent,
  Carousel,
  CarouselContent,
  CarouselItem,
} from '@kloudlite/ui'

const envs = [
  { name: 'development', region: 'ap-south-1', status: 'Running' },
  { name: 'staging', region: 'eu-west-1', status: 'Running' },
  { name: 'production', region: 'us-east-1', status: 'Degraded' },
]

export const FullWidthItem = () => (
  <Carousel className="w-72">
    <CarouselContent>
      {envs.map((e) => (
        <CarouselItem key={e.name}>
          <Card>
            <CardContent className="flex h-32 flex-col justify-center gap-2 p-6">
              <div className="font-semibold">{e.name}</div>
              <div className="text-sm text-muted-foreground">{e.region}</div>
              <Badge variant="secondary" className="w-24">
                {e.status}
              </Badge>
            </CardContent>
          </Card>
        </CarouselItem>
      ))}
    </CarouselContent>
  </Carousel>
)

export const HalfWidthItems = () => (
  <Carousel className="w-96" opts={{ align: 'start' }}>
    <CarouselContent>
      {envs.map((e) => (
        <CarouselItem key={e.name} style={{ flexBasis: '50%' }}>
          <Card>
            <CardContent className="flex h-24 flex-col justify-center p-6">
              <div className="text-sm font-semibold">{e.name}</div>
              <div className="text-sm text-muted-foreground">{e.region}</div>
            </CardContent>
          </Card>
        </CarouselItem>
      ))}
    </CarouselContent>
  </Carousel>
)

export const VerticalItem = () => (
  <Carousel orientation="vertical" className="w-72">
    <CarouselContent className="h-40">
      {envs.map((e) => (
        <CarouselItem key={e.name}>
          <Card>
            <CardContent className="flex h-36 flex-col justify-center p-6">
              <div className="font-semibold">{e.name}</div>
              <div className="text-sm text-muted-foreground">{e.region}</div>
            </CardContent>
          </Card>
        </CarouselItem>
      ))}
    </CarouselContent>
  </Carousel>
)
