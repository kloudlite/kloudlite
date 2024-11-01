import { NumberInput } from '@kloudlite/design-system/atoms/input';
import Radio from '@kloudlite/design-system/atoms/radio';
import SelectPrimitive from '@kloudlite/design-system/atoms/select-primitive';
import Popup from '@kloudlite/design-system/molecule/popup';
import { toast } from '@kloudlite/design-system/molecule/toast';
import CommonPopupHandle from '~/iotconsole/components/common-popup-handle';
import { NameIdView } from '~/iotconsole/components/name-id-view';
import { IDialogBase } from '~/iotconsole/components/types.d';
import { useIotConsoleApi } from '~/iotconsole/server/gql/api-provider';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { CredentialIn } from '~/root/src/generated/gql/server';
// import { IdSelector } from '~/iotconsole/components/id-selector';

type IDialog = IDialogBase<null>;
const expirationUnits = [
  { label: 'hours', value: 'h' },
  { label: 'days', value: 'd' },
  { label: 'weeks', value: 'w' },
  { label: 'months', value: 'm' },
  { label: 'years', value: 'y' },
];
const Root = (props: IDialog) => {
  const { setVisible } = props;
  const api = useIotConsoleApi();
  const reloadPage = useReload();

  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: {
        displayName: '',
        name: '',
        access: 'read_write' as CredentialIn['access'],
        unit: 'd' as CredentialIn['expiration']['unit'],
        value: '1',
        isNameError: false,
      },
      validationSchema: Yup.object({
        displayName: Yup.string().required(),
        name: Yup.string().required(),
        value: Yup.string().required('expiration time is required.'),
      }),
      onSubmit: async (val) => {
        try {
          const { errors: e } = await api.createCred({
            credential: {
              name: val.displayName,
              username: val.name,
              expiration: {
                unit: val.unit,
                value: parseInt(val.value, 10),
              },
              access: val.access,
            },
          });
          if (e) {
            throw e[0];
          }
          reloadPage();
          resetValues();
          toast.success('Credential created successfully');
          setVisible(false);
        } catch (err) {
          handleError(err);
        }
      },
    });
  return (
    <Popup.Form onSubmit={handleSubmit}>
      <Popup.Content>
        <div className="flex flex-col gap-3xl">
          <NameIdView
            resType="username"
            label="Name"
            displayName={values.displayName}
            name={values.name}
            errors={errors.name}
            handleChange={handleChange}
            nameErrorLabel="isNameError"
          />
          <Radio.Root
            label="Access"
            value={values.access}
            onChange={(value) => {
              handleChange('access')(dummyEvent(value));
            }}
          >
            <Radio.Item value="read_write">Read and Write</Radio.Item>
            <Radio.Item value="read">Read Only</Radio.Item>
          </Radio.Root>
          <div className="flex flex-row gap-3xl items-start">
            <div className="flex-1">
              <NumberInput
                label="Expiration time"
                value={values.value}
                onChange={handleChange('value')}
                error={!!errors.value}
                message={errors.value}
              />
            </div>
            <div className="flex-2">
              <SelectPrimitive.Root
                label="Unit"
                value={values.unit}
                onChange={handleChange('unit')}
              >
                {expirationUnits.map((eu) => (
                  <SelectPrimitive.Option key={eu.value} value={eu.value}>
                    {eu.label}
                  </SelectPrimitive.Option>
                ))}
              </SelectPrimitive.Root>
            </div>
          </div>
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
  );
};

const HandleCrCred = (props: IDialog) => {
  return (
    <CommonPopupHandle
      {...props}
      createTitle="Create credential"
      updateTitle="Update credential"
      root={Root}
    />
  );
};

export default HandleCrCred;
