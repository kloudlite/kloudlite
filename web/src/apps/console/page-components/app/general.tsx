import { useAppState } from '~/console/page-components/app-states';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { parseName } from '~/console/server/r-utils/common';
import { FadeIn } from '~/console/page-components/util';
import { NameIdView } from '~/console/components/name-id-view';
import { BottomNavigation, GitDetailRaw } from '~/console/components/commons';
import { registryHost } from '~/lib/configs/base-url.cjs';
import { useOutletContext } from '@remix-run/react';
import RepoSelector from '~/console/page-components/app/components';
import AppBuildIntegration from '~/console/page-components/app/app-build-integration';
import { keyconstants } from '~/console/server/r-utils/key-constants';
import { IEnvironmentContext } from '~/console/routes/_main+/$account+/$project+/env+/$environment+/_layout';
import { TextInput } from '~/components/atoms/input';
import { useEffect, useState } from 'react';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';
import { IGIT_PROVIDERS } from '~/console/hooks/use-git';
import ExtendedFilledTab from '~/console/components/extended-filled-tab';
import { getImageTag } from '~/console/routes/_main+/$account+/$project+/env+/$environment+/new-app/app-utils';
import { constants } from '~/console/server/utils/constants';
import HandleBuild from '~/console/routes/_main+/$account+/repo+/$repo+/builds/handle-builds';
import { ArrowClockwise, PencilSimple } from '~/console/components/icons';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import appFun from '~/console/routes/_main+/$account+/$project+/env+/$environment+/new-app/app-pre-submit';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { toast } from '~/components/molecule/toast';

const ExtraButton = ({
  onEdit,
  onTrigger,
}: {
  onEdit: () => void;
  onTrigger: () => void;
}) => {
  return (
    <ResourceExtraAction
      options={[
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
      ]}
    />
  );
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
  } = useAppState();

  const { project, account, environment } =
    useOutletContext<IEnvironmentContext>();
  const { performAction } = useUnsavedChanges();

  const [projectName, envName, accountName] = [
    parseName(project),
    parseName(environment),
    parseName(account),
  ];
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
      imageMode: 'default',
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
    },
    validationSchema: Yup.object({
      name: Yup.string().required(),
      displayName: Yup.string().required(),
      imageUrl: Yup.string(),
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
      // setBuildData(buildData);

      const formBuildData = () => {
        const imageTag = getImageTag({
          environment: envName,
          project: projectName,
          app: val.name,
        });
        return {
          buildClusterName: project.clusterName || '',
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
        if (!project.clusterName) {
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
          <ExtendedFilledTab
            value={values.imageMode}
            onChange={(e) => {
              handleChange('imageMode')(dummyEvent(e));
              if (!app.ciBuildId && e === 'git') {
                setIsEdited(true);
              }
              resetValues({
                ...values,
                imageMode: e,
              });
              if (e === 'default') {
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
          />

          {values.imageMode === 'default' && (
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
          )}

          {buildData?.name && values.imageMode === 'git' && !isEdited && (
            <GitDetailRaw
              provider={buildData.source.provider}
              repository={buildData.source.repository}
              branch={buildData.source.branch}
              extra={
                <div className="flex-1 flex justify-end">
                  <ExtraButton
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
            />
          )}
          {values.imageMode === 'git' && (isEdited || !buildData?.name) && (
            <AppBuildIntegration
              values={values}
              errors={errors}
              handleChange={handleChange}
            />
          )}
        </div>
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
