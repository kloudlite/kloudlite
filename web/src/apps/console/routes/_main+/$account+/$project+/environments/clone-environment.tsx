/* eslint-disable react/destructuring-assignment */
import { useParams } from '@remix-run/react';
import { Switch } from '~/components/atoms/switch';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import CommonPopupHandle from '~/console/components/common-popup-handle';
import { NameIdView } from '~/console/components/name-id-view';
import { IDialogBase } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IEnvironments } from '~/console/server/gql/queries/environment-queries';
import { ExtractNodeType, parseName } from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';

type IDialog = IDialogBase<ExtractNodeType<IEnvironments>>;

const Root = (props: IDialog) => {
  const { isUpdate, setVisible } = props;
  const api = useConsoleApi();
  const reloadPage = useReload();

  const { project } = useParams();

  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: {
        name: '',
        displayName: '',
        environmentRoutingMode: false,
      },
      validationSchema: Yup.object({
        name: Yup.string().required('Name is required.'),
        displayName: Yup.string().required(),
      }),
      onSubmit: async (val) => {
        if (isUpdate) {
          if (!project) {
            throw new Error('Project is required!.');
          }
          try {
            const { errors: e } = await api.cloneEnvironment({
              displayName: val.displayName,
              environmentRoutingMode: val.environmentRoutingMode
                ? 'public'
                : 'private',
              destinationEnvName: val.name,
              projectName: project,
              sourceEnvName: parseName(props.data),
            });
            if (e) {
              throw e[0];
            }
            resetValues();
            toast.success('Environment cloned successfully');
            setVisible(false);
            reloadPage();
          } catch (err) {
            handleError(err);
          }
        }
      },
    });
  return (
    <Popup.Form onSubmit={handleSubmit}>
      <Popup.Content>
        <div className="flex flex-col gap-3xl">
          <NameIdView
            displayName={values.displayName}
            name={values.name}
            resType="environment"
            errors={errors.name}
            label="Name"
            placeholder="Environment name"
            onChange={({ name, id }) => {
              handleChange('displayName')(dummyEvent(name));
              handleChange('name')(dummyEvent(id));
            }}
          />
          <Switch
            label="Public"
            checked={values.environmentRoutingMode}
            onChange={(val) => {
              handleChange('environmentRoutingMode')(dummyEvent(val));
            }}
          />
        </div>
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button closable content="Cancel" variant="basic" />
        <Popup.Button
          type="submit"
          content="Clone"
          variant="primary"
          loading={isLoading}
        />
      </Popup.Footer>
    </Popup.Form>
  );
};

const CloneEnvironment = (props: IDialog) => {
  return (
    <CommonPopupHandle
      {...props}
      root={Root}
      updateTitle="Clone Environment"
      createTitle="Clone Environment"
    />
  );
};

export default CloneEnvironment;
