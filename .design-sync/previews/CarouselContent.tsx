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

const steps = [
  { title: 'Create installation', body: 'Pick a region and a cluster size.' },
  { title: 'Install the operator', body: 'Run the generated command on your cluster.' },
  { title: 'Open a workspace', body: 'Start an environment and intercept a service.' },
]

const Slide = ({ title, body }: (typeof steps)[number]) => (
  <Card>
    <CardContent className="flex h-32 flex-col justify-center p-6">
      <div className="font-semibold">{title}</div>
      <div className="text-sm text-muted-foreground">{body}</div>
    </CardContent>
  </Card>
)

export const Horizontal = () => (
  <div style={{ padding: '0 3.5rem' }}>
    <Carousel className="w-72">
      <CarouselContent>
        {steps.map((s) => (
          <CarouselItem key={s.title}>
            <Slide {...s} />
          </CarouselItem>
        ))}
      </CarouselContent>
      <CarouselPrevious />
      <CarouselNext />
    </Carousel>
  </div>
)

export const VerticalContent = () => (
  <Carousel orientation="vertical" className="w-72">
    <CarouselContent className="h-40">
      {steps.map((s) => (
        <CarouselItem key={s.title}>
          <Slide {...s} />
        </CarouselItem>
      ))}
    </CarouselContent>
  </Carousel>
)

export const TwoPerView = () => (
  <Carousel className="w-96" opts={{ align: 'start' }}>
    <CarouselContent>
      {steps.map((s) => (
        <CarouselItem key={s.title} style={{ flexBasis: '50%' }}>
          <Slide {...s} />
        </CarouselItem>
      ))}
    </CarouselContent>
  </Carousel>
)
