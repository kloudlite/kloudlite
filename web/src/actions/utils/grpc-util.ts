import util from 'util';
import * as grpc from "@grpc/grpc-js";

type Promisified<T> = {
  [K in keyof T]: T[K] extends (
    request: infer Req,
    metadata: grpc.Metadata,
    options: Partial<grpc.CallOptions>,
    callback: (error: any, result: infer Res) => void
  ) => void
    ? (request: Req) => Promise<Res>
    : T[K] extends (...args: infer A) => any
    ? A extends [...infer B, (error: any, result: infer R) => void]
      ? (...args: B) => Promise<R>
      : T[K]
    : T[K];
};

export function createClient<T>(ClientClass: new (address: string, credentials: grpc.ChannelCredentials) => T, address: string): Promisified<T> {
  return promisifyAll(new ClientClass(address, grpc.credentials.createInsecure()));
}

export function grpcErrorToErr(error: any): Error {
  const grpcError = error as grpc.ServiceError;
  // grpcError.cause = grpcError.details;
  return new Error(grpcError.details);
}

export function promisifyAll<T>(obj: T) {
  const promisifiedObj = {} as any;
  for (const key in obj) {
    if (typeof obj[key] === 'function') {
      // Bind the method to its object.
      const fn = (obj as any)[key].bind(obj);
      // If the function expects (request, metadata, options, callback)
      if (fn.length === 4) {
        promisifiedObj[key] = (request: any) => {
          return new Promise((resolve, reject) => {
            fn(request, new grpc.Metadata(), {}, (error: any, result: any) => {
              if (error) {
                reject(error);
              } else {
                resolve(result);
              }
            });
          });
        };
      } else {
        // Fallback to util.promisify for other functions.
        promisifiedObj[key] = util.promisify(fn);
      }
    } else {
      promisifiedObj[key] = obj[key];
    }
  }
  return promisifiedObj as Promisified<T>;
}