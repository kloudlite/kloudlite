import { useLoaderData, useNavigate, useParams } from '@remix-run/react';
import { toast } from '~/components/molecule/toast';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { NameIdView } from '~/console/components/name-id-view';
import MultiStepProgressWrapper from '~/console/components/multi-step-progress-wrapper';
import MultiStepProgress, {
  useMultiStepProgress,
} from '~/console/components/multi-step-progress';
import { BottomNavigation } from '~/console/components/commons';
import FillerCloudProvider from '~/console/assets/filler-cloud-provider';
import { IRemixCtx } from '~/root/lib/types/common';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { defer } from '@remix-run/node';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import { LoadingPlaceHolder } from '~/console/components/loading';
import CodeView from '~/console/components/code-view';

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { cloudprovider: clusterName } = ctx.params;
    const { data, errors } = await GQLServerHandler(ctx.request).getByokCluster(
      {
        name: clusterName,
      }
    );

    if (errors) {
      return { redirect: '/teams', cluster: data };
    }

    return {
      cluster: data,
      redirect: '',
    };
  });
  return defer({ promise });
};

const Validator = ({ cluster }: { cluster: any }) => {
  const { a: accountName } = useParams();
  const { params } = useParams();
  console.log('ppp', params, cluster.metadata.name);
  const api = useConsoleApi();

  const rootUrl = `/${accountName}/infra/clusters`;

  const { data, isLoading, error } = useCustomSwr(
    cluster.metadata?.name || null,
    async () => {
      if (!cluster.metadata?.name) {
        throw new Error('Invalid cluster name');
      }
      return api.getBYOKClusterInstructions({
        name: cluster.metadata.name,
      });
    }
  );

  const navigate = useNavigate();
  // const { values, errors, handleSubmit, handleChange, isLoading } = useForm({
  //   initialValues: {
  //     name: '',
  //     displayName: '',
  //     isNameError: false,
  //   },
  //   validationSchema: Yup.object({
  //     name: Yup.string().required('id is required'),
  //     displayName: Yup.string().required('name is required'),
  //   }),
  //   onSubmit: async (val) => {
  //     try {
  //       const { errors: e } = await api.createBYOKCluster({
  //         cluster: {
  //           displayName: val.displayName,
  //           metadata: {
  //             name: val.name,
  //           },
  //         },
  //       });
  //       if (e) {
  //         throw e[0];
  //       }
  //       toast.success('Cluster attched successfully');
  //       navigate(rootUrl);
  //     } catch (err) {
  //       handleError(err);
  //     }
  //   },
  // });

  const { currentStep, jumpStep } = useMultiStepProgress({
    defaultStep: 3,
    totalSteps: 3,
  });

  return (
    <form>
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
          <MultiStepProgress.Step
            step={2}
            label="Attach Kubernetes Cluster"
            className="py-3xl flex flex-col gap-3xl
          "
          />
          <MultiStepProgress.Step
            step={3}
            label="Verify Your Attached Kubernetes Cluster"
          >
            <div className="flex flex-col gap-3xl">
              {error && (
                <span className="bodyMd-medium text-text-strong">
                  Error while fetching instructions
                </span>
              )}
              {isLoading ? (
                <LoadingPlaceHolder />
              ) : (
                data && (
                  <div className="flex flex-col gap-sm text-start ">
                    <span className="flex flex-wrap items-center gap-md py-lg">
                      Please follow below instruction for further steps
                    </span>

                    {data.map((d, index) => {
                      return (
                        <div
                          key={d.title}
                          className="flex flex-col gap-lg pb-2xl"
                        >
                          <span className="bodyMd-medium text-text-strong font-bold">
                            Step {`${index + 1}: ${d.title}`}
                          </span>
                          <CodeView
                            preClassName="!overflow-none text-wrap break-words"
                            copy
                            data={d.command || ''}
                          />
                        </div>
                      );
                    })}
                  </div>
                )
              )}
              <BottomNavigation
                secondaryButton={{
                  variant: 'outline',
                  content: 'Skip',
                  prefix: undefined,
                  onClick: () => {
                    navigate(rootUrl);
                  },
                }}
                primaryButton={{
                  variant: 'primary',
                  content: 'Create',
                  loading: isLoading,
                  type: 'submit',
                  onClick: () => {
                    navigate(rootUrl);
                  },
                }}
              />
            </div>
          </MultiStepProgress.Step>
          {/* <MultiStepProgress.Step step={3} label="Verify Cluster" /> */}
          {/* <MultiStepProgress.Step step={3} label="Validate cloud provider" /> */}
          {/* <MultiStepProgress.Step step={4} label="Setup first cluster" /> */}
        </MultiStepProgress.Root>
      </MultiStepProgressWrapper>
    </form>
  );
};

const ValidateAttachedCluster = () => {
  const { promise } = useLoaderData<typeof loader>();

  return (
    <LoadingComp data={promise}>
      {({ cluster }) => {
        if (cluster === null) {
          return null;
        }
        return <Validator cluster={cluster} />;
      }}
    </LoadingComp>
  );
};

export default ValidateAttachedCluster;
