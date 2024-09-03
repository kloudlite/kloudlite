'use client';
import { ArrowRightLg } from '@jengaicons/react';
import { Button } from 'kl-design-system/atoms/button';
import { useSearchParams } from 'next/navigation';


export const JoinWebinar = () => {

    const params = useSearchParams();
    const userData = params.get('userData');
    const decodedUserData = atob(userData || '');
    const searchParams = new URLSearchParams(decodedUserData);
    const name = searchParams.get('name');
    const email = searchParams.get('email');

    return (
        <div
            className='flex flex-col items-stretch gap-3xl'>
            <Button
                size="lg"
                variant="primary"
                content={<span className="bodyLg-medium">Join</span>}
                suffix={<ArrowRightLg />}
                block
                onClick={() => {
                    window.location.href = `/pages/meeting?email=${email}&name=${name}&meetingId=${process.env.NEXT_PUBLIC_DYTE_MEETING_ID}`
                }}
            />
        </div>
    )
}