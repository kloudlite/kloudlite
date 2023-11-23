import { NumberInput, TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { IdSelector } from '~/console/components/id-selector';
import { IDialog } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IBuildCaches } from '~/console/server/gql/queries/build-caches-queries';
import { ExtractNodeType } from '~/console/server/r-utils/common';
import { DIALOG_TYPE } from '~/console/utils/commons';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';

const HandleBuildCache = ({
  show,
  setShow,
}: IDialog<ExtractNodeType<IBuildCaches> | null, null>) => {
  const api = useConsoleApi();
  const reloadPage = useReload();

  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: {
        name: '',
        displayName: '',
        volumeSize: 0,
      },
      validationSchema: Yup.object({
        name: Yup.string().required(),
        displayName: Yup.string().required(),
        volumeSize: Yup.number().required(),
      }),
      onSubmit: async (val) => {
        try {
          const { errors: e } = await api.createBuildCache({
            buildCacheKey: {
              displayName: val.displayName,
              name: val.name,
              volumeSizeInGB: val.volumeSize,
            },
          });
          if (e) {
            throw e[0];
          }
          resetValues();
          toast.success('Build cache created successfully');
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
      <Popup.Header>Create build cache</Popup.Header>
      <Popup.Form onSubmit={handleSubmit}>
        <Popup.Content>
          <div className="flex flex-col gap-2xl">
            <TextInput
              value={values.displayName}
              label="Name"
              onChange={handleChange('displayName')}
              error={!!errors.displayName}
              message={errors.displayName}
            />
            {show?.type === DIALOG_TYPE.ADD && (
              <IdSelector
                resType="username"
                onChange={(v) => {
                  handleChange('name')(dummyEvent(v));
                }}
                name={values.displayName}
              />
            )}
            <NumberInput
              value={values.volumeSize}
              min={0}
              label="Size"
              error={!!errors.volumeSize}
              message={errors.volumeSize}
              onChange={handleChange('volumeSize')}
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
      </Popup.Form>
    </Popup.Root>
  );
};

export default HandleBuildCache;
