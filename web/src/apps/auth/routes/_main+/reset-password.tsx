import { Button } from '@kloudlite/design-system/atoms/button';
import { PasswordInput } from '@kloudlite/design-system/atoms/input';
import { Link, useLoaderData, useNavigate } from '@remix-run/react';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { toast } from '@kloudlite/design-system/molecule/toast';
import getQueries from '~/root/lib/server/helpers/get-queries';
import { cn } from '@kloudlite/design-system/utils';
import { redirect } from '@remix-run/node';
import { handleError } from '~/root/lib/utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import { useAuthApi } from '~/auth/server/gql/api-provider';
import { ArrowRight } from '@kloudlite/design-system/icons';
import Container from '~/auth/components/container';

const ForgetPassword = () => {
  const { token } = useLoaderData();
  const api = useAuthApi();

  const navigate = useNavigate();
  const { values, errors, handleChange, isLoading, handleSubmit } = useForm({
    initialValues: {
      password: '',
      c_password: '',
      token,
    },
    validationSchema: Yup.object({
      password: Yup.string().required(),
      c_password: Yup.string()
        .oneOf([Yup.ref('password'), ''], 'passwords must match')
        .required('confirm password is required'),
      token: Yup.string().required(),
    }),
    onSubmit: async (val) => {
      try {
        const { errors: e } = await api.resetPassword({
          password: val.password,
          token: val.token,
        });
        if (e) {
          throw e[0];
        }
        toast.success('password reset successfully done');
        navigate('/');
      } catch (err) {
        handleError(err);
      }
    },
  });
  return (
    <Container
      headerExtra={
        <Button
          variant="outline"
          content="Sign in"
          linkComponent={Link}
          to="/login"
        />
      }
    >
      <form
        className="flex flex-col gap-6xl md:w-[500px] px-3xl py-5xl md:px-9xl"
        onSubmit={handleSubmit}
      >
        <div className="flex flex-col gap-lg items-center text-center">
          <div className={cn('text-text-strong headingXl text-center')}>
            Reset password
          </div>
          <div className="bodyMd-medium text-text-soft">
            Your identity has been verified! Set your new password
          </div>
        </div>
        <div className="flex flex-col gap-3xl">
          <PasswordInput
            label="Password"
            size="lg"
            className="h-[48px]"
            value={values.password}
            error={!!errors.password}
            message={errors.password}
            onChange={handleChange('password')}
            placeholder="New password"
          />

          <PasswordInput
            label="Confirm Password"
            size="lg"
            className="h-[48px]"
            value={values.c_password}
            error={!!errors.c_password}
            message={errors.c_password}
            onChange={handleChange('c_password')}
            placeholder="Confirm password"
          />
          <Button
            size="lg"
            variant="primary"
            content={<span className="bodyLg-medium">Reset</span>}
            suffix={<ArrowRight />}
            block
            type="submit"
            loading={isLoading}
          />
        </div>
      </form>
    </Container>
  );
};

export const loader = async (ctx: IRemixCtx) => {
  const { token } = getQueries(ctx);
  if (!token) {
    return redirect('/reset-email-sent');
  }
  return { token };
};

export default ForgetPassword;
