'use client';
import { ArrowRightLg } from '@jengaicons/react';
import axios from 'axios';
//@ts-ignore
import { Button } from 'kl-design-system/atoms/button';
import { useEffect, useState } from 'react';

type UesrData = {
    id: string,
    email: string,
    verified: boolean,
    name: string,
    approved: boolean,
}

export const JoinWebinar = ({ userData }: { userData: UesrData }) => {

    const [meetingStatus, setMeetingStatus] = useState<boolean>(false);

    const getMeetingDetails = async () => {
        const token = btoa(`${process.env.NEXT_PUBLIC_DYTE_ORG_ID}:${process.env.NEXT_PUBLIC_DYTE_API_KEY}`);
        try {
            const { data: { success, data } } = await axios.get(
                `https://api.dyte.io/v2/meetings/${process.env.NEXT_PUBLIC_DYTE_MEETING_ID}`,
                {
                    headers: {
                        Authorization: `Basic ${token}`,
                    }
                }
            )
            if (!success) {
                throw new Error('Failed to get meeting details');
            }
            console.log("++++++>>>>>>>>>>>>>>", data);

            setMeetingStatus(data.status === 'ACTIVE');
        } catch (error) {
            console.error(error);
        }
    }

    useEffect(() => {
        (async () => {
            await getMeetingDetails();
        })();
    }, []);


    return (
        <div
            className='flex flex-col items-stretch gap-3xl'>
            {
                meetingStatus && (
                    <Button
                        size="lg"
                        variant="primary"
                        content={<span className="bodyLg-medium">Join</span>}
                        suffix={<ArrowRightLg />}
                        block
                        onClick={() => {
                            window.location.href = `/pages/meeting?email=${userData.email}&name=${userData.name}&meetingId=${process.env.NEXT_PUBLIC_DYTE_MEETING_ID}`
                        }}
                    />
                )
            }
        </div>
    )
}
