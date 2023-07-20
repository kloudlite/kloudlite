export const serverError = (errors) => {
  throw new Error(JSON.stringify(errors, null, 2));
};
