/* eslint-disable react/destructuring-assignment */
import { IDialogBase } from '~/console/components/types.d';

import {
  GitBranch,
  GithubLogoFill,
  GitlabLogoFill,
  MinusCircle,
  PencilSimple,
} from '@jengaicons/react';
import { useParams } from '@remix-run/react';
import AnimateHide from '~/components/atoms/animate-hide';
import { Checkbox } from '~/components/atoms/checkbox';
import Select from '~/components/atoms/select';
import { toast } from '~/components/molecule/toast';
import { cn, useMapper, uuid } from '~/components/utils';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { handleError } from '~/root/lib/utils/common';
import { ReactNode, useEffect, useState } from 'react';
import { TextArea, TextInput } from '~/components/atoms/input';
import { Button, IconButton } from '~/components/atoms/button';
import MultiStep, { useMultiStep } from '~/console/components/multi-step';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import Popup from '~/components/molecule/popup';
import GitRepoSelector from '~/console/components/git-repo-selector';
import CommonPopupHandle from '~/console/components/common-popup-handle';
import { IBuilds } from '~/console/server/gql/queries/build-queries';
import {
  ExtractNodeType,
  parseName,
  parseNodes,
} from '~/console/server/r-utils/common';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import KeyValuePair from '~/console/components/key-value-pair';

interface ISource {
  repo: string;
  branch: string;
  provider: 'github' | 'gitlab';
}

const BuildPlatforms = ({
  value,
  onChange,
}: {
  value?: Array<string>;
  onChange?(data: Array<string>): void;
}) => {
  const platforms = [
    { label: 'Arm', value: 'arm', checked: false },
    { label: 'x86', value: 'x86', checked: false },
    { label: 'x64', value: 'x64', checked: false },
  ];

  const [options, setOptions] = useState(platforms);

  useEffect(() => {
    setOptions((prev) =>
      prev.map((p) => {
        if (value?.includes(p.value)) {
          return { ...p, checked: true };
        }
        return { ...p, checked: false };
      })
    );
  }, [value]);

  useEffect(() => {
    onChange?.(options.filter((opt) => opt.checked).map((op) => op.value));
  }, [options]);

  return (
    <div className="flex flex-col gap-md">
      <span className="text-text-default bodyMd-medium">Platforms</span>
      <div className="flex flex-row items-center gap-xl">
        {options.map((bp) => {
          return (
            <Checkbox
              key={bp.label}
              label={bp.label}
              checked={bp.checked}
              onChange={(checked) => {
                setOptions((prev) =>
                  prev.map((p) => {
                    if (p.value === bp.value) {
                      return { ...p, checked: !!checked };
                    }
                    return p;
                  })
                );
              }}
            />
          );
        })}
      </div>
    </div>
  );
};

type IDialog = IDialogBase<ExtractNodeType<IBuilds>>;

