import { ArrowLeft, ArrowRight } from '@jengaicons/react';
import { useNavigate, useParams } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';
import { toast } from '~/components/molecule/toast';
import { useMapper } from '~/components/utils';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { useState } from 'react';
import { IdSelector } from '../components/id-selector';
import RawWrapper, { TitleBox } from '../components/raw-wrapper';
import { useConsoleApi } from '../server/gql/api-provider';
import { validateCloudProvider } from '../server/r-utils/common';
import { FadeIn } from './_.$account.$cluster.$project.$scope.$workspace.new-app/util';
import { AwsForm } from '../page-components/cloud-provider';
import { asyncPopupWindow } from '../utils/commons';

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
      accessKey: '',
      accessSecret: '',
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
            if (val.awsAccountId) {
              // return validateAccountIdAndPerform(async () => {
              // });

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
            }

            return api.createProviderSecret({
              secret: {
                displayName: val.displayName,
                metadata: {
                  name: val.name,
                },
                aws: {
                  accessKey: val.accessKey,
                  secretKey: val.accessSecret,
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
          <div className="flex flex-col gap-2xl">
            <TextInput
              label="Name"
              onChange={handleChange('displayName')}
              error={!!errors.displayName}
              message={errors.displayName}
              value={values.displayName}
              name="provider-secret-name"
            />
            <IdSelector
              name={values.displayName}
              resType="providersecret"
              onChange={(id) => {
                handleChange('name')({ target: { value: id } });
              }}
            />

            <Select
              error={!!errors.provider}
              message={errors.provider}
              value={values.provider}
              label="Provider"
              onChange={(value) => {
                handleChange('provider')(dummyEvent(value));
              }}
              options={async () => providers}
            />

            {values.provider.value === 'aws' && (
              <AwsForm
                {...{
                  values,
                  errors,
                  handleChange,
                }}
              />
            )}
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
