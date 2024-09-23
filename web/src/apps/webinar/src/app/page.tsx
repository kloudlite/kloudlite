import axios from 'axios';
//@ts-ignore
import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';
import { WebinarUI } from './orgs/webinar-ui';


export default async function Home() {

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

    const {
      data: { success, data },
    } = await axios.get(
      `https://api.dyte.io/v2/meetings/${process.env.DYTE_MEETING_ID}`,
      {
        headers: {
          Authorization: `Basic ${token}`,
        },
      },
    );
    if (!success) {
      throw new Error('Failed to get meeting details');
    }

    const userDetails = res.data.data;
    if (userDetails) {
      return (
        <main className="flex flex-col h-full">
          <WebinarUI userDetails={userDetails} meetingStatus={data.status} />
        </main>
      );
    }
    redirect(redirectUrl);
  } catch (e) {
    redirect(redirectUrl);
  }
}
