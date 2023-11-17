import { TextInput } from '~/components/atoms/input';
import { TitleBox } from '~/console/components/raw-wrapper';

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
  return (
    <>
      <TitleBox
        title="Build details"
        subtitle="The application streamlines project management through intuitive build tracking and collaboration tools."
      />
      <div className="flex flex-col">
        <div className="flex flex-col pb-3xl">
          <TextInput
            label="Build name"
            size="lg"
            value={values.displayName}
            onChange={handleChange('displayName')}
            error={!!errors.displayName}
            message={errors.displayName}
          />
          {/* <IdSelector
                  onChange={(v) => handleChange('name')(dummyEvent(v))}
                  name={values.displayName}
                  resType="cluster"
                  className="pt-2xl"
                /> */}
        </div>
        <TextInput
          error={!!errors.description}
          message={errors.description}
          label="Description"
          size="lg"
          value={values.description}
          onChange={handleChange('description')}
        />
      </div>
    </>
  );
};

export default BuildDetails;
