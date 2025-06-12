
import { SSOCheck } from './sso-check';

export default async function SSOLogin({searchParams}: { searchParams: Promise<{ [key: string]: string }> }) {
  const params = await searchParams;
  return <SSOCheck token={params.token} />
}