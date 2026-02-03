/**
 * Custom error classes for Kubernetes operations
 */

export class K8sError extends Error {
  constructor(
    message: string,
    public readonly code?: number,
    public readonly details?: unknown
  ) {
    super(message);
    this.name = 'K8sError';
    Object.setPrototypeOf(this, K8sError.prototype);
  }
}

export class NotFoundError extends K8sError {
  constructor(resource: string, name: string, namespace?: string) {
    const location = namespace ? `${namespace}/${name}` : name;
    super(`${resource} "${location}" not found`, 404);
    this.name = 'NotFoundError';
    Object.setPrototypeOf(this, NotFoundError.prototype);
  }
}

export class ConflictError extends K8sError {
  constructor(resource: string, name: string, namespace?: string) {
    const location = namespace ? `${namespace}/${name}` : name;
    super(`${resource} "${location}" already exists`, 409);
    this.name = 'ConflictError';
    Object.setPrototypeOf(this, ConflictError.prototype);
  }
}

export class ValidationError extends K8sError {
  constructor(message: string, public readonly errors?: string[]) {
    super(message, 400, errors);
    this.name = 'ValidationError';
    Object.setPrototypeOf(this, ValidationError.prototype);
  }
}

export class UnauthorizedError extends K8sError {
  constructor(message = 'Unauthorized') {
    super(message, 401);
    this.name = 'UnauthorizedError';
    Object.setPrototypeOf(this, UnauthorizedError.prototype);
  }
}

export class ForbiddenError extends K8sError {
  constructor(message = 'Forbidden') {
    super(message, 403);
    this.name = 'ForbiddenError';
    Object.setPrototypeOf(this, ForbiddenError.prototype);
  }
}

/**
 * Parse K8s API error response into custom error
 */
export function parseK8sError(err: unknown): K8sError {
  if (err instanceof K8sError) {
    return err;
  }

  if (err && typeof err === 'object') {
    const error = err as any;

    // Try to extract status code from various error formats
    let statusCode: number | undefined;
    let body: any;
    let message: string;

    // Check for statusCode in response or directly on error
    if (error.response?.statusCode || error.statusCode) {
      statusCode = error.response?.statusCode || error.statusCode;
      body = error.response?.body || error.body;
      message = body?.message || error.message || 'Unknown error';
    }
    // Check for Kubernetes Status response format (code in body)
    else if (error.body) {
      try {
        const parsedBody = typeof error.body === 'string' ? JSON.parse(error.body) : error.body;
        if (parsedBody.code) {
          statusCode = parsedBody.code;
          body = parsedBody;
          message = parsedBody.message || error.message || 'Unknown error';
        }
      } catch (e) {
        // Body is not valid JSON, continue with fallback
      }
    }

    // If we found a status code, return appropriate error
    if (statusCode) {
      switch (statusCode) {
        case 404:
          return new NotFoundError('Resource', message);
        case 409:
          return new ConflictError('Resource', message);
        case 401:
          return new UnauthorizedError(message);
        case 403:
          return new ForbiddenError(message);
        case 400:
          return new ValidationError(message);
        default:
          return new K8sError(message, statusCode, body);
      }
    }
  }

  const message = err instanceof Error ? err.message : 'Unknown error';
  return new K8sError(message);
}
