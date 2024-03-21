import { useParams } from '@remix-run/react';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { NameIdView } from '~/console/components/name-id-view';
import { IDialog } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { useReload } from '~/lib/client/helpers/reloader';
import useForm from '~/lib/client/hooks/use-form';
import Yup from '~/lib/server/helpers/yup';
import { handleError } from '~/lib/utils/common';

const HandleConfig = ({ show, setShow }: IDialog) => {
  const api = useConsoleApi();
  const reloadPage = useReload();
  const { project, environment } = useParams();
  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: {
        displayName: '',
        name: '',
        isNameError: false,
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
          const { errors: e } = await api.createConfig({
            envName: environment,
            projectName: project,
            config: {
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
          toast.success('Config created successfully');
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
        {show?.type === 'add' ? 'Add new config' : 'Edit config'}
      </Popup.Header>
      <form
        onSubmit={(e) => {
          if (!values.isNameError) {
            handleSubmit(e);
          } else {
            e.preventDefault();
          }
        }}
      >
        <Popup.Content>
          <NameIdView
            label="Name"
            placeholder="Enter config name"
            displayName={values.displayName}
            name={values.name}
            resType="config"
            errors={errors.name}
            handleChange={handleChange}
            nameErrorLabel="isNameError"
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

export default HandleConfig;
