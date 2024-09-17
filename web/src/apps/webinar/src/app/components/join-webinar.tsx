'use client';
import { ArrowRightLg } from '@jengaicons/react';
//@ts-ignore
import { Button } from 'kl-design-system/atoms/button';

type UesrData = {
    id: string,
    email: string,
    verified: boolean,
    name: string,
    approved: boolean,
}

export const JoinWebinar = ({ userData, meetingStatus }: { userData: UesrData, meetingStatus: string }) => {

    return (
        <div
            className='flex flex-col items-stretch gap-3xl'>
            <Button
                size="lg"
                variant="primary"
                content={<span className="bodyLg-medium">{meetingStatus === 'ACTIVE' ? 'Join' : 'Meeting is not active'}</span>}
                suffix={meetingStatus === 'ACTIVE' ? <ArrowRightLg /> : null}
                disabled={meetingStatus !== 'ACTIVE'}
                block
                onClick={() => {
                    window.location.href = `/pages/meeting?email=${userData.email}&name=${userData.name}&meetingId=${process.env.NEXT_PUBLIC_DYTE_MEETING_ID}`
                }}
            />
        </div>
    )
}
