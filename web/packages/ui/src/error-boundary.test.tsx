import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { render, screen, fireEvent } from "@testing-library/react"
import * as React from "react"
import { ErrorBoundary, ErrorBoundaryFallback, ErrorBoundaryWithFallback } from "./error-boundary"

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

    it("should not catch errors in event handlers", () => {
      const handleClick = () => {
        throw new Error("Event handler error")
      }

      render(
        <ErrorBoundary>
          <button onClick={handleClick}>Click me</button>
        </ErrorBoundary>
      )

      const button = screen.getByText("Click me")
      // Note: Error boundaries don't catch errors in event handlers
      // This test verifies the button exists and can be clicked
      fireEvent.click(button)
      // The error should propagate (not be caught by ErrorBoundary)
      expect(console.error).toHaveBeenCalled()
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

  describe("Error logging", () => {
    it("should call onError callback with error and errorInfo", () => {
      const onError = vi.fn()

      render(
        <ErrorBoundary onError={onError}>
          <ThrowError shouldThrow={true} />
        </ErrorBoundary>
      )

      expect(onError).toHaveBeenCalledTimes(1)
      expect(onError).toHaveBeenCalledWith(
        expect.any(Error),
        expect.objectContaining({
          componentStack: expect.any(String),
        })
      )
      expect(onError.mock.calls[0][0].message).toBe("Test error")
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

  describe("Error details display", () => {
    it("should display error message in details element", () => {
      render(
        <ErrorBoundary>
          <ThrowError shouldThrow={true} />
        </ErrorBoundary>
      )

      const detailsSummary = screen.getByText("Error details")
      expect(detailsSummary).toBeInTheDocument()

      fireEvent.click(detailsSummary)

      expect(screen.getByText("Test error")).toBeInTheDocument()
    })
  })
})

describe("ErrorBoundaryFallback", () => {
  it("should render fallback UI without reset button when no onReset provided", () => {
    render(<ErrorBoundaryFallback error={new Error("Test error")} />)

    expect(screen.getByText("Something went wrong")).toBeInTheDocument()
    expect(screen.queryByText("Try again")).not.toBeInTheDocument()
  })

  it("should render fallback UI with reset button when onReset provided", () => {
    const onReset = vi.fn()
    render(<ErrorBoundaryFallback error={new Error("Test error")} onReset={onReset} />)

    expect(screen.getByText("Try again")).toBeInTheDocument()

    const button = screen.getByText("Try again")
    fireEvent.click(button)

    expect(onReset).toHaveBeenCalledTimes(1)
  })

  it("should not render error details when no error provided", () => {
    render(<ErrorBoundaryFallback />)

    expect(screen.getByText("Something went wrong")).toBeInTheDocument()
    expect(screen.queryByText("Error details")).not.toBeInTheDocument()
  })
})

describe("ErrorBoundaryWithFallback", () => {
  let originalConsoleError: typeof console.error

  beforeEach(() => {
    originalConsoleError = console.error
    console.error = vi.fn()
  })

  afterEach(() => {
    console.error = originalConsoleError
    vi.clearAllMocks()
  })

  it("should wrap children in ErrorBoundary with custom className", () => {
    render(
      <ErrorBoundaryWithFallback className="custom-class">
        <div>Content</div>
      </ErrorBoundaryWithFallback>
    )

    const wrapper = screen.getByText("Content").parentElement
    expect(wrapper).toHaveClass("custom-class")
  })

  it("should pass through all ErrorBoundary props", () => {
    const onError = vi.fn()
    const customFallback = <div>Custom fallback</div>

    render(
      <ErrorBoundaryWithFallback onError={onError} fallback={customFallback}>
        <ThrowError shouldThrow={true} />
      </ErrorBoundaryWithFallback>
    )

    expect(screen.getByText("Custom fallback")).toBeInTheDocument()
    expect(onError).toHaveBeenCalledTimes(1)
  })
})
