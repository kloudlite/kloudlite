import {
  Envelope,
  GithubLogoFill,
  GitlabLogoFill,
  GoogleLogo,
} from '@jengaicons/react';
import {
  Link,
  useNavigate,
  useOutletContext,
  useSearchParams,
} from '@remix-run/react';
import { RECAPTCHA_SITE_KEY, mainUrl } from '~/auth/consts';
import { Button } from '~/components/atoms/button.jsx';
import { PasswordInput, TextInput } from '~/components/atoms/input.jsx';
import { ArrowLeft, ArrowRight } from '~/components/icons';
import { toast } from '~/components/molecule/toast';
import { cn } from '~/components/utils';
import grecaptcha from '~/root/lib/client/helpers/g-recaptcha';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import Container from '../../components/container';
import { IProviderContext } from './_layout';

const CustomGoogleIcon = (props: any) => {
  return <GoogleLogo {...props} weight={4} />;
};

const SignUpWithEmail = () => {
  const api = useAPIClient();
  const navigate = useNavigate();
  const { values, errors, handleChange, handleSubmit, isLoading } = useForm({
    initialValues: {
      name: '',
      email: '',
      password: '',
      c_password: '',
    },
    validationSchema: Yup.object({
      email: Yup.string().required().email(),
      name: Yup.string().trim().required(),
      password: Yup.string().trim().required(),
      c_password: Yup.string()
        .oneOf([Yup.ref('password'), ''], 'passwords must match')
        .required('confirm password is required'),
    }),
    onSubmit: async (v) => {
      try {
        const token = await grecaptcha.execute(RECAPTCHA_SITE_KEY, {
          action: 'login',
        });
        const { errors: _errors } = await api.signUpWithEmail({
          email: v.email,
          name: v.name,
          password: v.password,
          captchaToken: token,
        });
        if (_errors) {
          throw _errors[0];
        }
        toast.success('Signed up successfully');
        navigate('/');
      } catch (err) {
        handleError(err);
      }
    },
  });

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col items-stretch gap-3xl"
    >
      <div className="flex flex-col items-stretch gap-3xl">
        <TextInput
          name="name"
          value={values.name}
          error={!!errors.name}
          message={errors.name}
          onChange={handleChange('name')}
          label="Name"
          placeholder="Full name"
          size="lg"
          className="h-[48px]"
        />
        <TextInput
          name="email"
          value={values.email}
          error={!!errors.email}
          message={errors.email}
          onChange={handleChange('email')}
          label="Email"
          placeholder="ex: john@company.com"
          size="lg"
          className="h-[48px]"
        />
        <PasswordInput
          name="password"
          value={values.password}
          error={!!errors.password}
          onChange={handleChange('password')}
          label="Password"
          placeholder="XXXXXX"
          size="lg"
          message={errors.password}
          className="h-[48px]"
        />

        <PasswordInput
          value={values.c_password}
          error={!!errors.c_password}
          onChange={handleChange('c_password')}
          label="Confirm Password"
          placeholder="XXXXXX"
          size="lg"
          message={errors.c_password}
          className="h-[48px]"
        />
      </div>
      <Button
        loading={isLoading}
        size="lg"
        variant="primary"
        content={<span className="bodyLg-medium">Continue with email</span>}
        suffix={<ArrowRight />}
        block
        type="submit"
      />
    </form>
  );
};

const Signup = () => {
  const { githubLoginUrl, gitlabLoginUrl, googleLoginUrl } =
    useOutletContext<IProviderContext>();
  const [searchParams, _setSearchParams] = useSearchParams();
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
      <div className="flex flex-col gap-6xl md:w-[500px] px-3xl py-5xl md:px-9xl">
        <div className="flex flex-col gap-lg items-center text-center">
          <div className={cn('text-text-strong headingXl text-center')}>
            Create your Kloudlite.io account
          </div>
          <div className="bodyMd-medium text-text-soft">
            Get started for free. No credit card required.
          </div>
        </div>
        <div className="flex flex-col gap-3xl">
          <div className="flex flex-col items-stretch">
            {searchParams.get('mode') === 'email' ? (
              <SignUpWithEmail />
            ) : (
              <div className="flex flex-col items-stretch gap-3xl">
                <Button
                  size="lg"
                  variant="tertiary"
                  content={
                    <span className="bodyLg-medium">Continue with GitHub</span>
                  }
                  prefix={<GithubLogoFill />}
                  to={githubLoginUrl}
                  disabled={!githubLoginUrl}
                  block
                  linkComponent={Link}
                />
                <Button
                  size="lg"
                  variant="purple"
                  content={
                    <span className="bodyLg-medium">Continue with GitLab</span>
                  }
                  prefix={<GitlabLogoFill />}
                  to={gitlabLoginUrl}
                  disabled={!gitlabLoginUrl}
                  block
                  linkComponent={Link}
                />
                <Button
                  size="lg"
                  variant="primary"
                  content={
                    <span className="bodyLg-medium">Continue with Google</span>
                  }
                  prefix={<CustomGoogleIcon />}
                  to={googleLoginUrl}
                  disabled={!googleLoginUrl}
                  block
                  linkComponent={Link}
                />
              </div>
            )}
          </div>
          {searchParams.get('mode') === 'email' ? (
            <Button
              size="lg"
              variant="plain"
              content={
                <span className="bodyLg-medium">Other sign up options</span>
              }
              prefix={<ArrowLeft />}
              to="/signup"
              block
              linkComponent={Link}
            />
          ) : (
            <Button
              size="lg"
              variant="outline"
              content={
                <span className="bodyLg-medium">Continue with email</span>
              }
              prefix={<Envelope />}
              to="/signup/?mode=email"
              block
              linkComponent={Link}
            />
          )}
        </div>
        <div className="inline text-center">
          <span className="text-text-soft bodyLg">
            By continuing, you agree Kloudlite’s
          </span>
          <br />
          <span>
            <Button
              to={`${mainUrl}/legal/terms-of-services`}
              linkComponent={Link}
              className="!inline-block align-bottom"
              variant="plain"
              content={
                <span className="bodyLg-underline text-text-strong">
                  Terms of Service
                </span>
              }
            />
            <span> and </span>
            <Button
              to={`${mainUrl}/legal/privacy-policy`}
              linkComponent={Link}
              className="!inline-block align-bottom"
              variant="plain"
              content={
                <span className="bodyLg-underline text-text-strong">
                  Privacy Policy
                </span>
              }
            />
            .
          </span>
        </div>
      </div>
    </Container>
  );
};

export default Signup;
