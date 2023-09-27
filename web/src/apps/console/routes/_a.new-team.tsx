import { ArrowRight } from '@jengaicons/react';
import { useNavigate } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { toast } from '~/components/molecule/toast';
import { useMapper } from '~/components/utils';
import { useDataFromMatches } from '~/root/lib/client/hooks/use-custom-matches';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import { UserMe } from '~/root/lib/server/gql/saved-queries';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { useState } from 'react';
import { IdSelector } from '../components/id-selector';
import RawWrapper, { TitleBox } from '../components/raw-wrapper';
import { useConsoleApi } from '../server/gql/api-provider';
import { FadeIn } from './_.$account.$cluster.$project.$scope.$workspace.new-app/util';

const NewAccount = () => {
  const api = useConsoleApi();
  const navigate = useNavigate();
  const user = useDataFromMatches<UserMe>('user', {});
  const [isNameLoading, setIsNameLoading] = useState(false);
  const { values, handleSubmit, handleChange, errors, isLoading } = useForm({
    initialValues: {
      name: '',
      displayName: '',
    },
    validationSchema: Yup.object({
      name: Yup.string().required(),
      displayName: Yup.string().required(),
    }),
    onSubmit: async (v) => {
      if (isNameLoading) {
        toast.error('id is being checked, please wait');
        return;
      }
      try {
        const { errors: _errors } = await api.createAccount({
          account: {
            metadata: { name: v.name },
            spec: {},
            displayName: v.displayName,
            contactEmail: user.email,
          },
        });
        if (_errors) {
          throw _errors[0];
        }
        toast.success('account created');
        navigate(`/onboarding/${v.name}/invite-team-members`);
      } catch (err) {
        handleError(err);
      }
    },
  });

  const progressItems = [
    { label: 'Create Team', active: true, id: 1, completed: false },
    {
      label: 'Invite your Team Members',
      active: false,
      id: 2,
      completed: false,
    },
    {
      label: 'Add your Cloud Provider',
      active: false,
      id: 3,
      completed: false,
    },
    {
      label: 'Setup First Cluster',
      active: false,
      id: 4,
      completed: false,
    },
    {
      label: 'Create your project',
      active: false,
      id: 5,
      completed: false,
    },
  ];
  const items = useMapper(progressItems, (i) => {
    return {
      value: i.id,
      item: {
        ...i,
      },
    };
  });

  return (
    <RawWrapper
      title="Setup your Team!"
      subtitle="Simplify Collaboration and Enhance Productivity with Kloudlite
    teams."
      progressItems={items}
      rightChildren={
        <FadeIn onSubmit={handleSubmit}>
          <TitleBox
            title="Team name"
            subtitle="An assessment of the work, product, or performance."
          />
          <div className="flex flex-col">
            <TextInput
              size="lg"
              value={values.displayName}
              onChange={handleChange('displayName')}
              error={!!errors.displayName}
              message={errors.displayName}
              label="Name"
            />
            <IdSelector
              onLoad={(v) => setIsNameLoading(v)}
              name={values.displayName}
              onChange={(v) => handleChange('name')(dummyEvent(v))}
              resType="account"
              className="pt-2xl"
            />
          </div>
          <div className="flex flex-row justify-end">
            <Button
              variant="primary"
              content="Continue"
              suffix={<ArrowRight />}
              size="lg"
              loading={isLoading}
              type="submit"
            />
          </div>
        </FadeIn>
      }
    />
  );
};

export default NewAccount;
