"use client";
//@ts-ignore
import { Button } from 'kl-design-system/atoms/button';
//@ts-ignore
import { TextInput } from 'kl-design-system/atoms/input';
//@ts-ignore
import Popup from 'kl-design-system/molecule/popup';
//@ts-ignore
import { cn } from 'kl-design-system/utils';
// import Yup from '~/root/lib/server/helpers/yup';
// import { handleError } from '~/root/lib/utils/common';
import { useState } from 'react';
import Container from '../components/container';
import { JoinWebinar } from '../components/join-webinar';



export const WebinarUI = ({ userDetails, meetingStatus }: { userDetails: any, meetingStatus: string }) => {

    const [visible, setVisible] = useState(false);

    return (
        <Container
            headerExtra={
                <Button
                    variant="outline"
                    content="Register"
                    onClick={() => {
                        setVisible(true);
                    }}
                // linkComponent={Link}
                // to="/login"
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
                {visible && <HandleRegisterForm visible={visible} setVisible={setVisible} />}
            </div>
        </Container>
    )
}

const HandleRegisterForm = ({ visible, setVisible }: { visible: boolean, setVisible: (v: boolean) => void }) => {

    // const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    //     useForm({
    //         initialValues: {
    //             name: '',
    //             companyName: '',
    //             email: '',
    //         },

    //         // validationSchema: Yup.object({
    //         //     // name: Yup.string().required('id is required'),
    //         // }),
    //         validationSchema: null,

    //         onSubmit: async (val) => {
    //             try {
    //                 // if (!isUpdate) {
    //                 //     const { errors: e } = await api.createIotDeployment({
    //                 //         projectName: project.name,
    //                 //         deployment: {
    //                 //             name: val.name,
    //                 //             displayName: val.displayName,
    //                 //             CIDR: val.cidr,
    //                 //             exposedIps: val.exposedIps,
    //                 //             exposedDomains: val.exposedDomains,
    //                 //             exposedServices: val.exposedServices.map((service) => {
    //                 //                 return {
    //                 //                     name: service.name,
    //                 //                     ip: service.ip,
    //                 //                 };
    //                 //             }),
    //                 //         },
    //                 //     });
    //                 //     if (e) {
    //                 //         throw e[0];
    //                 //     }
    //                 // } else {
    //                 //     const { errors: e } = await api.updateIotDeployment({
    //                 //         projectName: project.name,
    //                 //         deployment: {
    //                 //             name: val.name,
    //                 //             displayName: val.displayName,
    //                 //             CIDR: val.cidr,
    //                 //             exposedIps: val.exposedIps,
    //                 //             exposedDomains: val.exposedDomains,
    //                 //             exposedServices: val.exposedServices.map((service) => {
    //                 //                 return {
    //                 //                     name: service.name,
    //                 //                     ip: service.ip,
    //                 //                 };
    //                 //             }),
    //                 //         },
    //                 //     });
    //                 //     if (e) {
    //                 //         throw e[0];
    //                 //     }
    //                 // }
    //                 // reloadPage();
    //                 resetValues();
    //                 // toast.success(
    //                 //     `deployment ${isUpdate ? 'updated' : 'created'} successfully`
    //                 // );
    //                 setVisible(false);
    //             } catch (err) {
    //                 // handleError(err);
    //             }
    //         },
    //     });

    return (
        <Popup.Form
        // onSubmit={(e) => {
        //     if (!values.isNameError) {
        //         handleSubmit(e);
        //     } else {
        //         e.preventDefault();
        //     }
        // }}
        >
            <Popup.Content>
                <div className="flex flex-col gap-2xl">
                    {/* <NameIdView
                        resType="project"
                        displayName={values.displayName}
                        name={values.name}
                        label="Deployment name"
                        placeholder="Enter deployment name"
                        errors={errors.name}
                        handleChange={handleChange}
                        nameErrorLabel="isNameError"
                        isUpdate={isUpdate}
                    /> */}
                    <TextInput
                        label="Full name"
                        size="lg"
                        placeholder="name"
                    // value={values.name}
                    // onChange={handleChange('name')}
                    />

                    {/* <Select
                        creatable
                        size="lg"
                        label="Exposed ips"
                        multiple
                        value={values.exposedIps || []}
                        options={async () => []}
                        onChange={(val, v) => {
                            handleChange('exposedIps')(dummyEvent(v));
                        }}
                        error={!!errors.exposedIps}
                        disableWhileLoading
                    /> */}

                    {/* <Select
                        creatable
                        size="lg"
                        label="Exposed domains"
                        multiple
                        value={values.exposedDomains || []}
                        options={async () => []}
                        onChange={(val, v) => {
                            handleChange('exposedDomains')(dummyEvent(v));
                        }}
                        error={!!errors.exposedDomains}
                        disableWhileLoading
                    /> */}

                    {/* <KeyValuePair
                        size="lg"
                        label="Exposed services"
                        value={values.exposedServices || []}
                        onChange={(items, _) => {
                            handleChange('exposedServices')(dummyEvent(items));
                        }}
                        keyLabel="name"
                        valueLabel="ip"
                        error={!!errors.exposedServices}
                        message={errors.exposedServices}
                    /> */}
                </div>
            </Popup.Content>
            <Popup.Footer>
                <Popup.Button closable content="Cancel" variant="basic" />
                <Popup.Button
                    // loading={isLoading}
                    type="submit"
                    // content={isUpdate ? 'Update' : 'Create'}
                    variant="primary"
                />
            </Popup.Footer>
        </Popup.Form>
    )

}