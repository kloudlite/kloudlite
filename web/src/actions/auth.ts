"use server";

import jwt from 'jsonwebtoken';
import { cookies } from "next/headers";
import { actionWrapper } from "./utils/action-wrapper";
import { grpcErrorToErr } from './utils/grpc-util';
import { authCli } from './_rpc-servers';

export const isLoggedIn = async () => {
  const cookieStore = await cookies();
  const userId = cookieStore.get("userId");
  if (!userId) return false;
  return userId.value !== "";
};

export const getCurrentUserId = async () => {
  const cookieStore = await cookies();
  const userId = cookieStore.get("userId");
  if (!userId) throw new Error("User is not logged in");
  return userId.value;
};

export const logout = async () => {
  const cookieStore = await cookies();
  cookieStore.delete("userId");
  return true;
};

export const login = actionWrapper(async (email: string, password: string) => {
  try {
    const res = await authCli.login({ email, password });
    const cookieStore = await cookies();
    cookieStore.set("userId", res.userId);
  } catch (error) {
    throw grpcErrorToErr(error);
  }
});

export const signup = actionWrapper(async (name: string, email: string, password: string) => {
  try {
    const res = await authCli.signup({ name, email, password });
    const cookieStore = await cookies();
    cookieStore.set("userId", res.userId);
    return true;
  } catch (error) {
    throw grpcErrorToErr(error);
  }
});

export const requestResetPassword = actionWrapper(async (email: string) => {
  try {
    const res = await authCli.requestResetPassword({ email });
    return res;
  } catch (error) {
    throw grpcErrorToErr(error);
  }
});

export const resetPassword = actionWrapper(async (newPassword: string, confirmPassword: string, token: string) => {
  if (newPassword !== confirmPassword) throw new Error("Passwords do not match");
  try {
    const res = await authCli.resetPassword({ newPassword, resetToken: token });
    return res;
  } catch (error) {
    throw grpcErrorToErr(error);
  }
});

export const verifyEmail = actionWrapper(async (token: string) => {
  try {
    const res = await authCli.verifyEmail({ verificationToken: token });
    return res;
  } catch (error) {
    throw grpcErrorToErr(error);
  }
});

export const resendEmailVerification = actionWrapper(async (email: string) => {
  try {
    const res = await authCli.resendEmailVerification({ email });
    return res;
  } catch (error) {
    throw grpcErrorToErr(error);
  }
});

export const loginWithSSO = actionWrapper(async (token: string) => {
  try {
    const data = await jwt.verify(token, process.env.JWT_SECRET || "") as { email: string, name: string };
    const res = await authCli.loginWithSso({ email: data.email, name: data.name });
    const cookieStore = await cookies();
    cookieStore.set("userId", res.userId);
  } catch (error) {
    if (error instanceof jwt.JsonWebTokenError) throw new Error("Invalid SSO token");
    throw grpcErrorToErr(error);
  }
});

export const loginWithOAuth = actionWrapper(async (name: string, email: string, provider: string) => {
  try {
    const res = await authCli.loginWithSso({ email, name });
    const cookieStore = await cookies();
    cookieStore.set("userId", res.userId);
  } catch (error) {
    if (error instanceof jwt.JsonWebTokenError) throw new Error("Invalid SSO token");
    throw grpcErrorToErr(error);
  }
});
