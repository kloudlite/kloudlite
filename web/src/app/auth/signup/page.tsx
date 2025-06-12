
import { notFound } from "next/navigation";
import SignupForm from "../_components/signup-form";

export default function Home() {
  if(process.env.SSO_AUTH == "true"){
    return notFound();
  }
  return <SignupForm />;
}
