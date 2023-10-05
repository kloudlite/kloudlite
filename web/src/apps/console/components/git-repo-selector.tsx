import { Search } from '@jengaicons/react';
import { useState } from 'react';
import { TextInput } from '~/components/atoms/input';
import { useMapper } from '~/components/utils';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import { useConsoleApi } from '../server/gql/api-provider';
import List from './list';

const GitRepoSelector = ({}) => {
  const api = useConsoleApi();

  const { data, error, isLoading } = useCustomSwr(
    'api/github_installations',
    async () => api.listGithubInstalltions({})
  );

  const options = useMapper(data || [], (d) => ({
    value: `${d.id!}`,
    render: () => <div>{d.account?.login}</div>,
  }));

  const [selectedAccount, setSelectedAccount] = useState(options[0]);

  const [repos, setRepos] = useState([
    { name: 'operator', private: false, updated_at: '161d ago' },
  ]);

  return (
    <div className="p-5xl flex flex-col gap-2xl">
      <div className="heading2xl text-text-strong">Import Git Repository</div>
      <div className="flex flex-row gap-lg items-center">
        <div className="flex-1">
          {/* <Select
            options={options}
            value={selectedAccount}
            onChange={() => {}}
          /> */}
        </div>
        <div className="flex-1">
          <TextInput placeholder="Search" prefixIcon={<Search />} />
        </div>
      </div>
      <List.Root>
        {/* {data?.data.map(({}) => {
          return <List.Row />;
        })} */}
      </List.Root>
    </div>
  );
};

export default GitRepoSelector;
