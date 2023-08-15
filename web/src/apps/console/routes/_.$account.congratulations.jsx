import { Button } from '~/components/atoms/button';

const Congratulations = () => {
  return (
    <div className="px-11xl pt-8xl pb-10xl flex items-center justify-center h-full">
      <div className="flex flex-col gap-6xl w-[450px] text-center">
        <div className="bg-surface-basic-active h-[200px]" />
        <div className="flex flex-col gap-3xl">
          <div className="heading4xl text-text-default">Congratulations 🚀</div>
          <div className="bodyLg text-text-soft block">
            You’ve successfully create your organization{' '}
            <span className="headingMd">Astroman</span> and deployed first
            project <span className="headingMd">Lobster Early</span>.
          </div>
          <Button content="Continue to dashboard" variant="basic" block />
        </div>
      </div>
    </div>
  );
};

export default Congratulations;
