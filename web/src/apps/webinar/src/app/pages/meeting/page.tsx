import { Suspense } from 'react';
import { Meeting } from "./components/meeting";

export default async function App() {

    // const response = await fetch(`https://auth-piyush.dev.kloudlite.io/api/get-envs`, {
    //     method: 'GET',
    //     cache: 'no-store'
    // });

    // const { dyteOrgId, dyteApiKey } = await response.json();

    // // Logging for debugging purposes
    // console.log('Dyte Org ID:', dyteOrgId, "ttt");
    // console.log('Dyte API Key:', dyteApiKey, "ttt");

    const dyteOrgId = process.env.DYTE_ORG_ID || "";
    const dyteApiKey = process.env.DYTE_API_KEY || "";


    return (
        <Suspense>
            <Meeting dyteOrgId={dyteOrgId} dyteApiKey={dyteApiKey} />
        </Suspense>
    );
}
