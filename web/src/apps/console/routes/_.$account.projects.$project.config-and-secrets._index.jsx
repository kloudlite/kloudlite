import { redirect } from '@remix-run/node';

export const loader = async (ctx) => {
  return redirect('config');
};
