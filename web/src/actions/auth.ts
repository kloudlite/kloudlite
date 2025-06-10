"use server";

import { AuthV2Client } from "@grpc/auth.v2";
import * as grpc from "@grpc/grpc-js";
import { promiseWrap } from "./grpc-wrapper";

const AUTHV2_SERVER_ADDRESS = process.env.AUTHV2_SERVER_ADDRESS || "localhost:50051";

const AuthClient = new AuthV2Client(
  AUTHV2_SERVER_ADDRESS,
  grpc.credentials.createInsecure()
);


export const login = async (email: string, password: string) => {
  const res = await promiseWrap(AuthClient.login)(
    {
      email,
      password
    }
  )
  console.log("Login response:", res);
}