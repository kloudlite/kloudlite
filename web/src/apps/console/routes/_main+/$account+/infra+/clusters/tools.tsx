import { ArrowsDownUp } from '~/console/components/icons';
import { useSearchParams } from '@remix-run/react';
import { useMemo } from 'react';
import OptionList from '~/components/atoms/option-list';
import Toolbar from '~/components/atoms/toolbar';
import CommonTools from '~/console/components/common-tools';

interface IFilterByClusterType {
  onChange: (data: string) => void;
  value: string;
}

const FilterByCLusterType = ({ onChange, value }: IFilterByClusterType) => {
  return (
    <OptionList.Root>
      <OptionList.Trigger>
        <Toolbar.Button
          content="Filter Cluster"
          variant="basic"
          prefix={<ArrowsDownUp />}
        />
      </OptionList.Trigger>
      <OptionList.Content>
        <OptionList.RadioGroup
          value={value}
          onValueChange={(e) => {
            onChange?.(e);
          }}
        >
          <OptionList.RadioGroupItem
            value="All"
            onClick={(e) => e.preventDefault()}
          >
            All
          </OptionList.RadioGroupItem>
          <OptionList.RadioGroupItem
            value="Normal"
            onClick={(e) => e.preventDefault()}
          >
            Kloudlite clusters
          </OptionList.RadioGroupItem>
          <OptionList.RadioGroupItem
            value="Byok"
            onClick={(e) => e.preventDefault()}
          >
            Attached Clusters
          </OptionList.RadioGroupItem>
        </OptionList.RadioGroup>
      </OptionList.Content>
    </OptionList.Root>
  );
};

const Tools = ({ onChange, value }: IFilterByClusterType) => {
  const [searchParams] = useSearchParams();
  const options = useMemo(
    () => [
      // {
      //   name: 'Provider',
      //   type: 'cloudProviderName',
      //   search: false,
      //   dataFetcher: async () => {
      //     return [
      //       { content: 'Amazon Web Services', value: 'aws' },
      //       { content: 'Digital Ocean', value: 'do' },
      //       { content: 'Google Cloud Platform', value: 'gcp' },
      //       { content: 'Microsoft Azure', value: 'azure' },
      //     ];
      //   },
      // },
      // {
      //   name: 'Region',
      //   type: 'region',
      //   search: false,
      //   dataFetcher: async () => {
      //     return [
      //       { content: 'Mumbai(ap-south-1)', value: 'ap-south-1' },
      //       { content: 'NY(ap-south-2)', value: 'do' },
      //     ];
      //   },
      // },
      // {
      //   name: 'Status',
      //   type: 'isReady',
      //   search: false,
      //   dataFetcher: async () => {
      //     return [
      //       { content: 'Running', value: true },
      //       { content: 'Error', value: false },
      //       // { content: 'Freezed', value: false, type: 'freezed' },
      //     ];
      //   },
      // },
    ],
    [searchParams]
  );
  return (
    <CommonTools
      {...{ options, noSort: true }}
      // commonToolPrefix={
      //   <FilterByCLusterType
      //     onChange={(e) => {
      //       onChange?.(e);
      //     }}
      //     value={value}
      //   />
      // }
    />
  );

  // const [searchParams] = useSearchParams();

  // const options = useMemo(() => [], [searchParams]);
};

export default Tools;
