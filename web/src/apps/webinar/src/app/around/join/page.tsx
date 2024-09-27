import axios from 'axios';
//@ts-ignore
import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';

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


export default async function Home(props: any) {
    const callbackUrl = props.searchParams.aroundUrl || process.env.CALLBACK_URL;;
    const redirectUrl = `${process.env.REDIRECT_URL}?callback=${callbackUrl}`;
    const aroundUrl = process.env.AROUND_MEETING_URL || "";

    const userDetails = await getUserDetails();

    if (userDetails) {
        redirect(aroundUrl);
    }
    redirect(redirectUrl);

}
