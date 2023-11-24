/* eslint-disable react/destructuring-assignment */
import { useParams } from '@remix-run/react';
import { TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import CommonPopupHandle from '~/console/components/common-popup-handle';
import { IDialogBase } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IDomains } from '~/console/server/gql/queries/domain-queries';
import { ExtractNodeType } from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';

type IDialog = IDialogBase<ExtractNodeType<IDomains>>;

const Root = (props: IDialog) => {
  const { isUpdate, setVisible } = props;
  const api = useConsoleApi();
  const reloadPage = useReload();
  const { cluster } = useParams();

  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: isUpdate
        ? {
            displayName: props.data.displayName,
            domainName: props.data.domainName,
          }
        : {
            displayName: '',
            domainName: '',
          },
      validationSchema: Yup.object({
        domainName: Yup.string().required('Domain name is required.'),
        displayName: Yup.string().required('Name is required.'),
      }),
      onSubmit: async (val) => {
        try {
          if (!isUpdate) {
            const { errors } = await api.createDomain({
              domainEntry: { ...val, clusterName: cluster! },
            });
            if (errors) {
              throw errors[0];
            }
          } else if (isUpdate && props.data) {
            const { errors } = await api.updateDomain({
              domainEntry: {
                clusterName: cluster!,
                displayName: val.displayName,
                domainName: props.data.domainName,
              },
            });
            if (errors) {
              throw errors[0];
            }
          }

          setVisible(false);
          reloadPage();
          toast.success(
            `Domain ${isUpdate ? 'updated' : 'created'} successfully`
          );
          // resetValues();
        } catch (err) {
          handleError(err);
        }
      },
    });

  return (
    <Popup.Form onSubmit={handleSubmit}>
      <Popup.Content>
        <div className="flex flex-col gap-2xl">
          <TextInput
            label="Name"
            value={values.displayName}
            onChange={handleChange('displayName')}
            error={!!errors.displayName}
            message={errors.displayName}
          />
          {!isUpdate && (
            <TextInput
              label="Domain name"
              value={values.domainName}
              onChange={handleChange('domainName')}
              error={!!errors.domainName}
              message={errors.domainName}
            />
          )}
        </div>
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button closable content="Cancel" variant="basic" />
        <Popup.Button
          type="submit"
          content={!isUpdate ? 'Create' : 'Update'}
          variant="primary"
          loading={isLoading}
        />
      </Popup.Footer>
    </Popup.Form>
  );
};

const HandleDomain = (props: IDialog) => {
  return (
    <CommonPopupHandle
      {...props}
      root={Root}
      createTitle="Add domain"
      updateTitle="Edit domain"
    />
  );
};
export default HandleDomain;
