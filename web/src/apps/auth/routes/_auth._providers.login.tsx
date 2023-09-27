import classNames from 'classnames';
import { Button } from '~/components/atoms/button';
import {
  ArrowRight,
  Envelope,
  EnvelopeFill,
  GithubLogoFill,
  GitlabLogoFill,
  GoogleLogo,
} from '@jengaicons/react';
import { useSearchParams, Link, useOutletContext } from '@remix-run/react';
import { PasswordInput, TextInput } from '~/components/atoms/input';
import { BrandLogo } from '~/components/branding/brand-logo.jsx';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { handleError } from '~/root/lib/utils/common';
import { toast } from '~/components/molecule/toast';
import Container from '../components/container';
import { useAuthApi } from '../server/gql/api-provider';
import { IProviderContext } from './_auth._providers';

const CustomGoogleIcon = (props: any) => {
  return <GoogleLogo {...props} weight={4} />;
};

const LoginWithEmail = () => {
  const api = useAuthApi();

  const reloadPage = useReload();
  const { values, errors, handleChange, handleSubmit, isLoading } = useForm({
    initialValues: {
      email: '',
      password: '',
    },
    validationSchema: Yup.object({
      email: Yup.string().required().email(),
      password: Yup.string().trim().required(),
    }),
    onSubmit: async (v) => {
      try {
        const { errors: _errors } = await api.login({
          email: v.email,
          password: v.password,
        });
        if (_errors) {
          throw _errors[0];
        }
        toast.success('logged in success fully');
        reloadPage();
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
      <TextInput
        value={values.email}
        error={!!errors.email}
        message={errors.email}
        onChange={handleChange('email')}
        label="Email"
        placeholder="ex: john@company.com"
        size="lg"
      />
      <PasswordInput
        value={values.password}
        error={!!errors.password}
        message={errors.ram}
        onChange={handleChange('password')}
        label="Password"
        placeholder="XXXXXX"
        size="lg"
        extra={
          <Button
            size="md"
            variant="primary-plain"
            content="Forgot password"
            to="/forgot-password"
            LinkComponent={Link}
          />
        }
      />
      <Button
        loading={isLoading}
        size="2xl"
        variant="primary"
        content={<span className="bodyLg-medium">Continue with Email</span>}
        prefix={<EnvelopeFill />}
        block
        type="submit"
      />
    </form>
  );
};

const Login = () => {
  const { githubLoginUrl, gitlabLoginUrl, googleLoginUrl } =
    useOutletContext<IProviderContext>();
  const [searchParams, _setSearchParams] = useSearchParams();

  return (
    <Container
      footer={{
        message: 'Don’t have an account?',
        to: '/signup',
        buttonText: 'Signup',
      }}
    >
      <div className="flex flex-col items-stretch justify-center gap-7xl md:w-[400px]">
        <div className="flex flex-col gap-5xl">
          <BrandLogo darkBg={false} size={60} />
          <div className="flex flex-col items-stretch gap-5xl border-b pb-5xl border-border-default">
            <div className="flex flex-col items-center md:px-7xl">
              <div
                className={classNames(
                  'text-text-strong heading3xl text-center'
                )}
              >
                Login to Kloudlite
              </div>
            </div>
            {searchParams.get('mode') === 'email' ? (
              <LoginWithEmail />
            ) : (
              <div className="flex flex-col items-stretch gap-3xl">
                <Button
                  size="2xl"
                  variant="tertiary"
                  content={
                    <span className="bodyLg-medium">Continue with GitHub</span>
                  }
                  prefix={<GithubLogoFill />}
                  to={githubLoginUrl}
                  disabled={!githubLoginUrl}
                  block
                  LinkComponent={Link}
                />
                <Button
                  size="2xl"
                  variant="purple"
                  content={
                    <span className="bodyLg-medium">Continue with GitLab</span>
                  }
                  prefix={<GitlabLogoFill />}
                  to={gitlabLoginUrl}
                  disabled={!gitlabLoginUrl}
                  block
                  LinkComponent={Link}
                />
                <Button
                  size="2xl"
                  variant="primary"
                  content={
                    <span className="bodyLg-medium">Continue with Google</span>
                  }
                  prefix={<CustomGoogleIcon />}
                  to={googleLoginUrl}
                  disabled={!googleLoginUrl}
                  block
                  LinkComponent={Link}
                />
              </div>
            )}
          </div>
          {searchParams.get('mode') === 'email' ? (
            <Button
              size="2xl"
              variant="outline"
              content={
                <span className="bodyLg-medium">Other Login options</span>
              }
              suffix={<ArrowRight />}
              to="/login"
              block
              LinkComponent={Link}
            />
          ) : (
            <Button
              size="2xl"
              variant="outline"
              content={
                <span className="bodyLg-medium">Continue with email</span>
              }
              prefix={<Envelope />}
              to="/login/?mode=email"
              block
              LinkComponent={Link}
            />
          )}
        </div>
        <div className="bodyMd text-text-soft text-center">
          By signing up, you agree to the{' '}
          <Link
            to="https://kloudlite.io/terms-and-conditions"
            className="underline"
          >
            Terms of Service
          </Link>{' '}
          and{' '}
          <Link className="underline" to="https://kloudlite.io/privacy-policy">
            Privacy Policy
          </Link>
          .
        </div>
      </div>
    </Container>
  );
};

export default Login;
