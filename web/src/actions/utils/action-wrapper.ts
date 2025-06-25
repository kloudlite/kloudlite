export function actionWrapper<T, Args extends unknown[]>(
  action: (...args: Args) => Promise<T>,
): (...args: Args) => Promise<[T|null, Error|null]> {
  return async (...args: Args) => {
    try {
      const result = await action(...args);
      return [result, null] as [T, null];
    } catch (error) {
      return [null, error as Error] as [null, Error];
    }
  };
}