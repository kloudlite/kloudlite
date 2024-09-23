"use client";
import { DyteProvider, useDyteClient } from '@dytesdk/react-web-core';
import axios from 'axios';
import { useParams } from 'next/navigation';
import { useEffect, useState } from 'react';
import eventsData from '~/root/lib/shared-statics/events.json';
import { MyMeetingUI } from '../../../orgs/my-meeting-ui';
//@ts-ignore

type UesrData = {
    id: string,
    email: string,
    verified: boolean,
    name: string,
    approved: boolean,
}

export const EventsUi = ({ userData, dyteOrgId, dyteApiKey }: { userData: UesrData, dyteOrgId: string, dyteApiKey: string }) => {
    const [meeting, initMeeting] = useDyteClient();
    const [authToken, setAuthToken] = useState('');

    const params = useParams();
    const selectedEventHash = params.eventHash as string;

    const handleJoinMeeting = async ({ name, email, meetingId }: { name: string, email: string, meetingId: string }) => {
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
        //@ts-ignore
        const selectedEvents = eventsData[selectedEventHash];
        const meetingId = selectedEvents.dyteMeetingId;
        const name = userData.name;
        const email = userData.email;

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