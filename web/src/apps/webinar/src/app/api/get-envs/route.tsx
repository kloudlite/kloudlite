import { NextResponse } from "next/server";

export async function GET() {
    // console.log("all-env", process.env)

    const dyteOrgId = process.env.NEXT_PUBLIC_DYTE_ORG_ID;
    const dyteApiKey = process.env.NEXT_PUBLIC_DYTE_API_KEY;
    const dyteMeetingId = process.env.NEXT_PUBLIC_DYTE_MEETING_ID;
    const marketApiUrl = process.env.NEXT_PUBLIC_MARKETING_API_URL;

    console.log("hello=====>>>", dyteOrgId, dyteApiKey, dyteMeetingId, marketApiUrl)


    return NextResponse.json({
        dyteOrgId,
        dyteApiKey,
        dyteMeetingId,
        marketApiUrl,
    });
}