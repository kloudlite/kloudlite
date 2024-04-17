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
import ExtendedFilledTab from '~/console/components/extended-filled-tab';
import { constants } from '~/console/server/utils/constants';
import { Button } from '~/components/atoms/button';
import { useEffect, useState } from 'react';
import { toast } from '~/components/molecule/toast';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import {
  ArrowClockwise,
  GitMerge,
  PencilSimple,
} from '~/console/components/icons';
import { IEnvironmentContext } from '../_layout';
import { getImageTag } from './app-utils';
import BuildSelectionDialog from './app-build-selection-dialog';

const ExtraButton = ({
  onNew,
  onExisting,
}: {
  onNew: () => void;
  onExisting: () => void;
}) => {
  return (
    <ResourceExtraAction
      options={[
        {
          label: 'Connect new',
          icon: <PencilSimple size={16} />,
          type: 'item',
          key: 'new',
          onClick: onNew,
        },
        {
          label: 'Choose from existing',
          icon: <ArrowClockwise size={16} />,
          type: 'item',
          key: 'existing',
          onClick: onExisting,
        },
      ]}
    />
  );
};

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
    existingBuildId,
    setExistingBuildID,
  } = useAppState();

  const { project, environment, account } =
    useOutletContext<IEnvironmentContext>();
  const [projectName, envName, accountName] = [
    parseName(project),
    parseName(environment),
    parseName(account),
  ];

  const [openBuildSelection, setOpenBuildSelection] = useState(false);

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
        const imageTag = getImageTag({
          environment: envName,
          project: projectName,
          app: val.name,
        });
        if (!existingBuildId) {
          resetBuildData();
        }

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
          if (existingBuildId) {
            setPage(2);
            markPageAsCompleted(1);
            return;
          }
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

  useEffect(() => {
    setExistingBuildID(null);
  }, [values.source]);
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
            <GitDetailRaw
              provider={buildData.source.provider}
              repository={buildData.source.repository}
              branch={buildData.source.branch}
              extra={
                <div className="flex-1 flex justify-end">
                  <ExtraButton
                    onNew={() => {
                      // @ts-ignore
                      setBuildData(null);
                      setExistingBuildID(null);
                    }}
                    onExisting={() => {
                      setOpenBuildSelection(true);
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

          {values.imageMode === 'git' && !buildData?.name && (
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
          {values.imageMode === 'git' && !buildData?.name && (
            <AppBuildIntegration
              values={values}
              errors={errors}
              handleChange={handleChange}
            />
          )}
        </div>

        <BuildSelectionDialog
          open={openBuildSelection}
          setOpen={setOpenBuildSelection}
          onChange={(e) => {
            if (e.build?.id) {
              setExistingBuildID(e.build?.id);
              setBuildData(e.build);
            } else {
              toast.error('Something went wrong');
            }
          }}
        />
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
