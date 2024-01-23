import { ArrowRight } from '@jengaicons/react';
import { useNavigate, useParams } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import MultiStepProgress, {
  useMultiStepProgress,
} from '~/console/components/multi-step-progress';
import MultiStepProgressWrapper from '~/console/components/multi-step-progress-wrapper';
import GitRepoSelector from '~/console/components/git-repo-selector';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import BuildDetails from './build-details';
import ReviewBuild from './review-build';

const NewBuild = () => {
  const navigate = useNavigate();

  const { account } = useParams();

  const { currentStep, nextStep, jumpStep } = useMultiStepProgress({
    defaultStep: 1,
    totalSteps: 3,
  });

  const api = useConsoleApi();
  const { repo } = useParams();

  const { values, errors, handleSubmit, handleChange, isLoading } = useForm({
    initialValues: {
      name: '',
      source: {
        branch: '',
        repository: '',
        provider: '',
      },
      tags: [],
      buildClusterName: '',
      advanceOptions: false,
      repository: repo || '',
      buildArgs: {},
      buildContexts: {},
      contextDir: '',
      dockerfilePath: '',
      dockerfileContent: '',
    },
    validationSchema: Yup.object({
      source: Yup.object().shape({
        branch: Yup.string().required('Branch is required'),
      }),
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
      if (!repo) {
        throw new Error('Repository is required!.');
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
                    tags: val.tags.map((t: any) => t.value),
                  },
                },
                resource: {
                  cpu: 500,
                  memoryInMb: 1000,
                },
              },
            },
          });
          if (e) {
            throw e[0];
          }
          navigate(`/${account}/repo/${repo}/builds`);
        } catch (err) {
          handleError(err);
        }
      };
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
              <GitRepoSelector
                onChange={(git) => {
                  handleChange('source')(
                    dummyEvent({
                      branch: git.branch,
                      repository: git.repo,
                      provider: git.provider,
                    })
                  );
                  console.log(git);
                }}
                error={errors?.['source.branch'] || ''}
              />
              <Button
                content="Next"
                variant="primary"
                type="submit"
                size="lg"
                suffix={<ArrowRight />}
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
              <Button
                content="Next"
                variant="primary"
                type="submit"
                size="lg"
                suffix={<ArrowRight />}
              />
            </div>
          </MultiStepProgress.Step>
          <MultiStepProgress.Step label="Review" step={3}>
            <div className="flex flex-col gap-3xl">
              <ReviewBuild values={values} onEdit={(step) => jumpStep(step)} />
              <Button
                content="Create"
                variant="primary"
                type="submit"
                size="lg"
                loading={isLoading}
                suffix={<ArrowRight />}
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
