import jwt from 'jsonwebtoken';
import ExternalLoginForm from '../_components/external-login-form';

export default function ExternalLoginPage() {
  async function generateToken(email:string, name:string) {
    "use server";
    return jwt.sign({ email, name }, process.env.JWT_SECRET||"");
  }
  return <ExternalLoginForm tokenGenerator={generateToken} />;
}