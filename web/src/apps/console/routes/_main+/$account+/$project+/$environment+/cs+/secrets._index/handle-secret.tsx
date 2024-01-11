import { useOutletContext, useParams } from '@remix-run/react';
import { TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { IdSelector } from '~/console/components/id-selector';
import { IDialog } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { parseTargetNs } from '~/console/server/r-utils/common';
import { keyconstants } from '~/console/server/r-utils/key-constants';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { NameIdView } from '~/console/components/name-id-view';
import { IEnvironmentContext } from '../../_layout';

const HandleSecret = ({ show, setShow }: IDialog) => {
  const api = useConsoleApi();
  const reloadPage = useReload();
  const { project, environment } = useParams();

  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: {
        displayName: '',
        name: '',
        nodeType: '',
      },
      validationSchema: Yup.object({
        displayName: Yup.string().required(),
        name: Yup.string().required(),
      }),
      onSubmit: async (val) => {
        if (!project || !environment) {
          throw new Error('Project and Environment is required!.');
        }
        try {
          const { errors: e } = await api.createSecret({
            envName: environment,
            projectName: project,
            secret: {
              metadata: {
                name: val.name,
              },
              displayName: val.displayName,
              data: {},
            },
          });
          if (e) {
            throw e[0];
          }
          reloadPage();
          resetValues();
          toast.success('Secret created successfully');
          setShow(null);
        } catch (err) {
          handleError(err);
        }
      },
    });

  return (
    <Popup.Root
      show={!!show}
      onOpenChange={(e) => {
        if (!e) {
          resetValues();
        }

        setShow(e);
      }}
    >
      <Popup.Header>
        {show?.type === 'add' ? 'Add new secret' : 'Edit secret'}
      </Popup.Header>
      <form onSubmit={handleSubmit}>
        <Popup.Content>
          <NameIdView
            label="Name"
            placeholder="Enter secret name"
            displayName={values.displayName}
            name={values.name}
            resType="secret"
            errors={errors.name}
            onChange={({ name, id }) => {
              handleChange('displayName')(dummyEvent(name));
              handleChange('name')(dummyEvent(id));
            }}
          />
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button closable content="Cancel" variant="basic" />
          <Popup.Button
            loading={isLoading}
            type="submit"
            content={show?.type === 'add' ? 'Create' : 'Update'}
            variant="primary"
          />
        </Popup.Footer>
      </form>
    </Popup.Root>
  );
};
export default HandleSecret;
