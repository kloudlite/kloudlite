/* eslint-disable react/destructuring-assignment */
import { toast } from 'react-toastify';
import { TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { IDialogBase } from '~/console/components/types.d';
import {
  ExtractNodeType,
  parseName,
  validateExternalAppRecordType,
} from '~/console/server/r-utils/common';
import { NameIdView } from '~/console/components/name-id-view';
import Select from '~/components/atoms/select';
import { IExternalApps } from '~/console/server/gql/queries/external-app-queries';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { useOutletContext } from '@remix-run/react';
import { IEnvironmentContext } from '../_layout';

type IDialog = IDialogBase<ExtractNodeType<IExternalApps>>;

const Root = (props: IDialog) => {
  const { setVisible, isUpdate } = props;

  const api = useConsoleApi();
  const reloadPage = useReload();
  const { environment } = useOutletContext<IEnvironmentContext>();

  const recordTypes = [
    {
      label: 'CNAME',
      value: 'CNAME',
      render: () => (
        <div className="flex flex-row gap-lg items-center">
          <div>C Name</div>
        </div>
      ),
    },
    {
      label: 'IP Addr',
      value: 'IPAddr',
      render: () => (
        <div className="flex flex-row gap-lg items-center">
          <div>IP Addr</div>
        </div>
      ),
    },
  ];

  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: isUpdate
        ? {
            displayName: props.data.displayName,
            name: parseName(props.data),
            recordType: props.data.spec?.recordType,
            record: props.data.spec?.record,
            isNameError: false,
          }
        : {
            name: '',
            displayName: '',
            recordType: recordTypes[0].value,
            record: '',
            isNameError: false,
          },
      validationSchema: Yup.object({
        name: Yup.string().required('id is required'),
        displayName: Yup.string().required('name is required'),
        record: Yup.string()
          .required()
          .when(['recordType'], ([recordType], schema) => {
            if (recordType === 'IPAddr') {
              return schema.matches(
                /^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/,
                'Please enter valid ip address'
              );
            }
            return schema;
          }),
      }),
      onSubmit: async (val) => {
        try {
          if (!isUpdate) {
            const { errors: e } = await api.createExternalApp({
              envName: parseName(environment),
              externalApp: {
                displayName: val.displayName,
                metadata: {
                  name: val.name,
                },
                spec: {
                  recordType: validateExternalAppRecordType(
                    val.recordType || ''
                  ),
                  record: val.record || '',
                },
              },
            });
            if (e) {
              throw e[0];
            }
          }
          reloadPage();
          resetValues();
          toast.success(
            `external app ${isUpdate ? 'updated' : 'created'} successfully`
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
            resType="app"
            displayName={values.displayName}
            name={values.name}
            label="External app name"
            placeholder="Enter external app name"
            errors={errors.name}
            handleChange={handleChange}
            nameErrorLabel="isNameError"
            isUpdate={isUpdate}
          />

          <Select
            error={!!errors.recordType}
            message={errors.recordType}
            value={values.recordType}
            label="Record Type"
            onChange={(_, value) => {
              handleChange('recordType')(dummyEvent(value));
            }}
            options={async () => recordTypes}
            disabled={isUpdate}
          />

          <TextInput
            label="Record"
            size="lg"
            placeholder="record"
            value={values.record}
            onChange={handleChange('record')}
            error={!!errors.record}
            message={errors.record}
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

const HandleExternalApp = (props: IDialog) => {
  const { isUpdate, setVisible, visible } = props;

  return (
    <Popup.Root show={visible} onOpenChange={(v) => setVisible(v)}>
      <Popup.Header>
        {isUpdate ? 'Edit External App' : 'Add External App'}
      </Popup.Header>
      {(!isUpdate || (isUpdate && props.data)) && <Root {...props} />}
    </Popup.Root>
  );
};

export default HandleExternalApp;
