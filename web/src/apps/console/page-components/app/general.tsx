import { useAppState } from '~/console/page-components/app-states';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { parseName } from '~/console/server/r-utils/common';
import { FadeIn } from '~/console/page-components/util';
import { NameIdView } from '~/console/components/name-id-view';
import { BottomNavigation, GitDetailRaw } from '~/console/components/commons';
// import { registryHost } from '~/lib/configs/base-url.cjs';
import { useOutletContext } from '@remix-run/react';
// import RepoSelector from '~/console/page-components/app/components';
import AppBuildIntegration from '~/console/page-components/app/app-build-integration';
import { TextInput } from '~/components/atoms/input';
import { useEffect, useState } from 'react';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';
import { IGIT_PROVIDERS } from '~/console/hooks/use-git';
// import ExtendedFilledTab from '~/console/components/extended-filled-tab';
import { constants } from '~/console/server/utils/constants';
import HandleBuild from '~/console/routes/_main+/$account+/repo+/$repo+/builds/handle-builds';
import {
  ArrowClockwise,
  GitMerge,
  PencilSimple,
} from '~/console/components/icons';
import ResourceExtraAction, {
  IResourceExtraItem,
} from '~/console/components/resource-extra-action';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { toast } from '~/components/molecule/toast';
import { Button } from '~/components/atoms/button';
import { keyconstants } from '~/console/server/r-utils/key-constants';
import { IEnvironmentContext } from '~/console/routes/_main+/$account+/env+/$environment+/_layout';
import { getImageTag } from '~/console/routes/_main+/$account+/env+/$environment+/new-app/app-utils';
import appFun from '~/console/routes/_main+/$account+/env+/$environment+/new-app/app-pre-submit';
import BuildSelectionDialog from '~/console/routes/_main+/$account+/env+/$environment+/new-app/app-build-selection-dialog';
import { Checkbox } from '~/components/atoms/checkbox';

const ExtraButton = ({
  onNew,
  onEdit,
  onTrigger,
  isExistingBuild,
}: {
  onNew: () => void;
  onEdit: () => void;
  onTrigger: () => void;
  isExistingBuild: boolean;
}) => {
  let options: IResourceExtraItem[] = [];

  if (isExistingBuild) {
    options = [
      {
        label: 'Connect new',
        icon: <PencilSimple size={16} />,
        type: 'item',
        key: 'new',
        onClick: onNew,
      },
      ...options,
    ];
  } else {
    options = [
      {
        label: 'Edit',
        icon: <PencilSimple size={16} />,
        type: 'item',
        key: 'edit',
        onClick: onEdit,
      },
      {
        label: 'Trigger',
        icon: <ArrowClockwise size={16} />,
        type: 'item',
        key: 'trigger',
        onClick: onTrigger,
      },
      ...options,
    ];
  }
  return <ResourceExtraAction options={options} />;
};

