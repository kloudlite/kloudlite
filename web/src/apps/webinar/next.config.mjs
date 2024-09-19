/** @type {import('next').NextConfig} */
const nextConfig = {
    env: {
        // customKey: process.env.keyName,
        NEXT_PUBLIC_DYTE_ORG_ID: process.env.NEXT_PUBLIC_DYTE_ORG_ID,
        NEXT_PUBLIC_DYTE_API_KEY: process.env.NEXT_PUBLIC_DYTE_API_KEY,
        NEXT_PUBLIC_DYTE_MEETING_ID: process.env.NEXT_PUBLIC_DYTE_MEETING_ID,
        NEXT_PUBLIC_MARKETING_API_URL: process.env.NEXT_PUBLIC_MARKETING_API_URL,
    },
};

export default nextConfig;
