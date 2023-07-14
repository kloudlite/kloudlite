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
import { useSearchParams, Link } from '@remix-run/react';
import { TextInput, PasswordInput } from '~/components/atoms/input.jsx';

const CustomGoogleIcon = (props) => {
  return <GoogleLogo {...props} weight={4} />;
};

const Signup = () => {
  const [searchParams, _setSearchParams] = useSearchParams();
  return (
    <div
      className={classNames(
        'flex flex-col items-center justify-center min-h-full'
      )}
    >
      <div
        className={classNames(
          'flex flex-1 flex-col items-center self-stretch justify-center px-3xl py-8xl border-b border-border-default'
        )}
      >
        <div className="flex flex-col items-stretch justify-center gap-5xl md:w-[400px]">
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
              <div className="text-text-soft bodySm text-center">
                To access your DevOps console, Please provide your login
                credentials.
              </div>
            </div>
            {searchParams.get('mode') === 'email' ? (
              <div className="flex flex-col items-stretch gap-3xl">
                <TextInput label="Name" placeholder="Full name" />
                <div className="flex flex-col gap-3xl items-stretch md:flex-row">
                  <TextInput label="Company Name" className="flex-1" />
                  {/* <NumberInput label={"Company Size"} className={"flex-1"} min={1} /> */}
                </div>
                <TextInput label="Email" placeholder="ex: john@company.com" />
                <PasswordInput label="Password" placeholder="XXXXXX" />
                <Button
                  size="large"
                  variant="primary"
                  content={
                    <span className="bodyLg-medium">Continue with Email</span>
                  }
                  prefix={EnvelopeFill}
                  block
                  LinkComponent={Link}
                />
              </div>
            ) : (
              <div className="flex flex-col items-stretch gap-3xl">
                <Button
                  size="large"
                  variant="basic"
                  content={
                    <span className="bodyLg-medium">Continue with GitHub</span>
                  }
                  prefix={GithubLogoFill}
                  href="https://google.com"
                  block
                  LinkComponent={Link}
                />
                <Button
                  size="large"
                  variant="primary"
                  content={
                    <span className="bodyLg-medium">Continue with GitLab</span>
                  }
                  prefix={GitlabLogoFill}
                  block
                  LinkComponent={Link}
                />
                <Button
                  size="large"
                  variant="secondary"
                  content={
                    <span className="bodyLg-medium">Continue with Google</span>
                  }
                  prefix={CustomGoogleIcon}
                  block
                  LinkComponent={Link}
                />
              </div>
            )}
          </div>
          {searchParams.get('mode') === 'email' ? (
            <Button
              size="large"
              variant="outline"
              content={
                <span className="bodyLg-medium">Other Signup options</span>
              }
              prefix={ArrowLeft}
              href="/signup"
              block
              LinkComponent={Link}
            />
          ) : (
            <Button
              size="large"
              variant="outline"
              content={<span className="bodyLg-medium">Signup with Email</span>}
              prefix={Envelope}
              href="/signup/?mode=email"
              block
              LinkComponent={Link}
            />
          )}

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
      </div>
      <div className="py-5xl px-3xl flex flex-row items-center justify-center self-stretch">
        <div className="bodyMd text-text-default">Already have an account?</div>
        <Button
          content="Login"
          variant="primary-plain"
          size="medium"
          href="/login"
          LinkComponent={Link}
        />
      </div>
    </div>
  );
};

export default Signup;
