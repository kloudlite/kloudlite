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
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { PasswordInput } from '~/components/atoms/input';
import FillerCloudProvider from '~/console/assets/filler-cloud-provider';
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
    () => parseName(cloudProvider) + isLoading,
    async () => {
      if (!parseName(cloudProvider)) {
        throw new Error('Invalid cloud provider name');
      }
      return api.checkAwsAccess({
        cloudproviderName: parseName(cloudProvider),
      });
    }
  );

  const { values, handleChange, errors, handleSubmit } = useForm({
    initialValues: {
      accessKey: '',
      secretKey: '',
    },
    validationSchema: Yup.object({
      accessKey: Yup.string().test(
        'provider',
        'access key is required',
        // @ts-ignores
        // eslint-disable-next-line react/no-this-in-sfc
        function (item) {
          return data?.result || item;
        }
      ),
      secretKey: Yup.string().test(
        'provider',
        'secret key is required',
        // eslint-disable-next-line func-names
        // @ts-ignore
        function (item) {
          return data?.result || item;
        }
      ),
    }),
    onSubmit: async (val) => {
      if (data?.result) {
        navigate(
          `/onboarding/${parseName(account)}/${parseName(
            cloudProvider
          )}/new-cluster`
        );
        return;
      }

      try {
        const { errors } = await api.updateProviderSecret({
          secret: {
            metadata: {
              name: parseName(cloudProvider),
            },
            cloudProviderName: cloudProvider.cloudProviderName,
            displayName: cloudProvider.displayName,
            aws: {
              authMechanism: 'secret_keys',
              authSecretKeys: {
                accessKey: val.accessKey,
                secretKey: val.secretKey,
              },
            },
          },
        });

        if (errors) {
          throw errors[0];
        }

        setIsLoading((s) => !s);
      } catch (err) {
        handleError(err);
      }
    },
  });

  const { currentStep, jumpStep } = useMultiStepProgress({
    defaultStep: 3,
    totalSteps: 4,
  });

  return (
    <div>
      <MultiStepProgressWrapper
        fillerImage={<FillerCloudProvider />}
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
            <form className="flex flex-col gap-3xl" onSubmit={handleSubmit}>
              <div className="bodyMd text-text-soft">
                Validate your cloud provider&apos;s credentials
              </div>
              {/* eslint-disable-next-line no-nested-ternary */}
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
                    <span className="bodyLg-medium">
                      {cloudProvider.displayName}
                    </span>
                    <span className="bodyMd-semibold text-text-primary">
                      ({parseName(cloudProvider)})
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

                  {!data?.result && (
                    <>
                      <div className="">
                        Once you have created the cloudformation stack, please
                        enter the access key and secret key below to validate
                        your cloud Provider, you can get the access key and
                        secret key from the output of the cloudformation stack.
                      </div>

                      <PasswordInput
                        name="accessKey"
                        onChange={handleChange('accessKey')}
                        error={!!errors.accessKey}
                        message={errors.accessKey}
                        value={values.accessKey}
                        label="Access Key"
                      />

                      <PasswordInput
                        name="secretKey"
                        onChange={handleChange('secretKey')}
                        error={!!errors.secretKey}
                        message={errors.secretKey}
                        value={values.secretKey}
                        label="Secret Key"
                      />
                    </>
                  )}
                </div>
              )}
              <BottomNavigation
                secondaryButton={
                  data?.result
                    ? undefined
                    : {
                        variant: 'outline',
                        content: 'Skip',
                        prefix: undefined,
                        onClick: () => {
                          navigate(
                            `/onboarding/${parseName(account)}/${parseName(
                              cloudProvider
                            )}/new-cluster`
                          );
                        },
                      }
                }
                primaryButton={{
                  variant: 'primary',
                  content: data?.result ? 'Continue' : 'Update',
                  type: 'submit',
                }}
              />
            </form>
          </MultiStepProgress.Step>
          <MultiStepProgress.Step step={4} label="Setup first cluster" />
        </MultiStepProgress.Root>
      </MultiStepProgressWrapper>
    </div>
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
