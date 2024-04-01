/* eslint-disable react/destructuring-assignment */
import { IDialogBase } from '~/console/components/types.d';
import { useOutletContext } from '@remix-run/react';
import Select from '~/components/atoms/select';
import { toast } from '~/components/molecule/toast';
import { useMapper } from '~/components/utils';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { handleError } from '~/root/lib/utils/common';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import Popup from '~/components/molecule/popup';
import CommonPopupHandle from '~/console/components/common-popup-handle';
import { IBuilds } from '~/console/server/gql/queries/build-queries';
import {
  ExtractNodeType,
  parseName,
  parseNodes,
} from '~/console/server/r-utils/common';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import Git from '~/console/components/git';
import { IGIT_PROVIDERS } from '~/console/hooks/use-git';
import MultiStep, { useMultiStep } from '~/console/components/multi-step';
import { TextInput } from '~/components/atoms/input';
import { GitDetail } from '~/console/components/commons';
import { IRepoContext } from '../_layout';
import AdvancedOptions from './advanced-options';

type IDialog = IDialogBase<
  Omit<
    ExtractNodeType<IBuilds>,
    | 'creationTime'
    | 'latestBuildRun'
    | 'errorMessages'
    | 'status'
    | 'markedForDeletion'
    | 'updateTime'
    | 'createdBy'
    | 'credUser'
    | 'lastUpdatedBy'
  > & {
    mode?: 'app' | 'build';
  }
>;

const Root = (props: IDialog) => {
  const { isUpdate, setVisible } = props;
  const api = useConsoleApi();
  const reloadPage = useReload();

  const { loginUrls, logins } = useOutletContext<IRepoContext>();

  const {
    data: clusters,
    error: errorCluster,
    isLoading: clusterLoading,
  } = useCustomSwr('/clusters', async () => api.listClusters({}));

  const clusterData = useMapper(parseNodes(clusters), (item) => {
    return {
      label: item.displayName,
      value: parseName(item),
      render: () => (
        <div className="flex flex-col">
          <div>{item.displayName}</div>
          <div className="bodySm text-text-soft">{parseName(item)}</div>
        </div>
      ),
    };
  });

  const { currentStep, onPrevious, onNext } = useMultiStep({
    defaultStep: isUpdate ? 2 : 1,
    totalSteps: 2,
  });

  const isAdvanceOptions = (data: any) => {
    if (!data) {
      return false;
    }
    return Object.values(data).some((d) => {
      return !!d;
    });
  };
  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: isUpdate
        ? {
            name: props.data.name,
            source: {
              branch: props.data.source.branch,
              repository: props.data.source.repository,
              provider: props.data.source.provider,
            },
            tags: props.data.spec.registry.repo.tags,
            buildClusterName: props.data.buildClusterName,
            repository: props.data.spec.registry.repo.name,
            advanceOptions:
              isAdvanceOptions(props.data.spec.buildOptions) ||
              (props.data.spec.caches || []).length > 0,
            ...props.data.spec.buildOptions,
            caches: props.data.spec.caches || [],
          }
        : {
            name: '',
            source: {
              branch: '',
              repository: '',
              provider: '' as IGIT_PROVIDERS,
            },
            tags: [],
            buildClusterName: '',
            advanceOptions: false,
            repository: '',
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
        const submit = async () => {
          try {
            if (isUpdate) {
              console.log('build data', props.data);
              const { errors: e } = await api.updateBuild({
                crUpdateBuildId: props.data.id,
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
                        name: val.repository,
                        tags: val.tags,
                      },
                    },
                    resource: {
                      cpu: 500,
                      memoryInMb: 1000,
                    },
                    caches: val.caches.map((v) => ({
                      path: v.path,
                      name: v.name,
                    })),
                  },
                },
              });
              if (e) {
                throw e[0];
              }
            }

            reloadPage();
            resetValues();
            toast.success(
              `Build ${isUpdate ? 'updated' : 'created'} successfully`
            );
            setVisible(false);
          } catch (err) {
            handleError(err);
          }
        };
        switch (currentStep) {
          case 1:
            onNext();
            break;
          case 2:
            await submit();
            break;
          default:
            break;
        }
      },
    });

  return (
    <Popup.Form onSubmit={handleSubmit}>
      <Popup.Content>
        <MultiStep.Root currentStep={currentStep}>
          <MultiStep.Step step={1}>
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
          </MultiStep.Step>
          <MultiStep.Step step={2}>
            <div className="flex flex-col gap-2xl">
              <GitDetail
                provider={values.source.provider}
                repository={values.source.repository}
                branch={values.source.branch}
                onEdit={onPrevious}
              />
              <div className="flex flex-col gap-xl rounded border border-border-default p-xl">
                <div className="flex flex-col gap-xl">
                  <TextInput
                    label="Name"
                    value={values.name}
                    onChange={handleChange('name')}
                    error={!!errors.name}
                    message={errors.name}
                    disabled={props.isUpdate && props.data.mode === 'app'}
                  />
                  <Select
                    label="Tags"
                    size="md"
                    creatable
                    multiple
                    value={values.tags}
                    options={async () =>
                      values.tags.map((t: string) => ({ label: t, value: t }))
                    }
                    onChange={(_, val) => {
                      handleChange('tags')(dummyEvent(val));
                    }}
                    error={!!errors.tags}
                    message={errors.tags}
                    disabled={props.isUpdate && props.data.mode === 'app'}
                  />

                  <Select
                    label="Clusters"
                    size="md"
                    value={values.buildClusterName}
                    options={async () => clusterData}
                    onChange={(_, val) => {
                      handleChange('buildClusterName')(dummyEvent(val));
                    }}
                    error={!!errors.buildClusterName || !!errorCluster}
                    message={
                      // eslint-disable-next-line no-nested-ternary
                      errors.buildClusterName
                        ? errors.buildClusterName
                        : errorCluster
                        ? 'Error loading clusters.'
                        : ''
                    }
                    loading={clusterLoading}
                    disabled={props.isUpdate && props.data.mode === 'app'}
                  />

                  <AdvancedOptions
                    values={values}
                    handleChange={handleChange}
                    errors={errors}
                  />
                </div>
              </div>
            </div>
          </MultiStep.Step>
        </MultiStep.Root>
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button closable content="Cancel" variant="basic" />
        <Popup.Button
          type="submit"
          content={currentStep === 1 ? 'Continue' : 'Update'}
          variant="primary"
          loading={isLoading}
        />
      </Popup.Footer>
    </Popup.Form>
  );
};

const HandleBuild = (props: IDialog) => {
  return (
    <CommonPopupHandle
      {...props}
      createTitle="Create build"
      updateTitle="Update build"
      root={Root}
    />
  );
};
export default HandleBuild;
