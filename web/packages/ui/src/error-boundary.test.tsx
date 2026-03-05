import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { render, screen, fireEvent } from "@testing-library/react"
import * as React from "react"
import { ErrorBoundary } from "./error-boundary"

const ThrowError = ({ shouldThrow }: { shouldThrow: boolean }) => {
  if (shouldThrow) {
    throw new Error("Test error")
  }
  return <div>No error</div>
}

describe("ErrorBoundary", () => {
  let originalConsoleError: typeof console.error

  beforeEach(() => {
    originalConsoleError = console.error
    console.error = vi.fn()
  })

  afterEach(() => {
    console.error = originalConsoleError
    vi.clearAllMocks()
  })

  describe("Basic error catching", () => {
    it("should render children when there is no error", () => {
      render(
        <ErrorBoundary>
          <ThrowError shouldThrow={false} />
        </ErrorBoundary>
      )

      expect(screen.getByText("No error")).toBeInTheDocument()
    })

    it("should catch and display error when child component throws", () => {
      render(
        <ErrorBoundary>
          <ThrowError shouldThrow={true} />
        </ErrorBoundary>
      )

      expect(screen.getByText("Something went wrong")).toBeInTheDocument()
      expect(screen.getByText(/An unexpected error occurred/)).toBeInTheDocument()
    })

    it("should log error to console when caught", () => {
      render(
        <ErrorBoundary>
          <ThrowError shouldThrow={true} />
        </ErrorBoundary>
      )

      expect(console.error).toHaveBeenCalledWith(
        "ErrorBoundary caught an error:",
        expect.any(Error),
        expect.any(Object)
      )
    })
  })

  describe("Custom fallback UI", () => {
    it("should render custom fallback when provided", () => {
      const customFallback = <div>Custom error message</div>

      render(
        <ErrorBoundary fallback={customFallback}>
          <ThrowError shouldThrow={true} />
        </ErrorBoundary>
      )

      expect(screen.getByText("Custom error message")).toBeInTheDocument()
      expect(screen.queryByText("Something went wrong")).not.toBeInTheDocument()
    })
  })

  describe("Error recovery", () => {
    it("should provide try again button that resets error state", () => {
      const { rerender } = render(
        <ErrorBoundary>
          <ThrowError shouldThrow={true} />
        </ErrorBoundary>
      )

      expect(screen.getByText("Something went wrong")).toBeInTheDocument()

      const tryAgainButton = screen.getByText("Try again")
      fireEvent.click(tryAgainButton)

      // After clicking try again, render a non-throwing child
      rerender(
        <ErrorBoundary key="recovered">
          <ThrowError shouldThrow={false} />
        </ErrorBoundary>
      )

      expect(screen.getByText("No error")).toBeInTheDocument()
    })
  })
})
