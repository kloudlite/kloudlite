import { useAppState } from '~/iotconsole/page-components/app-states';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { parseName } from '~/iotconsole/server/r-utils/common';
import { FadeIn } from '~/iotconsole/page-components/util';
import { BottomNavigation } from '~/iotconsole/components/commons';
import { registryHost } from '~/lib/configs/base-url.cjs';
import { useOutletContext } from '@remix-run/react';
import RepoSelector from '~/iotconsole/page-components/app/components';
import { NameIdView } from '~/iotconsole/components/name-id-view';
import { IDeviceBlueprintContext } from '../_layout';

const AppDetail = () => {
  console.log('use', useAppState());
  const { app, setApp, setPage, markPageAsCompleted, activeContIndex } =
    useAppState();

  const { deviceblueprint, account } =
    useOutletContext<IDeviceBlueprintContext>();
  const [accountName] = [
    project.name,
    deviceblueprint.name,
    parseName(account),
  ];

  // const [openBuildSelection, setOpenBuildSelection] = useState(false);

  const { values, errors, handleChange, handleSubmit, isLoading, setValues } =
    useForm({
      initialValues: {
        name: parseName(app),
        displayName: app.displayName,
        isNameError: false,
        imageUrl: app.spec.containers[activeContIndex]?.image || '',
        manualRepo: '',
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
