import axios from 'axios';
//@ts-ignore
import { cn } from 'kl-design-system/utils';
import { cookies } from 'next/headers';
import { redirect } from "next/navigation";
import { JoinWebinar } from './components/join-webinar';

// interface HomeProps {
//   userData: UserData | null; // Adjust this based on whether userData can be null
// }

export default async function Home() {

  const cookie = cookies().get("hotspot-session")
  const callbackUrl = "https://auth-piyush.dev.kloudlite.io";
  const redirectUrl = "https://auth.dev.kloudlite.io/login?callback=" + callbackUrl;
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
    const data = res.data.data;
    if (data) {
      return (
        <main className='flex flex-col h-full'>
          <div className='flex flex-1 flex-col md:items-center self-stretch justify-center px-3xl py-5xl md:py-9xl'>
            <div className='flex flex-col gap-3xl md:w-[500px] px-3xl py-5xl md:px-9xl'>
              <div className='flex flex-col items-stretch"'>
                <div className="flex flex-col gap-lg items-center pb-6xl text-center">
                  <div className={cn('text-text-strong headingXl text-center')}>
                    Join Kloudlite webinar
                  </div>
                  <div className="bodyMd-medium text-text-soft">
                    Join webinar and experience the power of Kloudlite
                  </div>
                </div>
                <JoinWebinar userData={data} />
                {data.email}
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
