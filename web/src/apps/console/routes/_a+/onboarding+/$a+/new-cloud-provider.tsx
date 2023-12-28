import { ArrowRight } from '@jengaicons/react';
import { useNavigate, useParams } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';
import { toast } from '~/components/molecule/toast';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { useState } from 'react';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { validateCloudProvider } from '~/console/server/r-utils/common';
import { IdSelector } from '~/console/components/id-selector';
import ProgressWrapper from '~/console/components/progress-wrapper';

const NewCloudProvider = () => {
  const { a: accountName } = useParams();
  const api = useConsoleApi();

  const providers = [{ label: 'Amazon Web Services', value: 'aws' }];

  const navigate = useNavigate();
  const [isNameLoading, _setIsNameLoading] = useState(false);
  const { values, errors, handleSubmit, handleChange, isLoading } = useForm({
    initialValues: {
      displayName: '',
      name: '',
      provider: providers[0],
      awsAccountId: '',
    },
    validationSchema: Yup.object({
      displayName: Yup.string().required(),
      name: Yup.string().required(),
      provider: Yup.object({
        label: Yup.string().required(),
        value: Yup.string().required(),
      }).required(),
    }),
    onSubmit: async (val) => {
      const addProvider = async () => {
        switch (val.provider.value) {
          case 'aws':
            return api.createProviderSecret({
              secret: {
                displayName: val.displayName,
                metadata: {
                  name: val.name,
                },
                aws: {
                  awsAccountId: val.awsAccountId,
                },
                cloudProviderName: validateCloudProvider(val.provider.value),
              },
            });

          default:
            throw new Error('invalid provider');
        }
      };

      try {
        if (isNameLoading) {
          toast.error('id is being checked, please wait');
          return;
        }

        const { errors: e } = await addProvider();
        if (e) {
          throw e[0];
        }

        toast.success('provider secret created successfully');

        navigate(`/onboarding/${accountName}/${val.name}/validate-cp`);
      } catch (err) {
        handleError(err);
      }
    },
  });

  const progressItems = [
    {
      label: 'Create Team',
      active: false,
      completed: true,
    },
    {
      label: 'Add your Cloud Provider',
      active: true,
      completed: false,
      children: (
        <form className="py-3xl flex flex-col gap-3xl" onSubmit={handleSubmit}>
          <div className="bodyMd text-text-soft">
            A cloud provider offers remote computing resources and services over
            the internet.
          </div>
          <div className="flex flex-col">
            <TextInput
              label="Name"
              onChange={handleChange('displayName')}
              error={!!errors.displayName}
              message={errors.displayName}
              value={values.displayName}
              name="provider-secret-name"
              size="lg"
            />
            <IdSelector
              name={values.displayName}
              resType="providersecret"
              onChange={(id) => {
                handleChange('name')({ target: { value: id } });
              }}
              className="pt-xl"
            />

            <div className="flex flex-col gap-3xl pt-3xl">
              <Select
                error={!!errors.provider}
                message={errors.provider}
                value={values.provider}
                size="lg"
                label="Provider"
                onChange={(value) => {
                  handleChange('provider')(dummyEvent(value));
                }}
                options={async () => providers}
              />

              {values.provider.value === 'aws' && (
                <TextInput
                  name="awsAccountId"
                  onChange={handleChange('awsAccountId')}
                  error={!!errors.awsAccountId}
                  message={errors.awsAccountId}
                  value={values.awsAccountId}
                  label="Account ID"
                  size="lg"
                />
              )}
            </div>
          </div>
          <div className="flex flex-row gap-xl justify-start">
            <Button
              loading={isLoading}
              type="submit"
              variant="primary"
              content="Continue"
              suffix={<ArrowRight />}
            />
          </div>
        </form>
      ),
    },
    {
      label: 'Validate Cloud Provider',
      active: false,
      completed: false,
    },
    {
      label: 'Setup First Cluster',
      active: false,
      completed: false,
    },
  ];

  return (
    <ProgressWrapper
      title="Setup your account!"
      subTitle="Simplify Collaboration and Enhance Productivity with Kloudlite
  teams"
      progressItems={{
        items: progressItems,
      }}
    />
  );
};

export default NewCloudProvider;
