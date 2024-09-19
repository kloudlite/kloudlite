"use client";
import { DyteProvider, useDyteClient } from '@dytesdk/react-web-core';
import axios from 'axios';
import { useSearchParams } from 'next/navigation';
import { useEffect, useState } from 'react';
import { MyMeetingUI } from '../../../orgs/my-meeting-ui';


export const Meeting = ({ dyteOrgId, dyteApiKey }: { dyteOrgId: string, dyteApiKey: string }) => {
    const [meeting, initMeeting] = useDyteClient();
    const [authToken, setAuthToken] = useState('');

    const params = useSearchParams();

    const handleJoinMeeting = async ({ name, email, meetingId }: { name: string, email: string, meetingId: string }) => {
        console.log("dyteOrgId", dyteOrgId, "kkk");
        console.log("dyteApiKey", dyteApiKey, "kkk");
        // const token = btoa(`${process.env.NEXT_PUBLIC_DYTE_ORG_ID}:${process.env.NEXT_PUBLIC_DYTE_API_KEY}`);
        const token = btoa(`${dyteOrgId}:${dyteApiKey}`);
        try {
            const { data: { success, data } } = await axios.post(
                `https://api.dyte.io/v2/meetings/${meetingId}/participants`,
                {
                    name,
                    picture: "https://i.imgur.com/test.jpg",
                    preset_name: "webinar_viewer",
                    custom_participant_id: email,
                },
                {
                    headers: {
                        Authorization: `Basic ${token}`,
                    },
                }
            );
            if (!success) {
                throw new Error('Failed to join meeting');
            }
            setAuthToken(data.token);
        } catch (error) {
            console.error(error);
        }
    };

    useEffect(() => {
        const meetingId = params.get('meetingId');
        const name = params.get('name');
        const email = params.get('email');

        if (meetingId && name && email) {
            (async () => {
                await handleJoinMeeting({ name, email, meetingId });
            })();
        }
    }, [params]);

    useEffect(() => {
        if (authToken) {
            initMeeting({
                authToken: authToken,
                defaults: {
                    audio: false,
                    video: false,
                },
            });
        }
    }, [authToken]);

    return (
        <DyteProvider value={meeting}>
            <MyMeetingUI />
        </DyteProvider>
    );
}