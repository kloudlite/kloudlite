import classNames from 'classnames';
import { BrandLogo } from '~/components/branding/brand-logo.jsx';
import { Button } from '~/components/atoms/button.jsx';
import {
  ArrowLeft,
  Envelope,
  EnvelopeFill,
  GithubLogoFill,
  GitlabLogoFill,
  GoogleLogo,
} from '@jengaicons/react';
import {
  useSearchParams,
  Link,
  useLoaderData,
  useNavigate,
} from '@remix-run/react';
import { TextInput, PasswordInput } from '~/components/atoms/input.jsx';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import logger from '~/root/lib/client/helpers/log';
import { assureNotLoggedIn } from '~/root/lib/server/helpers/minimal-auth';
import { toast } from '~/components/molecule/toast';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { handleError } from '~/root/lib/utils/common';
import { GQLServerHandler } from '../server/gql/saved-queries';
import Container from '../components/container';

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
        const { errors: _errors } = await api.signUpWithEmail({
          email: v.email,
          name: v.name,
          password: v.password,
        });
        if (_errors) {
          throw _errors[0];
        }
        toast.success('signed up successfully');
        navigate('/');
      } catch (err) {
        handleError(err);
      }
    },
  });

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col items-stretch gap-6xl"
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
        />
        <PasswordInput
          name="password"
          value={values.password}
          error={!!errors.password}
          onChange={handleChange('password')}
          label="Password"
          placeholder="XXXXXX"
          size="lg"
          message={
            <span className="bodySm text-text-soft">
              Must be atleast 8 character
            </span>
          }
        />

        <PasswordInput
          value={values.c_password}
          error={!!errors.c_password}
          onChange={handleChange('c_password')}
          label="Confirm Password"
          placeholder="XXXXXX"
          size="lg"
          message={
            <span className="bodySm text-text-soft">
              Both password must match
            </span>
          }
        />
      </div>

      <Button
        size="2xl"
        loading={isLoading}
        type="submit"
        variant="primary"
        content={<span className="bodyLg-medium">Continue with Email</span>}
        prefix={<EnvelopeFill />}
        block
        LinkComponent={Link}
      />
    </form>
  );
};

const Signup = () => {
  const { githubLoginUrl, gitlabLoginUrl, googleLoginUrl } = useLoaderData();
  const [searchParams, _setSearchParams] = useSearchParams();
  return (
    <Container
      footer={{
        message: 'Already have an account?',
        buttonText: 'Login',
        to: '/login',
      }}
    >
      <div className="flex flex-col items-stretch justify-center gap-7xl md:w-[400px]">
        <div className="flex flex-col gap-5xl">
          <BrandLogo darkBg={false} size={60} />
          <div className="flex flex-col items-stretch gap-5xl border-b pb-5xl border-border-default">
            <div className="flex flex-col gap-lg items-center md:px-7xl">
              <div
                className={classNames(
                  'text-text-strong heading3xl text-center'
                )}
              >
                Signup to Kloudlite
              </div>
            </div>
            {searchParams.get('mode') === 'email' ? (
              <SignUpWithEmail />
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
                <span className="bodyLg-medium">Other Signup options</span>
              }
              prefix={<ArrowLeft />}
              to="/signup"
              block
              LinkComponent={Link}
            />
          ) : (
            <Button
              size="2xl"
              variant="outline"
              content={<span className="bodyLg-medium">Signup with Email</span>}
              prefix={<Envelope />}
              to="/signup/?mode=email"
              block
              LinkComponent={() => Link}
            />
          )}
        </div>

        <div className="bodyMd text-text-soft text-center">
          By signing up, you agree to the{' '}
          <Link to="/terms" className="underline">
            Terms of Service
          </Link>{' '}
          and{' '}
          <Link className="underline" to="/privacy">
            Privacy Policy
          </Link>
          .
        </div>
      </div>
    </Container>
  );
};

const restActions = async (ctx: any) => {
  const { data, errors } = await GQLServerHandler(
    ctx.request
  ).loginPageInitUrls();
  if (errors) {
    logger.error(errors);
  }

  const {
    githubLoginUrl = null,
    gitlabLoginUrl = null,
    googleLoginUrl = null,
  } = data || {};
  return {
    githubLoginUrl,
    gitlabLoginUrl,
    googleLoginUrl,
  };
};

export const loader = async (ctx: any) =>
  (await assureNotLoggedIn(ctx)) || restActions(ctx);

export default Signup;
