import * as React from 'react'
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from '@kloudlite/ui'

const Frame = ({
  caption,
  children,
}: {
  caption: string
  children: React.ReactNode
}) => (
  <div className="w-full space-y-3 p-6">
    <p className="text-center text-sm text-muted-foreground">{caption}</p>
    {children}
  </div>
)

export const LeadingControl = () => (
  <Frame caption="Showing 21-40 of 148 deployments">
      <Pagination>
        <PaginationContent>
          <PaginationItem>
            <PaginationPrevious href="#" />
          </PaginationItem>
          <PaginationItem>
            <PaginationLink href="#">1</PaginationLink>
          </PaginationItem>
          <PaginationItem>
            <PaginationLink href="#" isActive>
              2
            </PaginationLink>
          </PaginationItem>
          <PaginationItem>
            <PaginationLink href="#">3</PaginationLink>
          </PaginationItem>
          <PaginationItem>
            <PaginationEllipsis />
          </PaginationItem>
          <PaginationItem>
            <PaginationLink href="#">8</PaginationLink>
          </PaginationItem>
          <PaginationItem>
            <PaginationNext href="#" />
          </PaginationItem>
        </PaginationContent>
      </Pagination>
  </Frame>
)

export const OnLastPage = () => (
  <Frame caption="Showing 141-148 of 148 deployments">
      <Pagination>
        <PaginationContent>
          <PaginationItem>
            <PaginationPrevious href="#" />
          </PaginationItem>
          <PaginationItem>
            <PaginationLink href="#">6</PaginationLink>
          </PaginationItem>
          <PaginationItem>
            <PaginationLink href="#">7</PaginationLink>
          </PaginationItem>
          <PaginationItem>
            <PaginationLink href="#" isActive>
              8
            </PaginationLink>
          </PaginationItem>
        </PaginationContent>
      </Pagination>
  </Frame>
)
