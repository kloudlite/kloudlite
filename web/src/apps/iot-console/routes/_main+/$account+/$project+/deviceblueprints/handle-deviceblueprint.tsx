/* eslint-disable react/destructuring-assignment */
import { toast } from 'react-toastify';
import { TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';
import Popup from '~/components/molecule/popup';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { Github__Com___Kloudlite___Api___Apps___Iot____Console___Internal___Entities__BluePrintType as deviceBluePrintType } from '~/root/src/generated/gql/server';
import { IDialogBase } from '~/iotconsole/components/types.d';
import { ExtractNodeType } from '~/iotconsole/server/r-utils/common';
import { NameIdView } from '~/iotconsole/components/name-id-view';
import { useIotConsoleApi } from '~/iotconsole/server/gql/api-provider';
import { IDeviceBlueprints } from '~/iotconsole/server/gql/queries/iot-device-blueprint-queries';
import { useOutletContext } from '@remix-run/react';
import { IProjectContext } from '../_layout';
import { deviceBlueprintTypes } from './blueprint-utils';

type IDialog = IDialogBase<ExtractNodeType<IDeviceBlueprints>>;

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
            deviceBlueprintType:
              props.data.bluePrintType || 'singleton_blueprint',
            version: props.data.version,
            isNameError: false,
          }
        : {
            name: '',
            displayName: '',
            deviceBlueprintType: 'singleton_blueprint',
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
            const { errors: e } = await api.createIotDeviceBlueprint({
              projectName: project.name,
              deviceBlueprint: {
                displayName: val.displayName,
                name: val.name,
                bluePrintType: val.deviceBlueprintType as deviceBluePrintType,
                version: val.version,
              },
            });
            if (e) {
              throw e[0];
            }
          } else if (isUpdate) {
            const { errors: e } = await api.updateIotDeviceBlueprint({
              projectName: project.name,
              deviceBlueprint: {
                displayName: val.displayName,
                name: val.name,
                bluePrintType: val.deviceBlueprintType as deviceBluePrintType,
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
            `device blueprint ${isUpdate ? 'updated' : 'created'} successfully`
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
            label="Blueprint name"
            placeholder="Enter blueprint name"
            errors={errors.name}
            handleChange={handleChange}
            nameErrorLabel="isNameError"
            isUpdate={isUpdate}
          />

          <Select
            label="Blueprint type"
            value={values.deviceBlueprintType}
            options={async () => deviceBlueprintTypes}
            onChange={(_, value) => {
              handleChange('deviceBlueprintType')(dummyEvent(value));
            }}
          />

          <TextInput
            label="Version"
            placeholder="Version"
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

const HandleDeviceBlueprint = (props: IDialog) => {
  const { isUpdate, setVisible, visible } = props;

  return (
    <Popup.Root show={visible} onOpenChange={(v) => setVisible(v)}>
      <Popup.Header>
        {isUpdate ? 'Edit device blueprint' : 'Add device blueprint'}
      </Popup.Header>
      {(!isUpdate || (isUpdate && props.data)) && <Root {...props} />}
    </Popup.Root>
  );
};

export default HandleDeviceBlueprint;
