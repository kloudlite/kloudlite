import { NonNullableString } from '~/root/lib/types/common';

interface Iplan {
  memoryPerCpu: number;
  name: string;
  description: string;
}

export type IcpuMode = 'shared' | 'dedicated';

export const plans: {
  shared: Iplan[];
  dedicated: Iplan[];
} = {
  shared: [
    {
      memoryPerCpu: 4,
      name: 'Essential Plan',
      description: 'This foundational package for your need',
    },
    {
      memoryPerCpu: 2,
      name: 'Standard offerings',
      description: 'This foundational package for your need',
    },
    {
      memoryPerCpu: 1,
      name: 'Memory-Boost packages',
      description: 'This foundational package for your need',
    },
  ],
  dedicated: [
    {
      memoryPerCpu: 4,
      name: 'Essential Plan',
      description: 'This foundational package for your need',
    },
    {
      memoryPerCpu: 2,
      name: 'Standard offerings',
      description: 'This foundational package for your need',
    },
    {
      memoryPerCpu: 1,
      name: 'Memory-Boost packages',
      description: 'This foundational package for your need',
    },
  ],
};
