import axios from 'axios';
//@ts-ignore
import { cookies } from 'next/headers';
import { redirect } from "next/navigation";
import { WebinarUI } from './orgs/webinar-ui';

type EnvVars = {
  dyteOrgId: string,
  dyteApiKey: string,
  dyteMeetingId: string,
  marketApiUrl: string,
}

export default async function Home() {

  // console.log("all-env", process.env)
  const cookie = cookies().get("hotspot-session")
  const callbackUrl = process.env.CALLBACK_URL;
  const redirectUrl = `${process.env.REDIRECT_URL}?callback=${callbackUrl}`;
  const token = btoa(`${process.env.DYTE_ORG_ID}:${process.env.DYTE_API_KEY}`);

  try {
    const res = await axios({
      url: `${process.env.AUTH_URL}/api` || 'https://auth.kloudlite.io/api',
      method: 'post',
      data: {
        method: 'whoAmI',
        args: [{}],
      },
      headers: {
        'Content-Type': 'application/json; charset=utf-8',
        connection: 'keep-alive',
        cookie: 'hotspot-session=' + cookie?.value + ';',
      },
    });

    const { data: { success, data } } = await axios.get(
      `https://api.dyte.io/v2/meetings/${process.env.DYTE_MEETING_ID}`,
      {
        headers: {
          Authorization: `Basic ${token}`,
        }
      }
    )
    if (!success) {
      throw new Error('Failed to get meeting details');
    }

    const userDetails = res.data.data;
    if (userDetails) {

      const envVars: EnvVars = {
        dyteOrgId: process.env.DYTE_ORG_ID as string,
        dyteApiKey: process.env.DYTE_API_KEY as string,
        dyteMeetingId: process.env.DYTE_MEETING_ID as string,
        marketApiUrl: process.env.MARKETING_API_URL as string,
      }

      return (
        <main className='flex flex-col h-full'>
          <WebinarUI userDetails={userDetails} meetingStatus={data.status} envVars={envVars} />
        </main >
      );
    }
    redirect(redirectUrl);
  } catch (e) {
    redirect(redirectUrl);
  }

}
