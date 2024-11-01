/* eslint-disable react/destructuring-assignment */
import { toast } from 'react-toastify';
import Popup from '@kloudlite/design-system/molecule/popup';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { IDialogBase } from '~/console/components/types.d';
import { ExtractNodeType, parseName } from '~/console/server/r-utils/common';
import { NameIdView } from '~/console/components/name-id-view';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IGlobalVpnDevices } from '~/console/server/gql/queries/global-vpn-queries';

type IDialog = IDialogBase<ExtractNodeType<IGlobalVpnDevices>>;

const Root = (props: IDialog) => {
  const { setVisible, isUpdate } = props;

  const api = useConsoleApi();
  const reloadPage = useReload();

  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: isUpdate
        ? {
            displayName: props.data.displayName,
            name: parseName(props.data),
            isNameError: false,
          }
        : {
            displayName: '',
            name: '',
            isNameError: false,
          },
      validationSchema: Yup.object({
        name: Yup.string().required('id is required'),
        displayName: Yup.string().required('name is required'),
      }),
      onSubmit: async (val) => {
        try {
          if (!isUpdate) {
            const { errors: e } = await api.createGlobalVpnDevice({
              gvpnDevice: {
                displayName: val.name,
                globalVPNName: 'default',
                metadata: {
                  name: val.name,
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
            `vpn device ${isUpdate ? 'updated' : 'created'} successfully`
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
            resType="cluster"
            displayName={values.displayName}
            name={values.name}
            label="Device name"
            placeholder="Enter device name"
            errors={errors.name}
            handleChange={handleChange}
            nameErrorLabel="isNameError"
            isUpdate={isUpdate}
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

const HandleGlobalVpnDevice = (props: IDialog) => {
  const { isUpdate, setVisible, visible } = props;

  return (
    <Popup.Root show={visible} onOpenChange={(v) => setVisible(v)}>
      <Popup.Header>{isUpdate ? 'Edit Device' : 'Add Device'}</Popup.Header>
      {(!isUpdate || (isUpdate && props.data)) && <Root {...props} />}
    </Popup.Root>
  );
};

export default HandleGlobalVpnDevice;
