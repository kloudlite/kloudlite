import { BrandLogo } from '~/components/branding/brand-logo';
import { Button } from '~/components/atoms/button';
import { PasswordInput } from '~/components/atoms/input';
import { GoogleReCaptcha } from 'react-google-recaptcha-v3';
import { Link, useLoaderData, useNavigate } from '@remix-run/react';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { toast } from '~/components/molecule/toast';
import getQueries from '~/root/lib/server/helpers/get-queries';
import { cn } from '~/components/utils';
import { redirect } from '@remix-run/node';
import { handleError } from '~/root/lib/utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import { useAuthApi } from '~/auth/server/gql/api-provider';
import { ArrowRight } from '~/components/icons';

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
    <div className={cn('flex flex-col items-center justify-center h-full')}>
      <form
        className={cn(
          'flex flex-1 flex-col items-center self-stretch justify-center px-3xl pb-5xl'
        )}
        onSubmit={handleSubmit}
      >
        <div className="flex flex-col items-stretch justify-center gap-5xl md:w-[400px]">
          <BrandLogo darkBg={false} size={60} />
          <div className="flex flex-col items-stretch gap-5xl pb-5xl">
            <div className="flex flex-col gap-lg items-center md:px-7xl">
              <div className={cn('text-text-strong heading3xl text-center')}>
                Reset password
              </div>
              <div className="text-text-soft bodySm text-center">
                Please provide the new password of your choice.
              </div>
            </div>
            <div className="flex flex-col items-stretch gap-3xl">
              <PasswordInput
                label="Password"
                size="lg"
                value={values.password}
                error={!!errors.password}
                message={errors.password}
                onChange={handleChange('password')}
              />

              <PasswordInput
                label="Confirm Password"
                size="lg"
                value={values.c_password}
                error={!!errors.c_password}
                message={errors.c_password}
                onChange={handleChange('c_password')}
              />
              <Button
                size="2xl"
                variant="primary"
                content={<span className="bodyLg-medium">Reset</span>}
                suffix={<ArrowRight />}
                block
                type="submit"
                loading={isLoading}
              />
            </div>
          </div>
        </div>
        <GoogleReCaptcha onVerify={() => {}} />
      </form>
      <div className="py-5xl px-3xl flex flex-row items-center justify-center self-stretch border-t border-border-default sticky bottom-0 bg-surface-basic-default">
        <div className="bodyMd text-text-default">Remember password?</div>
        <Button
          content="Login"
          variant="primary-plain"
          size="md"
          to="/login"
          linkComponent={Link}
        />
      </div>
    </div>
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
