/* eslint-disable react/destructuring-assignment */
import { NumberInput, TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import CommonPopupHandle from '~/console/components/common-popup-handle';
import { IdSelector } from '~/console/components/id-selector';
import { IDialogBase } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IBuildCaches } from '~/console/server/gql/queries/build-caches-queries';
import { ExtractNodeType } from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';

type IDialog = IDialogBase<ExtractNodeType<IBuildCaches>>;
const Root = (props: IDialog) => {
  const { isUpdate, setVisible } = props;
  const api = useConsoleApi();
  const reloadPage = useReload();

  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: isUpdate
        ? {
            name: props.data.name,
            displayName: props.data.displayName,
            volumeSize: props.data.volumeSizeInGB,
          }
        : {
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
          if (!isUpdate) {
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
          } else {
            const { errors: e } = await api.updateBuildCaches({
              crUpdateBuildCacheKeyId: props.data.id,
              buildCacheKey: {
                displayName: val.displayName,
                name: props.data.name,
                volumeSizeInGB: val.volumeSize,
              },
            });
            if (e) {
              throw e[0];
            }
          }
          // resetValues();
          toast.success(
            `Build cache ${isUpdate ? 'updated' : 'created'} successfully`
          );
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
        <div className="flex flex-col gap-2xl">
          <TextInput
            value={values.displayName}
            label="Name"
            onChange={handleChange('displayName')}
            error={!!errors.displayName}
            message={errors.displayName}
          />
          {!isUpdate && (
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
          content={isUpdate ? 'Update' : 'Create'}
          variant="primary"
          loading={isLoading}
        />
      </Popup.Footer>
    </Popup.Form>
  );
};

const HandleBuildCache = (props: IDialog) => {
  return (
    <CommonPopupHandle
      {...props}
      createTitle="Create build cache"
      updateTitle="Update build cache"
      root={Root}
    />
  );
};

export default HandleBuildCache;
