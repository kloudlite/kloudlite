import { CopySimple } from '@jengaicons/react';
import { useEffect } from 'react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { useOutletContext } from '@remix-run/react';
import { keyconstants } from '~/console/server/r-utils/key-constants';
import { useAppState } from '~/console/page-components/app-states';
import { IAppContext } from '../_.$account.$cluster.$project.$scope.$workspace.app.$app/route';

const SettingAdvance = () => {
  const { app, setApp } = useAppState();
  const { workspace } = useOutletContext<IAppContext>();

  const { values, errors, handleChange, submit } = useForm({
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
              ...(a.metadata.annotations || {}),
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

  return (
    <>
      <div className="rounded border border-border-default bg-surface-basic-default shadow-button flex flex-col">
        <div className="flex flex-col gap-3xl p-3xl">
          <div className="text-text-strong headingLg">Transfer</div>
          <div className="bodyMd text-text-default">
            Move your app to a different workspace seamlessly, avoiding any
            downtime or disruptions to workflows.
          </div>
        </div>
        <div className="bg-surface-basic-subdued p-3xl flex flex-row justify-end">
          <Button variant="basic" content="Transfer" />
        </div>
      </div>
      <div className="flex flex-col gap-3xl p-3xl rounded border border-border-critical bg-surface-basic-default shadow-button">
        <div className="text-text-strong headingLg">Delete Application</div>
        <div className="bodyMd text-text-default">
          Permanently remove your application and all of its contents from the
          “Lobster Early” project. This action is not reversible, so please
          continue with caution.
        </div>
        <Button content="Delete" variant="critical" />
      </div>
    </>
  );
};
export default SettingAdvance;
