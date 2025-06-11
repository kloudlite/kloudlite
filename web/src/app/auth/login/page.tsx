import LoginForm from "../_components/login-form";

export default function Home() {
  const withSSO = process.env.SSO_AUTH=== "true";
  const emailCommEnabled = process.env.EMAIL_COMM_ENABLED === "true";
  return <LoginForm withSSO={withSSO} emailCommEnabled={emailCommEnabled} />;
}
