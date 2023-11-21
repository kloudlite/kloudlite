import { TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { IDialog } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';

const HandleRepo = ({ show, setShow }: IDialog) => {
  const api = useConsoleApi();
  const reloadPage = useReload();

  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: {
        name: '',
      },
      validationSchema: Yup.object({
        name: Yup.string().required(),
      }),
      onSubmit: async (val) => {
        try {
          const { errors: e } = await api.createRepo({
            repository: {
              name: val.name,
            },
          });
          if (e) {
            throw e[0];
          }
          resetValues();
          toast.success('Repository created successfully');
          setShow(null);
          reloadPage();
        } catch (err) {
          handleError(err);
        }
      },
    });
  return (
    <Popup.Root
      show={show as any}
      onOpenChange={(e) => {
        if (!e) {
          resetValues();
        }

        setShow(e);
      }}
    >
      <Popup.Header>Create repository</Popup.Header>
      <form onSubmit={handleSubmit}>
        <Popup.Content>
          <div className="flex flex-col gap-3xl">
            <TextInput
              value={values.name}
              label="Name"
              onChange={handleChange('name')}
              error={!!errors.name}
              message={errors.name}
            />
          </div>
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button closable content="Cancel" variant="basic" />
          <Popup.Button
            type="submit"
            content="Create"
            variant="primary"
            loading={isLoading}
          />
        </Popup.Footer>
      </form>
    </Popup.Root>
  );
};

export default HandleRepo;
