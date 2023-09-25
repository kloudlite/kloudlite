import { Plus } from '@jengaicons/react';
import { Button } from '~/components/atoms/button';
import SecondarySubHeader from '~/console/components/secondary-sub-header';
import Wip from '~/root/lib/client/components/wip';

const Domain = () => {
  return (
    <div className="pt-3xl">
      <SecondarySubHeader
        title="Domain"
        action={
          <Button content="Add device" prefix={<Plus />} variant="primary" />
        }
      />
      <Wip />
    </div>
  );
};

export default Domain;
