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
import { NameIdView } from '~/console/components/name-id-view';
import MultiStepProgressWrapper from '~/console/components/multi-step-progress-wrapper';
import MultiStepProgress, {
  useMultiStepProgress,
} from '~/console/components/multi-step-progress';

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
      provider: providers[0].value,
      awsAccountId: '',
      isNameError: false,
    },
    validationSchema: Yup.object({
      displayName: Yup.string().required(),
      name: Yup.string().required(),
      provider: Yup.string().required(),
      awsAccountId: Yup.string().required('AccountId is required.'),
    }),
    onSubmit: async (val) => {
      const addProvider = async () => {
        switch (val.provider) {
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
                cloudProviderName: validateCloudProvider(val.provider),
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

  const { currentStep, jumpStep } = useMultiStepProgress({
    defaultStep: 2,
    totalSteps: 4,
  });

  return (
    <form onSubmit={handleSubmit}>
      <MultiStepProgressWrapper
        title="Setup your account!"
        subTitle="Simplify Collaboration and Enhance Productivity with Kloudlite
  teams"
      >
        <MultiStepProgress.Root
          currentStep={currentStep}
          editable={false}
          noJump={() => true}
          jumpStep={jumpStep}
        >
          <MultiStepProgress.Step
            step={1}
            label="Create team"
            className="py-3xl flex flex-col gap-3xl
            "
          />
          <MultiStepProgress.Step step={2} label="Add your cloud provider">
            <div className="flex flex-col gap-3xl">
              <div className="bodyMd text-text-soft">
                A cloud provider offers remote computing resources and services
                over the internet.
              </div>
              <div className="flex flex-col">
                <NameIdView
                  nameErrorLabel="isNameError"
                  resType="providersecret"
                  displayName={values.displayName}
                  name={values.name}
                  label="Name"
                  placeholder="Enter provider name"
                  errors={errors.name}
                  handleChange={handleChange}
                />
                <div className="flex flex-col gap-3xl pt-3xl">
                  <Select
                    error={!!errors.provider}
                    message={errors.provider}
                    value={values.provider}
                    size="lg"
                    label="Provider"
                    onChange={(_, value) => {
                      handleChange('provider')(dummyEvent(value));
                    }}
                    options={async () => providers}
                  />

                  {values.provider === 'aws' && (
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
            </div>
          </MultiStepProgress.Step>
          <MultiStepProgress.Step step={3} label="Validate cloud provider" />
          <MultiStepProgress.Step step={4} label="Setup first cluster" />
        </MultiStepProgress.Root>
      </MultiStepProgressWrapper>
    </form>
  );
};

export default NewCloudProvider;
