export const serverError = (errors: Error[]) => {
  throw new Error(JSON.stringify(errors, null, 2));
};
