import { ArrowLeft, ArrowRight } from '@jengaicons/react';
import { useNavigate, useParams } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import { PasswordInput, TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select-primitive';
import { toast } from '~/components/molecule/toast';
import { useMapper } from '~/components/utils';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { IdSelector } from '../components/id-selector';
import RawWrapper, { TitleBox } from '../components/raw-wrapper';
import { useConsoleApi } from '../server/gql/api-provider';
import { validateCloudProvider } from '../server/r-utils/common';
import { ensureAccountClientSide } from '../server/utils/auth-utils';
import { FadeIn } from './_.$account.$cluster.$project.$scope.$workspace.new-app/util';

const NewCloudProvider = () => {
  const { a: accountName } = useParams();
  const api = useConsoleApi();

  const navigate = useNavigate();
  const { values, errors, handleSubmit, handleChange, isLoading } = useForm({
    initialValues: {
      displayName: '',
      name: '',
      provider: 'aws',
      accessKey: '',
      accessSecret: '',
    },
    validationSchema: Yup.object({
      displayName: Yup.string().required(),
      name: Yup.string().required(),
      provider: Yup.string().required(),
      accessKey: Yup.string().required(),
      accessSecret: Yup.string().required(),
    }),
    onSubmit: async (val) => {
      try {
        console.log(val);
        ensureAccountClientSide({ account: accountName });
        const { errors: e } = await api.createProviderSecret({
          secret: {
            displayName: val.displayName,
            metadata: {
              name: val.name,
            },
            stringData: {
              accessKey: val.accessKey,
              accessSecret: val.accessSecret,
            },
            cloudProviderName: validateCloudProvider(val.provider),
          },
        });
        if (e) {
          throw e[0];
        }
        toast.success('provider secret created successfully');
        navigate(`/onboarding/${accountName}/${val.name}/new-cluster`);
      } catch (err) {
        handleError(err);
      }
    },
  });

  const progressItems = [
    { label: 'Create Team', active: true, id: 1, completed: false },
    {
      label: 'Invite your Team Members',
      active: true,
      id: 2,
      completed: false,
    },
    {
      label: 'Add your Cloud Provider',
      active: true,
      id: 3,
      completed: false,
    },
    {
      label: 'Setup First Cluster',
      active: false,
      id: 4,
      completed: false,
    },
    {
      label: 'Create your project',
      active: false,
      id: 5,
      completed: false,
    },
  ];

  const pItems = useMapper(progressItems, (i) => {
    return {
      value: i.id,
      item: {
        ...i,
      },
    };
  });

  return (
    <RawWrapper
      onProgressClick={() => {}}
      title="Integrate Cloud Provider"
      subtitle="Kloudlite will help you to develop and deploy cloud native
    applications easily."
      progressItems={pItems}
      rightChildren={
        <FadeIn onSubmit={handleSubmit}>
          <TitleBox
            title="Cloud provider details"
            subtitle="A cloud provider offers remote computing resources and services over the internet."
          />
          <div className="flex flex-col gap-3xl">
            <div className="flex flex-col">
              <TextInput
                label="Name"
                size="lg"
                value={values.displayName}
                onChange={handleChange('displayName')}
                error={!!errors.displayName}
                message={errors.displayName}
              />
              <IdSelector
                resType="providersecret"
                name={values.displayName}
                onChange={(v) => handleChange('name')(dummyEvent(v))}
                className="pt-2xl"
              />
            </div>

            <div className="flex flex-col gap-3xl">
              <Select.Root
                label="Provider"
                error={!!errors.provider}
                message={errors.provider}
                value={values.provider}
                onChange={(provider: string) => {
                  handleChange('provider')(dummyEvent(provider));
                }}
              >
                <Select.Option value="aws">Amazon Web Services</Select.Option>
              </Select.Root>

              <PasswordInput
                name="accessKey"
                label="Access Key ID"
                size="lg"
                onChange={handleChange('accessKey')}
                error={!!errors.accessKey}
                message={errors.accessKey}
                value={values.accessKey}
              />
              <PasswordInput
                name="accessSecret"
                label="Access Key Secret"
                size="lg"
                onChange={handleChange('accessSecret')}
                error={!!errors.accessSecret}
                message={errors.accessSecret}
                value={values.accessSecret}
              />
            </div>
          </div>
          <div className="flex flex-row gap-xl justify-end">
            <Button
              variant="outline"
              content="Back"
              prefix={<ArrowLeft />}
              size="lg"
            />
            <Button
              loading={isLoading}
              type="submit"
              variant="primary"
              content="Continue"
              suffix={<ArrowRight />}
              size="lg"
            />
          </div>
        </FadeIn>
      }
    />
  );
};

export default NewCloudProvider;
