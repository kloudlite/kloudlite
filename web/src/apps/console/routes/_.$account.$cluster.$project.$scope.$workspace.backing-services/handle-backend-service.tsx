import { useEffect, useState } from 'react';
import * as Chips from '~/components/atoms/chips';
import Select from '~/components/atoms/select';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { IDialog } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { parseName } from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';

const HandleBackendService = ({ show, setShow }: IDialog) => {
  const api = useConsoleApi();
  const reloadPage = useReload();

  const [validationSchema, setValidationSchema] = useState(
    Yup.object({
      nodePlan: Yup.string().required(),
    })
  );

  const {
    values,
    errors,
    handleSubmit,
    handleChange,
    isLoading,
    resetValues,
    setValues,
  } = useForm({
    initialValues: {
      displayName: '',
      name: '',
      provider: 'aws',
      accessKey: '',
      accessSecret: '',
    },
    validationSchema,

    onSubmit: async (val) => {
      try {
        if (show?.type === 'add') {
          toast.success('Backend service secret created successfully');
        } else {
          toast.success('Backend service secret updated successfully');
        }
        reloadPage();
        setShow(null);
        resetValues();
      } catch (err) {
        handleError(err);
      }
    },
  });

  useEffect(() => {
    // if (show && show.data && show.type === 'edit') {
    //   setValues((v) => ({
    //     ...v,
    //     accessSecret: show.data?.stringData?.accessSecret || '',
    //     accessKey: show.data?.stringData?.accessKey || '',
    //   }));
    //   setValidationSchema(
    //     // @ts-ignore
    //     Yup.object({
    //       displayName: Yup.string().trim().required(),
    //       accessSecret: Yup.string().trim().required(),
    //       accessKey: Yup.string().trim().required(),
    //       provider: Yup.string().required(),
    //     })
    //   );
    // }
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
        {show?.type === 'add'
          ? 'Add backing services'
          : 'Edit backing services'}
      </Popup.Header>
      <form onSubmit={handleSubmit}>
        <Popup.Content>
          <div className="flex flex-col gap-2xl">
            {show?.type === 'edit' && (
              <Chips.Chip
                {...{
                  item: { id: parseName(show.data) },
                  label: parseName(show.data),
                  prefix: 'Id:',
                  disabled: true,
                  type: 'BASIC',
                }}
              />
            )}

            {show?.type === 'add' && (
              <Select
                label="Node plan"
                options={[{ label: 'aws', value: 'aws' }]}
                value={undefined}
                placeholder="---Select---"
              />
            )}
          </div>
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button content="Cancel" variant="basic" closable />
          <Popup.Button
            loading={isLoading}
            type="submit"
            content={show?.type === 'add' ? 'Add' : 'Update'}
            variant="primary"
          />
        </Popup.Footer>
      </form>
    </Popup.Root>
  );
};

export default HandleBackendService;
