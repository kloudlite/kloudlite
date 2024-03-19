import { CopySimple } from '@jengaicons/react';
import { useEffect } from 'react';
import { TextInput } from '~/components/atoms/input';
import { Box } from '~/console/components/common-console-components';
import { useAppState } from '~/console/page-components/app-states';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { parseName } from '~/console/server/r-utils/common';
import AppWrapper from '~/console/page-components/app/app-wrapper';
import RepoSelector from '~/console/page-components/app/components';
import { registryHost } from '~/root/lib/configs/base-url.cjs';
import { useParams } from '@remix-run/react';

const SettingGeneral = () => {
  const { app, setApp, activeContIndex } = useAppState();

  const { account } = useParams();

  const { values, errors, handleChange, submit, setValues } = useForm({
    initialValues: {
      name: parseName(app),
      displayName: app.displayName,
      imageUrl: app.spec.containers[activeContIndex]?.image || '',
      manualRepo: '',
    },
    validationSchema: Yup.object({
      name: Yup.string().required(),
      displayName: Yup.string().required(),
      imageUrl: Yup.string(),
      manualRepo: Yup.string().when(['imageUrl'], ([imageUrl], schema, c) => {
        const regex = /^(\w+):(\w+)$/;
        if (!imageUrl) {
          if (!c.value) {
            return schema.required('Image is required');
          }
          return schema.matches(regex, 'Invalid image format');
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
            name: val.name,
            annotations: {
              ...(a.metadata?.annotations || {}),
            },
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
    },
  });

  useEffect(() => {
    submit();
  }, [values]);

  return (
    <AppWrapper title="General">
      <Box title="Application detail">
        <div className="flex flex-row items-center gap-3xl">
          <div className="flex-1">
            <TextInput
              label="Application name"
              error={!!errors.displayName}
              message={errors.displayName}
              onChange={handleChange('displayName')}
              value={values.displayName}
            />
          </div>
          <div className="flex-1">
            <TextInput
              label="Application ID"
              value={values.name}
              suffixIcon={<CopySimple />}
              disabled
            />
          </div>
        </div>
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
      </Box>
    </AppWrapper>
  );
};
export default SettingGeneral;