const Root = (props: IDialog) => {
  const { isUpdate, setVisible } = props;
  const api = useConsoleApi();
  const reloadPage = useReload();

  const { data: clusters, error: errorCluster } = useCustomSwr(
    '/clusters',
    async () => api.listClusters({})
  );

  const clusterData = useMapper(parseNodes(clusters), (item) => {
    console.log(errorCluster);

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

  useEffect(() => {
    if (errorCluster) {
      toast.error('Error loading clusters.');
    }
  }, [errorCluster]);

  const [source, setSource] = useState<ISource | null>(null);

  useEffect(() => {
    if (isUpdate) {
      setSource({ ...props.data.source, repo: props.data.source.repository });
    }
  }, [isUpdate]);

  const { currentStep, onNext, onPrevious } = useMultiStep({
    defaultStep: isUpdate ? 2 : 1,
    totalSteps: 2,
  });

  const { repo } = useParams();

  const isAdvanceOptions = (data: any) => {
    if (!data) {
      return false;
    }
    return Object.values(data).some((d) => {
      return !!d;
    });
  };
  const { values, errors, handleChange, handleSubmit, resetValues } = useForm({
    initialValues: isUpdate
      ? {
          name: props.data.name,
          tags:
            props.data.spec.registry.repo.tags.map((t) => ({
              label: t,
              value: t,
            })) || [],
          buildClusterName: '',
          repository: repo,
          advanceOptions: isAdvanceOptions(props.data.spec.buildOptions),
          ...props.data.spec.buildOptions,
          ...(props.data.spec.buildOptions?.buildArgs || props),
        }
      : {
          name: '',
          tags: [],
          buildClusterName: '',
          advanceOptions: false,
          repository: repo,
          buildArgs: {},
          buildContexts: {},
          contextDir: '',
          dockerfilePath: '',
          dockerfileContent: '',
        },
    validationSchema: Yup.object({
      name: Yup.string().required(),
      buildClusterName: Yup.string().required(),
      tags: Yup.array()
        .required()
        .test('is-valid', 'Tags is required', (value) => {
          return value.length > 0;
        }),
      // buildArgs: Yup.object().when('advanceOptions', {
      //   is: true,
      //   then: (schema) =>
      //     schema
      //       .required()
      //       .test('is-valid', 'Build args is required', (value) => {
      //         return Object.keys(value).length > 0;
      //       }),
      // }),
      // buildContexts: Yup.object().when('advanceOptions', {
      //   is: true,
      //   then: (schema) =>
      //     schema
      //       .required()
      //       .test('is-valid', 'Build contexts is required', (value) => {
      //         return Object.keys(value).length > 0;
      //       }),
      // }),
    }),
    onSubmit: async (val) => {
      if (source) {
        try {
          if (!isUpdate) {
            const { errors: e } = await api.createBuild({
              build: {
                name: val.name,
                buildClusterName: val.buildClusterName,
                source: {
                  branch: source.branch!,
                  repository: source.repo!,
                  provider: source.provider!,
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
          } else {
            const { errors: e } = await api.updateBuild({
              crUpdateBuildId: props.data.id,
              build: {
                name: val.name,
                buildClusterName: val.buildClusterName,
                source: {
                  branch: source.branch!,
                  repository: source.repo!,
                  provider: source.provider!,
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
      }
    },
  });

  const getProviderLogo = (provider: string) => {
    const logoSize = 24;
    switch (provider) {
      case 'github':
        return <GithubLogoFill size={logoSize} />;
      case 'gitlab':
        return <GitlabLogoFill size={logoSize} />;
      default:
        return null;
    }
  };

  return (
    <Popup.Form onSubmit={handleSubmit}>
      <Popup.Content>
        <MultiStep.Root currentStep={currentStep}>
          <MultiStep.Step step={1}>
            <div className="p-xl !pt-0">
              <GitRepoSelector
                onImport={(val) => {
                  setSource({ ...val, branch: val.branch! });
                  onNext();
                }}
              />
            </div>
          </MultiStep.Step>
          <MultiStep.Step step={2}>
            <div className="flex flex-col gap-2xl">
              <div className="flex flex-col gap-xl rounded border border-border-default p-xl">
                <div className="flex flex-row gap-3xl items-center justify-between">
                  <div className="flex flex-row items-center gap-lg">
                    {getProviderLogo(source?.provider || '')}{' '}
                    <div className="bodyMd-medium">{source?.repo}</div>
                  </div>
                  <div className="flex flex-row items-center gap-lg">
                    <GitBranch size={16} />{' '}
                    <div className="bodyMd-medium">{source?.branch}</div>
                  </div>

                  <div className="self-end pt-lg">
                    <Button
                      content="Change"
                      variant="basic"
                      prefix={<PencilSimple />}
                      size="sm"
                      onClick={() => {
                        onPrevious();
                      }}
                    />
                  </div>
                </div>
              </div>
              <div className="flex flex-col gap-xl rounded border border-border-default p-xl">
                <div className="headingXl text-text-default">Target</div>
                <div className="flex flex-col gap-xl">
                  <TextInput
                    label="Name"
                    value={values.name}
                    onChange={handleChange('name')}
                    error={!!errors.name}
                    message={errors.name}
                  />
                  <Select
                    label="Tags"
                    size="md"
                    creatable
                    multiple
                    value={values.tags}
                    options={async () => values.tags}
                    onChange={(val) => {
                      handleChange('tags')(dummyEvent(val));
                    }}
                    error={!!errors.tags}
                    message={errors.tags}
                  />

                  <Select
                    label="Clusters"
                    size="md"
                    value={clusterData.find(
                      (cd) => cd.value === values.buildClusterName
                    )}
                    options={async () => clusterData}
                    onChange={(val) => {
                      handleChange('buildClusterName')(dummyEvent(val.value));
                    }}
                    error={!!errors.buildClusterName}
                    message={errors.buildClusterName}
                  />

                  <Checkbox
                    label="Advance options"
                    checked={values.advanceOptions}
                    onChange={(check) => {
                      handleChange('advanceOptions')(dummyEvent(!!check));
                    }}
                  />
                  {values.advanceOptions && (
                    <div className="flex flex-col gap-xl">
                      <KeyValuePair
                        label="Build args"
                        value={Object.entries(values.buildArgs || {}).map(
                          ([key, value]) => ({ key, value })
                        )}
                        onChange={(_, items) => {
                          handleChange('buildArgs')(dummyEvent(items));
                        }}
                        error={!!errors.buildArgs}
                        message={errors.buildArgs}
                      />
                      <KeyValuePair
                        label="Build contexts"
                        value={Object.entries(values.buildContexts || {}).map(
                          ([key, value]) => ({ key, value })
                        )}
                        onChange={(_, items) => {
                          handleChange('buildContexts')(dummyEvent(items));
                        }}
                        error={!!errors.buildContexts}
                        message={errors.buildContexts}
                      />
                      <TextInput
                        label="Context dir"
                        value={values.contextDir}
                        onChange={handleChange('contextDir')}
                      />
                      <TextInput
                        label="Docker file path"
                        value={values.dockerfilePath}
                        onChange={handleChange('dockerfilePath')}
                      />
                      <TextArea
                        label="Docker file content"
                        value={values.dockerfileContent}
                        onChange={handleChange('dockerfileContent')}
                        resize={false}
                        rows="6"
                      />
                      <BuildPlatforms
                        onChange={(data) => {
                          console.log(data);
                        }}
                      />
                    </div>
                  )}
                </div>
              </div>
            </div>
          </MultiStep.Step>
        </MultiStep.Root>
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button closable content="Cancel" variant="basic" />
        {currentStep > 1 && (
          <Popup.Button
            type="submit"
            content={!isUpdate ? 'Create' : 'Update'}
            variant="primary"
          />
        )}
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
