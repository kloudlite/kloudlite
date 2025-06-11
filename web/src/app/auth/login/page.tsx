import LoginForm from "../_components/login-form";

export default function Home() {
  const withSSO = process.env.SSO_AUTH=== "true";
  return <LoginForm withSSO={withSSO} />;
}
