/* eslint-disable react/destructuring-assignment */
import { toast } from 'react-toastify';
import { TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { IDialogBase } from '~/iotconsole/components/types.d';
import { ExtractNodeType } from '~/iotconsole/server/r-utils/common';
import { NameIdView } from '~/iotconsole/components/name-id-view';
import { useIotConsoleApi } from '~/iotconsole/server/gql/api-provider';
import { useOutletContext } from '@remix-run/react';
import { IDevices } from '~/iotconsole/server/gql/queries/iot-device-queries';
import { IProjectContext } from '../../../_layout';
import { IDeploymentContext } from '../_layout';

type IDialog = IDialogBase<ExtractNodeType<IDevices>>;

const Root = (props: IDialog) => {
  const { setVisible, isUpdate } = props;

  const api = useIotConsoleApi();
  const reloadPage = useReload();
  const { project } = useOutletContext<IProjectContext>();
  const { deployment } = useOutletContext<IDeploymentContext>();

  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: isUpdate
        ? {
            displayName: props.data.displayName,
            name: props.data.name,
            ip: props.data.ip,
            podCIDR: props.data.podCIDR,
            publicKey: props.data.publicKey,
            serviceCIDR: props.data.serviceCIDR,
            version: props.data.version,
            isNameError: false,
          }
        : {
            name: '',
            displayName: '',
            ip: '',
            podCIDR: '',
            publicKey: '',
            serviceCIDR: '',
            version: '',
            isNameError: false,
          },
      validationSchema: Yup.object({
        name: Yup.string().required('id is required'),
        displayName: Yup.string().required('name is required'),
      }),
      onSubmit: async (val) => {
        try {
          if (!isUpdate) {
            console.log('ttt', deployment.name, project.name);
            const { errors: e } = await api.createIotDevice({
              projectName: project.name,
              deploymentName: deployment.name,
              device: {
                name: val.name,
                displayName: val.displayName,
                ip: val.ip,
                podCIDR: val.podCIDR,
                publicKey: val.publicKey,
                serviceCIDR: val.serviceCIDR,
                version: val.version,
              },
            });
            if (e) {
              throw e[0];
            }
          } else if (isUpdate) {
            const { errors: e } = await api.updateIotDevice({
              projectName: project.name,
              deploymentName: deployment.name,
              device: {
                name: val.name,
                displayName: val.displayName,
                ip: val.ip,
                podCIDR: val.podCIDR,
                publicKey: val.publicKey,
                serviceCIDR: val.serviceCIDR,
                version: val.version,
              },
            });
            if (e) {
              throw e[0];
            }
          }
          reloadPage();
          resetValues();
          toast.success(
            `device ${isUpdate ? 'updated' : 'created'} successfully`
          );
          setVisible(false);
        } catch (err) {
          handleError(err);
        }
      },
    });

  return (
    <Popup.Form
      onSubmit={(e) => {
        if (!values.isNameError) {
          handleSubmit(e);
        } else {
          e.preventDefault();
        }
      }}
    >
      <Popup.Content>
        <div className="flex flex-col gap-2xl">
          <NameIdView
            resType="project"
            displayName={values.displayName}
            name={values.name}
            label="Device name"
            placeholder="Enter device name"
            errors={errors.name}
            handleChange={handleChange}
            nameErrorLabel="isNameError"
            isUpdate={isUpdate}
          />
          <TextInput
            label="IP"
            size="lg"
            placeholder="ip"
            value={values.ip}
            onChange={handleChange('ip')}
          />
          <TextInput
            label="Public Key"
            size="lg"
            placeholder="public key"
            value={values.publicKey}
            onChange={handleChange('publicKey')}
          />
          <TextInput
            label="Pod CIDR"
            size="lg"
            placeholder="pod cidr"
            value={values.podCIDR}
            onChange={handleChange('podCIDR')}
          />

          <TextInput
            label="Service CIDR"
            size="lg"
            placeholder="service cidr"
            value={values.serviceCIDR}
            onChange={handleChange('serviceCIDR')}
          />
          <TextInput
            label="Version"
            size="lg"
            placeholder="version"
            value={values.version}
            onChange={handleChange('version')}
          />
        </div>
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button closable content="Cancel" variant="basic" />
        <Popup.Button
          loading={isLoading}
          type="submit"
          content={isUpdate ? 'Update' : 'Create'}
          variant="primary"
        />
      </Popup.Footer>
    </Popup.Form>
  );
};

const HandleDevice = (props: IDialog) => {
  const { isUpdate, setVisible, visible } = props;

  return (
    <Popup.Root show={visible} onOpenChange={(v) => setVisible(v)}>
      <Popup.Header>{isUpdate ? 'Edit device' : 'Add device'}</Popup.Header>
      {(!isUpdate || (isUpdate && props.data)) && <Root {...props} />}
    </Popup.Root>
  );
};

export default HandleDevice;
