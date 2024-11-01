/* eslint-disable react/destructuring-assignment */
import { toast } from 'react-toastify';
import { TextInput } from '@kloudlite/design-system/atoms/input';
import Popup from '@kloudlite/design-system/molecule/popup';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { IDialogBase } from '~/iotconsole/components/types.d';
import { ExtractNodeType } from '~/iotconsole/server/r-utils/common';
import { NameIdView } from '~/iotconsole/components/name-id-view';
import { useIotConsoleApi } from '~/iotconsole/server/gql/api-provider';
import { useOutletContext } from '@remix-run/react';
import { IDeployments } from '~/iotconsole/server/gql/queries/iot-deployment-queries';
import KeyValuePair from '~/iotconsole/components/key-value-pair';
import Select from '@kloudlite/design-system/atoms/select';
import { useEffect } from 'react';
import { IProjectContext } from '../_layout';

type IDialog = IDialogBase<ExtractNodeType<IDeployments>>;

const Root = (props: IDialog) => {
  const { setVisible, isUpdate } = props;

  const api = useIotConsoleApi();
  const reloadPage = useReload();
  const { project } = useOutletContext<IProjectContext>();

  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: isUpdate
        ? {
            displayName: props.data.displayName,
            name: props.data.name,
            cidr: props.data.CIDR,
            exposedServices: props.data.exposedServices,
            exposedIps: props.data.exposedIps,
            exposedDomains: props.data.exposedDomains,
            isNameError: false,
          }
        : {
            name: '',
            displayName: '',
            cidr: '',
            exposedServices: [],
            exposedIps: [],
            exposedDomains: [],
            isNameError: false,
          },
      validationSchema: Yup.object({
        name: Yup.string().required('id is required'),
        displayName: Yup.string().required('name is required'),
      }),
      onSubmit: async (val) => {
        try {
          if (!isUpdate) {
            const { errors: e } = await api.createIotDeployment({
              projectName: project.name,
              deployment: {
                name: val.name,
                displayName: val.displayName,
                CIDR: val.cidr,
                exposedIps: val.exposedIps,
                exposedDomains: val.exposedDomains,
                exposedServices: val.exposedServices.map((service) => {
                  return {
                    name: service.name,
                    ip: service.ip,
                  };
                }),
              },
            });
            if (e) {
              throw e[0];
            }
          } else if (isUpdate) {
            const { errors: e } = await api.updateIotDeployment({
              projectName: project.name,
              deployment: {
                name: val.name,
                displayName: val.displayName,
                CIDR: val.cidr,
                exposedIps: val.exposedIps,
                exposedDomains: val.exposedDomains,
                exposedServices: val.exposedServices.map((service) => {
                  return {
                    name: service.name,
                    ip: service.ip,
                  };
                }),
              },
            });
            if (e) {
              throw e[0];
            }
          }
          reloadPage();
          resetValues();
          toast.success(
            `deployment ${isUpdate ? 'updated' : 'created'} successfully`
          );
          setVisible(false);
        } catch (err) {
          handleError(err);
        }
      },
    });

  useEffect(() => {}, [values.exposedIps]);

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
            label="Deployment name"
            placeholder="Enter deployment name"
            errors={errors.name}
            handleChange={handleChange}
            nameErrorLabel="isNameError"
            isUpdate={isUpdate}
          />
          <TextInput
            label="CIDR"
            size="lg"
            placeholder="cidr"
            value={values.cidr}
            onChange={handleChange('cidr')}
          />

          <Select
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
          />

          <Select
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
          />

          <KeyValuePair
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

const HandleDeployment = (props: IDialog) => {
  const { isUpdate, setVisible, visible } = props;

  return (
    <Popup.Root show={visible} onOpenChange={(v) => setVisible(v)}>
      <Popup.Header>
        {isUpdate ? 'Edit deployment' : 'Add deployment'}
      </Popup.Header>
      {(!isUpdate || (isUpdate && props.data)) && <Root {...props} />}
    </Popup.Root>
  );
};

export default HandleDeployment;
