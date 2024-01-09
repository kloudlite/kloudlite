import { ArrowRight } from '@jengaicons/react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { IdSelector } from '~/console/components/id-selector';
import { useAppState } from '~/console/page-components/app-states';
import { keyconstants } from '~/console/server/r-utils/key-constants';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { parseName } from '~/console/server/r-utils/common';
import { FadeIn } from '~/console/page-components/util';

const AppDetail = () => {
  const { app, setApp, setPage, markPageAsCompleted } = useAppState();

  const { values, errors, handleChange, handleSubmit, isLoading } = useForm({
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
            annotations: {
              ...(a.metadata?.annotations || {}),
              [keyconstants.description]: val.description || '',
            },
            name: val.name,
          },
          displayName: val.displayName,
        };
      });
      setPage('Compute');
      markPageAsCompleted('Application details');
    },
  });

  return (
    <FadeIn onSubmit={handleSubmit} className="py-3xl">
      <div className="bodyMd text-text-soft">
        The application streamlines project management through intuitive task
        tracking and collaboration tools.
      </div>
      <div className="flex flex-col">
        <div className="flex flex-col pb-3xl">
          <TextInput
            label="Application name"
            size="lg"
            value={values.displayName}
            onChange={handleChange('displayName')}
            error={!!errors.displayName}
            message={errors.displayName}
          />
          <IdSelector
            onChange={(v) => handleChange('name')(dummyEvent(v))}
            name={values.displayName}
            resType="app"
            className="pt-2xl"
          />
        </div>
        <TextInput
          error={!!errors.description}
          message={errors.description}
          label="Description"
          size="lg"
          value={values.description}
          onChange={handleChange('description')}
        />
      </div>
      <div className="flex flex-row gap-xl justify-end items-center">
        <Button
          loading={isLoading}
          type="submit"
          content="Save & Continue"
          suffix={<ArrowRight />}
          variant="primary"
          // onClick={next}
        />
      </div>
    </FadeIn>
  );
};

export default AppDetail;
