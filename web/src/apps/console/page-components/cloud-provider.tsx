import { useState } from 'react';
import { PasswordInput, TextInput } from '~/components/atoms/input';
import { Button } from '~/components/atoms/button';
import { InfoLabel } from '../routes/_.$account.$cluster.$project.$scope.$workspace.new-app/util';

type IHMap = {
  [key: string]: any;
};

export const AwsForm = ({
  errors,
  values,
  handleChange,
}: {
  errors: IHMap;
  values: IHMap;
  handleChange: (name: string) => (e: any) => void;
}) => {
  const [withAccId, setWithAccId] = useState(true);

  return (
    <>
      {withAccId ? (
        <div className="flex-1">
          <TextInput
            name="awsAccountId"
            onChange={handleChange('awsAccountId')}
            error={!!errors.awsAccountId}
            message={errors.awsAccountId}
            value={values.awsAccountId}
            label="Account ID"
          />
        </div>
      ) : (
        <>
          <PasswordInput
            name="accessKey"
            onChange={handleChange('accessKey')}
            error={!!errors.accessKey}
            message={errors.accessKey}
            value={values.accessKey}
            label={
              <InfoLabel
                info={
                  <div>
                    <p>
                      Provide access key and secret key to access your AWS
                      account. <br />
                      We need these creds with following permissions: <br />
                    </p>
                    <ul className="px-md">
                      <li>ec2</li>
                      <li>s3</li>
                      <li>spotFleetTaggingRole</li>
                    </ul>
                  </div>
                }
                label="Access Key ID"
              />
            }
          />
          <PasswordInput
            name="accessSecret"
            label="Access Key Secret"
            onChange={handleChange('accessSecret')}
            error={!!errors.accessSecret}
            message={errors.accessSecret}
            value={values.accessSecret}
          />
        </>
      )}

      <div className="flex">
        <Button
          onClick={() => {
            return setWithAccId((s) => !s);
          }}
          variant="primary-plain"
          content={
            withAccId ? 'Use access creds instead' : 'Use account id instead'
          }
        />
      </div>
    </>
  );
};
