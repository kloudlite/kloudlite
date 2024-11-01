import { Link } from '@remix-run/react';
import { RECAPTCHA_SITE_KEY } from '~/auth/consts';
import { useAuthApi } from '~/auth/server/gql/api-provider';
import { Button } from '@kloudlite/design-system/atoms/button';
import { TextInput } from '@kloudlite/design-system/atoms/input';
import { ArrowRight } from '@kloudlite/design-system/icons';
import { toast } from '@kloudlite/design-system/molecule/toast';
import { cn } from '@kloudlite/design-system/utils';
import grecaptcha from '~/root/lib/client/helpers/g-recaptcha';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
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
        const token = await grecaptcha.execute(RECAPTCHA_SITE_KEY, {
          action: 'login',
        });
        const { errors: e } = await api.requestResetPassword({
          email: val.email,
          // @ts-ignore
          captchaToken: token,
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
            Forgot password
          </div>
          <div className="bodyMd-medium text-text-soft">
            Enter your registered email below to receive password reset
            instruction
          </div>
        </div>
        <div className="flex flex-col items-stretch gap-3xl">
          <TextInput
            label="Email"
            placeholder="ex: john@company.com"
            size="lg"
            className="h-[48px]"
            value={values.email}
            error={!!errors.email}
            message={errors.email}
            onChange={handleChange('email')}
          />
          <Button
            size="lg"
            variant="primary"
            content={<span className="bodyLg-medium">Send instructions</span>}
            suffix={<ArrowRight />}
            block
            type="submit"
            linkComponent={Link}
            loading={isLoading}
          />
        </div>
        <div className="text-center">
          <span className="text-text-soft bodyLg">
            Remember password?{' '}
            <Button
              variant="plain"
              className="!inline-block align-bottom"
              content={
                <span className="bodyLg-underline text-text-strong">Login</span>
              }
              to="/login"
              linkComponent={Link}
            />
          </span>
        </div>
      </form>
    </Container>
  );
};

export default ForgetPassword;
