import { Plus } from '@jengaicons/react';
import { Link } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import { SubHeader } from '~/components/organisms/sub-header';

const Config = () => {
  return (
    <SubHeader
      backUrl="../config-and-secrets/configs"
      LinkComponent={Link}
      title="kloud-root-ca.crt"
      actions={
        <Button variant="basic" content="Add new config" prefix={Plus} />
      }
    />
  );
};

export default Config;
