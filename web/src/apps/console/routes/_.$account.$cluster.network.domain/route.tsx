import { Plus } from '@jengaicons/react';
import { Button } from '~/components/atoms/button';
import Wip from '~/root/lib/client/components/wip';

const Domain = () => {
  return (
    <div className="pt-3xl">
      <div className="flex flex-row items-center min-h-[36px]">
        <div className="headingXl text-text-strong flex-1">Domain</div>
        <div>
          <Button
            content="Add domain"
            prefix={<Plus />}
            variant="primary"
            onClick={() => {}}
          />
        </div>
      </div>
      <Wip />
    </div>
  );
};

export default Domain;
