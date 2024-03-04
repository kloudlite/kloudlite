/* eslint-disable react/no-unescaped-entities */
/* eslint-disable no-nested-ternary */
import { IRemixCtx } from '~/root/lib/types/common';
import { useLoaderData, useNavigate, useOutletContext } from '@remix-run/react';
import { defer } from '@remix-run/node';
import { useState } from 'react';
import { toast } from '~/components/molecule/toast';
import { handleError } from '~/root/lib/utils/common';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import { Badge } from '~/components/atoms/badge';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { parseName } from '~/console/server/r-utils/common';
import { LoadingPlaceHolder } from '~/console/components/loading';
import CodeView from '~/console/components/code-view';
import { asyncPopupWindow } from '~/console/utils/commons';
import MultiStepProgressWrapper from '~/console/components/multi-step-progress-wrapper';
import MultiStepProgress, {
  useMultiStepProgress,
} from '~/console/components/multi-step-progress';
import { Check } from '~/console/components/icons';
import { BottomNavigation } from '~/console/components/commons';
import { IAccountContext } from '../../../../_main+/$account+/_layout';

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { cloudprovider: cp } = ctx.params;
    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).getProviderSecret({
      name: cp,
    });

    if (errors) {
      return { redirect: '/teams', cloudProvider: data };
    }

    return {
      cloudProvider: data,
      redirect: '',
    };
  });
  return defer({ promise });
};

const Validator = ({ cloudProvider }: { cloudProvider: any }) => {
  const { account } = useOutletContext<IAccountContext>();
  const navigate = useNavigate();

  const api = useConsoleApi();
  const checkAwsAccess = async () => {
    const { data, errors } = await api.checkAwsAccess({
      cloudproviderName: cloudProvider.metadata?.name || '',
    });
    if (errors) {
      throw errors[0];
    }
    return data;
  };

  const [isLoading, setIsLoading] = useState(false);

  const { data, isLoading: il } = useCustomSwr(
    () => cloudProvider.metadata!.name + isLoading,
    async () => {
      if (!cloudProvider.metadata!.name) {
        throw new Error('Invalid cloud provider name');
      }
      return api.checkAwsAccess({
        cloudproviderName: cloudProvider.metadata.name,
      });
    }
  );

  const { currentStep, jumpStep } = useMultiStepProgress({
    defaultStep: 3,
    totalSteps: 4,
  });

  return (
    <form>
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
          <MultiStepProgress.Step step={2} label="Add your cloud provider" />
          <MultiStepProgress.Step step={3} label="Validate cloud provider">
            <div className="flex flex-col gap-3xl">
              <div className="bodyMd text-text-soft">
                Validate your cloud provider's credentials
              </div>
              {il ? (
                <div className="py-2xl">
                  <LoadingPlaceHolder
                    title="Validating the cloudformation stack"
                    height={100}
                  />
                </div>
              ) : data?.result ? (
                <div className="py-2xl">
                  <Badge type="success" icon={<Check />}>
                    Your Credential is valid
                  </Badge>
                </div>
              ) : (
                <div className="flex flex-col gap-3xl p-xl border border-border-default rounded">
                  <div className="flex gap-xl items-center">
                    <span>Account ID</span>
                    <span className="bodyMd-semibold text-text-primary">
                      {cloudProvider.aws?.awsAccountId}
                    </span>
                  </div>
                  <div className="flex flex-col gap-2xl text-start">
                    <CodeView copy data={data?.installationUrl || ''} />

                    <span className="">
                      visit the link above or
                      <button
                        className="inline-block mx-lg text-text-primary hover:underline"
                        onClick={async () => {
                          setIsLoading(true);
                          try {
                            await asyncPopupWindow({
                              url: data?.installationUrl || '',
                            });

                            const res = await checkAwsAccess();

                            if (res.result) {
                              toast.success(
                                'Aws account validated successfully'
                              );
                            } else {
                              toast.error('Aws account validation failed');
                            }
                          } catch (err) {
                            handleError(err);
                          }

                          setIsLoading(false);
                        }}
                      >
                        click here
                      </button>
                      to create AWS cloudformation stack
                    </span>
                  </div>
                </div>
              )}
              <BottomNavigation
                primaryButton={{
                  variant: 'primary',
                  content: data?.result ? 'Next' : 'Skip',
                  onClick: () => {
                    navigate(
                      `/onboarding/${parseName(account)}/${parseName(
                        cloudProvider
                      )}/new-cluster`
                    );
                  },
                }}
              />
            </div>
          </MultiStepProgress.Step>
          <MultiStepProgress.Step step={4} label="Setup first cluster" />
        </MultiStepProgress.Root>
      </MultiStepProgressWrapper>
    </form>
  );
};

const _NewCluster = () => {
  const { promise } = useLoaderData<typeof loader>();

  return (
    <LoadingComp data={promise}>
      {({ cloudProvider }) => {
        if (cloudProvider === null) {
          return null;
        }
        return <Validator cloudProvider={cloudProvider} />;
      }}
    </LoadingComp>
  );
};

export default _NewCluster;
