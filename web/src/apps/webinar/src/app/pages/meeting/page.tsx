//@ts-ignore
import { cookies } from 'next/headers';
import { Suspense } from 'react';
import { Meeting } from "./components/meeting";

export default async function App() {
    // NOTE: cookie is defining here just because to get environment variables,
    // as next js is not allowing to access environment variables in server side if we will not make server call.
    const cookie = cookies().get("hotspot-session")

    return (
        <Suspense>
            <Meeting dyteOrgId={process.env.DYTE_ORG_ID || ""} dyteApiKey={process.env.DYTE_API_KEY || ""} />
        </Suspense>
    );

}

