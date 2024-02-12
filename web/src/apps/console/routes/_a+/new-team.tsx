import { ArrowRight } from '@jengaicons/react';
import { useNavigate } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import { toast } from '~/components/molecule/toast';
import { useDataFromMatches } from '~/root/lib/client/hooks/use-custom-matches';
import useForm from '~/root/lib/client/hooks/use-form';
import { UserMe } from '~/root/lib/server/gql/saved-queries';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import ProgressWrapper from '~/console/components/progress-wrapper';
import { NameIdView } from '~/console/components/name-id-view';

const NewAccount = () => {
  const api = useConsoleApi();
  const navigate = useNavigate();
  const user = useDataFromMatches<UserMe>('user', {});
  const { values, handleChange, errors, isLoading, handleSubmit } = useForm({
    initialValues: {
      name: '',
      displayName: '',
      isNameError: false,
    },
    validationSchema: Yup.object({
      name: Yup.string().required(),
      displayName: Yup.string().required(),
    }),
    onSubmit: async (v) => {
      try {
        const { errors: _errors } = await api.createAccount({
          account: {
            metadata: { name: v.name },
            displayName: v.displayName,
            contactEmail: user.email,
          },
        });
        if (_errors) {
          throw _errors[0];
        }
        toast.success('account created');
        navigate(`/onboarding/${v.name}/new-cloud-provider`);
      } catch (err) {
        handleError(err);
      }
    },
  });

  const progressItems = [
    {
      label: 'Create Team',
      active: true,
      completed: false,
      children: (
        <form className="py-3xl flex flex-col gap-3xl" onSubmit={handleSubmit}>
          <NameIdView
            label="Name"
            resType="account"
            name={values.name}
            displayName={values.displayName}
            errors={errors.name}
            handleChange={handleChange}
            nameErrorLabel="isNameError"
          />
          <div className="flex flex-row gap-xl justify-start">
            <Button
              variant="primary"
              content="Continue"
              suffix={<ArrowRight />}
              size="md"
              loading={isLoading}
              type="submit"
            />
          </div>
        </form>
      ),
    },
    {
      label: 'Add your Cloud Provider',
      active: false,
      completed: false,
    },
    {
      label: 'Validate Cloud Provider',
      active: false,
      completed: false,
    },
    {
      label: 'Setup First Cluster',
      active: false,
      completed: false,
    },
  ];

  return (
    <ProgressWrapper
      title="Setup your account!"
      subTitle="Simplify Collaboration and Enhance Productivity with Kloudlite
  teams"
      progressItems={{
        items: progressItems,
      }}
    />
  );
};

export default NewAccount;
