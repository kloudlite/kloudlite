import { useAppState } from '~/console/page-components/app-states';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { parseName } from '~/console/server/r-utils/common';
import { FadeIn } from '~/console/page-components/util';
import { NameIdView } from '~/console/components/name-id-view';
import { BottomNavigation, GitDetail } from '~/console/components/commons';
import { registryHost } from '~/lib/configs/base-url.cjs';
import { useOutletContext } from '@remix-run/react';
import RepoSelector from '~/console/page-components/app/components';
import AppBuildIntegration from '~/console/page-components/app/app-build-integration';
import { keyconstants } from '~/console/server/r-utils/key-constants';
import ExtendedFilledTab from '~/console/components/extended-filled-tab';
import { constants } from '~/console/server/utils/constants';
import { IEnvironmentContext } from '../_layout';
import { getImageTag } from './app-utils';

const AppDetail = () => {
  const {
    app,
    setApp,
    setPage,
    setBuildData,
    buildData,
    resetBuildData,
    markPageAsCompleted,
    activeContIndex,
  } = useAppState();

  const { project, environment, account } =
    useOutletContext<IEnvironmentContext>();
  const [projectName, envName, accountName] = [
    parseName(project),
    parseName(environment),
    parseName(account),
  ];

  const { values, errors, handleChange, handleSubmit, isLoading, setValues } =
    useForm({
      initialValues: {
        name: parseName(app),
        displayName: app.displayName,
        isNameError: false,
        imageMode:
          app.metadata?.annotations?.[keyconstants.appImageMode] || 'default',
        imageUrl: app.spec.containers[activeContIndex]?.image || '',
        manualRepo: '',
        source: {
          branch: buildData?.source.branch,
          repository: buildData?.source.repository,
          provider: buildData?.source.provider,
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
            const regex = /[a-z0-9-/.]+[:][a-z0-9-.]+[a-z0-9]/;
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
        resetBuildData();

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
          const imageTag = getImageTag({
            environment: envName,
            project: projectName,
            app: val.name,
          });
          setBuildData({
            name: imageTag,
            buildClusterName: project.clusterName,
            source: {
              branch: val.source.branch,
              provider: val.source.provider === 'github' ? 'github' : 'gitlab',
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
                  name: constants.defaultAppRepoNameOnly,
                  tags: [imageTag],
                },
              },
              resource: {
                cpu: 500,
                memoryInMb: 1000,
              },
            },
          });
        }
        setPage(2);
        markPageAsCompleted(1);
      },
    });

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
      <div className="bodyMd text-text-soft">
        The application streamlines project management through intuitive task
        tracking and collaboration tools.
      </div>
      <div className="flex flex-col gap-5xl">
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
        <div className="flex flex-col gap-xl">
          <ExtendedFilledTab
            value={values.imageMode}
            onChange={(e) => {
              handleChange('imageMode')(dummyEvent(e));
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
          {buildData?.name && values.imageMode === 'git' && (
            <GitDetail
              provider={buildData.source.provider}
              repository={buildData.source.repository}
              branch={buildData.source.branch}
              onEdit={() => {
                resetBuildData();
              }}
            />
          )}
          {values.imageMode === 'git' && !buildData?.name && (
            <AppBuildIntegration
              values={values}
              errors={errors}
              handleChange={handleChange}
            />
          )}
        </div>
      </div>
      <BottomNavigation
        primaryButton={{
          loading: isLoading,
          type: 'submit',
          content: 'Save & Continue',
          variant: 'primary',
        }}
      />
    </FadeIn>
  );
};

export default AppDetail;
