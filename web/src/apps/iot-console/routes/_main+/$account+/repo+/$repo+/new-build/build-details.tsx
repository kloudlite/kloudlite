import { useEffect, useRef } from 'react';
import { TextInput } from '@kloudlite/design-system/atoms/input';
import Select from '@kloudlite/design-system/atoms/select';
import { dummyEvent } from '~/root/lib/client/hooks/use-form';
import AdvancedOptions from '../builds/advanced-options';

const BuildDetails = ({
  values,
  handleChange,
  errors,
}: {
  values: { [key: string]: any };
  handleChange: (key: string) => (e: { target: { value: any } }) => void;
  errors: {
    [key: string]: string | undefined;
  };
}) => {
  const ref = useRef<HTMLInputElement>(null);
  useEffect(() => {
    ref.current?.focus();
  }, [ref.current]);

  return (
    <div className="flex flex-col gap-3xl">
      <TextInput
        ref={ref}
        label="Build name"
        size="lg"
        value={values.name}
        onChange={handleChange('name')}
        error={!!errors.name}
        message={errors.name}
      />
      <Select
        label="Tags"
        size="lg"
        placeholder="Add tags"
        creatable
        multiple
        value={values.tags}
        options={async () =>
          values.tags.map((t: string) => ({ label: t, value: t }))
        }
        onChange={(_, val) => {
          handleChange('tags')(dummyEvent(val));
        }}
        error={!!errors.tags}
        message={errors.tags}
      />

      <AdvancedOptions
        values={values}
        handleChange={handleChange}
        errors={errors}
      />
    </div>
  );
};

export default BuildDetails;
