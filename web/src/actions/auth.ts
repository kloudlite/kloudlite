"use server";

import { AuthV2Client, LoginRequest, LoginResponse } from "@grpc/auth.v2";
import util from "util"
import * as grpc from "@grpc/grpc-js";
import { promiseWrap } from "./grpc-wrapper";

const AUTHV2_SERVER_ADDRESS = process.env.AUTHV2_SERVER_ADDRESS || "localhost:50051";

const AuthClient = new AuthV2Client(
  AUTHV2_SERVER_ADDRESS,
  grpc.credentials.createInsecure()
);

const serverMethods = {
  login: promiseWrap<LoginRequest, LoginResponse>(AuthClient.login.bind(AuthClient))
};

export const login = async (email: string, password: string) => {
  try {
    const res = await serverMethods.login({
      email, password
    });
    return res.userId
  } catch (error) {
    console.error("Login failed:", error);
  }
}