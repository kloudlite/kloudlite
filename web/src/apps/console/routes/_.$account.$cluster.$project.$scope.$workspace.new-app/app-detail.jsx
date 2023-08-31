import { ArrowLeft, ArrowRight } from '@jengaicons/react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { IdSelector } from '~/console/components/id-selector';

const AppDetail = () => {
  return (
    <>
      <div className="flex flex-col gap-lg">
        <div className="headingXl text-text-default">Application details</div>
        <div className="bodySm text-text-soft">
          The application streamlines project management through intuitive task
          tracking and collaboration tools.
        </div>
      </div>
      <div className="flex flex-col gap-3xl">
        <TextInput label="Application name" size="lg" />
        <IdSelector name="app" />
        <TextInput label="Description" size="lg" />
      </div>
    </>
  );
};

export default AppDetail;
