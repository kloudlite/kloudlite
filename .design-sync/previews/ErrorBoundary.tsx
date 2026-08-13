import * as React from 'react'
import {
  Alert,
  AlertDescription,
  AlertTitle,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  ErrorBoundary,
} from '@kloudlite/ui'

const BrokenPanel = (): React.ReactElement => {
  throw new Error('Failed to load cluster metrics: upstream returned 503')
}

export const HealthyChild = () => (
  <ErrorBoundary>
    <Card className="w-96">
      <CardHeader>
        <CardTitle>Cluster metrics</CardTitle>
        <CardDescription>kloudlite-dev &middot; ap-south-1</CardDescription>
      </CardHeader>
      <CardContent className="text-sm text-muted-foreground">
        The boundary is transparent while its subtree renders successfully.
      </CardContent>
    </Card>
  </ErrorBoundary>
)

export const DefaultFallback = () => (
  <div className="w-96">
    <ErrorBoundary>
      <BrokenPanel />
    </ErrorBoundary>
  </div>
)

export const CustomFallback = () => (
  <div className="w-96">
    <ErrorBoundary
      fallback={
        <Alert variant="destructive">
          <AlertTitle>Metrics unavailable</AlertTitle>
          <AlertDescription>
            <p>
              We could not reach the metrics API for kloudlite-dev. Workloads are unaffected.
            </p>
            <Button size="sm" variant="outline">
              Retry
            </Button>
          </AlertDescription>
        </Alert>
      }
    >
      <BrokenPanel />
    </ErrorBoundary>
  </div>
)
