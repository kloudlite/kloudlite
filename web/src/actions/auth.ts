"use server";

import { AuthV2Client, LoginRequest, LoginResponse, SignupRequest, SignupResponse } from "@grpc/auth.v2";
import util from "util"
import * as grpc from "@grpc/grpc-js";
import { promiseWrap } from "./grpc-wrapper";
import { cookies } from "next/headers";

const AUTHV2_SERVER_ADDRESS = process.env.AUTHV2_SERVER_ADDRESS || "localhost:50051";

const AuthClient = new AuthV2Client(
  AUTHV2_SERVER_ADDRESS,
  grpc.credentials.createInsecure()
);

const serverMethods = {
  login: promiseWrap<LoginRequest, LoginResponse>(AuthClient.login.bind(AuthClient)),
  signup: promiseWrap<SignupRequest, SignupResponse>(AuthClient.signup.bind(AuthClient))
};

export const login = async (email: string, password: string) => {
  try {
    console.log("Attempting to login with email:", email);
    const res = await serverMethods.login({
      email, password
    });
    const cookieStore = await cookies()
    cookieStore.set("userId", res.userId)
    return [true, null];
  } catch (error) {
    return [false, util.inspect(error)];
  }
}


export const signup = async (name:string, email: string, password: string) => {
  try {
    const res = await serverMethods.signup({
      name, email, password
    });
    const cookieStore = await cookies()
    cookieStore.set("userId", res.userId)
    return [true, null];
  } catch (error) {
    return [false, util.inspect(error)];
  }
}