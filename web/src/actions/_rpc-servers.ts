import { AuthClient } from "@grpc/auth.external";
import { createClient } from "./utils/grpc-util";
import { AccountsClient } from "@grpc/accounts.external";

const AUTH_SERVER_ADDRESS = process.env.AUTH_SERVER_ADDRESS || "localhost:50051";
export const authCli = createClient(AuthClient, AUTH_SERVER_ADDRESS);

const ACCOUNT_SERVER_ADDRESS = process.env.ACCOUNT_SERVER_ADDRESS || "localhost:50052";
export const accountsCli = createClient(AccountsClient, ACCOUNT_SERVER_ADDRESS);