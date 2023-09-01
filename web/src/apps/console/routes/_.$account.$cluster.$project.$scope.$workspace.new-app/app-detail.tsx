import { useState } from 'react';
import { TextInput } from '~/components/atoms/input';
import { IdSelector } from '~/console/components/id-selector';

const AppDetail = () => {
  const [name, setName] = useState<string>('');
  const [description, setDescription] = useState<string>('');
  return (
    <>
      <div className="flex flex-col gap-lg">
        <div className="headingXl text-text-default">Application details</div>
        <div className="bodyMd text-text-soft">
          The application streamlines project management through intuitive task
          tracking and collaboration tools.
        </div>
      </div>
      <div className="flex flex-col gap-3xl">
        <TextInput
          label="Application name"
          size="lg"
          value={name}
          onChange={({ target }) => {
            setName(target.value);
          }}
        />
        <IdSelector name="app" />
        <TextInput
          label="Description"
          size="lg"
          value={description}
          onChange={({ target }) => {
            setDescription(target.value);
          }}
        />
      </div>
    </>
  );
};

export default AppDetail;
