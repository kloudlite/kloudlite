import { useLoaderData } from '@remix-run/react';

type UnionToIntersection<U> = (U extends any ? (k: U) => void : never) extends (
  k: infer I
) => void
  ? I
  : never;

const useExtLoaderData = <T extends (...args: any) => Promise<any>>() =>
  useLoaderData() as UnionToIntersection<Awaited<ReturnType<T>>>;

export default useExtLoaderData;
