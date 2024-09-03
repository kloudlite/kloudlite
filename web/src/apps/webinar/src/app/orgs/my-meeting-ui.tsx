
import { DyteMeeting } from '@dytesdk/react-ui-kit';
import { useDyteMeeting } from '@dytesdk/react-web-core';
import React from 'react';

export const MyMeetingUI = () => {
    const { meeting } = useDyteMeeting();
    return (
        <DyteMeeting mode='fill' meeting={meeting} showSetupScreen={false} />
    );
}
