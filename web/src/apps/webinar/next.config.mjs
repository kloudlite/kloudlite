/** @type {import('next').NextConfig} */
const nextConfig = {
    env: {
        customKey: process.env.keyName, // pulls from .env file
    },
};

export default nextConfig;
