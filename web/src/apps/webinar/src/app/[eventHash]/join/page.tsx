import axios from 'axios';
import { Suspense } from 'react';
import { EventsUi } from './components/events-ui';
//@ts-ignore
import { cookies } from 'next/headers';
import { redirect } from "next/navigation";


type EnvVars = {
    dyteOrgId: string,
    dyteApiKey: string,
    dyteMeetingId: string,
    marketApiUrl: string,
}

export default async function App(props: any) {

    const dyteOrgId = process.env.DYTE_ORG_ID || "";
    const dyteApiKey = process.env.DYTE_API_KEY || "";

    const cookie = cookies().get("hotspot-session")
    const callbackUrl = props.searchParams.eventHashUrl || process.env.CALLBACK_URL;
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
                <Suspense>
                    <EventsUi userData={userDetails} dyteOrgId={dyteOrgId} dyteApiKey={dyteApiKey} />
                </Suspense>

            );
        }
        redirect(redirectUrl);
    } catch (e) {
        redirect(redirectUrl);
    }
}
