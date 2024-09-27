import axios from 'axios';
//@ts-ignore
import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';
import { WebinarUI } from './orgs/webinar-ui';

async function getUserDetails() {
  const cookie = cookies().get("hotspot-session")

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
    if (!res.status) {
      console.log("failed to get user details");
    }
    return res.data.data;

  } catch (error) {
    console.log("error", error);
    return null;
  }

}

async function getMeetingDetails() {
  const token = btoa(`${process.env.DYTE_ORG_ID}:${process.env.DYTE_API_KEY}`);

  try {
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
      console.log("failed to get meeting details");
    }
    return data;
  } catch (error) {
    console.log("error", error);
    return null;
  }
}


export default async function Home() {
  const callbackUrl = process.env.CALLBACK_URL;
  const redirectUrl = `${process.env.REDIRECT_URL}?callback=${callbackUrl}`;

  const userDetails = await getUserDetails();
  const meetingDetails = await getMeetingDetails();

  if (userDetails) {
    return (
      <main className="flex flex-col h-full">
        <WebinarUI userDetails={userDetails} meetingStatus={meetingDetails.status} />
      </main>
    );
  }
  redirect(redirectUrl);

}
