import { useAppState } from '~/console/page-components/app-states';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { parseName } from '~/console/server/r-utils/common';
import { FadeIn } from '~/console/page-components/util';
import { NameIdView } from '~/console/components/name-id-view';
import { BottomNavigation } from '~/console/components/commons';
import { registryHost } from '~/lib/configs/base-url.cjs';
import { useParams } from '@remix-run/react';
import RepoSelector from '~/console/page-components/app/components';

const AppDetail = () => {
  const { app, setApp, setPage, markPageAsCompleted, activeContIndex } =
    useAppState();

  const { account } = useParams();

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
        manualRepo: Yup.string().when(['imageUrl'], ([imageUrl], schema) => {
          const regex = /^(\w+):(\w+)$/;

          if (!imageUrl) {
            return schema.required().matches(regex, 'Invalid image format');
          }
          return schema;
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
                  ...(a.spec.containers?.[0] || {}),
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
      <div className="flex flex-col gap-3xl">
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
        <RepoSelector
          tag={values.imageUrl.split(':')[1]}
          repo={
            values.imageUrl
              .replace(`${registryHost}/${account}/`, '')
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
              dummyEvent(`${registryHost}/${account}/${repo}:${tag}`)
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
