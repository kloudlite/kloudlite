import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { toast } from '~/components/molecule/toast';
import { useNavigate } from '@remix-run/react';
import { useAPIClient } from '../server/utils/api-provider';

const NewAccount = () => {
  const api = useAPIClient();
  const navigate = useNavigate();
  const { values, handleSubmit, handleChange, errors, isLoading } = useForm({
    initialValues: {
      name: '',
      displayName: 'temp',
    },
    validationSchema: Yup.object({
      name: Yup.string().required(),
    }),
    onSubmit: async (v) => {
      try {
        const { errors: _errors } = await api.createAccount({
          name: v.name,
          displayName: v.displayName,
        });
        if (_errors) {
          throw _errors[0];
        }
        toast.success('account created');
        navigate('/accounts');
      } catch (err) {
        toast.error(err.message);
      }
    },
  });
  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col gap-2xl justify-center items-center p-12xl"
    >
      <span>Create Account</span>
      <TextInput
        value={values.name}
        onChange={handleChange('name')}
        error={errors.name}
        label="Name"
      />
      <Button loading={isLoading} type="submit" content="create account" />
    </form>
  );
};

export default NewAccount;
