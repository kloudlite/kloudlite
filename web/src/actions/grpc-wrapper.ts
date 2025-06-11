import { ClientUnaryCall, ServiceError } from "@grpc/grpc-js";

export function promiseWrap<Req, Res>(fn: (request: Req, callback: (error: ServiceError | null, response: Res) => void) => ClientUnaryCall): (req: Req) => Promise<Res> {
  return (req) => {
    return new Promise<Res>((resolve, reject) => {
      fn(req, (error, response) => {
        if (error) {
          reject(error);
        } else {
          resolve(response);
        }
      });
    })
  };
}