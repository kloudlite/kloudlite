import { Suspense } from 'react';
import { Meeting } from "./components/meeting";

export default async function App() {

    const dyteOrgId = process.env.DYTE_ORG_ID || "";
    const dyteApiKey = process.env.DYTE_API_KEY || "";


    return (
        <Suspense>
            <Meeting dyteOrgId={dyteOrgId} dyteApiKey={dyteApiKey} />
        </Suspense>
    );
}
