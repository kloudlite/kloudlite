import { BrandLogo } from '~/components/branding/brand-logo.jsx';
import { Button } from '~/components/atoms/button';
import { ArrowRight } from '@jengaicons/react';
import { TextInput } from '~/components/atoms/input';
import { GoogleReCaptcha } from 'react-google-recaptcha-v3';
import { Link } from '@remix-run/react';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { toast } from '~/components/molecule/toast';
import { cn } from '~/components/utils';
import { handleError } from '~/root/lib/utils/common';
import { useAuthApi } from '~/auth/server/gql/api-provider';
import Container from '../../components/container';

const ForgetPassword = () => {
  const api = useAuthApi();
  const { values, errors, handleChange, isLoading, handleSubmit } = useForm({
    initialValues: {
      email: '',
    },
    validationSchema: Yup.object({
      email: Yup.string().required().email(),
    }),
    onSubmit: async (val) => {
      try {
        const { errors: e } = await api.requestResetPassword({
          email: val.email,
        });
        if (e) {
          throw e[0];
        }
        toast.success('reset link sent on your email address');
      } catch (err) {
        handleError(err);
      }
    },
  });
  return (
    <Container
      footer={{
        message: 'Remember password?',
        buttonText: 'Login',
        to: '/login',
      }}
    >
      <form
        className={cn(
          'flex flex-col items-center self-stretch justify-center '
        )}
        onSubmit={handleSubmit}
      >
        <div className="flex flex-col items-stretch justify-center gap-5xl md:w-[400px]">
          <BrandLogo darkBg={false} size={60} />
          <div className="flex flex-col items-stretch gap-5xl pb-5xl">
            <div className="flex flex-col gap-lg items-center md:px-7xl">
              <div className={cn('text-text-strong heading3xl text-center')}>
                Forgot password
              </div>
              <div className="text-text-soft bodySm text-center">
                Enter your registered email below to receive password reset
                instruction.
              </div>
            </div>
            <div className="flex flex-col items-stretch gap-3xl">
              <TextInput
                label="Email"
                placeholder="ex: john@company.com"
                size="lg"
                value={values.email}
                error={!!errors.email}
                message={errors.email}
                onChange={handleChange('email')}
              />
              <Button
                size="2xl"
                variant="primary"
                content={
                  <span className="bodyLg-medium">Send instructions</span>
                }
                suffix={<ArrowRight />}
                block
                type="submit"
                LinkComponent={Link}
                loading={isLoading}
              />
            </div>
          </div>
        </div>
        <GoogleReCaptcha onVerify={() => {}} />
      </form>
    </Container>
  );
};

export default ForgetPassword;
