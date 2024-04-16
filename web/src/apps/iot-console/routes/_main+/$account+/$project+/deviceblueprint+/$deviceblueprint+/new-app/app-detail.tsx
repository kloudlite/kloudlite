import { useAppState } from '~/iotconsole/page-components/app-states';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { parseName } from '~/iotconsole/server/r-utils/common';
import { FadeIn } from '~/iotconsole/page-components/util';
import { BottomNavigation } from '~/iotconsole/components/commons';
import { registryHost } from '~/lib/configs/base-url.cjs';
import { useOutletContext } from '@remix-run/react';
import RepoSelector from '~/iotconsole/page-components/app/components';
import { keyconstants } from '~/iotconsole/server/r-utils/key-constants';
import { useState } from 'react';
import ResourceExtraAction from '~/iotconsole/components/resource-extra-action';
import { ArrowClockwise, PencilSimple } from '~/iotconsole/components/icons';
import { NameIdView } from '~/iotconsole/components/name-id-view';
import { IDeviceBlueprintContext } from '../_layout';

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
    buildData,
    markPageAsCompleted,
    activeContIndex,
  } = useAppState();

  const { project, deviceblueprint, account } =
    useOutletContext<IDeviceBlueprintContext>();
  const [projectName, envName, accountName] = [
    project.name,
    deviceblueprint.name,
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
                  image: val.imageUrl || val.manualRepo,
                  name: 'container-0',
                },
              ],
            },
          };
        });

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
          resType="project"
          errors={errors.name}
          label="Application name"
          placeholder="Enter application name"
          handleChange={handleChange}
          nameErrorLabel="isNameError"
        />
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
