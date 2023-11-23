import { TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import CommonPopupHandle from '~/console/components/common-popup-handle';
import { IDialogBase } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IRepos } from '~/console/server/gql/queries/repo-queries';
import { ExtractNodeType } from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';

type IDialog = IDialogBase<ExtractNodeType<IRepos>>;

const Root = (props: IDialog) => {
  const { isUpdate, setVisible } = props;
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
          setVisible(false);
          reloadPage();
        } catch (err) {
          handleError(err);
        }
      },
    });
  return (
    <Popup.Form onSubmit={handleSubmit}>
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
          content={isUpdate ? 'Update' : 'Create'}
          variant="primary"
          loading={isLoading}
        />
      </Popup.Footer>
    </Popup.Form>
  );
};

const HandleRepo = (props: IDialog) => {
  return (
    <CommonPopupHandle
      {...props}
      root={Root}
      updateTitle="Edit repository"
      createTitle="Add repository"
    />
  );
};

export default HandleRepo;
