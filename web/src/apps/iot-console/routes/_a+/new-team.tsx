import { useNavigate } from '@remix-run/react';
import { Button } from '@kloudlite/design-system/atoms/button';
import { toast } from '@kloudlite/design-system/molecule/toast';
import FillerCreateTeam from '~/iotconsole/assets/filler-create-team';
import { BottomNavigation } from '~/iotconsole/components/commons';
import { SignOut } from '~/iotconsole/components/icons';
import MultiStepProgress, {
  useMultiStepProgress,
} from '~/iotconsole/components/multi-step-progress';
import MultiStepProgressWrapper from '~/iotconsole/components/multi-step-progress-wrapper';
import { NameIdView } from '~/iotconsole/components/name-id-view';
import { useIotConsoleApi } from '~/iotconsole/server/gql/api-provider';
import { useExternalRedirect } from '~/root/lib/client/helpers/use-redirect';
import { useDataFromMatches } from '~/root/lib/client/hooks/use-custom-matches';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import useForm from '~/root/lib/client/hooks/use-form';
import { authBaseUrl } from '~/root/lib/configs/base-url.cjs';
import { UserMe } from '~/root/lib/server/gql/saved-queries';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';

const NewAccount = () => {
  const api = useIotConsoleApi();
  const navigate = useNavigate();
  const user = useDataFromMatches<UserMe>('user', {});

  const { data: accountsData } = useCustomSwr('/list_accounts', async () => {
    return api.listAccounts({});
  });

  const { values, handleChange, errors, isLoading, handleSubmit } = useForm({
    initialValues: {
      name: '',
      displayName: '',
      isNameError: false,
    },
    validationSchema: Yup.object({
      name: Yup.string().required(),
      displayName: Yup.string().required(),
    }),
    onSubmit: async (v) => {
      try {
        const { errors: _errors } = await api.createAccount({
          account: {
            metadata: { name: v.name },
            displayName: v.displayName,
            contactEmail: user.email,
          },
        });
        if (_errors) {
          throw _errors[0];
        }
        toast.success('account created');
        navigate(`/onboarding/${v.name}/new-cloud-provider`);
      } catch (err) {
        handleError(err);
      }
    },
  });

  const { currentStep, jumpStep } = useMultiStepProgress({
    defaultStep: 1,
    totalSteps: 4,
  });
  const eNavigate = useExternalRedirect();

  return (
    <form onSubmit={handleSubmit}>
      <MultiStepProgressWrapper
        fillerImage={<FillerCreateTeam />}
        title="Setup your account!"
        action={
          accountsData?.length === 0 && (
            <Button
              variant="plain"
              suffix={<SignOut />}
              size="sm"
              content="Sign Out"
              onClick={() => {
                eNavigate(`${authBaseUrl}/logout`);
              }}
            />
          )
        }
        subTitle="Simplify Collaboration and Enhance Productivity with Kloudlite
  teams"
      >
        <MultiStepProgress.Root
          currentStep={currentStep}
          editable={false}
          noJump={() => true}
          jumpStep={jumpStep}
        >
          <MultiStepProgress.Step step={1} label="Create team">
            <div className="flex flex-col gap-3xl">
              <NameIdView
                label="Name"
                resType="account"
                name={values.name}
                displayName={values.displayName}
                errors={errors.name}
                handleChange={handleChange}
                nameErrorLabel="isNameError"
              />
              <BottomNavigation
                primaryButton={{
                  variant: 'primary',
                  content: 'Next',
                  loading: isLoading,
                  type: 'submit',
                }}
              />
            </div>
          </MultiStepProgress.Step>
          <MultiStepProgress.Step step={2} label="Add your cloud provider" />
          <MultiStepProgress.Step step={3} label="Validate cloud provider" />
          <MultiStepProgress.Step step={4} label="Setup first cluster" />
        </MultiStepProgress.Root>
      </MultiStepProgressWrapper>
    </form>
  );
};

export default NewAccount;
