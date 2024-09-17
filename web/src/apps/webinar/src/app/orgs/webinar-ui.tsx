"use client";
//@ts-ignore
import { Button } from 'kl-design-system/atoms/button';
//@ts-ignore
import { cn } from 'kl-design-system/utils';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import Container from '../components/container';
import { JoinWebinar } from '../components/join-webinar';

export const WebinarUI = ({ userDetails, meetingStatus }: { userDetails: any, meetingStatus: string }) => {
    return (
        <Container
            headerExtra={
                <Button
                    variant="outline"
                    content="Register"
                    // linkComponent={Link}
                    to="/login"
                />
            }
        >
            <div className='flex flex-1 flex-col md:items-center self-stretch justify-center px-3xl py-5xl md:py-9xl'>
                <div className='flex flex-col gap-3xl md:w-[500px] px-3xl py-5xl md:px-9xl'>
                    <div className="flex flex-col items-stretch gap-lg">
                        <div className="flex flex-col gap-lg items-center pb-6xl text-center">
                            <div className={cn('text-text-strong headingXl text-center')}>
                                Join Kloudlite webinar
                            </div>
                            <div className="bodyMd-medium text-text-soft">
                                Join webinar and experience the power of Kloudlite
                            </div>
                        </div>
                        <JoinWebinar userData={userDetails} meetingStatus={meetingStatus} />
                    </div>
                </div>
            </div>
        </Container>
    )
}

const HandleRegisterForm = ({ visible, setVisible }: { visible: boolean, setVisible: (v: boolean) => void }) => {

    const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
        useForm({
            initialValues: {
                name: '',
                companyName: '',
                email: '',
            },

            validationSchema: Yup.object({
                name: Yup.string().required('id is required'),
            }),
            onSubmit: async (val) => {
                try {
                    // if (!isUpdate) {
                    //     const { errors: e } = await api.createIotDeployment({
                    //         projectName: project.name,
                    //         deployment: {
                    //             name: val.name,
                    //             displayName: val.displayName,
                    //             CIDR: val.cidr,
                    //             exposedIps: val.exposedIps,
                    //             exposedDomains: val.exposedDomains,
                    //             exposedServices: val.exposedServices.map((service) => {
                    //                 return {
                    //                     name: service.name,
                    //                     ip: service.ip,
                    //                 };
                    //             }),
                    //         },
                    //     });
                    //     if (e) {
                    //         throw e[0];
                    //     }
                    // } else {
                    //     const { errors: e } = await api.updateIotDeployment({
                    //         projectName: project.name,
                    //         deployment: {
                    //             name: val.name,
                    //             displayName: val.displayName,
                    //             CIDR: val.cidr,
                    //             exposedIps: val.exposedIps,
                    //             exposedDomains: val.exposedDomains,
                    //             exposedServices: val.exposedServices.map((service) => {
                    //                 return {
                    //                     name: service.name,
                    //                     ip: service.ip,
                    //                 };
                    //             }),
                    //         },
                    //     });
                    //     if (e) {
                    //         throw e[0];
                    //     }
                    // }
                    // reloadPage();
                    resetValues();
                    // toast.success(
                    //     `deployment ${isUpdate ? 'updated' : 'created'} successfully`
                    // );
                    setVisible(false);
                } catch (err) {
                    handleError(err);
                }
            },
        });

}