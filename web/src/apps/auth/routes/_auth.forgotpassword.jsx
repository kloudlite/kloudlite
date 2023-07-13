import classNames from 'classnames';
import { BrandLogo } from '~/components/branding/brand-logo.jsx';
import { Button } from '~/components/atoms/button.jsx';
import { ArrowRight } from '@jengaicons/react';
import { TextInput } from '~/components/atoms/input.jsx';
import { GoogleReCaptcha } from 'react-google-recaptcha-v3';
import { Link } from '@remix-run/react';

export default ({ }) => {
  const onsubmit = (e) => {
    e.preventDefault();
  };
  return (
    <div
      className={classNames('flex flex-col items-center justify-center h-full')}
    >
      <form
        className={classNames(
          'flex flex-1 flex-col items-center self-stretch justify-center px-3xl pb-5xl border-b border-border-default'
        )}
        onSubmit={onsubmit}
      >
        <div className="flex flex-col items-stretch justify-center gap-5xl md:w-[400px]">
          <BrandLogo darkBg={false} size={60} />
          <div className="flex flex-col items-stretch gap-5xl pb-5xl">
            <div className="flex flex-col gap-lg items-center md:px-7xl">
              <div
                className={classNames(
                  'text-text-strong heading3xl text-center'
                )}
              >
                Forgot password
              </div>
              <div className="text-text-soft bodySm text-center">
                Enter your registered email below to receive password reset
                instruction.
              </div>
            </div>
            <div className="flex flex-col items-stretch gap-3xl">
              <TextInput label="Email" placeholder="ex: john@company.com" />
              <Button
                size="large"
                variant="primary"
                content={
                  <span className="bodyLg-medium">Send instructions</span>
                }
                DisclosureComp={ArrowRight}
                block
                type="submit"
                LinkComponent={Link}
              />
            </div>
          </div>
        </div>
        <GoogleReCaptcha onVerify={(e) => { }} />
      </form>
      <div className="py-5xl px-3xl flex flex-row items-center justify-center self-stretch">
        <div className="bodyMd text-text-default">Remember password?</div>
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
