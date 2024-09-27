import { useNavigate, useOutletContext, useParams } from '@remix-run/react';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import MultiStepProgress, {
  useMultiStepProgress,
} from '~/iotconsole/components/multi-step-progress';
import MultiStepProgressWrapper from '~/iotconsole/components/multi-step-progress-wrapper';
import { useIotConsoleApi } from '~/iotconsole/server/gql/api-provider';
import { BottomNavigation } from '~/iotconsole/components/commons';
import { toast } from '~/components/molecule/toast';
import Git from '~/iotconsole/components/git';
import { IGIT_PROVIDERS } from '~/iotconsole/hooks/use-git';
import { constants } from '~/iotconsole/server/utils/constants';
import ReviewBuild from './review-build';
import BuildDetails from './build-details';
import { IRepoContext } from '../_layout';

const NewBuild = () => {
  const { loginUrls, logins, repoName } = useOutletContext<IRepoContext>();

  const navigate = useNavigate();

  const { account } = useParams();

  const { currentStep, nextStep, jumpStep } = useMultiStepProgress({
    defaultStep: 1,
    totalSteps: 3,
  });

  const api = useIotConsoleApi();

  const { values, errors, handleSubmit, handleChange, isLoading } = useForm({
    initialValues: {
      name: '',
      source: {
        branch: '',
        repository: '',
        provider: '',
      },
      tags: [],
      buildClusterName: constants.kloudliteClusterName,
      advanceOptions: false,
      repository: repoName || '',
      buildArgs: {},
      buildContexts: {},
      contextDir: '',
      dockerfilePath: '',
      dockerfileContent: '',
      isGitLoading: false,
      caches: [],
    },
    validationSchema: Yup.object({
      source: Yup.object()
        .shape({
          branch: Yup.string().required('Branch is required'),
        })
        .required('Branch is required'),
      name: Yup.string().test('required', 'Name is required', (v) => {
        return !(currentStep === 2 && !v);
      }),
      buildClusterName: Yup.string().test(
        'required',
        'Build cluster name is required',
        (v) => {
          return !(currentStep === 2 && !v);
        }
      ),
      tags: Yup.array().test('required', 'Tags is required', (value = []) => {
        return !(currentStep === 2 && !(value.length > 0));
      }),
    }),
    onSubmit: async (val) => {
      if (!repoName) {
        toast.error('Repository is required!.');
        return;
      }
      const submit = async () => {
        try {
          const { errors: e } = await api.createBuild({
            build: {
              name: val.name,
              buildClusterName: val.buildClusterName,
              source: {
                branch: val.source.branch,
                provider:
                  val.source.provider === 'github' ? 'github' : 'gitlab',
                repository: val.source.repository,
              },
              spec: {
                ...{
                  ...(val.advanceOptions
                    ? {
                        buildOptions: {
                          buildArgs: val.buildArgs,
                          buildContexts: val.buildContexts,
                          contextDir: val.contextDir,
                          dockerfileContent: val.dockerfileContent,
                          dockerfilePath: val.dockerfilePath,
                          targetPlatforms: [],
                        },
                      }
                    : {}),
                },
                registry: {
                  repo: {
                    name: val.repository || '',
                    tags: val.tags,
                  },
                },
                resource: {
                  cpu: 500,
                  memoryInMb: 1000,
                },
                caches: val.caches.map((c: any) => {
                  return { name: c.name, path: c.path };
                }),
              },
            },
          });
          if (e) {
            throw e[0];
          }
          navigate(`/${account}/repo/${btoa(repoName)}/builds`);
        } catch (err) {
          handleError(err);
        }
      };

      try {
        switch (currentStep) {
          case 1:
            nextStep();
            break;
          case 2:
            nextStep();
            break;
          default:
            await submit();
            break;
        }
      } catch (err) {
        handleError(err);
      }
    },
  });

  return (
    <form onSubmit={handleSubmit}>
      <MultiStepProgressWrapper
        title="Create new Build"
        subTitle="Create your build under to repo effortlessly"
        backButton={{
          content: 'Back to builds',
          to: '../builds',
        }}
      >
        <MultiStepProgress.Root currentStep={currentStep} jumpStep={jumpStep}>
          <MultiStepProgress.Step label="Import git repository" step={1}>
            <div className="flex flex-col gap-3xl">
              <Git
                logins={logins}
                loginUrls={loginUrls}
                error={errors?.['source.branch'] || ''}
                onChange={(git) => {
                  handleChange('source')(
                    dummyEvent({
                      branch: git.branch,
                      repository: git.repo,
                      provider: git.provider,
                    })
                  );
                }}
                value={{
                  branch: values.source.branch,
                  repo: values.source.repository,
                  provider:
                    (values.source.provider as IGIT_PROVIDERS) || 'github',
                }}
              />
              <BottomNavigation
                primaryButton={{
                  type: 'submit',
                  disabled: !values.source.branch,
                  content: 'Next',
                }}
              />
            </div>
          </MultiStepProgress.Step>
          <MultiStepProgress.Step label="Build configurations" step={2}>
            <div className="flex flex-col gap-3xl">
              <BuildDetails
                errors={errors}
                values={values}
                handleChange={handleChange}
              />
              <BottomNavigation
                primaryButton={{
                  type: 'submit',
                  disabled: !values.source.branch,
                  content: 'Next',
                }}
              />
            </div>
          </MultiStepProgress.Step>
          <MultiStepProgress.Step label="Review" step={3}>
            <div className="flex flex-col gap-3xl">
              <ReviewBuild values={values} onEdit={(step) => jumpStep(step)} />
              <BottomNavigation
                primaryButton={{
                  type: 'submit',
                  content: 'Create',
                  loading: isLoading,
                }}
              />
            </div>
          </MultiStepProgress.Step>
        </MultiStepProgress.Root>
      </MultiStepProgressWrapper>
    </form>
  );
};

export default NewBuild;

export const handle = {
  noMainLayout: true,
};