const AppGeneral = ({ mode = 'new' }: { mode: 'edit' | 'new' }) => {
  const {
    app,
    readOnlyApp,
    setApp,
    setPage,
    setBuildData,
    buildData,
    markPageAsCompleted,
    activeContIndex,
    setExistingBuildID,
    existingBuildId,
    setContainer,
  } = useAppState();

  const { account, environment } = useOutletContext<IEnvironmentContext>();
  const { performAction } = useUnsavedChanges();

  const [envName, accountName] = [parseName(environment), parseName(account)];

  const [openBuildSelection, setOpenBuildSelection] = useState(false);

  // only for edit mode
  const [isEdited, setIsEdited] = useState(!app.ciBuildId);
  const [showBuildEdit, setShowBuildEdit] = useState(false);

  const api = useConsoleApi();

  const {
    values,
    errors,
    handleChange,
    handleSubmit,
    isLoading,
    setValues,
    submit,
    resetValues,
  } = useForm({
    initialValues: {
      name: parseName(app),
      displayName: app.displayName,
      isNameError: false,
      imageMode:
        readOnlyApp.metadata?.annotations[keyconstants.appImageMode] ||
        'default',
      imageUrl: readOnlyApp?.spec.containers[activeContIndex]?.image || '',
      manualRepo: '',
      source: {
        branch: readOnlyApp?.build?.source.branch,
        repository: readOnlyApp?.build?.source.repository,
        provider: readOnlyApp?.build?.source.provider,
      },
      advanceOptions: false,
      buildArgs: {},
      buildContexts: {},
      contextDir: '',
      dockerfilePath: '',
      dockerfileContent: '',
      isGitLoading: false,

      imagePullPolicy:
        readOnlyApp?.spec.containers[activeContIndex]?.imagePullPolicy ||
        'Always',
    },
    validationSchema: Yup.object({
      name: Yup.string().required(),
      displayName: Yup.string().required(),
      imageUrl: Yup.string().matches(
        constants.dockerImageFormatRegex,
        'Invalid image format'
      ),
      manualRepo: Yup.string().when(
        ['imageUrl', 'imageMode'],
        ([imageUrl, imageMode], schema) => {
          const regex = /^(\w+):(\w+)$/;
          if (imageMode === 'git') {
            return schema;
          }
          if (!imageUrl) {
            return schema.required().matches(regex, 'Invalid image format');
          }
          return schema;
        }
      ),
      imageMode: Yup.string().required(),
      source: Yup.object()
        .shape({})
        .test('is-empty', 'Branch is required.', (v, c) => {
          // @ts-ignoredfgdfg
          if (!v?.branch && c.parent.imageMode === 'git') {
            return false;
          }
          return true;
        }),
    }),

    onSubmit: async (val) => {
      const imageTag = getImageTag({
        environment: envName,
        app: val.name,
      });

      const formBuildData = () => {
        return {
          buildClusterName: environment.clusterName || '',
          name: imageTag,
          source: {
            branch: val.source.branch!,
            provider: (val.source.provider === 'github'
              ? 'github'
              : 'gitlab') as IGIT_PROVIDERS,
            repository: val.source.repository!,
          },
          spec: {
            ...{
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
                  : {
                      buildOptions: null,
                    }),
              },
            },
            registry: {
              repo: {
                name: constants.defaultAppRepoNameOnly,
                tags: [imageTag],
              },
            },
            resource: {
              cpu: 500,
              memoryInMb: 1000,
            },
          },
        };
      };

      setApp((a) => {
        return {
          ...a,
          metadata: {
            ...a.metadata,
            annotations: {
              ...(a.metadata?.annotations || {}),
              [keyconstants.appImageMode]: val.imageMode,
            },
            name: val.name,
          },
          displayName: val.displayName,
          spec: {
            ...a.spec,
            containers: [
              {
                ...a.spec.containers?.[0],
                image: val.imageUrl || val.manualRepo,
                name: 'container-0',
              },
            ],
          },
        };
      });

      if (val.imageMode === 'git') {
        if (!environment.clusterName) {
          throw new Error('Cluster name is required');
        }
        if (
          !val.source.provider ||
          !val.source.branch ||
          !val.source.repository
        ) {
          throw new Error('Source is required');
        }
        if (isEdited) {
          // @ts-ignore
          setBuildData(formBuildData());
        } else {
          // @ts-ignore
          setBuildData(buildData);
        }
      }
      setPage(2);
      markPageAsCompleted(1);
    },
  });

  /** ---- Only for edit mode in settings ----* */
  useEffect(() => {
    if (mode === 'edit') {
      submit();
    }
  }, [values, mode]);

  useEffect(() => {
    if (performAction === 'discard-changes') {
      if (app.ciBuildId) {
        setIsEdited(false);
      }
      resetValues();
      // @ts-ignore
      setBuildData(readOnlyApp?.build);
    } else if (performAction === 'init') {
      setIsEdited(false);
    }
  }, [performAction]);

  useEffect(() => {
    resetValues();
  }, [readOnlyApp]);

  useEffect(() => {
    if (
      values.imagePullPolicy !==
      readOnlyApp.spec.containers[activeContIndex].imagePullPolicy
    ) {
      setContainer((s) => {
        return {
          ...s,
          imagePullPolicy: values.imagePullPolicy,
        };
      });
    }
  }, [values.imagePullPolicy]);

  return (
    <FadeIn
      onSubmit={(e) => {
        if (!values.isNameError) {
          handleSubmit(e);
        } else {
          e.preventDefault();
        }
      }}
    >
      {mode === 'new' && (
        <div className="bodyMd text-text-soft">
          The application streamlines project management through intuitive task
          tracking and collaboration tools.
        </div>
      )}
      <div className="flex flex-col gap-5xl">
        {mode === 'new' ? (
          <NameIdView
            displayName={values.displayName}
            name={values.name}
            resType="app"
            errors={errors.name}
            label="Application name"
            placeholder="Enter application name"
            handleChange={handleChange}
            nameErrorLabel="isNameError"
          />
        ) : (
          <TextInput
            value={values.displayName}
            label="Name"
            size="lg"
            onChange={handleChange('displayName')}
          />
        )}
        <div className="flex flex-col gap-xl">
          {/* <ExtendedFilledTab
            value={values.imageMode}
            onChange={(e) => {
              handleChange('imageMode')(dummyEvent(e));
              if (!app.ciBuildId && e === 'git') {
                if (!existingBuildId) {
                  setIsEdited(true);
                } else {
                  setIsEdited(false);
                }
              }
              // resetValues({
              //   ...values,
              //   imageMode: e,
              // });
              if (e === 'default' && !existingBuildId) {
                // @ts-ignore
                setBuildData(readOnlyApp?.build);
              }
            }}
            items={[
              { label: 'Container repo', value: 'default' },
              {
                label: 'Git repo',
                value: 'git',
              },
            ]}
            size="sm"
          /> */}

          <TextInput
            size="lg"
            label="Image name"
            placeholder="Enter Image name"
            value={values.imageUrl}
            onChange={handleChange('imageUrl')}
            error={!!errors.imageUrl}
            message={errors.imageUrl}
          />

          {/* {values.imageMode === 'default' && (
            <RepoSelector
              tag={values.imageUrl.split(':')[1]}
              repo={
                values.imageUrl
                  .replace(`${registryHost}/${accountName}/`, '')
                  .split(':')[0]
              }
              onClear={() => {
                setValues((v) => {
                  return {
                    ...v,
                    imageUrl: '',
                    manualRepo: '',
                  };
                });
              }}
              textValue={values.manualRepo}
              onTextChanged={(e) => {
                handleChange('manualRepo')(e);
                handleChange('imageUrl')(dummyEvent(''));
              }}
              onValueChange={({ repo, tag }) => {
                handleChange('imageUrl')(
                  dummyEvent(`${registryHost}/${accountName}/${repo}:${tag}`)
                );
              }}
              error={errors.manualRepo}
            />
          )} */}

          {buildData?.name && values.imageMode === 'git' && !isEdited && (
            <GitDetailRaw
              provider={buildData.source.provider}
              repository={buildData.source.repository}
              branch={buildData.source.branch}
              extra={
                <div className="flex-1 flex justify-end">
                  <ExtraButton
                    isExistingBuild={!!existingBuildId}
                    onNew={() => {
                      setIsEdited(true);
                      setExistingBuildID(null);
                    }}
                    onEdit={() => {
                      setShowBuildEdit(true);
                    }}
                    onTrigger={async () => {
                      if (readOnlyApp.ciBuildId) {
                        const res = await appFun.triggerBuild({
                          api,
                          buildId: readOnlyApp.ciBuildId,
                        });
                        if (res) {
                          toast.info('Build triggered successfully');
                        } else {
                          toast.info('Build trigger failed');
                        }
                      }
                    }}
                  />
                </div>
              }
            >
              {existingBuildId && (
                <div className="text-text-soft bodySm pt-lg flex flex-col gap-md">
                  <div className="bodyMd-medium text-text-default">Build</div>
                  <div className="flex flex-row items-center gap-xl">
                    <GitMerge size={16} />
                    <span>{buildData.name}</span>
                  </div>
                </div>
              )}
            </GitDetailRaw>
          )}

          {values.imageMode === 'git' && (isEdited || !buildData?.name) && (
            <div className="flex flex-col gap-lg items-center pt-lg">
              <Button
                content="Choose from existing builds"
                variant="primary-outline"
                onClick={() => {
                  setOpenBuildSelection(true);
                }}
              />
              <span className="pl-3xl text-text-soft">or</span>
            </div>
          )}

          {values.imageMode === 'git' && (isEdited || !buildData?.name) && (
            <AppBuildIntegration
              values={values}
              errors={errors}
              handleChange={handleChange}
            />
          )}

          <BuildSelectionDialog
            open={openBuildSelection}
            setOpen={setOpenBuildSelection}
            onChange={(e) => {
              if (e.build?.id) {
                setExistingBuildID(e.build?.id);
                setBuildData({
                  buildClusterName: e.build.buildClusterName,
                  name: e.build.name,
                  source: e.build.source,
                  spec: {
                    registry: e.build.spec.registry,
                    resource: e.build.spec.resource,
                  },
                });
                setIsEdited(false);
              } else {
                toast.error('Something went wrong');
              }
            }}
          />
        </div>

        <Checkbox
          label="Always pull image on restart"
          checked={values.imagePullPolicy === 'Always'}
          onChange={(val) => {
            const imagePullPolicy = val ? 'Always' : 'IfNotPresent';
            handleChange('imagePullPolicy')(dummyEvent(imagePullPolicy));
          }}
        />
      </div>
      {mode === 'new' && (
        <BottomNavigation
          primaryButton={{
            loading: isLoading,
            type: 'submit',
            content: 'Save & Continue',
            variant: 'primary',
          }}
        />
      )}
      <HandleBuild
        {...{
          isUpdate: true,
          data: { mode: 'app', ...readOnlyApp.build! },
          visible: showBuildEdit,
          setVisible: () => setShowBuildEdit(false),
        }}
      />
    </FadeIn>
  );
};

export default AppGeneral;
