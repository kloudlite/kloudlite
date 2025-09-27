import * as grpc from '@grpc/grpc-js';

import { AuthClient } from '@/grpc/auth.external';

let authClient: AuthClient | null = null;

export function getAuthClient(): AuthClient {
  if (!authClient) {
    // Use 127.0.0.1 instead of localhost to avoid IPv6 issues
    // Default to port 3001 if not specified
    const backendUrl = process.env.BACKEND_URL || '127.0.0.1:3001';
    authClient = new AuthClient(
      backendUrl,
      grpc.credentials.createInsecure()
    );
  }
  return authClient;
}

// Helper function to promisify gRPC calls
export function promisifyGrpcCall<TRequest, TResponse>(
  method: (request: TRequest, callback: (error: grpc.ServiceError | null, response: TResponse) => void) => grpc.ClientUnaryCall,
  request: TRequest
): Promise<TResponse> {
  return new Promise((resolve, reject) => {
    method.call(authClient, request, (error, response) => {
      if (error) {
        reject(error);
      } else {
        resolve(response);
      }
    });
  });
}