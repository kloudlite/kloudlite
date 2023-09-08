import { TextInput } from '~/components/atoms/input';
import { IdSelector } from '~/console/components/id-selector';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { keyconstants } from '~/console/server/r-utils/key-constants';
import { Button } from '~/components/atoms/button';
import { ArrowRight } from '@jengaicons/react';
import { useOutletContext } from '@remix-run/react';
import { useAppState } from './states';
import { FadeIn } from './util';
import { IWorkspaceContext } from '../_.$account.$cluster.$project.$scope.$workspace/route';

const AppDetail = () => {
  const { app, setApp, setPage } = useAppState();
  const { workspace } = useOutletContext<IWorkspaceContext>();

  const { values, errors, handleChange, handleSubmit, isLoading } = useForm({
    initialValues: {
      name: app.metadata.name,
      displayName: app.displayName,
      description: app.metadata.annotations?.[keyconstants.description] || '',
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
            namespace: workspace.spec?.targetNamespace,
            annotations: {
              [keyconstants.description]: val.description,
            },
          },
          displayName: val.name,
        };
      });
      setPage('compute');
    },
  });

  return (
    <FadeIn onSubmit={handleSubmit}>
      <div className="flex flex-col gap-lg">
        <div className="headingXl text-text-default">Application details</div>
        <div className="bodyMd text-text-soft">
          The application streamlines project management through intuitive task
          tracking and collaboration tools.
        </div>
      </div>
      <div className="flex flex-col gap-3xl">
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
        />
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
