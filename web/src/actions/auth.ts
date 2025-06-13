"use server";

import jwt from 'jsonwebtoken';
import { AuthV2Client, LoginRequest, LoginResponse, LoginWithSSORequest, LoginWithSSOResponse, RequestResetPasswordRequest, RequestResetPasswordResponse, ResetPasswordRequest, ResetPasswordResponse, SignupRequest, SignupResponse, VerifyEmailRequest, VerifyEmailResponse } from "@grpc/auth.v2";
import util from "util"
import * as grpc from "@grpc/grpc-js";
import { promiseWrap } from "./utils/grpc-wrapper";
import { cookies } from "next/headers";
import { actionWrapper } from "./utils/action-wrapper";

const AUTHV2_SERVER_ADDRESS = process.env.AUTHV2_SERVER_ADDRESS || "localhost:50051";

const AuthClient = new AuthV2Client(
  AUTHV2_SERVER_ADDRESS,
  grpc.credentials.createInsecure()
);

const serverMethods = {
  login: promiseWrap<LoginRequest, LoginResponse>(AuthClient.login.bind(AuthClient)),
  loginWithSSO: promiseWrap<LoginWithSSORequest, LoginWithSSOResponse>(AuthClient.loginWithSso.bind(AuthClient)),
  signup: promiseWrap<SignupRequest, SignupResponse>(AuthClient.signup.bind(AuthClient)),
  requestResetPassword: promiseWrap<RequestResetPasswordRequest, RequestResetPasswordResponse>(AuthClient.requestResetPassword.bind(AuthClient)),
  resetPassword: promiseWrap<ResetPasswordRequest, ResetPasswordResponse>(AuthClient.resetPassword.bind(AuthClient)),
  verifyEmail: promiseWrap<VerifyEmailRequest, VerifyEmailResponse>(AuthClient.verifyEmail.bind(AuthClient)),
  resendEmailVerification: promiseWrap<{ email: string }, { success: boolean }>(AuthClient.resendEmailVerification.bind(AuthClient)),
};

export const isLoggedIn = async () => {
  const cookieStore = await cookies();
  const userId = cookieStore.get("userId");
  if (!userId) {
    return false;
  }
  return userId.value !== "";
};

export const getUserId = async () => {
  const cookieStore = await cookies();
  const userId = cookieStore.get("userId");
  if (!userId) {
    throw new Error("User is not logged in");
  }
  return userId.value;
}

export const logout = async () => {
  const cookieStore = await cookies();
  cookieStore.delete("userId");
  return true;
};

export const login = actionWrapper(async (email: string, password: string) => {
  try {
    const res = await serverMethods.login({
      email, password
    });
    const cookieStore = await cookies()
    cookieStore.set("userId", res.userId);
  } catch (error) {
    throw new Error((error as grpc.ServiceError).details);
  }
});

export const signup = async (name: string, email: string, password: string) => {
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
};

export const requestResetPassword = actionWrapper(async (email: string) => {
  try {
    const res = await serverMethods.requestResetPassword({ email });
    return res;
  } catch (error) {
    throw new Error((error as grpc.ServiceError).details);
  }
});

export const resetPassword = actionWrapper(async (newPassword: string, confirmPassword: string, token: string) => {
  if (newPassword !== confirmPassword) {
    throw new Error("Passwords do not match");
  }
  try {
    const res = await serverMethods.resetPassword({ newPassword, resetToken: token });
    return res;
  } catch (error) {
    throw new Error((error as grpc.ServiceError).details);
  }
});

export const verifyEmail = actionWrapper(async (token: string) => {
  try {
    const res = await serverMethods.verifyEmail({ verificationToken: token });
    return res;
  } catch (error) {
    throw new Error((error as grpc.ServiceError).details);
  }
});

export const resendEmailVerification = actionWrapper(async (email: string) => {
  try {
    const res = await serverMethods.resendEmailVerification({ email });
    return res;
  } catch (error) {
    throw new Error((error as grpc.ServiceError).details);
  }
});

export const loginWithSSO = actionWrapper(async (token: string) => {
  try {
    const data = await jwt.verify(token, process.env.JWT_SECRET || "") as { email: string, name: string };
    const res = await serverMethods.loginWithSSO({
      email: data.email, name: data.name
    });
    const cookieStore = await cookies()
    cookieStore.set("userId", res.userId);
  } catch (error) {
    if (error instanceof jwt.JsonWebTokenError) {
      throw new Error("Invalid SSO token");
    }
    throw new Error((error as grpc.ServiceError).details);
  }
})


export const loginWithOAuth = actionWrapper(async (name:string, email:string, provider: string) => {
  try {
    const res = await serverMethods.loginWithSSO({
      email: email, name: name
    });
    const cookieStore = await cookies()
    cookieStore.set("userId", res.userId);
  } catch (error) {
    if (error instanceof jwt.JsonWebTokenError) {
      throw new Error("Invalid SSO token");
    }
    throw new Error((error as grpc.ServiceError).details);
  }
})