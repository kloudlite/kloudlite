import {
  type ServiceError,
} from "@grpc/grpc-js";


export function promiseWrap<Req, Res>(
  fn: (
    request: Req,
    callback: (error: ServiceError | null, response: Res) => void
  ) => void,
) {
  return (request:Req)=>{
    return new Promise((resolve, reject) => {
      fn(request, (err, res) => {
        if (err) {
          reject(err);
        } else {
          resolve(res!);
        }
      });
    })
  };
}