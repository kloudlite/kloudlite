import React, { useMemo, ReactNode } from 'react';
import { useSearchParams } from '@remix-run/react';
import CommonTools from '~/console/components/common-tools';
import { IToolsProps } from '~/console/server/utils/common';

interface IOption {
  label: string;
  value: string;
  render?: ReactNode;
}

interface IGroup<T> {
  label: string;
  options: T;
}

// Define a generic type parameter for SelectProps
interface SelectProps<T, A extends boolean | undefined = undefined> {
  options: T[] & (IOption[] | IGroup<IOption[] | T[]>[]);
  isMulti?: A;
  value: A extends true ? T[] : T; // Ensure value matches the type of options
}

const Select = <T, A extends boolean | undefined = undefined>({
  options,
  isMulti = false,
  value,
}: SelectProps<T, A>) => {
  // Your component logic here

  return <div />;
};

const Tools = ({ viewMode, setViewMode }: IToolsProps) => {
  const [searchParams] = useSearchParams();

  const options = useMemo(
    () => [
      {
        name: 'Status',
        type: 'status',
        search: false,
        dataFetcher: async (_: string) => {
          return [
            {
              content: 'running',
              value: true,
            },
            {
              content: 'error',
              value: false,
            },
          ];
        },
      },
    ],
    [searchParams]
  );
  const x = [{ label: 'a', options: [{ value: 'a', label: 'b', c: 'v' }] }];
  return (
    <>
      <CommonTools {...{ viewMode, setViewMode, options }} />
      <Select
        isMulti
        options={[{ label: 'a', value: 'b', c: 'a' }]}
        value={[{ label: 'a', value: 'b', c: 'a' }]}
      />
    </>
  );
};
export default Tools;
