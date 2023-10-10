import { useParams } from '@remix-run/react';
import { useEffect } from 'react';
import { TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { IDialog } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IDomains } from '~/console/server/gql/queries/domain-queries';
import { ExtractNodeType } from '~/console/server/r-utils/common';
import { DIALOG_TYPE } from '~/console/utils/commons';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';

const HandleDomain = ({
  show,
  setShow,
}: IDialog<ExtractNodeType<IDomains> | null>) => {
  const api = useConsoleApi();
  const reloadPage = useReload();
  const { cluster } = useParams();

  const {
    values,
    errors,
    handleChange,
    handleSubmit,
    resetValues,
    setValues,
    isLoading,
  } = useForm({
    initialValues: {
      displayName: '',
      domainName: '',
      clusterName: cluster!,
    },
    validationSchema: Yup.object({
      domainName: Yup.string().required('Domain name is required.'),
      displayName: Yup.string().required('Name is required.'),
      clusterName: Yup.string().required(),
    }),
    onSubmit: async (val) => {
      try {
        if (show?.type === DIALOG_TYPE.ADD) {
          const { errors } = await api.createDomain({
            domainEntry: val,
          });
          if (errors) {
            throw errors[0];
          }
        } else if (show?.data) {
          const { errors } = await api.updateDomain({
            domainEntry: {
              clusterName: cluster!,
              displayName: val.displayName,
              domainName: val.domainName,
            },
          });
          if (errors) {
            throw errors[0];
          }
        }

        reloadPage();
        resetValues();
        toast.success('Credential created successfully');
        setShow(null);
      } catch (err) {
        handleError(err);
      }
    },
  });

  useEffect(() => {
    if (show && show.data && show.type === DIALOG_TYPE.EDIT) {
      setValues((v) => ({
        ...v,
        displayName: show.data?.displayName || '',
        domainName: show.data?.domainName || '',
      }));
    }
  }, [show]);
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
      <Popup.Header>
        {show?.type === DIALOG_TYPE.ADD ? 'Add new domain' : 'Edit domain'}
      </Popup.Header>
      <form onSubmit={handleSubmit}>
        <Popup.Content>
          <div className="flex flex-col gap-2xl">
            <TextInput
              label="Name"
              value={values.displayName}
              onChange={handleChange('displayName')}
              error={!!errors.displayName}
              message={errors.displayName}
            />
            <TextInput
              label="Domain name"
              value={values.domainName}
              onChange={handleChange('domainName')}
              error={!!errors.domainName}
              message={errors.domainName}
            />
          </div>
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button closable content="Cancel" variant="basic" />
          <Popup.Button
            type="submit"
            content={show?.type === DIALOG_TYPE.ADD ? 'Create' : 'Update'}
            variant="primary"
            loading={isLoading}
          />
        </Popup.Footer>
      </form>
    </Popup.Root>
  );
};

export default HandleDomain;
