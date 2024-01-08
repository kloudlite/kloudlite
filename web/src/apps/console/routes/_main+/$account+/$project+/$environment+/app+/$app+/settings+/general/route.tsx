import { CopySimple } from '@jengaicons/react';
import { useEffect } from 'react';
import { TextInput } from '~/components/atoms/input';
import { Box } from '~/console/components/common-console-components';
import { useAppState } from '~/console/page-components/app-states';
import { keyconstants } from '~/console/server/r-utils/key-constants';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { parseName } from '~/console/server/r-utils/common';
import Wrapper from '~/console/components/wrapper';
import { Button } from '~/components/atoms/button';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';

const SettingGeneral = () => {
  const { app, setApp } = useAppState();
  const { setPerformAction, hasChanges, loading } = useUnsavedChanges();

  const { values, errors, handleChange, submit, resetValues } = useForm({
    initialValues: {
      name: parseName(app),
      displayName: app.displayName,
      description: app.metadata?.annotations?.[keyconstants.description] || '',
    },
    validationSchema: Yup.object({
      name: Yup.string().required(),
      displayName: Yup.string().required(),
      description: Yup.string(),
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
              [keyconstants.description]: val.description,
            },
          },
          displayName: val.displayName,
        };
      });
    },
  });

  useEffect(() => {
    submit();
  }, [values]);

  useEffect(() => {
    if (!hasChanges) {
      resetValues();
    }
  }, [hasChanges]);

  return (
    <div>
      <Wrapper
        secondaryHeader={{
          title: 'General',
          action: hasChanges && (
            <div className="flex flex-row items-center gap-lg">
              <Button
                disabled={loading}
                variant="basic"
                content="Discard changes"
                onClick={() => setPerformAction('discard-changes')}
              />
              <Button
                disabled={loading}
                content={loading ? 'Committing changes.' : 'View changes'}
                loading={loading}
                onClick={() => setPerformAction('view-changes')}
              />
            </div>
          ),
        }}
      >
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
          <TextInput
            label="Description"
            error={!!errors.description}
            message={errors.description}
            value={values.description}
            onChange={handleChange('description')}
          />
        </Box>
      </Wrapper>
    </div>
  );
};
export default SettingGeneral;
