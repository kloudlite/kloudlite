import Wrapper from '~/console/components/wrapper';

import Wip from '~/console/components/wip';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import { useParams } from '@remix-run/react';

const Tabs = () => {
  const { account, repo } = useParams();
  return (
    <CommonTabs
      backButton={{
        to: `/${account}/container-registry/repo/${repo}/builds/`,
        label: 'Build configs',
      }}
    />
  );
};

export const handle = () => {
  return {
    navbar: <Tabs />,
  };
};
const BuildRuns = () => {
  return (
    <Wrapper
      header={{
        title: 'Build runs',
      }}
    >
      <Wip />
    </Wrapper>
  );
};

export default BuildRuns;
