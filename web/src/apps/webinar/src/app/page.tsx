import axios from 'axios';
//@ts-ignore
import { cn } from 'kl-design-system/utils';
import { cookies } from 'next/headers';
import { redirect } from "next/navigation";
import { JoinWebinar } from './components/join-webinar';


export default async function Home() {

  const cookie = cookies().get("hotspot-session")
  // const callbackUrl = "https://auth-piyush.dev.kloudlite.io";
  // const redirectUrl = "https://auth.dev.kloudlite.io/login?callback=" + callbackUrl;
  const callbackUrl = process.env.NEXT_PUBLIC_Callback_URL;
  const redirectUrl = `${process.env.NEXT_PUBLIC_REDIRECT_URL}?callback=${callbackUrl}`;
  const token = btoa(`${process.env.NEXT_PUBLIC_DYTE_ORG_ID}:${process.env.NEXT_PUBLIC_DYTE_API_KEY}`);

  try {
    const res = await axios({
      url: `${process.env.NEXT_PUBLIC_AUTH_URL}/api` || 'https://auth.kloudlite.io/api',
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
      `https://api.dyte.io/v2/meetings/${process.env.NEXT_PUBLIC_DYTE_MEETING_ID}`,
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
      return (
        <main className='flex flex-col h-full'>
          <div className='flex flex-1 flex-col md:items-center self-stretch justify-center px-3xl py-5xl md:py-9xl'>
            <div className='flex flex-col gap-3xl md:w-[500px] px-3xl py-5xl md:px-9xl'>
              <div className="flex flex-col items-stretch gap-lg">
                <div className="flex flex-col gap-lg items-center pb-6xl text-center">
                  <div className={cn('text-text-strong headingXl text-center')}>
                    Join Kloudlite webinar
                  </div>
                  <div className="bodyMd-medium text-text-soft">
                    Join webinar and experience the power of Kloudlite
                  </div>
                </div>
                <JoinWebinar userData={userDetails} meetingStatus={data.status} />
              </div>
            </div>
          </div>
        </main >
      );
    }
    redirect(redirectUrl);
  } catch (e) {
    redirect(redirectUrl);
  }

}
